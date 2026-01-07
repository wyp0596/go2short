package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/warren/go2short/internal/cache"
	"github.com/warren/go2short/internal/config"
	"github.com/warren/go2short/internal/events"
	"github.com/warren/go2short/internal/handler"
	"github.com/warren/go2short/internal/link"
	"github.com/warren/go2short/internal/logger"
	_ "github.com/warren/go2short/internal/metrics" // register metrics
	"github.com/warren/go2short/internal/middleware"
	"github.com/warren/go2short/internal/redirect"
	"github.com/warren/go2short/internal/store"
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
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	api := r.Group("/api")
	api.POST("/links", linkHandler.Create)

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
