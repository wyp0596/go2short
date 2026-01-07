package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/wyp0596/go2short/internal/config"
)

type Link struct {
	Code       string
	LongURL    string
	CreatedAt  time.Time
	ExpiresAt  sql.NullTime
	IsDisabled bool
}

type Store struct {
	db *sql.DB
}

func New(cfg *config.Config) (*Store, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &Store{db: db}, nil
}

// GetLink fetches a link by code. Returns nil if not found.
func (s *Store) GetLink(ctx context.Context, code string) (*Link, error) {
	var link Link
	err := s.db.QueryRowContext(ctx,
		`SELECT code, long_url, created_at, expires_at, is_disabled
		 FROM links WHERE code = $1`,
		code,
	).Scan(&link.Code, &link.LongURL, &link.CreatedAt, &link.ExpiresAt, &link.IsDisabled)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// CreateLink inserts a new link. Returns error if code already exists.
func (s *Store) CreateLink(ctx context.Context, code, longURL string, expiresAt *time.Time) error {
	var exp sql.NullTime
	if expiresAt != nil {
		exp = sql.NullTime{Time: *expiresAt, Valid: true}
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO links (code, long_url, expires_at) VALUES ($1, $2, $3)`,
		code, longURL, exp,
	)
	return err
}

// InsertClickEvents bulk inserts click events.
func (s *Store) InsertClickEvents(ctx context.Context, events []ClickEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO click_events (code, ts, ip_hash, ua_hash, referer) VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		if _, err := stmt.ExecContext(ctx, e.Code, e.Timestamp, e.IPHash, e.UAHash, e.Referer); err != nil {
			return err
		}
	}

	return tx.Commit()
}

type ClickEvent struct {
	Code      string
	Timestamp time.Time
	IPHash    string
	UAHash    string
	Referer   string
}

func (s *Store) Close() error {
	return s.db.Close()
}
