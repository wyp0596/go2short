# Short URL (Go) 项目协作与架构约定

> 目标：容器内运行、资源占用小、重定向高并发低延迟、可水平扩展、可观测、易演进。


## 1. 项目概览

### 1.1 核心能力
- **短链创建**：生成 `code` → 存储 `code -> long_url`
- **短链跳转**：`GET /{code}` → 302 重定向（**最薄路径**）
- **点击统计**：重定向只投递事件，统计写库异步完成
- **管理后台**：基础 CRUD、黑白名单、流量观察（可选）

### 1.2 非功能目标（SLO/约束）
- 重定向路径：P50 < 5ms（应用内，不含公网网络），P99 尽量稳定
- Redis 命中率：目标 > 95%（视业务）
- Postgres：保证可持续写入与可回溯查询（点击表分区）
- 任何重定向请求 **不做同步写库**（禁止）


## 2. 技术栈与组件

### 2.1 技术栈
- Backend：Go 1.25.x、Gin、Ent（管理面/非关键路径）
- Database：PostgreSQL 15+
- Cache/Queue：Redis 7+（缓存 + Streams 事件）
- Frontend：Vue 3.4+、Vite 5+、TailwindCSS（管理台）
- Gateway：Nginx/Caddy/Traefik（TLS、HTTP/2、基础限流）

### 2.2 进程/服务拆分（Docker）
- `gateway`：TLS 终止、连接复用、压缩、基础限流
- `app`：API + Redirect + 点击事件消费（Go Gin + 内置 goroutine）
- `postgres`：主库
- `redis`：缓存 + Streams

> 单进程 MVP：app 内启动消费协程批量写库；需要时可拆成独立 worker 镜像。
> 水平扩展：`app` 可多副本；Redis/Postgres 逐步演进到 HA/托管。


## 3. 请求路径与关键原则

### 3.1 重定向路径（性能关键）
- 路由：`GET /:code`
- 流程：
  1) 校验 `code` 合法性（长度/字符集）
  2) **Redis GET 命中直接返回** long_url
  3) Miss：查 Postgres `links` 表 → 回填 Redis
  4) 返回 301/302（默认 302，可配置）
  5) 发送点击事件到 Redis Streams（**异步**，不阻塞）

**禁止事项**
- 禁止在重定向请求里写 Postgres（点击统计、风控写入等）
- 禁止在重定向请求里执行复杂 ORM 装载、联表、N+1
- 禁止在重定向请求里做外部 HTTP 调用

### 3.2 创建短链路径（可接受较重）
- 路由：`POST /api/links`
- 校验：URL 合法性、长度限制、协议白名单（http/https）
- 生成 code：见 4.1
- 写库：Postgres 为准（source of truth）
- 回填缓存：成功写库后写 Redis

### 3.3 管理后台路径（非关键）
- 查询、禁用、删除策略按产品需求实现
- 管理路径可使用 Ent 全量能力


## 4. 数据模型与存储策略

### 4.1 Code 生成策略
默认方案：**随机 base62 + 唯一约束冲突重试**
- 优点：实现简单、可多实例扩展
- 规则：
  - `code` 长度：建议 7~10（按碰撞概率与业务量调）
  - 字符集：base62（0-9a-zA-Z）
  - 生成失败：遇到唯一冲突重试（限制重试次数）

备选：Snowflake/Segment ID → base62（未来需要可控递增或更低冲突时再做）

### 4.2 Postgres 表（建议）
#### `links`
- `code` TEXT PRIMARY KEY / UNIQUE
- `long_url` TEXT NOT NULL
- `created_at` TIMESTAMPTZ NOT NULL
- `expires_at` TIMESTAMPTZ NULL（可选）
- `is_disabled` BOOLEAN NOT NULL DEFAULT false（可选）
- 索引：
  - PK/UNIQUE(`code`)
  - 可选索引：`created_at`（管理台分页）

#### `click_events`（写多读少，建议分区）
- `id` BIGSERIAL（或 UUID）
- `code` TEXT NOT NULL
- `ts` TIMESTAMPTZ NOT NULL
- `ip_hash` TEXT NULL（脱敏后）
- `ua_hash` TEXT NULL
- `referer` TEXT NULL
- `country` TEXT NULL（可选）
- 分区：按 `ts` 日/周分区
- 索引：
  - `code, ts`
  - `ts`（按时间查询）

> 重定向路径不依赖 click_events，写入由 worker 批量完成。

### 4.3 Redis 使用规范
- 缓存 Key：
  - `su:link:{code}` → long_url（string）
- TTL：
  - 默认可不设 TTL（依赖 maxmemory + volatile/LRU），或设长 TTL
- **负缓存**：
  - `su:miss:{code}` → 1（短 TTL，例如 30~120s）
  - Miss 时先查负缓存，避免穿透打爆 DB
- Streams：
  - Stream：`su:clicks`
  - Consumer Group：`su-worker`


## 5. 异步事件与 Worker

### 5.1 事件格式（建议 JSON）
字段建议：
- `code`
- `ts`（RFC3339 或 unix ms）
- `ip_hash`（可选）
- `ua_hash`（可选）
- `referer`（可选）
- `req_id`（用于追踪）

### 5.2 消费协程行为（app 内启动）
- app 启动时 `go consumeClicks()` 启动消费协程
- 从 Redis Streams 消费（consumer group）
- 批量聚合写入 Postgres（减少事务与连接压力）
- 出错重试与 DLQ（可选：用另一个 stream 存失败事件）
- 保障：至少一次（at-least-once），写库侧幂等按需求处理


## 6. 性能与并发控制

### 6.1 连接池与超时（必须）
- Redis：连接池、超时（读写/总超时）
- Postgres：连接池（最大连接数受控），超时（statement timeout）
- HTTP：
  - server read/write timeout
  - keep-alive 开启
  - gin 中间件避免额外开销（重定向路径最小化）

### 6.2 限流与防护
- gateway 层：按 IP 基础限流
- app 层：可选按 IP / code 限流（防刷）
- 对不存在 code：负缓存 + 限流


## 7. 可观测性（必须）

### 7.1 日志
- 结构化日志（JSON）
- 重定向路径日志采样（避免高 QPS 爆日志）
- 统一字段：`req_id`、`code`、`status`、`latency_ms`

### 7.2 Metrics（Prometheus 风格）
建议指标：
- `redirect_requests_total{status}`
- `redirect_latency_ms_bucket`
- `redis_cache_hits_total` / `redis_cache_misses_total`
- `db_queries_total` / `db_latency_ms_bucket`
- `click_events_enqueued_total`
- 消费协程：
  - `worker_batch_insert_total`
  - `stream_lag`（消费延迟）

### 7.3 Tracing（可选但推荐）
- OpenTelemetry
- 仅关键路径采样（避免性能回退）


## 8. API 规范

### 8.1 Redirect
- `GET /:code`
- 302 默认（可配置 301）
- 失败：
  - 不存在：404 或跳转到自定义落地页（产品决定）
  - 禁用/过期：410 或落地页

### 8.2 Create Link
- `POST /api/links`
- Request：
  - `long_url` 必填
  - 可选：`expires_at`、`custom_code`
- Response：
  - `code`
  - `short_url`
  - `created_at`


## 9. 代码组织建议（单仓库）

```

/cmd
/app         # gin server + click event consumer
/internal
/config
/http        # handlers/middlewares
/domain      # business logic
/repo        # db access (ent + raw sql)
/cache       # redis access
/events      # stream producer/consumer
/observability
/web           # vue admin
/migrations
/docker
docker-compose.yml

```

> 重定向查询可使用 raw SQL/轻量 repo，管理面用 Ent。


## 10. 迁移与运维

### 10.1 数据库迁移
- 迁移脚本必须可重复执行
- 分区表管理：提供定时脚本/Job（例如每天创建未来 N 天分区）

### 10.2 Docker/Compose 基线
- 服务健康检查（healthcheck）
- 资源限制（cpu/mem）在测试环境显式配置
- 环境变量集中管理（.env）


## 11. 安全与合规

- URL 校验：只允许 http/https
- 防 SSRF：创建短链时禁止内网 IP（可选但建议）
- 隐私：IP/UA 建议哈希化存储（可加盐）
- 管理后台鉴权：JWT / Session（视需求）


## 12. Claude Code 协作规则（写代码时遵循）

### 12.1 基本准则
- KISS / DRY / 最小 diff
- 关键路径优先：重定向路径不引入额外依赖与复杂逻辑
- 任何变更需附带：
  - 配置项默认值
  - 错误处理与超时
  - 基础指标/日志埋点（关键模块）

### 12.2 Review Checklist
- 重定向路径是否只做：校验 → cache → db → redirect → enqueue？
- 是否引入同步写库/外部调用？
- Redis miss 是否有负缓存与回填？
- DB 查询是否走索引？是否避免 N+1？
- Worker 是否批量写入？失败是否可重试？
- 是否增加/破坏 metrics/logs？
- 配置是否可通过 env 注入？

### 12.3 交付要求
- 新增接口必须有：
  - OpenAPI/接口说明（README 或 docs）
  - 单元测试（核心逻辑）
- 性能敏感改动必须提供：
  - 基本压测方法与结果摘要（本地或 CI）


## 13. 默认配置建议（可调整）

- `REDIRECT_STATUS_CODE=302`
- `CODE_LENGTH=8`
- `REDIS_KEY_PREFIX=su`
- `NEGATIVE_CACHE_TTL_SECONDS=60`
- `STREAM_NAME=su:clicks`
- `STREAM_GROUP=su-worker`
- `WORKER_BATCH_SIZE=500`
- `WORKER_FLUSH_INTERVAL_MS=200`
- `HTTP_READ_TIMEOUT=5s`
- `HTTP_WRITE_TIMEOUT=5s`
- `DB_MAX_OPEN_CONNS=20`
- `DB_MAX_IDLE_CONNS=10`
- `REDIS_DIAL_TIMEOUT=200ms`
- `REDIS_RW_TIMEOUT=200ms`


## 14. 明确的“不做”清单（避免过度设计）
- 不引入 Kafka/ES/ClickHouse 作为第一版必选
- 不做复杂风控/画像作为重定向同步逻辑
- 不做多租户/分库分表（除非已明确需求）
- 不在重定向请求里记录全量日志

