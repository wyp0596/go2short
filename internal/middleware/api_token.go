package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wyp0596/go2short/internal/store"
)

type APITokenMiddleware struct {
	store *store.Store
}

func NewAPITokenMiddleware(s *store.Store) *APITokenMiddleware {
	return &APITokenMiddleware{store: s}
}

// HashToken returns SHA256 hash of token.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// RequireAPIToken is a Gin middleware that validates API token.
// Accepts: Authorization: Bearer <token> or X-API-Key: <token>
func (m *APITokenMiddleware) RequireAPIToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing api token"})
			return
		}

		tokenHash := HashToken(token)
		apiToken, err := m.store.GetAPITokenByHash(c.Request.Context(), tokenHash)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "auth check failed"})
			return
		}
		if apiToken == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api token"})
			return
		}

		// Update last used (async, don't block request)
		go m.store.UpdateAPITokenLastUsed(context.Background(), apiToken.ID)

		c.Set("api_token_id", apiToken.ID)
		c.Set("api_token_name", apiToken.Name)
		// Set userID from token (null means global/admin token)
		if apiToken.UserID.Valid {
			c.Set("userID", int(apiToken.UserID.Int32))
		} else {
			c.Set("userID", 0) // Global token, no user restriction
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	// Try Authorization: Bearer <token>
	auth := c.GetHeader("Authorization")
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}
	// Try X-API-Key header
	if key := c.GetHeader("X-API-Key"); key != "" {
		return key
	}
	return ""
}
