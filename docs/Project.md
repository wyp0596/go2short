# Architecture & Internals

> Deep dive into go2short's design decisions and implementation details.
> For quick start, see [README](../README.md).

---

## Design Principles

1. **Redirect path is sacred** - Zero sync DB writes, minimal middleware
2. **Simple > Clever** - No over-abstraction, no premature optimization
3. **Explicit > Magic** - Configuration via env, no hidden behavior

---

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Gateway   │────▶│     App     │────▶│    Redis    │
│  (Nginx)    │     │   (Go/Gin)  │     │  Cache+MQ   │
└─────────────┘     └──────┬──────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │  Postgres   │
                    └─────────────┘
```

**Single binary**: App includes redirect handler + click event consumer goroutine.

---

## Critical Path: Redirect

```
GET /:code
    │
    ▼
┌─────────────────────────────┐
│ 1. Validate code (base62)   │
├─────────────────────────────┤
│ 2. Redis GET su:link:{code} │──▶ HIT ──▶ 302 + async enqueue
├─────────────────────────────┤
│ 3. MISS: Check negative     │
│    cache su:miss:{code}     │──▶ HIT ──▶ 404
├─────────────────────────────┤
│ 4. Query Postgres           │
├─────────────────────────────┤
│ 5. Found: backfill Redis    │
│    Not found: set negative  │
│    cache                    │
├─────────────────────────────┤
│ 6. Enqueue click event      │
│    (non-blocking)           │
└─────────────────────────────┘
```

### Hard Rules

| Forbidden in redirect path | Why |
|---------------------------|-----|
| Sync DB write | Latency, connection exhaustion |
| External HTTP call | Unpredictable latency |
| Complex ORM operations | N+1, memory allocation |
| Full request logging | I/O bottleneck at high QPS |

### Performance Targets

- P50 latency: < 5ms (app only)
- Redis hit rate: > 95%
- Zero sync writes on redirect

---

## Data Model

### Postgres

```sql
CREATE TABLE links (
    code        TEXT PRIMARY KEY,
    long_url    TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ,
    is_disabled BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE click_events (
    id          BIGSERIAL,
    code        TEXT NOT NULL,
    ts          TIMESTAMPTZ NOT NULL,
    ip_hash     TEXT,
    ua_hash     TEXT,
    referer     TEXT,
    PRIMARY KEY (id, ts)
) PARTITION BY RANGE (ts);

CREATE INDEX idx_clicks_code_ts ON click_events (code, ts);

CREATE TABLE api_tokens (
    id           SERIAL PRIMARY KEY,
    token_hash   VARCHAR(64) NOT NULL UNIQUE,
    name         VARCHAR(100) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    disabled     BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_api_tokens_hash ON api_tokens (token_hash) WHERE NOT disabled;
```

### Redis Keys

| Key Pattern | Type | TTL | Purpose |
|-------------|------|-----|---------|
| `su:link:{code}` | string | LRU | URL cache |
| `su:miss:{code}` | string | 60s | Negative cache |
| `su:clicks` | stream | - | Click event queue |
| `su:ratelimit:{ip}` | string | 60s | Rate limit counter |

---

## Code Generation

Random base62 + unique constraint retry (max 3 attempts):

```go
const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
```

---

## Async Click Processing

```
Redirect ──XADD──▶ Redis Streams ──batch──▶ Consumer ──bulk insert──▶ Postgres
```

**Consumer behavior**:
- Batch size: 500 events
- Flush interval: 200ms
- At-least-once delivery

---

## Full Configuration

```bash
# Server
HTTP_ADDR=:8080
HTTP_READ_TIMEOUT=5s
HTTP_WRITE_TIMEOUT=5s
TRUSTED_PROXIES=127.0.0.1,172.16.0.0/12

# Redirect
REDIRECT_STATUS_CODE=302
CODE_LENGTH=8

# Redis
REDIS_ADDR=localhost:6379
REDIS_DIAL_TIMEOUT=200ms
REDIS_RW_TIMEOUT=200ms
REDIS_KEY_PREFIX=su
NEGATIVE_CACHE_TTL=60s

# Postgres
DATABASE_URL=postgres://user:pass@localhost:5432/shorturl
DB_MAX_OPEN_CONNS=20
DB_MAX_IDLE_CONNS=10

# Worker
STREAM_NAME=su:clicks
STREAM_GROUP=su-worker
WORKER_BATCH_SIZE=500
WORKER_FLUSH_INTERVAL=200ms

# Admin
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
ADMIN_TOKEN_TTL=24h

# Rate Limiting
RATE_LIMIT_REQUESTS=60
RATE_LIMIT_WINDOW=60s
```

---

## Observability

### Metrics (Prometheus)

```
redirect_requests_total{status="302|404|410"}
redirect_latency_seconds_bucket{le="0.005|0.01|0.05|0.1"}
cache_hits_total
cache_misses_total
click_events_processed_total
```

### Logging

Structured JSON, sampled on redirect path.

---

## Project Structure

```
go2short/
├── cmd/app/           # main.go
├── internal/
│   ├── config/        # env loading
│   ├── handler/       # HTTP handlers
│   ├── redirect/      # redirect logic (perf critical)
│   ├── link/          # link CRUD
│   ├── cache/         # Redis operations
│   ├── store/         # Postgres operations
│   ├── events/        # stream producer/consumer
│   └── middleware/    # auth, rate limiting
├── migrations/        # SQL migrations
├── web/               # Vue 3 admin (embedded)
└── docs/
```

---

## Security

- [x] URL validation: http/https only
- [x] Block private IP ranges (SSRF prevention)
- [x] Hash IP/UA before storage
- [x] API Token auth (SHA256 hashed)
- [x] Rate limiting
- [x] No secrets in logs

---

## Not Doing (v1)

- No Kafka/ES/ClickHouse
- No multi-tenancy
- No sharding
- No real-time analytics in redirect path

---

## Review Checklist

Before merging redirect-path changes:

1. Sync DB write? **REJECT**
2. External call? **REJECT**
3. Negative cache handled? **REQUIRED**
4. Metrics updated? **REQUIRED**
