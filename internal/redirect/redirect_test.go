package redirect

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/wyp0596/go2short/internal/store"
)

// --- Mock implementations ---

type mockStore struct {
	links map[string]*store.Link
	err   error
}

func (m *mockStore) GetLink(_ context.Context, code string) (*store.Link, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.links[code], nil
}

type mockCache struct {
	urls    map[string]string
	misses  map[string]bool
	err     error
	setURLs []string // track SetURL calls
}

func (m *mockCache) GetURL(_ context.Context, code string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.urls[code], nil
}

func (m *mockCache) SetURL(_ context.Context, code, url string) error {
	if m.err != nil {
		return m.err
	}
	m.urls[code] = url
	m.setURLs = append(m.setURLs, code)
	return nil
}

func (m *mockCache) IsMiss(_ context.Context, code string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.misses[code], nil
}

func (m *mockCache) SetMiss(_ context.Context, code string) error {
	if m.err != nil {
		return m.err
	}
	m.misses[code] = true
	return nil
}

// --- Tests ---

func TestIsValidCode(t *testing.T) {
	s := &Service{codeLength: 8}

	tests := []struct {
		code string
		want bool
	}{
		// Valid codes
		{"abcdef", true},
		{"ABC123xyz", true},
		{"123456789012", true}, // max 12

		// Invalid: too short
		{"abc", false},
		{"abcde", false},

		// Invalid: too long
		{"abcdefghijklm", false},

		// Invalid: bad chars
		{"abc-def", false},
		{"abc_123", false},
		{"abc 123", false},
	}

	for _, tt := range tests {
		if got := s.isValidCode(tt.code); got != tt.want {
			t.Errorf("isValidCode(%q) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestResolve(t *testing.T) {
	ctx := context.Background()

	t.Run("cache hit", func(t *testing.T) {
		mc := &mockCache{
			urls:   map[string]string{"abc123": "https://example.com"},
			misses: make(map[string]bool),
		}
		ms := &mockStore{links: make(map[string]*store.Link)}
		svc := NewService(mc, ms, 8)

		result, err := svc.Resolve(ctx, "abc123")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result.StatusCode != 302 {
			t.Errorf("expected 302, got %d", result.StatusCode)
		}
		if !result.CacheHit {
			t.Error("expected CacheHit=true")
		}
		if result.URL != "https://example.com" {
			t.Errorf("expected https://example.com, got %q", result.URL)
		}
	})

	t.Run("negative cache hit", func(t *testing.T) {
		mc := &mockCache{
			urls:   make(map[string]string),
			misses: map[string]bool{"notfnd": true},
		}
		ms := &mockStore{links: make(map[string]*store.Link)}
		svc := NewService(mc, ms, 8)

		result, err := svc.Resolve(ctx, "notfnd")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result.StatusCode != 404 {
			t.Errorf("expected 404, got %d", result.StatusCode)
		}
		if !result.CacheHit {
			t.Error("expected CacheHit=true for negative cache")
		}
	})

	t.Run("db hit and cache backfill", func(t *testing.T) {
		mc := &mockCache{
			urls:   make(map[string]string),
			misses: make(map[string]bool),
		}
		ms := &mockStore{links: map[string]*store.Link{
			"dbcode": {Code: "dbcode", LongURL: "https://db.example.com"},
		}}
		svc := NewService(mc, ms, 8)

		result, err := svc.Resolve(ctx, "dbcode")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result.StatusCode != 302 {
			t.Errorf("expected 302, got %d", result.StatusCode)
		}
		if result.CacheHit {
			t.Error("expected CacheHit=false for DB lookup")
		}
		if result.URL != "https://db.example.com" {
			t.Errorf("expected https://db.example.com, got %q", result.URL)
		}
		// Verify cache was backfilled
		if mc.urls["dbcode"] != "https://db.example.com" {
			t.Error("cache should be backfilled")
		}
	})

	t.Run("not found sets negative cache", func(t *testing.T) {
		mc := &mockCache{
			urls:   make(map[string]string),
			misses: make(map[string]bool),
		}
		ms := &mockStore{links: make(map[string]*store.Link)}
		svc := NewService(mc, ms, 8)

		result, err := svc.Resolve(ctx, "nolink")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result.StatusCode != 404 {
			t.Errorf("expected 404, got %d", result.StatusCode)
		}
		// Verify negative cache was set
		if !mc.misses["nolink"] {
			t.Error("negative cache should be set")
		}
	})

	t.Run("disabled link returns 410", func(t *testing.T) {
		mc := &mockCache{
			urls:   make(map[string]string),
			misses: make(map[string]bool),
		}
		ms := &mockStore{links: map[string]*store.Link{
			"disabled": {Code: "disabled", LongURL: "https://example.com", IsDisabled: true},
		}}
		svc := NewService(mc, ms, 8)

		result, err := svc.Resolve(ctx, "disabled")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result.StatusCode != 410 {
			t.Errorf("expected 410, got %d", result.StatusCode)
		}
	})

	t.Run("expired link returns 410", func(t *testing.T) {
		mc := &mockCache{
			urls:   make(map[string]string),
			misses: make(map[string]bool),
		}
		pastTime := time.Now().Add(-24 * time.Hour)
		ms := &mockStore{links: map[string]*store.Link{
			"expired": {
				Code:      "expired",
				LongURL:   "https://example.com",
				ExpiresAt: sql.NullTime{Time: pastTime, Valid: true},
			},
		}}
		svc := NewService(mc, ms, 8)

		result, err := svc.Resolve(ctx, "expired")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result.StatusCode != 410 {
			t.Errorf("expected 410, got %d", result.StatusCode)
		}
	})

	t.Run("invalid code format returns 404", func(t *testing.T) {
		mc := &mockCache{
			urls:   make(map[string]string),
			misses: make(map[string]bool),
		}
		ms := &mockStore{links: make(map[string]*store.Link)}
		svc := NewService(mc, ms, 8)

		result, err := svc.Resolve(ctx, "bad-code")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result.StatusCode != 404 {
			t.Errorf("expected 404 for invalid code, got %d", result.StatusCode)
		}
	})

	t.Run("too short code returns 404", func(t *testing.T) {
		mc := &mockCache{
			urls:   make(map[string]string),
			misses: make(map[string]bool),
		}
		ms := &mockStore{links: make(map[string]*store.Link)}
		svc := NewService(mc, ms, 8)

		result, err := svc.Resolve(ctx, "abc")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if result.StatusCode != 404 {
			t.Errorf("expected 404, got %d", result.StatusCode)
		}
	})
}
