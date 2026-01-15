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
	UserID     sql.NullInt32
}

type User struct {
	ID           int
	Email        string
	PasswordHash sql.NullString
	Provider     string
	ProviderID   sql.NullString
	CreatedAt    time.Time
	LastLoginAt  sql.NullTime
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
// userID can be nil for system/admin created links.
func (s *Store) CreateLink(ctx context.Context, code, longURL string, expiresAt *time.Time, userID *int) error {
	var exp sql.NullTime
	if expiresAt != nil {
		exp = sql.NullTime{Time: *expiresAt, Valid: true}
	}
	var uid sql.NullInt32
	if userID != nil {
		uid = sql.NullInt32{Int32: int32(*userID), Valid: true}
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO links (code, long_url, expires_at, user_id) VALUES ($1, $2, $3, $4)`,
		code, longURL, exp, uid,
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
		`INSERT INTO click_events (code, ts, ip, ua, device_type, browser, os, referer) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		if _, err := stmt.ExecContext(ctx, e.Code, e.Timestamp, e.IP, e.UA, e.DeviceType, e.Browser, e.OS, e.Referer); err != nil {
			return err
		}
	}

	return tx.Commit()
}

type ClickEvent struct {
	Code       string
	Timestamp  time.Time
	IP         string
	UA         string
	DeviceType string
	Browser    string
	OS         string
	Referer    string
}

func (s *Store) Close() error {
	return s.db.Close()
}

// CreateUser creates a new user and returns the ID.
func (s *Store) CreateUser(ctx context.Context, email, passwordHash, provider, providerID string) (int, error) {
	var pwHash sql.NullString
	if passwordHash != "" {
		pwHash = sql.NullString{String: passwordHash, Valid: true}
	}
	var provID sql.NullString
	if providerID != "" {
		provID = sql.NullString{String: providerID, Valid: true}
	}

	var id int
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, provider, provider_id) VALUES ($1, $2, $3, $4) RETURNING id`,
		email, pwHash, provider, provID,
	).Scan(&id)
	return id, err
}

// GetUserByEmail returns user by email.
func (s *Store) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, provider, provider_id, created_at, last_login_at FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.LastLoginAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByProvider returns user by OAuth provider and provider ID.
func (s *Store) GetUserByProvider(ctx context.Context, provider, providerID string) (*User, error) {
	var u User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, provider, provider_id, created_at, last_login_at FROM users WHERE provider = $1 AND provider_id = $2`,
		provider, providerID,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.LastLoginAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByID returns user by ID.
func (s *Store) GetUserByID(ctx context.Context, id int) (*User, error) {
	var u User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, provider, provider_id, created_at, last_login_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.LastLoginAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateUserLastLogin updates the last_login_at timestamp.
func (s *Store) UpdateUserLastLogin(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET last_login_at = NOW() WHERE id = $1`, id)
	return err
}

// ListLinks returns paginated links with optional search.
// userID nil means all links (admin), otherwise filter by user.
func (s *Store) ListLinks(ctx context.Context, search string, limit, offset int, userID *int) ([]Link, int, error) {
	var total int
	var rows *sql.Rows
	var err error

	if userID == nil {
		// Admin: all links
		if search != "" {
			pattern := "%" + search + "%"
			err = s.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM links WHERE code ILIKE $1 OR long_url ILIKE $1`, pattern).Scan(&total)
			if err != nil {
				return nil, 0, err
			}
			rows, err = s.db.QueryContext(ctx,
				`SELECT code, long_url, created_at, expires_at, is_disabled, user_id FROM links
				 WHERE code ILIKE $1 OR long_url ILIKE $1
				 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, pattern, limit, offset)
		} else {
			err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM links`).Scan(&total)
			if err != nil {
				return nil, 0, err
			}
			rows, err = s.db.QueryContext(ctx,
				`SELECT code, long_url, created_at, expires_at, is_disabled, user_id FROM links
				 ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
		}
	} else {
		// Regular user: only their links
		if search != "" {
			pattern := "%" + search + "%"
			err = s.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM links WHERE user_id = $1 AND (code ILIKE $2 OR long_url ILIKE $2)`, *userID, pattern).Scan(&total)
			if err != nil {
				return nil, 0, err
			}
			rows, err = s.db.QueryContext(ctx,
				`SELECT code, long_url, created_at, expires_at, is_disabled, user_id FROM links
				 WHERE user_id = $1 AND (code ILIKE $2 OR long_url ILIKE $2)
				 ORDER BY created_at DESC LIMIT $3 OFFSET $4`, *userID, pattern, limit, offset)
		} else {
			err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM links WHERE user_id = $1`, *userID).Scan(&total)
			if err != nil {
				return nil, 0, err
			}
			rows, err = s.db.QueryContext(ctx,
				`SELECT code, long_url, created_at, expires_at, is_disabled, user_id FROM links
				 WHERE user_id = $1
				 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, *userID, limit, offset)
		}
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.Code, &l.LongURL, &l.CreatedAt, &l.ExpiresAt, &l.IsDisabled, &l.UserID); err != nil {
			return nil, 0, err
		}
		links = append(links, l)
	}
	return links, total, rows.Err()
}

// UpdateLink updates a link's long_url and expires_at.
// userID nil means admin (can update any), otherwise only user's own links.
func (s *Store) UpdateLink(ctx context.Context, code, longURL string, expiresAt *time.Time, userID *int) error {
	var exp sql.NullTime
	if expiresAt != nil {
		exp = sql.NullTime{Time: *expiresAt, Valid: true}
	}
	var result sql.Result
	var err error
	if userID == nil {
		result, err = s.db.ExecContext(ctx,
			`UPDATE links SET long_url = $1, expires_at = $2 WHERE code = $3`,
			longURL, exp, code)
	} else {
		result, err = s.db.ExecContext(ctx,
			`UPDATE links SET long_url = $1, expires_at = $2 WHERE code = $3 AND user_id = $4`,
			longURL, exp, code, *userID)
	}
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
// userID nil means admin, otherwise only user's own links.
func (s *Store) SetLinkDisabled(ctx context.Context, code string, disabled bool, userID *int) error {
	var result sql.Result
	var err error
	if userID == nil {
		result, err = s.db.ExecContext(ctx,
			`UPDATE links SET is_disabled = $1 WHERE code = $2`, disabled, code)
	} else {
		result, err = s.db.ExecContext(ctx,
			`UPDATE links SET is_disabled = $1 WHERE code = $2 AND user_id = $3`, disabled, code, *userID)
	}
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
// userID nil means admin, otherwise only user's own links.
func (s *Store) DeleteLink(ctx context.Context, code string, userID *int) error {
	var result sql.Result
	var err error
	if userID == nil {
		result, err = s.db.ExecContext(ctx, `DELETE FROM links WHERE code = $1`, code)
	} else {
		result, err = s.db.ExecContext(ctx, `DELETE FROM links WHERE code = $1 AND user_id = $2`, code, *userID)
	}
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
// userID nil means admin, otherwise verifies link belongs to user.
func (s *Store) GetLinkStats(ctx context.Context, code string, days int, userID *int) (*LinkClickStats, error) {
	// Verify link ownership if not admin
	if userID != nil {
		var count int
		err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM links WHERE code = $1 AND user_id = $2`, code, *userID).Scan(&count)
		if err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, sql.ErrNoRows
		}
	}

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
// userID nil means admin (all), otherwise filter by user.
func (s *Store) GetOverviewStats(ctx context.Context, userID *int) (*OverviewStats, error) {
	var stats OverviewStats
	if userID == nil {
		// Admin: global stats
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM links`).Scan(&stats.TotalLinks); err != nil {
			return nil, err
		}
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM links WHERE is_disabled = false AND (expires_at IS NULL OR expires_at > NOW())`).Scan(&stats.ActiveLinks); err != nil {
			return nil, err
		}
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM click_events`).Scan(&stats.TotalClicks); err != nil {
			return nil, err
		}
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM click_events WHERE ts >= DATE_TRUNC('day', NOW())`).Scan(&stats.TodayClicks); err != nil {
			return nil, err
		}
	} else {
		// User: user's stats
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM links WHERE user_id = $1`, *userID).Scan(&stats.TotalLinks); err != nil {
			return nil, err
		}
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM links WHERE user_id = $1 AND is_disabled = false AND (expires_at IS NULL OR expires_at > NOW())`, *userID).Scan(&stats.ActiveLinks); err != nil {
			return nil, err
		}
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM click_events WHERE code IN (SELECT code FROM links WHERE user_id = $1)`, *userID).Scan(&stats.TotalClicks); err != nil {
			return nil, err
		}
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM click_events WHERE code IN (SELECT code FROM links WHERE user_id = $1) AND ts >= DATE_TRUNC('day', NOW())`, *userID).Scan(&stats.TodayClicks); err != nil {
			return nil, err
		}
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
// userID nil means admin (all), otherwise filter by user.
func (s *Store) GetTopLinks(ctx context.Context, limit, days int, userID *int) ([]TopLink, error) {
	var rows *sql.Rows
	var err error
	if userID == nil {
		rows, err = s.db.QueryContext(ctx,
			`SELECT l.code, l.long_url, COUNT(c.id) as clicks
			 FROM links l
			 LEFT JOIN click_events c ON l.code = c.code AND c.ts >= NOW() - INTERVAL '1 day' * $2
			 GROUP BY l.code, l.long_url
			 ORDER BY clicks DESC
			 LIMIT $1`, limit, days)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT l.code, l.long_url, COUNT(c.id) as clicks
			 FROM links l
			 LEFT JOIN click_events c ON l.code = c.code AND c.ts >= NOW() - INTERVAL '1 day' * $2
			 WHERE l.user_id = $3
			 GROUP BY l.code, l.long_url
			 ORDER BY clicks DESC
			 LIMIT $1`, limit, days, *userID)
	}
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
// userID nil means admin (all), otherwise filter by user.
func (s *Store) GetClickTrend(ctx context.Context, days int, userID *int) ([]DayClick, error) {
	var rows *sql.Rows
	var err error
	if userID == nil {
		rows, err = s.db.QueryContext(ctx,
			`SELECT DATE(ts) as day, COUNT(*) as clicks
			 FROM click_events
			 WHERE ts >= NOW() - INTERVAL '1 day' * $1
			 GROUP BY DATE(ts)
			 ORDER BY day ASC`, days)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT DATE(ts) as day, COUNT(*) as clicks
			 FROM click_events
			 WHERE ts >= NOW() - INTERVAL '1 day' * $1 AND code IN (SELECT code FROM links WHERE user_id = $2)
			 GROUP BY DATE(ts)
			 ORDER BY day ASC`, days, *userID)
	}
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

// APIToken represents an API token for external access.
type APIToken struct {
	ID         int
	TokenHash  string
	Name       string
	CreatedAt  time.Time
	LastUsedAt sql.NullTime
	Disabled   bool
	UserID     sql.NullInt32
}

// CreateAPIToken inserts a new API token. Returns the ID.
// userID can be nil for admin-created global tokens.
func (s *Store) CreateAPIToken(ctx context.Context, tokenHash, name string, userID *int) (int, error) {
	var uid sql.NullInt32
	if userID != nil {
		uid = sql.NullInt32{Int32: int32(*userID), Valid: true}
	}
	var id int
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO api_tokens (token_hash, name, user_id) VALUES ($1, $2, $3) RETURNING id`,
		tokenHash, name, uid,
	).Scan(&id)
	return id, err
}

// GetAPITokenByHash returns token by hash. Returns nil if not found or disabled.
func (s *Store) GetAPITokenByHash(ctx context.Context, tokenHash string) (*APIToken, error) {
	var t APIToken
	err := s.db.QueryRowContext(ctx,
		`SELECT id, token_hash, name, created_at, last_used_at, disabled, user_id
		 FROM api_tokens WHERE token_hash = $1 AND disabled = false`,
		tokenHash,
	).Scan(&t.ID, &t.TokenHash, &t.Name, &t.CreatedAt, &t.LastUsedAt, &t.Disabled, &t.UserID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateAPITokenLastUsed updates last_used_at timestamp.
func (s *Store) UpdateAPITokenLastUsed(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE api_tokens SET last_used_at = NOW() WHERE id = $1`, id)
	return err
}

// ListAPITokens returns API tokens.
// userID nil means admin (all), otherwise filter by user.
func (s *Store) ListAPITokens(ctx context.Context, userID *int) ([]APIToken, error) {
	var rows *sql.Rows
	var err error
	if userID == nil {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, token_hash, name, created_at, last_used_at, disabled, user_id
			 FROM api_tokens ORDER BY created_at DESC`)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, token_hash, name, created_at, last_used_at, disabled, user_id
			 FROM api_tokens WHERE user_id = $1 ORDER BY created_at DESC`, *userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []APIToken
	for rows.Next() {
		var t APIToken
		if err := rows.Scan(&t.ID, &t.TokenHash, &t.Name, &t.CreatedAt, &t.LastUsedAt, &t.Disabled, &t.UserID); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

// DeleteAPIToken removes an API token.
// userID nil means admin (can delete any), otherwise only user's own tokens.
func (s *Store) DeleteAPIToken(ctx context.Context, id int, userID *int) error {
	var result sql.Result
	var err error
	if userID == nil {
		result, err = s.db.ExecContext(ctx, `DELETE FROM api_tokens WHERE id = $1`, id)
	} else {
		result, err = s.db.ExecContext(ctx, `DELETE FROM api_tokens WHERE id = $1 AND user_id = $2`, id, *userID)
	}
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
