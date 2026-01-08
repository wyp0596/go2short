package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyp0596/go2short/internal/cache"
	"github.com/wyp0596/go2short/internal/config"
	"github.com/wyp0596/go2short/internal/middleware"
	"github.com/wyp0596/go2short/internal/store"
)

type AdminHandler struct {
	store   *store.Store
	cache   *cache.Cache
	auth    *middleware.AuthMiddleware
	cfg     *config.Config
	baseURL string
}

func NewAdminHandler(s *store.Store, c *cache.Cache, auth *middleware.AuthMiddleware, cfg *config.Config) *AdminHandler {
	return &AdminHandler{
		store:   s,
		cache:   c,
		auth:    auth,
		cfg:     cfg,
		baseURL: cfg.BaseURL,
	}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// Login handles admin login.
func (h *AdminHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Username != h.cfg.AdminUsername || req.Password != h.cfg.AdminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate random token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	token := hex.EncodeToString(b)

	// Save token to Redis
	if err := h.auth.SaveToken(c.Request.Context(), token, h.cfg.AdminTokenTTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save token"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{Token: token})
}

// Logout handles admin logout.
func (h *AdminHandler) Logout(c *gin.Context) {
	token, _ := c.Get("token")
	if t, ok := token.(string); ok {
		h.auth.DeleteToken(c.Request.Context(), t)
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

type linkResponse struct {
	Code       string     `json:"code"`
	ShortURL   string     `json:"short_url"`
	LongURL    string     `json:"long_url"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	IsDisabled bool       `json:"is_disabled"`
}

type linksResponse struct {
	Links []linkResponse `json:"links"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// ListLinks returns paginated links.
func (h *AdminHandler) ListLinks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	links, total, err := h.store.ListLinks(c.Request.Context(), search, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list links"})
		return
	}

	resp := linksResponse{
		Links: make([]linkResponse, 0, len(links)),
		Total: total,
		Page:  page,
		Limit: limit,
	}

	for _, l := range links {
		lr := linkResponse{
			Code:       l.Code,
			ShortURL:   h.baseURL + "/" + l.Code,
			LongURL:    l.LongURL,
			CreatedAt:  l.CreatedAt,
			IsDisabled: l.IsDisabled,
		}
		if l.ExpiresAt.Valid {
			lr.ExpiresAt = &l.ExpiresAt.Time
		}
		resp.Links = append(resp.Links, lr)
	}

	c.JSON(http.StatusOK, resp)
}

type updateLinkRequest struct {
	LongURL   string     `json:"long_url" binding:"required,url"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// UpdateLink updates a link.
func (h *AdminHandler) UpdateLink(c *gin.Context) {
	code := c.Param("code")
	var req updateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.store.UpdateLink(c.Request.Context(), code, req.LongURL, req.ExpiresAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update link"})
		return
	}

	// Update cache
	h.cache.SetURL(c.Request.Context(), code, req.LongURL)

	c.JSON(http.StatusOK, gin.H{"message": "link updated"})
}

// DeleteLink removes a link.
func (h *AdminHandler) DeleteLink(c *gin.Context) {
	code := c.Param("code")

	err := h.store.DeleteLink(c.Request.Context(), code)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "link deleted"})
}

type disableLinkRequest struct {
	Disabled bool `json:"disabled"`
}

// SetLinkDisabled enables or disables a link.
func (h *AdminHandler) SetLinkDisabled(c *gin.Context) {
	code := c.Param("code")
	var req disableLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.store.SetLinkDisabled(c.Request.Context(), code, req.Disabled)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "link updated"})
}

// GetLinkStats returns click statistics for a link.
func (h *AdminHandler) GetLinkStats(c *gin.Context) {
	code := c.Param("code")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	stats, err := h.store.GetLinkStats(c.Request.Context(), code, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetOverviewStats returns overall statistics.
func (h *AdminHandler) GetOverviewStats(c *gin.Context) {
	stats, err := h.store.GetOverviewStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTopLinks returns top links by click count.
func (h *AdminHandler) GetTopLinks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if limit < 1 || limit > 100 {
		limit = 10
	}
	if days < 1 || days > 365 {
		days = 30
	}

	links, err := h.store.GetTopLinks(c.Request.Context(), limit, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get top links"})
		return
	}

	// Add short_url to response
	type topLinkResponse struct {
		Code       string `json:"code"`
		ShortURL   string `json:"short_url"`
		LongURL    string `json:"long_url"`
		ClickCount int    `json:"click_count"`
	}
	resp := make([]topLinkResponse, 0, len(links))
	for _, l := range links {
		resp = append(resp, topLinkResponse{
			Code:       l.Code,
			ShortURL:   h.baseURL + "/" + l.Code,
			LongURL:    l.LongURL,
			ClickCount: l.ClickCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"links": resp})
}

// GetClickTrend returns click trend for the last N days.
func (h *AdminHandler) GetClickTrend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	trend, err := h.store.GetClickTrend(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get trend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trend": trend})
}

type createTokenRequest struct {
	Name string `json:"name" binding:"required"`
}

type createTokenResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

// CreateAPIToken creates a new API token and returns the plaintext token (only once).
func (h *AdminHandler) CreateAPIToken(c *gin.Context) {
	var req createTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	// Generate random token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	token := hex.EncodeToString(b)
	tokenHash := middleware.HashToken(token)

	id, err := h.store.CreateAPIToken(c.Request.Context(), tokenHash, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	c.JSON(http.StatusCreated, createTokenResponse{
		ID:    id,
		Name:  req.Name,
		Token: token,
	})
}

type apiTokenResponse struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// ListAPITokens returns all API tokens (without token values).
func (h *AdminHandler) ListAPITokens(c *gin.Context) {
	tokens, err := h.store.ListAPITokens(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tokens"})
		return
	}

	resp := make([]apiTokenResponse, 0, len(tokens))
	for _, t := range tokens {
		r := apiTokenResponse{
			ID:        t.ID,
			Name:      t.Name,
			CreatedAt: t.CreatedAt,
		}
		if t.LastUsedAt.Valid {
			r.LastUsedAt = &t.LastUsedAt.Time
		}
		resp = append(resp, r)
	}

	c.JSON(http.StatusOK, gin.H{"tokens": resp})
}

// DeleteAPIToken removes an API token.
func (h *AdminHandler) DeleteAPIToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token id"})
		return
	}

	err = h.store.DeleteAPIToken(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token deleted"})
}
