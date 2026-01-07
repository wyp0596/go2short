# go2short

A minimal, high-performance URL shortener built with Go.

## Features

- **Fast redirects**: P50 < 5ms, Redis-first with Postgres fallback
- **Async analytics**: Click events via Redis Streams, no sync writes on redirect
- **Simple deployment**: Single binary with embedded admin UI
- **Horizontally scalable**: Stateless app, scale with replicas
- **Admin dashboard**: Link management, click stats, built with Vue 3

## Quick Start

```bash
# One command to start everything
docker compose up -d --build

# Admin dashboard
open http://localhost:8080/admin/
# Default login: admin / admin123

# Create a short link (API)
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very/long/path"}'

# Use it
curl -L http://localhost:8080/{code}
```

### Local Development

```bash
# Start dependencies only
docker compose up -d postgres redis

# Build frontend (required for embedding)
cd web && npm install && npm run build && cd ..

# Run the app locally
go run ./cmd/app
```

### Build from Source

```bash
# Build frontend
cd web && npm install && npm run build && cd ..

# Build Go binary (includes embedded frontend)
go build -o go2short ./cmd/app

# Run
./go2short
```

## Architecture

```
Request → Gateway → App → Redis (cache) → Postgres (store)
                      ↓
              Redis Streams → Worker → Postgres (analytics)
```

**Key constraint**: Redirect path does zero sync database writes.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8080` | Listen address |
| `REDIRECT_STATUS_CODE` | `302` | Redirect status |
| `CODE_LENGTH` | `8` | Generated code length |
| `REDIS_ADDR` | `localhost:6379` | Redis connection |
| `DATABASE_URL` | - | Postgres connection |
| `ADMIN_USERNAME` | `admin` | Admin login username |
| `ADMIN_PASSWORD` | `admin123` | Admin login password |
| `ADMIN_TOKEN_TTL` | `24h` | Admin session duration |

See [docs/Project.md](docs/Project.md) for full configuration.

## API

### Redirect
```
GET /:code → 302 redirect
```

### Create Link
```
POST /api/links
{"long_url": "https://..."}

→ {"code": "abc123", "short_url": "https://go2.sh/abc123"}
```

## Documentation

- [Architecture & Specification](docs/Project.md)
- [中文文档](README_zh.md)

## Tech Stack

- Go 1.22+ / Gin
- PostgreSQL 15+
- Redis 7+
- Vue 3 (admin, optional)

## License

MIT
