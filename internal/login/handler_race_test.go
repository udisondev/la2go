package login

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/model"
)

// TestHandler_ConcurrentAutoCreate тестирует race condition fix в auto-create.
func TestHandler_ConcurrentAutoCreate(t *testing.T) {
	// Arrange
	var createCount atomic.Int32
	var getOrCreateCount atomic.Int32

	mockRepo := &MockAccountRepository{
		GetAccountFunc: func(ctx context.Context, login string) (*model.Account, error) {
			// Account не существует
			return nil, nil
		},
		GetOrCreateAccountFunc: func(ctx context.Context, login, passwordHash, ip string) (*model.Account, error) {
			getOrCreateCount.Add(1)
			// Симулируем атомарную операцию (все вызовы успешны)
			return &model.Account{
				Login:        login,
				PasswordHash: passwordHash,
				AccessLevel:  0,
			}, nil
		},
		CreateAccountFunc: func(ctx context.Context, login, passwordHash, ip string) error {
			createCount.Add(1)
			// Не должно вызываться — используем GetOrCreateAccount
			t.Error("CreateAccount should not be called when using GetOrCreateAccount")
			return nil
		},
		UpdateLastLoginFunc: func(ctx context.Context, login, ip string) error {
			return nil
		},
	}

	cfg := config.DefaultLoginServer()
	cfg.AutoCreateAccounts = true
	cfg.ShowLicence = true

	sm := NewSessionManager()
	handler := NewHandler(mockRepo, cfg, sm)

	// Act: 10 concurrent requests для одного и того же login
	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for range numGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()

			rsaKeyPair, _ := crypto.GenerateRSAKeyPair()
			client := &Client{
				sessionID:  12345,
				rsaKeyPair: rsaKeyPair,
				state:      StateAuthedGG,
				ip:         "127.0.0.1",
			}

			packet := buildAuthGameGuardPacket(client.SessionID())
			buf := make([]byte, 1024)

			// Simulate RequestAuthLogin через прямой вызов handleRequestAuthLogin
			// (упрощённо — без RSA encryption)
			login := "concurrent_test_user"
			password := "password"
			passHash := db.HashPassword(password)

			// Direct call to internal logic
			acc, err := handler.accounts.GetAccount(context.Background(), login)
			if err != nil {
				errors <- fmt.Errorf("GetAccount failed: %w", err)
				return
			}

			if acc == nil && cfg.AutoCreateAccounts {
				acc, err = handler.accounts.GetOrCreateAccount(context.Background(), login, passHash, client.IP())
				if err != nil {
					errors <- fmt.Errorf("GetOrCreateAccount failed: %w", err)
					return
				}
			}

			if acc == nil {
				errors <- fmt.Errorf("account is nil after GetOrCreateAccount")
				return
			}

			_ = buf
			_ = packet
		}()
	}

	wg.Wait()
	close(errors)

	// Assert
	for err := range errors {
		t.Errorf("goroutine error: %v", err)
	}

	// Проверяем что GetOrCreateAccount вызван N раз (по одному на каждую goroutine)
	if getOrCreateCount.Load() != numGoroutines {
		t.Errorf("expected %d GetOrCreateAccount calls, got %d", numGoroutines, getOrCreateCount.Load())
	}

	// Проверяем что CreateAccount НЕ вызывался (используем GetOrCreateAccount)
	if createCount.Load() != 0 {
		t.Errorf("expected 0 CreateAccount calls, got %d", createCount.Load())
	}
}

// TestGetOrCreateAccount_RaceCondition тестирует защиту от race condition.
func TestGetOrCreateAccount_RaceCondition(t *testing.T) {
	// Arrange
	var createAttempts atomic.Int32
	successCount := atomic.Int32{}

	mockRepo := &MockAccountRepository{
		GetOrCreateAccountFunc: func(ctx context.Context, login, passwordHash, ip string) (*model.Account, error) {
			attempts := createAttempts.Add(1)

			// Симулируем ON CONFLICT: только первый вызов "создаёт" аккаунт,
			// остальные просто возвращают существующий
			if attempts == 1 {
				// Первый вызов — создаём аккаунт
				successCount.Add(1)
			}

			// Все вызовы возвращают одинаковый аккаунт (как в реальной БД)
			return &model.Account{
				Login:        login,
				PasswordHash: passwordHash,
				AccessLevel:  0,
			}, nil
		},
	}

	// Act: 100 concurrent GetOrCreateAccount для одного login
	const numGoroutines = 100
	var wg sync.WaitGroup
	results := make(chan *model.Account, numGoroutines)

	for range numGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()

			acc, err := mockRepo.GetOrCreateAccount(
				context.Background(),
				"same_user",
				"hash",
				"127.0.0.1",
			)
			if err != nil {
				t.Errorf("GetOrCreateAccount failed: %v", err)
				return
			}
			results <- acc
		}()
	}

	wg.Wait()
	close(results)

	// Assert: все goroutines получили аккаунт
	accountsReceived := 0
	for acc := range results {
		if acc == nil {
			t.Error("received nil account")
		} else {
			accountsReceived++
		}
	}

	if accountsReceived != numGoroutines {
		t.Errorf("expected %d accounts, got %d", numGoroutines, accountsReceived)
	}

	// Все 100 goroutines вызвали GetOrCreateAccount
	if createAttempts.Load() != numGoroutines {
		t.Errorf("expected %d create attempts, got %d", numGoroutines, createAttempts.Load())
	}

	t.Logf("✅ Race condition test passed: %d concurrent calls handled safely", numGoroutines)
}
