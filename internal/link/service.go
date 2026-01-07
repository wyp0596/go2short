package link

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/wyp0596/go2short/internal/cache"
	"github.com/wyp0596/go2short/internal/store"
)

const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	ErrInvalidURL  = errors.New("invalid URL")
	ErrURLTooLong  = errors.New("URL too long (max 2048)")
	ErrBlockedIP   = errors.New("URL points to private IP")
	ErrCodeTaken   = errors.New("custom code already taken")
	ErrInvalidCode = errors.New("invalid custom code")
	ErrMaxRetries  = errors.New("failed to generate unique code")
)

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

type CreateRequest struct {
	LongURL    string
	ExpiresAt  *time.Time
	CustomCode string
}

type CreateResult struct {
	Code      string
	CreatedAt time.Time
}

func (s *Service) Create(ctx context.Context, req *CreateRequest) (*CreateResult, error) {
	// Validate URL
	if err := s.validateURL(req.LongURL); err != nil {
		return nil, err
	}

	// Determine code
	code := req.CustomCode
	if code != "" {
		if !isValidCode(code) {
			return nil, ErrInvalidCode
		}
		// Check if custom code exists
		existing, err := s.store.GetLink(ctx, code)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrCodeTaken
		}
	} else {
		// Generate random code with retry
		var err error
		code, err = s.generateUniqueCode(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Create link
	if err := s.store.CreateLink(ctx, code, req.LongURL, req.ExpiresAt); err != nil {
		return nil, err
	}

	// Pre-warm cache
	_ = s.cache.SetURL(ctx, code, req.LongURL)

	return &CreateResult{
		Code:      code,
		CreatedAt: time.Now(),
	}, nil
}

func (s *Service) validateURL(rawURL string) error {
	if len(rawURL) > 2048 {
		return ErrURLTooLong
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return ErrInvalidURL
	}

	// Only http/https
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrInvalidURL
	}

	// Check for private IPs
	host := u.Hostname()
	if isPrivateHost(host) {
		return ErrBlockedIP
	}

	return nil
}

func (s *Service) generateUniqueCode(ctx context.Context) (string, error) {
	for i := 0; i < 3; i++ {
		code := generateCode(s.codeLength)
		existing, err := s.store.GetLink(ctx, code)
		if err != nil {
			return "", err
		}
		if existing == nil {
			return code, nil
		}
	}
	return "", ErrMaxRetries
}

func generateCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func isValidCode(code string) bool {
	if len(code) < 6 || len(code) > 12 {
		return false
	}
	for _, c := range code {
		if !strings.ContainsRune(charset, c) {
			return false
		}
	}
	return true
}

func isPrivateHost(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		// Try resolving hostname
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			return false // let it fail later if unresolvable
		}
		ip = ips[0]
	}

	// Check private ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}

	return false
}
