package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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

// ValidateToken checks if token exists in Redis and returns userID and isAdmin.
// Returns userID=0 for admin, userID>0 for regular user.
func (m *AuthMiddleware) ValidateToken(ctx context.Context, token string) (userID int, isAdmin bool, err error) {
	data, err := m.redis.Get(ctx, m.tokenKey(token)).Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	// Parse "userID:isAdmin" format
	parts := strings.Split(data, ":")
	if len(parts) != 2 {
		return 0, false, nil
	}

	userID, _ = strconv.Atoi(parts[0])
	isAdmin = parts[1] == "true"
	return userID, isAdmin, nil
}

// SaveToken stores token in Redis with userID and isAdmin info.
// userID=0 for admin users.
func (m *AuthMiddleware) SaveToken(ctx context.Context, token string, userID int, isAdmin bool, ttl time.Duration) error {
	data := fmt.Sprintf("%d:%v", userID, isAdmin)
	return m.redis.Set(ctx, m.tokenKey(token), data, ttl).Err()
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
		userID, isAdmin, err := m.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "auth check failed"})
			return
		}
		if userID == 0 && !isAdmin {
			// Token not found or invalid format
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("token", token)
		c.Set("userID", userID)
		c.Set("isAdmin", isAdmin)
		c.Next()
	}
}
