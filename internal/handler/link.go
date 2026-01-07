package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyp0596/go2short/internal/link"
)

type LinkHandler struct {
	service *link.Service
	baseURL string
}

func NewLinkHandler(s *link.Service, baseURL string) *LinkHandler {
	return &LinkHandler{
		service: s,
		baseURL: baseURL,
	}
}

type createRequest struct {
	LongURL    string  `json:"long_url" binding:"required"`
	ExpiresAt  *string `json:"expires_at,omitempty"`
	CustomCode *string `json:"custom_code,omitempty"`
}

type createResponse struct {
	Code      string `json:"code"`
	ShortURL  string `json:"short_url"`
	CreatedAt string `json:"created_at"`
}

func (h *LinkHandler) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Parse expires_at
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expires_at format"})
			return
		}
		expiresAt = &t
	}

	// Get custom code
	customCode := ""
	if req.CustomCode != nil {
		customCode = *req.CustomCode
	}

	result, err := h.service.Create(c.Request.Context(), &link.CreateRequest{
		LongURL:    req.LongURL,
		ExpiresAt:  expiresAt,
		CustomCode: customCode,
	})

	if err != nil {
		switch err {
		case link.ErrInvalidURL:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid URL (http/https only)"})
		case link.ErrURLTooLong:
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL too long (max 2048)"})
		case link.ErrBlockedIP:
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL points to private IP"})
		case link.ErrCodeTaken:
			c.JSON(http.StatusConflict, gin.H{"error": "custom code already taken"})
		case link.ErrInvalidCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid custom code (6-12 chars, base62)"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusCreated, createResponse{
		Code:      result.Code,
		ShortURL:  h.baseURL + "/" + result.Code,
		CreatedAt: result.CreatedAt.Format(time.RFC3339),
	})
}
