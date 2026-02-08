package login

import (
	"context"

	"github.com/udisondev/la2go/internal/model"
)

// AccountRepository определяет интерфейс для работы с аккаунтами.
// Используется для dependency injection в тестах.
type AccountRepository interface {
	// GetAccount возвращает аккаунт по логину.
	// Возвращает nil, nil если аккаунт не найден.
	GetAccount(ctx context.Context, login string) (*model.Account, error)

	// CreateAccount создаёт новый аккаунт с указанным паролем и IP.
	CreateAccount(ctx context.Context, login, passwordHash, ip string) error

	// GetOrCreateAccount атомарно получает существующий или создаёт новый аккаунт.
	// Thread-safe: использует INSERT ... ON CONFLICT для защиты от race conditions.
	// Всегда возвращает аккаунт (существующий или только что созданный).
	GetOrCreateAccount(ctx context.Context, login, passwordHash, ip string) (*model.Account, error)

	// UpdateLastLogin обновляет last_active и last_ip при успешном логине.
	UpdateLastLogin(ctx context.Context, login, ip string) error
}
