package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyp0596/go2short/internal/logger"
	"github.com/wyp0596/go2short/internal/store"
)

type Consumer struct {
	client        *redis.Client
	store         *store.Store
	streamName    string
	groupName     string
	consumerName  string
	batchSize     int
	flushInterval time.Duration
	buffer        []store.ClickEvent
	stopCh        chan struct{}
}

func NewConsumer(
	client *redis.Client,
	s *store.Store,
	streamName, groupName, consumerName string,
	batchSize int,
	flushInterval time.Duration,
) *Consumer {
	return &Consumer{
		client:        client,
		store:         s,
		streamName:    streamName,
		groupName:     groupName,
		consumerName:  consumerName,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		buffer:        make([]store.ClickEvent, 0, batchSize),
		stopCh:        make(chan struct{}),
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	// Create consumer group if not exists
	err := c.client.XGroupCreateMkStream(ctx, c.streamName, c.groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}

	go c.run(ctx)
	return nil
}

func (c *Consumer) Stop() {
	close(c.stopCh)
}

func (c *Consumer) run(ctx context.Context) {
	ticker := time.NewTicker(c.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.flush(context.Background()) // flush remaining on shutdown
			return
		case <-c.stopCh:
			c.flush(context.Background())
			return
		case <-ticker.C:
			c.flush(ctx)
		default:
			c.consume(ctx)
		}
	}
}

func (c *Consumer) consume(ctx context.Context) {
	streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    c.groupName,
		Consumer: c.consumerName,
		Streams:  []string{c.streamName, ">"},
		Count:    int64(c.batchSize - len(c.buffer)),
		Block:    100 * time.Millisecond,
	}).Result()

	if err == redis.Nil || len(streams) == 0 {
		return
	}
	if err != nil {
		logger.Error("XReadGroup error", logger.Err(err))
		return
	}

	for _, stream := range streams {
		for _, msg := range stream.Messages {
			data, ok := msg.Values["data"].(string)
			if !ok {
				continue
			}

			var event ClickEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				logger.Error("failed to unmarshal event", logger.Err(err))
				continue
			}

			c.buffer = append(c.buffer, store.ClickEvent{
				Code:      event.Code,
				Timestamp: event.Timestamp,
				IPHash:    event.IPHash,
				UAHash:    event.UAHash,
				Referer:   event.Referer,
			})

			// ACK message
			c.client.XAck(ctx, c.streamName, c.groupName, msg.ID)
		}
	}

	if len(c.buffer) >= c.batchSize {
		c.flush(ctx)
	}
}

func (c *Consumer) flush(ctx context.Context) {
	if len(c.buffer) == 0 {
		return
	}

	events := make([]store.ClickEvent, len(c.buffer))
	copy(events, c.buffer)
	c.buffer = c.buffer[:0]

	if err := c.store.InsertClickEvents(ctx, events); err != nil {
		logger.Error("failed to insert click events", logger.Err(err), logger.Extra("count", len(events)))
	}
}
