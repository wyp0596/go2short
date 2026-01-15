package link

import (
	"context"
	"strings"
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

func (m *mockStore) CreateLink(_ context.Context, code, longURL string, expiresAt *time.Time, _ *int) error {
	if m.err != nil {
		return m.err
	}
	m.links[code] = &store.Link{Code: code, LongURL: longURL}
	return nil
}

type mockCache struct {
	urls map[string]string
	err  error
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
	return nil
}

// --- Tests ---

func TestIsValidCode(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		// Valid codes
		{"abcdef", true},       // min length 6
		{"AbCdEf123456", true}, // max length 12
		{"abc123XYZ", true},

		// Invalid: too short
		{"abc", false},
		{"abcde", false},

		// Invalid: too long
		{"abcdefghijklm", false},

		// Invalid: bad chars
		{"abc-def", false},
		{"abc_def", false},
		{"abc def", false},
		{"abc!@#", false},
	}

	for _, tt := range tests {
		if got := isValidCode(tt.code); got != tt.want {
			t.Errorf("isValidCode(%q) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestGenerateCode(t *testing.T) {
	lengths := []int{6, 8, 12}
	for _, length := range lengths {
		code := generateCode(length)
		if len(code) != length {
			t.Errorf("generateCode(%d) returned len %d", length, len(code))
		}
		if !isValidCode(code) {
			t.Errorf("generateCode(%d) returned invalid code: %s", length, code)
		}
	}
}

func TestIsPrivateHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		// Private IPv4
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.0.1", true},
		{"192.168.255.255", true},
		{"169.254.1.1", true},

		// Public IPv4
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"203.0.113.1", false},

		// Private IPv6
		{"::1", true},
		{"fc00::1", true},
		{"fe80::1", true},

		// Public IPv6
		{"2001:4860:4860::8888", false},
	}

	for _, tt := range tests {
		if got := isPrivateHost(tt.host); got != tt.want {
			t.Errorf("isPrivateHost(%q) = %v, want %v", tt.host, got, tt.want)
		}
	}
}

func TestValidateURL(t *testing.T) {
	s := &Service{}

	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		{"valid http", "http://example.com", nil},
		{"valid https", "https://example.com/path?q=1", nil},
		{"invalid scheme ftp", "ftp://example.com", ErrInvalidURL},
		{"invalid scheme empty", "example.com", ErrInvalidURL},
		{"invalid scheme javascript", "javascript:alert(1)", ErrInvalidURL},
		{"too long", "https://example.com/" + strings.Repeat("a", 2048), ErrURLTooLong},
		{"private ip 127", "http://127.0.0.1", ErrBlockedIP},
		{"private ip 10", "http://10.0.0.1/path", ErrBlockedIP},
		{"private ip 192", "http://192.168.1.1", ErrBlockedIP},
		{"private localhost", "http://localhost", ErrBlockedIP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.validateURL(tt.url)
			if err != tt.wantErr {
				t.Errorf("validateURL(%q) = %v, want %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("success with random code", func(t *testing.T) {
		ms := &mockStore{links: make(map[string]*store.Link)}
		mc := &mockCache{urls: make(map[string]string)}
		svc := NewService(mc, ms, 8)

		result, err := svc.Create(ctx, &CreateRequest{LongURL: "https://example.com"})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if len(result.Code) != 8 {
			t.Errorf("expected code length 8, got %d", len(result.Code))
		}
		if ms.links[result.Code] == nil {
			t.Error("link not stored")
		}
	})

	t.Run("success with custom code", func(t *testing.T) {
		ms := &mockStore{links: make(map[string]*store.Link)}
		mc := &mockCache{urls: make(map[string]string)}
		svc := NewService(mc, ms, 8)

		result, err := svc.Create(ctx, &CreateRequest{
			LongURL:    "https://example.com",
			CustomCode: "mycode1",
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if result.Code != "mycode1" {
			t.Errorf("expected code 'mycode1', got %q", result.Code)
		}
	})

	t.Run("custom code already taken", func(t *testing.T) {
		ms := &mockStore{links: map[string]*store.Link{"taken1": {}}}
		mc := &mockCache{urls: make(map[string]string)}
		svc := NewService(mc, ms, 8)

		_, err := svc.Create(ctx, &CreateRequest{
			LongURL:    "https://example.com",
			CustomCode: "taken1",
		})
		if err != ErrCodeTaken {
			t.Errorf("expected ErrCodeTaken, got %v", err)
		}
	})

	t.Run("invalid custom code", func(t *testing.T) {
		ms := &mockStore{links: make(map[string]*store.Link)}
		mc := &mockCache{urls: make(map[string]string)}
		svc := NewService(mc, ms, 8)

		_, err := svc.Create(ctx, &CreateRequest{
			LongURL:    "https://example.com",
			CustomCode: "bad-code",
		})
		if err != ErrInvalidCode {
			t.Errorf("expected ErrInvalidCode, got %v", err)
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		ms := &mockStore{links: make(map[string]*store.Link)}
		mc := &mockCache{urls: make(map[string]string)}
		svc := NewService(mc, ms, 8)

		_, err := svc.Create(ctx, &CreateRequest{LongURL: "ftp://example.com"})
		if err != ErrInvalidURL {
			t.Errorf("expected ErrInvalidURL, got %v", err)
		}
	})
}

func TestBatchCreate(t *testing.T) {
	ctx := context.Background()
	ms := &mockStore{links: make(map[string]*store.Link)}
	mc := &mockCache{urls: make(map[string]string)}
	svc := NewService(mc, ms, 8)

	requests := []BatchCreateRequest{
		{LongURL: "https://example1.com"},
		{LongURL: "https://example2.com"},
		{LongURL: "ftp://invalid.com"}, // invalid
	}

	results := svc.BatchCreate(ctx, requests)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// First two should succeed
	if results[0].Error != nil {
		t.Errorf("results[0] should succeed, got %v", results[0].Error)
	}
	if results[1].Error != nil {
		t.Errorf("results[1] should succeed, got %v", results[1].Error)
	}
	// Third should fail
	if results[2].Error != ErrInvalidURL {
		t.Errorf("results[2] should be ErrInvalidURL, got %v", results[2].Error)
	}
}

func TestGetLongURL(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		ms := &mockStore{links: map[string]*store.Link{
			"abc123": {Code: "abc123", LongURL: "https://example.com"},
		}}
		mc := &mockCache{urls: make(map[string]string)}
		svc := NewService(mc, ms, 8)

		url, err := svc.GetLongURL(ctx, "abc123")
		if err != nil {
			t.Fatalf("GetLongURL failed: %v", err)
		}
		if url != "https://example.com" {
			t.Errorf("expected https://example.com, got %q", url)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ms := &mockStore{links: make(map[string]*store.Link)}
		mc := &mockCache{urls: make(map[string]string)}
		svc := NewService(mc, ms, 8)

		url, err := svc.GetLongURL(ctx, "notexist")
		if err != nil {
			t.Fatalf("GetLongURL failed: %v", err)
		}
		if url != "" {
			t.Errorf("expected empty string, got %q", url)
		}
	})
}
