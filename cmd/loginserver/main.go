package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/login"
)

const ConfigPath = "config/loginserver.yaml"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		slog.Info("shutting down", "signal", sig)
		cancel()
	}()

	if err := run(ctx); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Configure slog
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	slog.Info("la2go login server starting")

	// Load config
	cfgPath := ConfigPath
	if p := os.Getenv("LA2GO_CONFIG"); p != "" {
		cfgPath = p
	}
	cfg, err := config.LoadLoginServer(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	slog.Info("config loaded", "bind", cfg.BindAddress, "port", cfg.Port, "auto_create", cfg.AutoCreateAccounts)

	// Connect to database
	database, err := db.New(ctx, cfg.Database.DSN())
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer database.Close()
	slog.Info("database connected")

	// Run migrations
	if err := db.RunMigrations(ctx, cfg.Database.DSN()); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	slog.Info("database migrations applied")

	// Start login server
	server, err := login.NewServer(cfg, database)
	if err != nil {
		return fmt.Errorf("creating login server: %w", err)
	}

	if err := server.Run(ctx); err != nil {
		return fmt.Errorf("starting login server: %w", err)
	}

	return nil
}
