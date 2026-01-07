package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/warren/go2short/internal/events"
	"github.com/warren/go2short/internal/metrics"
	"github.com/warren/go2short/internal/redirect"
)

type RedirectHandler struct {
	service  *redirect.Service
	producer *events.Producer
}

func NewRedirectHandler(s *redirect.Service, p *events.Producer) *RedirectHandler {
	return &RedirectHandler{
		service:  s,
		producer: p,
	}
}

func (h *RedirectHandler) Handle(c *gin.Context) {
	start := time.Now()
	code := c.Param("code")

	result, err := h.service.Resolve(c.Request.Context(), code)
	if err != nil {
		metrics.RedirectRequests.WithLabelValues("500").Inc()
		c.Status(http.StatusInternalServerError)
		return
	}

	// Record cache metrics
	if result.CacheHit {
		metrics.CacheHits.Inc()
	} else {
		metrics.CacheMisses.Inc()
	}

	// Record latency
	metrics.RedirectLatency.WithLabelValues().Observe(time.Since(start).Seconds())
	metrics.RedirectRequests.WithLabelValues(strconv.Itoa(result.StatusCode)).Inc()

	switch result.StatusCode {
	case 302:
		// Async enqueue click event
		h.producer.EnqueueAsync(&events.ClickEvent{
			Code:      code,
			Timestamp: time.Now().UTC(),
			IPHash:    hashString(c.ClientIP()),
			UAHash:    hashString(c.GetHeader("User-Agent")),
			Referer:   c.GetHeader("Referer"),
			ReqID:     c.GetHeader("X-Request-ID"),
		})
		metrics.ClickEventsEnqueued.Inc()
		c.Redirect(http.StatusFound, result.URL)

	case 404:
		c.Status(http.StatusNotFound)

	case 410:
		c.Status(http.StatusGone)

	default:
		c.Status(http.StatusInternalServerError)
	}
}

func hashString(s string) string {
	if s == "" {
		return ""
	}
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8]) // first 8 bytes = 16 hex chars
}
