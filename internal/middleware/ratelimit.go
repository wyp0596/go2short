package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis  *redis.Client
	prefix string
	limit  int
	window time.Duration
}

func NewRateLimiter(client *redis.Client, prefix string, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:  client,
		prefix: prefix,
		limit:  limit,
		window: window,
	}
}

func (r *RateLimiter) key(identifier string) string {
	return r.prefix + ":ratelimit:" + identifier
}

// Limit returns a Gin middleware that rate limits by client IP.
func (r *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ip := c.ClientIP()
		key := r.key(ip)

		// Sliding window using Redis INCR + EXPIRE
		count, err := r.redis.Incr(ctx, key).Result()
		if err != nil {
			// Fail open on Redis error
			c.Next()
			return
		}

		// Set expiry on first request
		if count == 1 {
			r.redis.Expire(ctx, key, r.window)
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(r.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(max(0, r.limit-int(count))))

		if int(count) > r.limit {
			ttl, _ := r.redis.TTL(ctx, key).Result()
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}
