package auth

import (
	"testing"

	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

func TestGoogleConfig(t *testing.T) {
	cfg := GoogleConfig("client-id", "client-secret", "https://example.com/callback")

	if cfg.ClientID != "client-id" {
		t.Errorf("ClientID = %q, want %q", cfg.ClientID, "client-id")
	}
	if cfg.ClientSecret != "client-secret" {
		t.Errorf("ClientSecret = %q, want %q", cfg.ClientSecret, "client-secret")
	}
	if cfg.RedirectURL != "https://example.com/callback" {
		t.Errorf("RedirectURL = %q, want %q", cfg.RedirectURL, "https://example.com/callback")
	}
	if cfg.Endpoint != google.Endpoint {
		t.Error("Endpoint should be google.Endpoint")
	}
	// Check scopes
	if len(cfg.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(cfg.Scopes))
	}
}

func TestGitHubConfig(t *testing.T) {
	cfg := GitHubConfig("gh-client", "gh-secret", "https://example.com/gh/callback")

	if cfg.ClientID != "gh-client" {
		t.Errorf("ClientID = %q, want %q", cfg.ClientID, "gh-client")
	}
	if cfg.ClientSecret != "gh-secret" {
		t.Errorf("ClientSecret = %q, want %q", cfg.ClientSecret, "gh-secret")
	}
	if cfg.RedirectURL != "https://example.com/gh/callback" {
		t.Errorf("RedirectURL = %q, want %q", cfg.RedirectURL, "https://example.com/gh/callback")
	}
	if cfg.Endpoint != github.Endpoint {
		t.Error("Endpoint should be github.Endpoint")
	}
	// Check scopes
	if len(cfg.Scopes) != 1 || cfg.Scopes[0] != "user:email" {
		t.Errorf("expected scope [user:email], got %v", cfg.Scopes)
	}
}
