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

### 创建短链（需要 API Token）
```
POST /api/links
Authorization: Bearer <api_token>

{"long_url": "https://..."}

→ {"code": "abc123", "short_url": "https://go2.sh/abc123"}
```

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
