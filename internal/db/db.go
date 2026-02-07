package db

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/model"
)

// DB wraps a pgx connection pool for account operations.
type DB struct {
	pool *pgxpool.Pool
}

// New connects to PostgreSQL and returns a DB handle.
func New(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return &DB{pool: pool}, nil
}

// Close closes the database connection pool.
func (d *DB) Close() {
	d.pool.Close()
}

// Pool returns the underlying pgx pool (for goose migrations).
func (d *DB) Pool() *pgxpool.Pool {
	return d.pool
}

// HashPassword hashes a password with SHA-1 and returns Base64 encoding.
// This matches the L2J algorithm: SHA1(password) -> Base64.
func HashPassword(password string) string {
	h := sha1.New()
	h.Write([]byte(password))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// GetAccount retrieves an account by login.
// Returns nil, nil if the account does not exist.
func (d *DB) GetAccount(ctx context.Context, login string) (*model.Account, error) {
	login = strings.ToLower(login)
	var acc model.Account
	err := d.pool.QueryRow(ctx,
		`SELECT login, password, access_level, last_server, last_ip, last_active
		 FROM accounts WHERE login = $1`, login,
	).Scan(&acc.Login, &acc.PasswordHash, &acc.AccessLevel, &acc.LastServer, &acc.LastIP, &acc.LastActive)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("querying account %q: %w", login, err)
	}
	return &acc, nil
}

// CreateAccount inserts a new account with the given password hash.
func (d *DB) CreateAccount(ctx context.Context, login, passwordHash, ip string) error {
	login = strings.ToLower(login)
	_, err := d.pool.Exec(ctx,
		`INSERT INTO accounts (login, password, last_active, access_level, last_ip)
		 VALUES ($1, $2, $3, 0, $4)`,
		login, passwordHash, time.Now(), ip,
	)
	if err != nil {
		return fmt.Errorf("creating account %q: %w", login, err)
	}
	slog.Info("auto-created account", "login", login)
	return nil
}

// UpdateLastLogin updates last_active and last_ip on successful login.
func (d *DB) UpdateLastLogin(ctx context.Context, login, ip string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE accounts SET last_active = $1, last_ip = $2 WHERE login = $3`,
		time.Now(), ip, strings.ToLower(login),
	)
	if err != nil {
		return fmt.Errorf("updating last login for %q: %w", login, err)
	}
	return nil
}

// UpdateLastServer updates the last_server field for the account.
func (d *DB) UpdateLastServer(ctx context.Context, login string, serverID int) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE accounts SET last_server = $1 WHERE login = $2`,
		serverID, strings.ToLower(login),
	)
	if err != nil {
		return fmt.Errorf("updating last server for %q: %w", login, err)
	}
	return nil
}
