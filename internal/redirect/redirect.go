package redirect

import (
	"context"
	"regexp"
	"time"

	"github.com/warren/go2short/internal/cache"
	"github.com/warren/go2short/internal/store"
)

var base62Regex = regexp.MustCompile(`^[0-9a-zA-Z]+$`)

type Result struct {
	URL        string
	StatusCode int // 302, 404, 410
	CacheHit   bool
}

type Service struct {
	cache      *cache.Cache
	store      *store.Store
	codeLength int
}

func NewService(c *cache.Cache, s *store.Store, codeLength int) *Service {
	return &Service{
		cache:      c,
		store:      s,
		codeLength: codeLength,
	}
}

// Resolve looks up a short code and returns the target URL.
// Flow: cache -> negative cache -> database -> backfill cache
func (s *Service) Resolve(ctx context.Context, code string) (*Result, error) {
	// 1. Validate code format
	if !s.isValidCode(code) {
		return &Result{StatusCode: 404}, nil
	}

	// 2. Check Redis cache
	url, err := s.cache.GetURL(ctx, code)
	if err != nil {
		return nil, err
	}
	if url != "" {
		return &Result{URL: url, StatusCode: 302, CacheHit: true}, nil
	}

	// 3. Check negative cache
	isMiss, err := s.cache.IsMiss(ctx, code)
	if err != nil {
		return nil, err
	}
	if isMiss {
		return &Result{StatusCode: 404, CacheHit: true}, nil
	}

	// 4. Query database
	link, err := s.store.GetLink(ctx, code)
	if err != nil {
		return nil, err
	}

	// 5. Not found -> set negative cache
	if link == nil {
		_ = s.cache.SetMiss(ctx, code) // ignore error, non-critical
		return &Result{StatusCode: 404}, nil
	}

	// 6. Check disabled/expired
	if link.IsDisabled {
		return &Result{StatusCode: 410}, nil
	}
	if link.ExpiresAt.Valid && link.ExpiresAt.Time.Before(time.Now()) {
		return &Result{StatusCode: 410}, nil
	}

	// 7. Backfill cache
	_ = s.cache.SetURL(ctx, code, link.LongURL) // ignore error, non-critical

	return &Result{URL: link.LongURL, StatusCode: 302}, nil
}

func (s *Service) isValidCode(code string) bool {
	if len(code) < 6 || len(code) > 12 {
		return false
	}
	return base62Regex.MatchString(code)
}
