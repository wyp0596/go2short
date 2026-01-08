# go2short

极简高性能短链服务，Go 构建。

## 特性

- **极速重定向**：P50 < 5ms，Redis 优先 + Postgres 兜底
- **异步统计**：点击事件走 Redis Streams，重定向零同步写库
- **部署简单**：单二进制，内嵌管理界面
- **水平扩展**：无状态应用，加副本即可
- **管理后台**：链接管理、点击统计，Vue 3 构建

## 快速开始

```bash
# 一键启动
docker compose up -d --build

# 管理后台
open http://localhost:8080/admin/
# 默认账号：admin / admin123
```

### 通过 API 创建短链

API 访问需要 API Token。在管理后台的「API Tokens」页面创建，然后：

```bash
# 创建短链
curl -X POST http://localhost:8080/api/links \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"long_url": "https://example.com/very/long/path"}'

# 使用
curl -L http://localhost:8080/{code}
```

### 本地开发

```bash
# 仅启动依赖
docker compose up -d postgres redis

# 构建前端（嵌入需要）
cd web && npm install && npm run build && cd ..

# 本地运行应用
go run ./cmd/app
```

### 源码构建

```bash
# 构建前端
cd web && npm install && npm run build && cd ..

# 构建 Go 二进制（包含嵌入的前端）
go build -o go2short ./cmd/app

# 运行
./go2short
```

## 架构

```
请求 → 网关 → 应用 → Redis（缓存）→ Postgres（存储）
                 ↓
         Redis Streams → Worker → Postgres（统计）
```

**核心约束**：重定向路径禁止同步写库。

## 配置

| 变量 | 默认值 | 说明 |
|-----|-------|-----|
| `HTTP_ADDR` | `:8080` | 监听地址 |
| `BASE_URL` | `http://localhost:8080` | 生成短链的基础 URL |
| `TRUSTED_PROXIES` | - | 可信代理 IP（逗号分隔，如 `127.0.0.1,172.16.0.0/12`） |
| `REDIRECT_STATUS_CODE` | `302` | 重定向状态码 |
| `CODE_LENGTH` | `8` | 短码长度 |
| `REDIS_ADDR` | `localhost:6379` | Redis 地址 |
| `DATABASE_URL` | - | Postgres 连接串 |
| `ADMIN_USERNAME` | `admin` | 管理员用户名 |
| `ADMIN_PASSWORD` | `admin123` | 管理员密码 |
| `ADMIN_TOKEN_TTL` | `24h` | 登录会话时长 |

完整配置见 [docs/Project.md](docs/Project.md)。

## API

### 重定向
```
GET /:code → 302 跳转
```

### 二维码
```
GET /:code/qr?size=256 → PNG 图片

# size: 128-1024（默认 256）
```

### 创建短链（需要 API Token）
```
POST /api/links
Authorization: Bearer <api_token>

{"long_url": "https://...", "custom_code": "mycode", "expires_at": "2025-12-31T23:59:59Z"}

→ {"code": "abc123", "short_url": "https://go2short.go2f.cn/abc123", "created_at": "..."}
```

### 批量创建（需要 API Token）
```
POST /api/links/batch
Authorization: Bearer <api_token>

{"items": [{"long_url": "https://..."}, {"long_url": "https://...", "custom_code": "foo"}]}

→ {"results": [{"index": 0, "code": "abc123", "short_url": "..."}, ...]}

# 单次最多 100 条
```

### 短链预览（需要 API Token）
```
GET /api/links/:code/preview
Authorization: Bearer <api_token>

→ {"code": "abc123", "long_url": "https://..."}
```

### 频率限制
- 创建接口：60 次/分钟/IP
- 响应头：`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

### API Token 管理（管理员）
```
POST   /api/admin/tokens     → 创建 token（明文仅返回一次）
GET    /api/admin/tokens     → 列出所有 token
DELETE /api/admin/tokens/:id → 删除 token
```

## 文档

- [架构与规范](docs/Project.md)
- [English](README.md)

## 技术栈

- Go 1.22+ / Gin
- PostgreSQL 15+
- Redis 7+
- Vue 3（管理台，可选）

## 许可

MIT
