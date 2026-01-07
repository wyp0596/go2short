package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyp0596/go2short/internal/config"
)

type Cache struct {
	client *redis.Client
	prefix string
	negTTL time.Duration
}

func New(cfg *config.Config) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisRWTimeout,
		WriteTimeout: cfg.RedisRWTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RedisDialTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Cache{
		client: client,
		prefix: cfg.RedisKeyPrefix,
		negTTL: cfg.NegativeCacheTTL,
	}, nil
}

func (c *Cache) linkKey(code string) string {
	return c.prefix + ":link:" + code
}

func (c *Cache) missKey(code string) string {
	return c.prefix + ":miss:" + code
}

// GetURL returns the long URL for a code. Returns empty string if not found.
func (c *Cache) GetURL(ctx context.Context, code string) (string, error) {
	val, err := c.client.Get(ctx, c.linkKey(code)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// SetURL caches a code -> URL mapping.
func (c *Cache) SetURL(ctx context.Context, code, url string) error {
	return c.client.Set(ctx, c.linkKey(code), url, 0).Err()
}

// IsMiss checks if code is in negative cache.
func (c *Cache) IsMiss(ctx context.Context, code string) (bool, error) {
	exists, err := c.client.Exists(ctx, c.missKey(code)).Result()
	return exists > 0, err
}

// SetMiss marks a code as not found (negative cache).
func (c *Cache) SetMiss(ctx context.Context, code string) error {
	return c.client.Set(ctx, c.missKey(code), "1", c.negTTL).Err()
}

// Client returns the underlying redis client for stream operations.
func (c *Cache) Client() *redis.Client {
	return c.client
}

func (c *Cache) Close() error {
	return c.client.Close()
}
