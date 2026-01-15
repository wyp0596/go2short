package link

import (
	"context"
	"time"

	"github.com/wyp0596/go2short/internal/store"
)

// Storer defines the store operations needed by link service.
type Storer interface {
	GetLink(ctx context.Context, code string) (*store.Link, error)
	CreateLink(ctx context.Context, code, longURL string, expiresAt *time.Time, userID *int) error
}

// Cacher defines the cache operations needed by link service.
type Cacher interface {
	GetURL(ctx context.Context, code string) (string, error)
	SetURL(ctx context.Context, code, url string) error
}
