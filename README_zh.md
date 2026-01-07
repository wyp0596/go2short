# go2short

极简高性能短链服务，Go 构建。

## 特性

- **极速重定向**：P50 < 5ms，Redis 优先 + Postgres 兜底
- **异步统计**：点击事件走 Redis Streams，重定向零同步写库
- **部署简单**：单二进制，Docker 友好
- **水平扩展**：无状态应用，加副本即可

## 快速开始

```bash
# 一键启动
docker compose up -d --build

# 创建短链
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{"long_url": "https://example.com/very/long/path"}'

# 使用
curl -L http://localhost:8080/{code}
```

### 本地开发

```bash
# 仅启动依赖
docker compose up -d postgres redis

# 本地运行应用
go run cmd/app/main.go
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
| `REDIRECT_STATUS_CODE` | `302` | 重定向状态码 |
| `CODE_LENGTH` | `8` | 短码长度 |
| `REDIS_ADDR` | `localhost:6379` | Redis 地址 |
| `DATABASE_URL` | - | Postgres 连接串 |

完整配置见 [docs/Project.md](docs/Project.md)。

## API

### 重定向
```
GET /:code → 302 跳转
```

### 创建短链
```
POST /api/links
{"long_url": "https://..."}

→ {"code": "abc123", "short_url": "https://go2.sh/abc123"}
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
