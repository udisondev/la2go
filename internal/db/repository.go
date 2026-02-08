package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/model"
)

// PostgresAccountRepository реализует AccountRepository для PostgreSQL.
type PostgresAccountRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAccountRepository создаёт новый PostgreSQL repository.
func NewPostgresAccountRepository(pool *pgxpool.Pool) *PostgresAccountRepository {
	return &PostgresAccountRepository{pool: pool}
}

// GetAccount возвращает аккаунт по логину.
// Возвращает nil, nil если аккаунт не найден.
func (r *PostgresAccountRepository) GetAccount(ctx context.Context, login string) (*model.Account, error) {
	login = strings.ToLower(login)
	var acc model.Account
	err := r.pool.QueryRow(ctx,
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

// CreateAccount создаёт новый аккаунт с указанным паролем и IP.
func (r *PostgresAccountRepository) CreateAccount(ctx context.Context, login, passwordHash, ip string) error {
	login = strings.ToLower(login)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO accounts (login, password, last_active, access_level, last_ip)
		 VALUES ($1, $2, $3, 0, $4)`,
		login, passwordHash, time.Now(), ip,
	)
	if err != nil {
		return fmt.Errorf("creating account %q: %w", login, err)
	}
	return nil
}

// GetOrCreateAccount атомарно получает существующий или создаёт новый аккаунт.
// Thread-safe: использует INSERT ... ON CONFLICT DO NOTHING для защиты от race conditions.
// Всегда возвращает аккаунт (существующий или только что созданный).
func (r *PostgresAccountRepository) GetOrCreateAccount(ctx context.Context, login, passwordHash, ip string) (*model.Account, error) {
	login = strings.ToLower(login)

	// Попытка создать аккаунт (если уже существует — ON CONFLICT игнорирует)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO accounts (login, password, last_active, access_level, last_ip)
		 VALUES ($1, $2, $3, 0, $4)
		 ON CONFLICT (login) DO NOTHING`,
		login, passwordHash, time.Now(), ip,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting account %q: %w", login, err)
	}

	// Получаем аккаунт (гарантированно существует после INSERT ON CONFLICT)
	acc, err := r.GetAccount(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("getting account after insert %q: %w", login, err)
	}
	if acc == nil {
		return nil, fmt.Errorf("account %q not found after insert (unexpected)", login)
	}

	return acc, nil
}

// UpdateLastLogin обновляет last_active и last_ip при успешном логине.
func (r *PostgresAccountRepository) UpdateLastLogin(ctx context.Context, login, ip string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE accounts SET last_active = $1, last_ip = $2 WHERE login = $3`,
		time.Now(), ip, strings.ToLower(login),
	)
	if err != nil {
		return fmt.Errorf("updating last login for %q: %w", login, err)
	}
	return nil
}
