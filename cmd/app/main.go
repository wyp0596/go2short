package main

import (
	"context"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wyp0596/go2short/internal/cache"
	"github.com/wyp0596/go2short/internal/config"
	"github.com/wyp0596/go2short/internal/events"
	"github.com/wyp0596/go2short/internal/handler"
	"github.com/wyp0596/go2short/internal/link"
	"github.com/wyp0596/go2short/internal/logger"
	_ "github.com/wyp0596/go2short/internal/metrics" // register metrics
	"github.com/wyp0596/go2short/internal/middleware"
	"github.com/wyp0596/go2short/internal/redirect"
	"github.com/wyp0596/go2short/internal/store"
	"github.com/wyp0596/go2short/web"
)

func main() {
	cfg := config.Load()

	// Initialize cache
	c, err := cache.New(cfg)
	if err != nil {
		logger.Error("failed to connect to Redis", logger.Err(err))
		os.Exit(1)
	}
	defer c.Close()

	// Initialize store
	s, err := store.New(cfg)
	if err != nil {
		logger.Error("failed to connect to Postgres", logger.Err(err))
		os.Exit(1)
	}
	defer s.Close()

	// Initialize services
	redirectService := redirect.NewService(c, s, cfg.CodeLength)
	linkService := link.NewService(c, s, cfg.CodeLength)
	producer := events.NewProducer(c.Client(), cfg.StreamName)
	redirectHandler := handler.NewRedirectHandler(redirectService, producer)
	linkHandler := handler.NewLinkHandler(linkService, cfg.BaseURL)

	// Initialize admin
	authMiddleware := middleware.NewAuthMiddleware(c.Client(), cfg.RedisKeyPrefix)
	adminHandler := handler.NewAdminHandler(s, c, authMiddleware, cfg)

	// Start click event consumer
	consumer := events.NewConsumer(
		c.Client(),
		s,
		cfg.StreamName,
		cfg.StreamGroup,
		"worker-1",
		cfg.WorkerBatchSize,
		cfg.WorkerFlushInterval,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := consumer.Start(ctx); err != nil {
		logger.Error("failed to start consumer", logger.Err(err))
		os.Exit(1)
	}

	// Setup router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	// Routes
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin/")
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	api := r.Group("/api")
	api.POST("/links", linkHandler.Create)

	// Admin routes
	admin := api.Group("/admin")
	admin.POST("/login", adminHandler.Login)
	// Protected admin routes
	adminAuth := admin.Group("")
	adminAuth.Use(authMiddleware.RequireAuth())
	adminAuth.POST("/logout", adminHandler.Logout)
	adminAuth.GET("/links", adminHandler.ListLinks)
	adminAuth.PUT("/links/:code", adminHandler.UpdateLink)
	adminAuth.DELETE("/links/:code", adminHandler.DeleteLink)
	adminAuth.PATCH("/links/:code/disable", adminHandler.SetLinkDisabled)
	adminAuth.GET("/links/:code/stats", adminHandler.GetLinkStats)
	adminAuth.GET("/stats/overview", adminHandler.GetOverviewStats)
	adminAuth.GET("/stats/top-links", adminHandler.GetTopLinks)
	adminAuth.GET("/stats/trend", adminHandler.GetClickTrend)

	// Serve embedded frontend
	distFS, _ := fs.Sub(web.DistFS, "dist")
	r.GET("/admin", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin/")
	})
	r.GET("/admin/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		// Try to serve the file
		if filepath != "/" {
			if data, err := fs.ReadFile(distFS, filepath[1:]); err == nil {
				contentType := "application/octet-stream"
				if strings.HasSuffix(filepath, ".html") {
					contentType = "text/html; charset=utf-8"
				} else if strings.HasSuffix(filepath, ".js") {
					contentType = "application/javascript"
				} else if strings.HasSuffix(filepath, ".css") {
					contentType = "text/css"
				}
				c.Data(http.StatusOK, contentType, data)
				return
			}
		}
		// Fallback to index.html for SPA routing
		data, _ := fs.ReadFile(distFS, "index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// Redirect (must be last - catches all other paths)
	r.GET("/:code", redirectHandler.Handle)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down")
		cancel()
		consumer.Stop()
	}()

	logger.Info("server started", logger.Extra("addr", cfg.HTTPAddr))
	if err := r.Run(cfg.HTTPAddr); err != nil {
		logger.Error("server failed", logger.Err(err))
		os.Exit(1)
	}
}
