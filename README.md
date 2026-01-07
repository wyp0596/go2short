# go2short

A minimal, high-performance URL shortener built with Go.

## Features

- **Fast redirects**: P50 < 5ms, Redis-first with Postgres fallback
- **Async analytics**: Click events via Redis Streams, no sync writes on redirect
- **Simple deployment**: Single binary, Docker-ready
- **Horizontally scalable**: Stateless app, scale with replicas

## Quick Start

```bash
# Start dependencies
docker-compose up -d postgres redis

# Run the app
go run cmd/app/main.go

# Create a short link
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{"long_url": "https://example.com/very/long/path"}'

# Use it
curl -v http://localhost:8080/{code}
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
