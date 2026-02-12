package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/udisondev/la2go/internal/db/migrations"
)

var gooseOnce sync.Once

// RunMigrations runs goose migrations on the given DSN.
func RunMigrations(ctx context.Context, dsn string) error {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("opening sql connection for migrations: %w", err)
	}
	defer sqlDB.Close()

	var dialectErr error
	gooseOnce.Do(func() {
		goose.SetBaseFS(migrations.FS)
		dialectErr = goose.SetDialect("postgres")
	})
	if dialectErr != nil {
		return fmt.Errorf("setting goose dialect: %w", dialectErr)
	}
	if err := goose.UpContext(ctx, sqlDB, "."); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	return nil
}
