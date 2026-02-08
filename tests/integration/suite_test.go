package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/udisondev/la2go/internal/db"
)

// IntegrationSuite — базовый suite для интеграционных тестов.
// Автоматически запускает PostgreSQL через testcontainers.
type IntegrationSuite struct {
	suite.Suite
	db              *db.DB
	ctx             context.Context
	postgresContainer *postgres.PostgresContainer
}

// SetupSuite выполняется один раз перед всеми тестами в suite.
func (s *IntegrationSuite) SetupSuite() {
	s.ctx = context.Background()

	// Если DB_ADDR задан вручную — используем его (для CI/CD)
	dbAddr := os.Getenv("DB_ADDR")
	if dbAddr == "" {
		// Запускаем PostgreSQL через testcontainers
		var err error
		s.postgresContainer, err = postgres.Run(s.ctx,
			"postgres:17-alpine",
			postgres.WithDatabase("la2go_test"),
			postgres.WithUsername("la2go"),
			postgres.WithPassword("testpass"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2),
			),
		)
		if err != nil {
			s.T().Fatalf("failed to start postgres container: %v", err)
		}

		dbAddr, err = s.postgresContainer.ConnectionString(s.ctx, "sslmode=disable")
		if err != nil {
			s.T().Fatalf("failed to get connection string: %v", err)
		}
	}

	// Run migrations first
	if err := db.RunMigrations(s.ctx, dbAddr); err != nil {
		s.T().Fatalf("failed to run migrations: %v", err)
	}

	var err error
	s.db, err = db.New(s.ctx, dbAddr)
	if err != nil {
		s.T().Fatalf("failed to connect to database: %v", err)
	}
}

// SetupTest выполняется перед каждым тестом для очистки данных.
func (s *IntegrationSuite) SetupTest() {
	// Очищаем тестовые данные
	if err := s.cleanupTestData(); err != nil {
		s.T().Fatalf("failed to cleanup test data: %v", err)
	}
}

// TearDownSuite выполняется один раз после всех тестов в suite.
func (s *IntegrationSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}

	// Останавливаем testcontainer (если запускали)
	if s.postgresContainer != nil {
		if err := testcontainers.TerminateContainer(s.postgresContainer); err != nil {
			s.T().Logf("failed to terminate postgres container: %v", err)
		}
	}
}

// cleanupTestData очищает тестовые данные из базы.
func (s *IntegrationSuite) cleanupTestData() error {
	// Удаляем тестовые аккаунты
	query := "DELETE FROM accounts WHERE login LIKE 'test%' OR login LIKE 'user%'"
	if _, err := s.db.Pool().Exec(s.ctx, query); err != nil {
		return fmt.Errorf("failed to cleanup test accounts: %w", err)
	}

	// Удаляем тестовые game server записи (если таблица существует)
	query = "DELETE FROM game_servers WHERE server_id >= 100"
	_, _ = s.db.Pool().Exec(s.ctx, query) // Игнорируем ошибку если таблицы нет

	// Удаляем тестовые spawns (если таблица существует)
	query = "DELETE FROM spawns WHERE template_id >= 1000"
	_, _ = s.db.Pool().Exec(s.ctx, query)

	// Удаляем тестовые npc templates (если таблица существует)
	query = "DELETE FROM npc_templates WHERE template_id >= 1000"
	_, _ = s.db.Pool().Exec(s.ctx, query)

	return nil
}

// TestIntegrationSuite — entry point для запуска IntegrationSuite.
// Можно расширить другими suite через embedding.
func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}

	suite.Run(t, new(IntegrationSuite))
}
