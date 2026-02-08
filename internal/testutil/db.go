package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/udisondev/la2go/internal/db/migrations"
)

// SetupTestDB создаёт PostgreSQL testcontainer, применяет миграции и возвращает pool.
// Автоматически cleanup при завершении теста.
func SetupTestDB(tb testing.TB) *pgxpool.Pool {
	tb.Helper()
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
		tb.Fatalf("starting postgres container: %v", err)
	}

	tb.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	// Получаем DSN
	host, err := container.Host(ctx)
	if err != nil {
		tb.Fatalf("getting container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		tb.Fatalf("getting container port: %v", err)
	}
	dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())

	// Подключаемся через pgxpool
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		tb.Fatalf("connecting to test db: %v", err)
	}
	tb.Cleanup(func() { pool.Close() })

	// Применяем миграции через goose
	if err := runMigrations(ctx, pool); err != nil {
		tb.Fatalf("running migrations: %v", err)
	}

	return pool
}

// runMigrations применяет embedded миграции через goose.
func runMigrations(_ context.Context, pool *pgxpool.Pool) error {
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
