package integration

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/udisondev/la2go/internal/testutil"
)

// DatabaseSuite тестирует database операции.
type DatabaseSuite struct {
	IntegrationSuite
}

// TestAccountCRUD тестирует создание, чтение, обновление аккаунта.
func (s *DatabaseSuite) TestAccountCRUD() {
	ctx := s.ctx
	login := "testuser1"
	password := testutil.Fixtures.ValidHash
	ip := "127.0.0.1"

	// Create
	err := s.db.CreateAccount(ctx, login, password, ip)
	s.Require().NoError(err, "CreateAccount должен успешно создать аккаунт")

	// Read
	acc, err := s.db.GetAccount(ctx, login)
	s.Require().NoError(err)
	s.Require().NotNil(acc)
	s.Equal(login, acc.Login)
	s.Equal(password, acc.PasswordHash)
	s.Equal(ip, acc.LastIP)
	s.Equal(0, acc.AccessLevel)

	// Update LastLogin
	newIP := "192.168.1.1"
	err = s.db.UpdateLastLogin(ctx, login, newIP)
	s.Require().NoError(err)

	// Verify update
	acc, err = s.db.GetAccount(ctx, login)
	s.Require().NoError(err)
	s.Equal(newIP, acc.LastIP)
}

// TestAccountNotFound тестирует получение несуществующего аккаунта.
func (s *DatabaseSuite) TestAccountNotFound() {
	acc, err := s.db.GetAccount(s.ctx, "nonexistent_user")
	s.Require().NoError(err)
	s.Nil(acc, "несуществующий аккаунт должен вернуть nil")
}

// TestCreateAccountDuplicate тестирует создание дубликата аккаунта.
func (s *DatabaseSuite) TestCreateAccountDuplicate() {
	ctx := s.ctx
	login := "testuser2"
	password := testutil.Fixtures.ValidHash
	ip := "127.0.0.1"

	// Первое создание
	err := s.db.CreateAccount(ctx, login, password, ip)
	s.Require().NoError(err)

	// Попытка создать дубликат
	err = s.db.CreateAccount(ctx, login, password, ip)
	s.Error(err, "создание дубликата должно вернуть ошибку")
}

// TestConcurrentAccountCreation тестирует concurrent создание одного аккаунта.
// Должна сработать UNIQUE constraint в БД.
func (s *DatabaseSuite) TestConcurrentAccountCreation() {
	login := "testuser_concurrent"
	password := testutil.Fixtures.ValidHash
	ip := "127.0.0.1"

	const goroutines = 10
	var wg sync.WaitGroup
	errChan := make(chan error, goroutines)

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			err := s.db.CreateAccount(ctx, login, password, ip)
			errChan <- err
		}()
	}

	wg.Wait()
	close(errChan)

	// Должна быть ровно одна успешная операция
	successCount := 0
	errorCount := 0

	for err := range errChan {
		if err == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	s.Equal(1, successCount, "только один goroutine должен успешно создать аккаунт")
	s.Equal(goroutines-1, errorCount, "остальные должны получить ошибку")
}

// TestUpdateLastServer тестирует обновление last_server.
func (s *DatabaseSuite) TestUpdateLastServer() {
	ctx := s.ctx
	login := "testuser_server"
	password := testutil.Fixtures.ValidHash
	ip := "127.0.0.1"

	// Создаём аккаунт
	err := s.db.CreateAccount(ctx, login, password, ip)
	s.Require().NoError(err)

	// Обновляем last_server
	err = s.db.UpdateLastServer(ctx, login, 1)
	s.Require().NoError(err)

	// Проверяем
	acc, err := s.db.GetAccount(ctx, login)
	s.Require().NoError(err)
	s.Equal(1, acc.LastServer)
}

// TestUpdateLastLoginNonexistent тестирует обновление несуществующего аккаунта.
// В текущей реализации это не вызывает ошибку, просто не обновляет строки.
func (s *DatabaseSuite) TestUpdateLastLoginNonexistent() {
	err := s.db.UpdateLastLogin(s.ctx, "nonexistent", "127.0.0.1")
	s.NoError(err, "обновление несуществующего аккаунта не должно вызывать ошибку")
}

// ============================================================================
// Phase 3: Database Integration Tests
// ============================================================================

// TestGetAccountUpdateLastLoginIntegration тестирует интеграцию GetAccount + UpdateLastLogin.
func (s *DatabaseSuite) TestGetAccountUpdateLastLoginIntegration() {
	login := "integration_user"
	hash := "test_hash_integration"

	// Create account
	err := s.db.CreateAccount(s.ctx, login, hash, "127.0.0.1")
	s.Require().NoError(err)

	// Get account
	acc, err := s.db.GetAccount(s.ctx, login)
	s.Require().NoError(err)
	s.NotNil(acc)

	initialLastActive := acc.LastActive

	// Update last login (PostgreSQL NOW() обновит last_active)
	err = s.db.UpdateLastLogin(s.ctx, login, "192.168.1.1")
	s.Require().NoError(err)

	// Get account again
	acc, err = s.db.GetAccount(s.ctx, login)
	s.Require().NoError(err)
	s.NotNil(acc)

	// Verify updates (last_active должен быть обновлён или равен, PostgreSQL NOW() гарантирует корректное время)
	s.True(acc.LastActive.After(initialLastActive) || acc.LastActive.Equal(initialLastActive),
		"last_active should be updated or equal (depends on timestamp precision)")
	s.Equal("192.168.1.1", acc.LastIP, "last_ip should be updated")
}

// TestCreateAccountAutoCreateIntegration тестирует auto-create flow.
func (s *DatabaseSuite) TestCreateAccountAutoCreateIntegration() {
	login := "autocreate_integration"
	hash := "autocreate_hash"

	// Verify doesn't exist
	acc, err := s.db.GetAccount(s.ctx, login)
	s.Require().NoError(err)
	s.Nil(acc)

	// Create account (simulating auto-create)
	err = s.db.CreateAccount(s.ctx, login, hash, "10.0.0.1")
	s.Require().NoError(err)

	// Verify exists
	acc, err = s.db.GetAccount(s.ctx, login)
	s.Require().NoError(err)
	s.NotNil(acc)
	s.Equal(login, acc.Login)
	s.Equal(hash, acc.PasswordHash)
	s.Equal("10.0.0.1", acc.LastIP)
	s.Equal(0, acc.AccessLevel)
}

// TestUpdateLastServerIntegration тестирует UpdateLastServer + GetAccount.
func (s *DatabaseSuite) TestUpdateLastServerIntegration() {
	login := "lastserver_user"
	hash := "test_hash"

	// Create account
	err := s.db.CreateAccount(s.ctx, login, hash, "127.0.0.1")
	s.Require().NoError(err)

	// Update last server
	err = s.db.UpdateLastServer(s.ctx, login, 5)
	s.Require().NoError(err)

	// Get account and verify
	acc, err := s.db.GetAccount(s.ctx, login)
	s.Require().NoError(err)
	s.NotNil(acc)
	s.Equal(5, acc.LastServer)

	// Update to different server
	err = s.db.UpdateLastServer(s.ctx, login, 10)
	s.Require().NoError(err)

	acc, err = s.db.GetAccount(s.ctx, login)
	s.Require().NoError(err)
	s.Equal(10, acc.LastServer)
}

// TestDatabaseSuite запускает DatabaseSuite.
func TestDatabaseSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}
	t.Parallel()

	suite.Run(t, new(DatabaseSuite))
}
