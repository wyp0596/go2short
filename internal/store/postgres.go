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

// ListLinks returns paginated links with optional search.
func (s *Store) ListLinks(ctx context.Context, search string, limit, offset int) ([]Link, int, error) {
	var total int
	var rows *sql.Rows
	var err error

	if search != "" {
		pattern := "%" + search + "%"
		err = s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM links WHERE code ILIKE $1 OR long_url ILIKE $1`, pattern).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
		rows, err = s.db.QueryContext(ctx,
			`SELECT code, long_url, created_at, expires_at, is_disabled FROM links
			 WHERE code ILIKE $1 OR long_url ILIKE $1
			 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, pattern, limit, offset)
	} else {
		err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM links`).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
		rows, err = s.db.QueryContext(ctx,
			`SELECT code, long_url, created_at, expires_at, is_disabled FROM links
			 ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.Code, &l.LongURL, &l.CreatedAt, &l.ExpiresAt, &l.IsDisabled); err != nil {
			return nil, 0, err
		}
		links = append(links, l)
	}
	return links, total, rows.Err()
}

// UpdateLink updates a link's long_url and expires_at.
func (s *Store) UpdateLink(ctx context.Context, code, longURL string, expiresAt *time.Time) error {
	var exp sql.NullTime
	if expiresAt != nil {
		exp = sql.NullTime{Time: *expiresAt, Valid: true}
	}
	result, err := s.db.ExecContext(ctx,
		`UPDATE links SET long_url = $1, expires_at = $2 WHERE code = $3`,
		longURL, exp, code)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetLinkDisabled enables or disables a link.
func (s *Store) SetLinkDisabled(ctx context.Context, code string, disabled bool) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE links SET is_disabled = $1 WHERE code = $2`, disabled, code)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteLink removes a link.
func (s *Store) DeleteLink(ctx context.Context, code string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM links WHERE code = $1`, code)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// LinkClickStats holds click statistics for a link.
type LinkClickStats struct {
	TotalClicks int       `json:"total_clicks"`
	DailyClicks []DayClick `json:"daily_clicks"`
}

type DayClick struct {
	Date   string `json:"date"`
	Clicks int    `json:"clicks"`
}

// GetLinkStats returns click statistics for a link.
func (s *Store) GetLinkStats(ctx context.Context, code string, days int) (*LinkClickStats, error) {
	var total int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM click_events WHERE code = $1`, code).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT DATE(ts) as day, COUNT(*) as clicks FROM click_events
		 WHERE code = $1 AND ts >= NOW() - INTERVAL '1 day' * $2
		 GROUP BY DATE(ts) ORDER BY day DESC`, code, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var daily []DayClick
	for rows.Next() {
		var d DayClick
		var date time.Time
		if err := rows.Scan(&date, &d.Clicks); err != nil {
			return nil, err
		}
		d.Date = date.Format("2006-01-02")
		daily = append(daily, d)
	}
	return &LinkClickStats{TotalClicks: total, DailyClicks: daily}, rows.Err()
}

// OverviewStats holds overall statistics.
type OverviewStats struct {
	TotalLinks       int `json:"total_links"`
	ActiveLinks      int `json:"active_links"`
	TotalClicks      int `json:"total_clicks"`
	TodayClicks      int `json:"today_clicks"`
}

// GetOverviewStats returns overall statistics.
func (s *Store) GetOverviewStats(ctx context.Context) (*OverviewStats, error) {
	var stats OverviewStats
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM links`).Scan(&stats.TotalLinks)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM links WHERE is_disabled = false AND (expires_at IS NULL OR expires_at > NOW())`).Scan(&stats.ActiveLinks)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM click_events`).Scan(&stats.TotalClicks)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM click_events WHERE ts >= DATE_TRUNC('day', NOW())`).Scan(&stats.TodayClicks)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// TopLink holds a link with its click count.
type TopLink struct {
	Code       string `json:"code"`
	LongURL    string `json:"long_url"`
	ClickCount int    `json:"click_count"`
}

// GetTopLinks returns the top N links by click count in the last N days.
func (s *Store) GetTopLinks(ctx context.Context, limit, days int) ([]TopLink, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT l.code, l.long_url, COUNT(c.id) as clicks
		 FROM links l
		 LEFT JOIN click_events c ON l.code = c.code AND c.ts >= NOW() - INTERVAL '1 day' * $2
		 GROUP BY l.code, l.long_url
		 ORDER BY clicks DESC
		 LIMIT $1`, limit, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []TopLink
	for rows.Next() {
		var l TopLink
		if err := rows.Scan(&l.Code, &l.LongURL, &l.ClickCount); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

// GetClickTrend returns daily click counts for the last N days across all links.
func (s *Store) GetClickTrend(ctx context.Context, days int) ([]DayClick, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT DATE(ts) as day, COUNT(*) as clicks
		 FROM click_events
		 WHERE ts >= NOW() - INTERVAL '1 day' * $1
		 GROUP BY DATE(ts)
		 ORDER BY day ASC`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trend []DayClick
	for rows.Next() {
		var d DayClick
		var date time.Time
		if err := rows.Scan(&date, &d.Clicks); err != nil {
			return nil, err
		}
		d.Date = date.Format("2006-01-02")
		trend = append(trend, d)
	}
	return trend, rows.Err()
}
