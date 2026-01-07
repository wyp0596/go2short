package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type ClickEvent struct {
	Code      string    `json:"code"`
	Timestamp time.Time `json:"ts"`
	IPHash    string    `json:"ip_hash"`
	UAHash    string    `json:"ua_hash"`
	Referer   string    `json:"referer"`
	ReqID     string    `json:"req_id"`
}

type Producer struct {
	client     *redis.Client
	streamName string
}

func NewProducer(client *redis.Client, streamName string) *Producer {
	return &Producer{
		client:     client,
		streamName: streamName,
	}
}

// Enqueue adds a click event to the stream. Non-blocking, fire-and-forget.
func (p *Producer) Enqueue(ctx context.Context, event *ClickEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: p.streamName,
		Values: map[string]interface{}{"data": string(data)},
	}).Err()
}

// EnqueueAsync fires event in a goroutine. Does not block caller.
func (p *Producer) EnqueueAsync(event *ClickEvent) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		_ = p.Enqueue(ctx, event) // ignore error, non-critical path
	}()
}
