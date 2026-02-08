package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/udisondev/la2go/internal/db/migrations"
)

// testPool — shared connection pool для всех benchmarks
var testPool *pgxpool.Pool

// TestMain настраивает окружение для всех tests/benchmarks в package db
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Запускаем PostgreSQL 16 testcontainer
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("starting postgres container: %v", err)
	}
	defer func() {
		_ = container.Terminate(ctx)
	}()

	// Получаем DSN
	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("getting container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("getting container port: %v", err)
	}
	dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())

	// Подключаемся через pgxpool
	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connecting to test db: %v", err)
	}
	defer testPool.Close()

	// Применяем миграции через goose
	if err := runMigrations(testPool); err != nil {
		log.Fatalf("running migrations: %v", err)
	}

	// Запускаем все tests/benchmarks
	code := m.Run()
	os.Exit(code)
}

// setupTestDB возвращает shared pool для benchmarks.
// Очищает таблицы перед каждым benchmark для изоляции.
func setupTestDB(tb testing.TB) *pgxpool.Pool {
	tb.Helper()

	// Очищаем таблицы для изоляции между benchmarks
	ctx := context.Background()
	queries := []string{
		"TRUNCATE items CASCADE",
		"TRUNCATE characters CASCADE",
		"TRUNCATE accounts CASCADE",
	}

	for _, query := range queries {
		if _, err := testPool.Exec(ctx, query); err != nil {
			tb.Logf("cleanup warning: %v", err) // non-fatal
		}
	}

	return testPool
}

// runMigrations применяет embedded миграции через goose.
func runMigrations(pool *pgxpool.Pool) error {
	// goose требует *sql.DB, получаем его из pgxpool
	connConfig := pool.Config().ConnConfig
	connStr := stdlib.RegisterConnConfig(connConfig)
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("opening sql.DB: %w", err)
	}
	defer sqlDB.Close()

	// Устанавливаем базовую директорию для goose (не используется для embedded FS)
	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting goose dialect: %w", err)
	}

	// Применяем миграции из embedded FS
	if err := goose.Up(sqlDB, "."); err != nil {
		return fmt.Errorf("running goose up: %w", err)
	}

	return nil
}
