package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/warren/go2short/internal/logger"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		reqID := c.GetHeader("X-Request-ID")

		c.Next()

		latency := float64(time.Since(start).Microseconds()) / 1000.0 // ms
		status := c.Writer.Status()

		// Skip health/metrics endpoints
		if path == "/health" || path == "/metrics" {
			return
		}

		logger.Info("request",
			logger.ReqID(reqID),
			logger.Status(status),
			logger.Latency(latency),
			logger.Extra("method", c.Request.Method),
			logger.Extra("path", path),
		)
	}
}
