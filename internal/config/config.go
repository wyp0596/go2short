package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	// Server
	HTTPAddr         string
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	BaseURL          string
	TrustedProxies   []string

	// Admin
	AdminUsername string
	AdminPassword string
	AdminTokenTTL time.Duration

	// Redirect
	RedirectStatusCode int
	CodeLength         int

	// Redis
	RedisURL         string
	RedisAddr        string
	RedisDialTimeout time.Duration
	RedisRWTimeout   time.Duration
	RedisKeyPrefix   string
	NegativeCacheTTL time.Duration

	// Postgres
	DatabaseURL    string
	DBMaxOpenConns int
	DBMaxIdleConns int

	// Worker
	StreamName          string
	StreamGroup         string
	WorkerBatchSize     int
	WorkerFlushInterval time.Duration
}

func Load() *Config {
	return &Config{
		HTTPAddr:            getHTTPAddr(),
		HTTPReadTimeout:     getDuration("HTTP_READ_TIMEOUT", 5*time.Second),
		HTTPWriteTimeout:    getDuration("HTTP_WRITE_TIMEOUT", 5*time.Second),
		BaseURL:             getEnv("BASE_URL", "http://localhost:8080"),
		TrustedProxies:      getStringSlice("TRUSTED_PROXIES", nil),
		AdminUsername:       getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:       getEnv("ADMIN_PASSWORD", "admin123"),
		AdminTokenTTL:       getDuration("ADMIN_TOKEN_TTL", 24*time.Hour),
		RedirectStatusCode:  getInt("REDIRECT_STATUS_CODE", 302),
		CodeLength:          getInt("CODE_LENGTH", 8),
		RedisURL:            getEnv("REDIS_URL", ""),
		RedisAddr:           getEnv("REDIS_ADDR", "localhost:6379"),
		RedisDialTimeout:    getDuration("REDIS_DIAL_TIMEOUT", 200*time.Millisecond),
		RedisRWTimeout:      getDuration("REDIS_RW_TIMEOUT", 200*time.Millisecond),
		RedisKeyPrefix:      getEnv("REDIS_KEY_PREFIX", "su"),
		NegativeCacheTTL:    getDuration("NEGATIVE_CACHE_TTL", 60*time.Second),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/go2short?sslmode=disable"),
		DBMaxOpenConns:      getInt("DB_MAX_OPEN_CONNS", 20),
		DBMaxIdleConns:      getInt("DB_MAX_IDLE_CONNS", 10),
		StreamName:          getEnv("STREAM_NAME", "su:clicks"),
		StreamGroup:         getEnv("STREAM_GROUP", "su-worker"),
		WorkerBatchSize:     getInt("WORKER_BATCH_SIZE", 500),
		WorkerFlushInterval: getDuration("WORKER_FLUSH_INTERVAL", 200*time.Millisecond),
	}
}

func getHTTPAddr() string {
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		return addr
	}
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":8080"
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}

func getInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return defaultVal
}

func getStringSlice(key string, defaultVal []string) []string {
	if v := os.Getenv(key); v != "" {
		var result []string
		for _, s := range strings.Split(v, ",") {
			if s = strings.TrimSpace(s); s != "" {
				result = append(result, s)
			}
		}
		return result
	}
	return defaultVal
}
