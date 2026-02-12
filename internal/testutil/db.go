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
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/udisondev/la2go/internal/db/migrations"
)

// SetupTestDB создаёт PostgreSQL testcontainer, применяет миграции и возвращает pool.
// Использует модуль postgres с BasicWaitStrategies (log occurrence(2) + port check).
// Автоматически cleanup при завершении теста.
func SetupTestDB(tb testing.TB) *pgxpool.Pool {
	tb.Helper()
	ctx := context.Background()

	// Запускаем PostgreSQL 16 через специализированный модуль
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		tb.Fatalf("starting postgres container: %v", err)
	}

	tb.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			tb.Logf("terminating postgres container: %v", err)
		}
	})

	// Получаем DSN через встроенный метод контейнера
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		tb.Fatalf("getting connection string: %v", err)
	}

	// Подключаемся через pgxpool
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		tb.Fatalf("connecting to test db: %v", err)
	}
	tb.Cleanup(func() { pool.Close() })

	// Применяем миграции через goose
	if err := runMigrations(pool); err != nil {
		tb.Fatalf("running migrations: %v", err)
	}

	return pool
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
