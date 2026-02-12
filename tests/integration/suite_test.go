package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/udisondev/la2go/internal/db"
)

// IntegrationSuite — базовый suite для интеграционных тестов.
// PostgreSQL контейнер создаётся один раз в TestMain, каждый suite получает
// изолированную schema через acquireSchema().
type IntegrationSuite struct {
	suite.Suite
	db  *db.DB
	ctx context.Context
}

// SetupSuite выполняется один раз перед всеми тестами в suite.
func (s *IntegrationSuite) SetupSuite() {
	s.ctx = context.Background()

	// Если DB_ADDR задан вручную — используем его (для CI/CD)
	dbAddr := os.Getenv("DB_ADDR")
	if dbAddr == "" {
		dbAddr = acquireSchema(s.T())
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
	// Контейнер терминируется в TestMain, schema удаляется через t.Cleanup
}

// cleanupTestData очищает все данные из тестовых таблиц.
func (s *IntegrationSuite) cleanupTestData() error {
	_, err := s.db.Pool().Exec(s.ctx,
		"TRUNCATE TABLE accounts, characters, items, character_skills CASCADE")
	if err != nil {
		return fmt.Errorf("truncating test tables: %w", err)
	}
	return nil
}

// TestIntegrationSuite — entry point для запуска IntegrationSuite.
func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}

	suite.Run(t, new(IntegrationSuite))
}
