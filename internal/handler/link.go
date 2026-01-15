package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"
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

	// Get userID from API token (set by api_token middleware)
	var userID *int
	if uid, ok := c.Get("userID"); ok {
		if id, ok := uid.(int); ok && id > 0 {
			userID = &id
		}
	}

	result, err := h.service.Create(c.Request.Context(), &link.CreateRequest{
		LongURL:    req.LongURL,
		ExpiresAt:  expiresAt,
		CustomCode: customCode,
		UserID:     userID,
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

// QRCode generates a QR code image for a short link.
func (h *LinkHandler) QRCode(c *gin.Context) {
	code := c.Param("code")
	size := 256 // default size

	if s := c.Query("size"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			size = n
		}
	}
	// Clamp size
	if size < 128 {
		size = 128
	}
	if size > 1024 {
		size = 1024
	}

	url := h.baseURL + "/" + code
	png, err := qrcode.Encode(url, qrcode.Medium, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate QR code"})
		return
	}

	c.Data(http.StatusOK, "image/png", png)
}

// Preview returns the target URL without redirecting.
func (h *LinkHandler) Preview(c *gin.Context) {
	code := c.Param("code")

	longURL, err := h.service.GetLongURL(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if longURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     code,
		"long_url": longURL,
	})
}

type batchCreateItem struct {
	LongURL    string  `json:"long_url" binding:"required"`
	ExpiresAt  *string `json:"expires_at,omitempty"`
	CustomCode *string `json:"custom_code,omitempty"`
}

type batchCreateRequest struct {
	Items []batchCreateItem `json:"items" binding:"required,min=1,max=100"`
}

type batchCreateResultItem struct {
	Index    int    `json:"index"`
	Code     string `json:"code,omitempty"`
	ShortURL string `json:"short_url,omitempty"`
	Error    string `json:"error,omitempty"`
}

// BatchCreate creates multiple links in one request.
func (h *LinkHandler) BatchCreate(c *gin.Context) {
	var req batchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Get userID from API token
	var userID *int
	if uid, ok := c.Get("userID"); ok {
		if id, ok := uid.(int); ok && id > 0 {
			userID = &id
		}
	}

	// Convert to service requests
	serviceReqs := make([]link.BatchCreateRequest, len(req.Items))
	for i, item := range req.Items {
		var expiresAt *time.Time
		if item.ExpiresAt != nil {
			t, err := time.Parse(time.RFC3339, *item.ExpiresAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid expires_at format",
					"index": i,
				})
				return
			}
			expiresAt = &t
		}
		customCode := ""
		if item.CustomCode != nil {
			customCode = *item.CustomCode
		}
		serviceReqs[i] = link.BatchCreateRequest{
			LongURL:    item.LongURL,
			ExpiresAt:  expiresAt,
			CustomCode: customCode,
			UserID:     userID,
		}
	}

	results := h.service.BatchCreate(c.Request.Context(), serviceReqs)

	response := make([]batchCreateResultItem, len(results))
	for i, r := range results {
		if r.Error != nil {
			response[i] = batchCreateResultItem{
				Index: r.Index,
				Error: r.Error.Error(),
			}
		} else {
			response[i] = batchCreateResultItem{
				Index:    r.Index,
				Code:     r.Code,
				ShortURL: h.baseURL + "/" + r.Code,
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"results": response})
}
