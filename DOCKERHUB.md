# go2short

A minimal, high-performance URL shortener built with Go.

## Features

- Fast redirects (P50 < 5ms)
- Redis-first with Postgres fallback
- Async analytics via Redis Streams
- Built-in admin dashboard
- Single binary deployment

## Quick Start

```bash
docker run -d --name go2short \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://user:pass@host:5432/go2short" \
  -e REDIS_URL="redis://:pass@host:6379" \
  -e BASE_URL="https://your-domain.com" \
  -e ADMIN_PASSWORD="your-password" \
  warren0596/go2short:latest
```

Admin dashboard: `http://localhost:8080/admin/`

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_URL` | Redis connection string |
| `BASE_URL` | Base URL for short links |
| `ADMIN_PASSWORD` | Admin login password |

## Links

- [GitHub](https://github.com/wyp0596/go2short)
- [Documentation](https://github.com/wyp0596/go2short#readme)
