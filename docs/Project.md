# go2short - Architecture & Specification

> A minimal, high-performance URL shortener. Container-ready, horizontally scalable.

---

## Design Principles

1. **Redirect path is sacred** - Zero sync DB writes, minimal middleware
2. **Simple > Clever** - No over-abstraction, no premature optimization
3. **Explicit > Magic** - Configuration via env, no hidden behavior
4. **Observability built-in** - Metrics and structured logs from day one

---

## Architecture Overview

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Gateway   │────▶│     App     │────▶│    Redis    │
│  (Nginx)    │     │   (Go/Gin)  │     │  Cache+MQ   │
└─────────────┘     └──────┬──────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │  Postgres   │
                    │   (Data)    │
                    └─────────────┘
```

**Single binary MVP**: App process includes redirect handler + click event consumer goroutine.

---

## Critical Path: Redirect

```
GET /:code
    │
    ▼
┌─────────────────────────────┐
│ 1. Validate code (length,   │
│    charset: base62)         │
├─────────────────────────────┤
│ 2. Redis GET su:link:{code} │──▶ HIT ──▶ 302 + async enqueue
├─────────────────────────────┤
│ 3. MISS: Check negative     │
│    cache su:miss:{code}     │──▶ HIT ──▶ 404
├─────────────────────────────┤
│ 4. Query Postgres links     │
│    WHERE code = ?           │
├─────────────────────────────┤
│ 5. Found: backfill Redis,   │
│    return 302               │
│    Not found: set negative  │
│    cache, return 404        │
├─────────────────────────────┤
│ 6. Enqueue click event      │
│    (non-blocking)           │
└─────────────────────────────┘
```

### Hard Rules (Non-Negotiable)

| Forbidden in redirect path | Why |
|---------------------------|-----|
| Sync DB write | Latency, connection exhaustion |
| External HTTP call | Unpredictable latency |
| Complex ORM operations | N+1, memory allocation |
| Full request logging | I/O bottleneck at high QPS |

### Performance Targets

- P50 latency: < 5ms (app only, excludes network)
- Redis hit rate: > 95%
- Zero sync writes on redirect

---

## Data Model

### Postgres

```sql
-- links: source of truth
CREATE TABLE links (
    code        TEXT PRIMARY KEY,
    long_url    TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ,
    is_disabled BOOLEAN NOT NULL DEFAULT FALSE
);

-- click_events: partitioned by time
CREATE TABLE click_events (
    id          BIGSERIAL,
    code        TEXT NOT NULL,
    ts          TIMESTAMPTZ NOT NULL,
    ip_hash     TEXT,
    ua_hash     TEXT,
    referer     TEXT,
    PRIMARY KEY (id, ts)
) PARTITION BY RANGE (ts);

-- Index for aggregation queries
CREATE INDEX idx_clicks_code_ts ON click_events (code, ts);
```

### Redis Keys

| Key Pattern | Type | TTL | Purpose |
|-------------|------|-----|---------|
| `su:link:{code}` | string | none/LRU | URL cache |
| `su:miss:{code}` | string | 60s | Negative cache |
| `su:clicks` | stream | - | Click event queue |

---

## Code Generation

**Strategy**: Random base62 + unique constraint retry

```go
const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const codeLen = 8 // configurable

func generateCode() string {
    b := make([]byte, codeLen)
    for i := range b {
        b[i] = charset[rand.Intn(len(charset))]
    }
    return string(b)
}
```

On unique constraint violation: retry (max 3 attempts).

---

## Async Click Processing

```
┌──────────┐    XADD    ┌──────────┐   batch   ┌──────────┐
│ Redirect │──────────▶│  Redis   │──────────▶│ Consumer │
│ Handler  │           │ Streams  │           │ Goroutine│
└──────────┘           └──────────┘           └────┬─────┘
                                                   │
                                              bulk insert
                                                   │
                                                   ▼
                                            ┌──────────┐
                                            │ Postgres │
                                            └──────────┘
```

**Event payload** (JSON):
```json
{
  "code": "abc123",
  "ts": "2024-01-01T00:00:00Z",
  "ip_hash": "sha256...",
  "ua_hash": "sha256...",
  "referer": "https://...",
  "req_id": "uuid"
}
```

**Consumer behavior**:
- Batch size: 500 events
- Flush interval: 200ms (whichever comes first)
- At-least-once delivery
- Failed events: log + optional DLQ stream

---

## API Specification

### Redirect

```
GET /:code

Success: 302 Found
Location: {long_url}

Not found: 404
Disabled/Expired: 410
```

### Create Link

```
POST /api/links
Content-Type: application/json

Request:
{
  "long_url": "https://example.com/very/long/path",
  "expires_at": "2024-12-31T23:59:59Z",  // optional
  "custom_code": "mycode"                 // optional
}

Response: 201 Created
{
  "code": "abc123",
  "short_url": "https://go2.sh/abc123",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Validation**:
- URL: http/https only, max 2048 chars
- Code: base62, 6-12 chars
- Block internal IPs (SSRF prevention)

---

## Configuration

All via environment variables:

```bash
# Server
HTTP_ADDR=:8080
HTTP_READ_TIMEOUT=5s
HTTP_WRITE_TIMEOUT=5s

# Redirect behavior
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
```

---

## Observability

### Metrics (Prometheus)

```
# Redirect
redirect_requests_total{status="302|404|410"}
redirect_latency_seconds_bucket{le="0.005|0.01|0.05|0.1"}

# Cache
cache_hits_total
cache_misses_total

# Database
db_queries_total{operation="select|insert"}
db_latency_seconds_bucket

# Worker
click_events_enqueued_total
click_events_processed_total
stream_lag_messages
```

### Logging

Structured JSON, sampled on redirect path:

```json
{
  "level": "info",
  "ts": "2024-01-01T00:00:00Z",
  "req_id": "uuid",
  "path": "/abc123",
  "status": 302,
  "latency_ms": 2,
  "cache": "hit"
}
```

---

## Project Structure

```
go2short/
├── cmd/
│   └── app/              # main.go
├── internal/
│   ├── config/           # env loading
│   ├── handler/          # HTTP handlers
│   ├── redirect/         # redirect logic (performance critical)
│   ├── link/             # link CRUD (can use ORM)
│   ├── cache/            # Redis operations
│   ├── store/            # Postgres operations
│   ├── events/           # stream producer/consumer
│   └── metrics/          # Prometheus collectors
├── migrations/           # SQL migrations
├── web/                  # Vue admin (optional)
├── docs/
└── docker-compose.yml
```

---

## Security Checklist

- [ ] URL validation: http/https only
- [ ] Block private IP ranges (10.x, 172.16-31.x, 192.168.x)
- [ ] Hash IP/UA before storage
- [ ] Admin API behind auth
- [ ] Rate limiting at gateway
- [ ] No secrets in logs

---

## Not Doing (v1)

- No Kafka/ES/ClickHouse
- No multi-tenancy
- No sharding
- No real-time analytics in redirect path
- No complex fraud detection (async only)

---

## Review Checklist

Before merging any redirect-path change:

1. Does it add sync DB write? **REJECT**
2. Does it add external call? **REJECT**
3. Is negative cache handled? **REQUIRED**
4. Are metrics updated? **REQUIRED**
5. Is the change under 50 lines? **PREFERRED**
