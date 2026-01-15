package redirect

import (
	"context"

	"github.com/wyp0596/go2short/internal/store"
)

// Storer defines the store operations needed by redirect service.
type Storer interface {
	GetLink(ctx context.Context, code string) (*store.Link, error)
}

// Cacher defines the cache operations needed by redirect service.
type Cacher interface {
	GetURL(ctx context.Context, code string) (string, error)
	SetURL(ctx context.Context, code, url string) error
	IsMiss(ctx context.Context, code string) (bool, error)
	SetMiss(ctx context.Context, code string) error
}
