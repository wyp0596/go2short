package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type AuthMiddleware struct {
	redis  *redis.Client
	prefix string
}

func NewAuthMiddleware(client *redis.Client, prefix string) *AuthMiddleware {
	return &AuthMiddleware{redis: client, prefix: prefix}
}

func (m *AuthMiddleware) tokenKey(token string) string {
	return m.prefix + ":token:" + token
}

// ValidateToken checks if token exists in Redis.
func (m *AuthMiddleware) ValidateToken(ctx context.Context, token string) (bool, error) {
	exists, err := m.redis.Exists(ctx, m.tokenKey(token)).Result()
	return exists > 0, err
}

// SaveToken stores token in Redis with TTL.
func (m *AuthMiddleware) SaveToken(ctx context.Context, token string, ttl time.Duration) error {
	return m.redis.Set(ctx, m.tokenKey(token), "1", ttl).Err()
}

// DeleteToken removes token from Redis.
func (m *AuthMiddleware) DeleteToken(ctx context.Context, token string) error {
	return m.redis.Del(ctx, m.tokenKey(token)).Err()
}

// RequireAuth is a Gin middleware that validates Bearer token.
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		token := parts[1]
		valid, err := m.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "auth check failed"})
			return
		}
		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("token", token)
		c.Next()
	}
}
