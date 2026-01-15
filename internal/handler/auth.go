package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wyp0596/go2short/internal/auth"
	"github.com/wyp0596/go2short/internal/config"
	"github.com/wyp0596/go2short/internal/middleware"
	"github.com/wyp0596/go2short/internal/store"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	cfg          *config.Config
	store        *store.Store
	auth         *middleware.AuthMiddleware
	googleConfig *oauth2.Config
	githubConfig *oauth2.Config
}

func NewAuthHandler(cfg *config.Config, s *store.Store, authMw *middleware.AuthMiddleware) *AuthHandler {
	h := &AuthHandler{
		cfg:   cfg,
		store: s,
		auth:  authMw,
	}
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		h.googleConfig = auth.GoogleConfig(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.BaseURL+"/api/auth/google/callback")
	}
	if cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "" {
		h.githubConfig = auth.GitHubConfig(cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.BaseURL+"/api/auth/github/callback")
	}
	return h
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Register handles user registration with email+password.
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Check if user exists
	existing, err := h.store.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	// Hash password
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Create user
	userID, err := h.store.CreateUser(c.Request.Context(), req.Email, hash, "email", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Generate token and login
	token, err := h.generateAndSaveToken(c, userID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

type userLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles user login with email+password.
func (h *AuthHandler) Login(c *gin.Context) {
	var req userLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := h.store.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Check password (only for email provider users)
	if user.Provider != "email" || !user.PasswordHash.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "please login with " + user.Provider})
		return
	}
	if !auth.CheckPassword(req.Password, user.PasswordHash.String) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Update last login
	_ = h.store.UpdateUserLastLogin(c.Request.Context(), user.ID)

	// Generate token
	token, err := h.generateAndSaveToken(c, user.ID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GoogleRedirect redirects to Google OAuth.
func (h *AuthHandler) GoogleRedirect(c *gin.Context) {
	if h.googleConfig == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Google OAuth not configured"})
		return
	}
	state := generateState()
	// Store state in cookie for validation
	c.SetCookie("oauth_state", state, 300, "/", "", false, true)
	url := h.googleConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles Google OAuth callback.
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	if h.googleConfig == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Google OAuth not configured"})
		return
	}

	// Validate state
	state, _ := c.Cookie("oauth_state")
	if state == "" || state != c.Query("state") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	// Exchange code for token
	oauthToken, err := h.googleConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to exchange code"})
		return
	}

	// Get user info from Google
	client := h.googleConfig.Client(c.Request.Context(), oauthToken)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user info"})
		return
	}

	// Find or create user
	token, err := h.findOrCreateOAuthUser(c, "google", userInfo.ID, userInfo.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process user"})
		return
	}

	// Redirect to frontend with token
	c.Redirect(http.StatusTemporaryRedirect, "/admin?token="+token)
}

// GitHubRedirect redirects to GitHub OAuth.
func (h *AuthHandler) GitHubRedirect(c *gin.Context) {
	if h.githubConfig == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "GitHub OAuth not configured"})
		return
	}
	state := generateState()
	c.SetCookie("oauth_state", state, 300, "/", "", false, true)
	url := h.githubConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubCallback handles GitHub OAuth callback.
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	if h.githubConfig == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "GitHub OAuth not configured"})
		return
	}

	// Validate state
	state, _ := c.Cookie("oauth_state")
	if state == "" || state != c.Query("state") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	// Exchange code for token
	oauthToken, err := h.githubConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to exchange code"})
		return
	}

	// Get user info from GitHub
	client := h.githubConfig.Client(c.Request.Context(), oauthToken)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user info"})
		return
	}

	// GitHub might not return email, try emails endpoint
	if userInfo.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if json.NewDecoder(emailResp.Body).Decode(&emails) == nil {
				for _, e := range emails {
					if e.Primary {
						userInfo.Email = e.Email
						break
					}
				}
			}
		}
	}

	if userInfo.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not get email from GitHub"})
		return
	}

	// Find or create user
	token, err := h.findOrCreateOAuthUser(c, "github", fmt.Sprintf("%d", userInfo.ID), userInfo.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process user"})
		return
	}

	// Redirect to frontend with token
	c.Redirect(http.StatusTemporaryRedirect, "/admin?token="+token)
}

func (h *AuthHandler) findOrCreateOAuthUser(c *gin.Context, provider, providerID, email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	// Try to find by provider+providerID
	user, err := h.store.GetUserByProvider(c.Request.Context(), provider, providerID)
	if err != nil {
		return "", err
	}

	if user == nil {
		// Try to find by email
		user, err = h.store.GetUserByEmail(c.Request.Context(), email)
		if err != nil {
			return "", err
		}
	}

	var userID int
	if user == nil {
		// Create new user
		userID, err = h.store.CreateUser(c.Request.Context(), email, "", provider, providerID)
		if err != nil {
			return "", err
		}
	} else {
		userID = user.ID
		_ = h.store.UpdateUserLastLogin(c.Request.Context(), userID)
	}

	return h.generateAndSaveToken(c, userID, false)
}

func (h *AuthHandler) generateAndSaveToken(c *gin.Context, userID int, isAdmin bool) (string, error) {
	token := generateToken()
	err := h.auth.SaveToken(c.Request.Context(), token, userID, isAdmin, h.cfg.AdminTokenTTL)
	if err != nil {
		return "", err
	}
	return token, nil
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// OAuthEnabled returns available OAuth providers.
func (h *AuthHandler) OAuthEnabled(c *gin.Context) {
	providers := make([]string, 0)
	if h.googleConfig != nil {
		providers = append(providers, "google")
	}
	if h.githubConfig != nil {
		providers = append(providers, "github")
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}
