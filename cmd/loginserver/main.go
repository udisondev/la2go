package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/gslistener"
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

	// Create GameServer table
	gsTable := gameserver.NewGameServerTable(database)
	slog.Info("GameServer table initialized")

	// Create login server (clients on :2106)
	loginServer, err := login.NewServer(cfg, database)
	if err != nil {
		return fmt.Errorf("creating login server: %w", err)
	}

	// Create gslistener server (GameServers on :9013)
	gsListener, err := gslistener.NewServer(cfg, database, gsTable, loginServer.SessionManager())
	if err != nil {
		return fmt.Errorf("creating gslistener server: %w", err)
	}

	// Run both servers in parallel
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		slog.Info("starting login server")
		if err := loginServer.Run(gctx); err != nil {
			return fmt.Errorf("login server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Info("starting gslistener server")
		if err := gsListener.Run(gctx); err != nil {
			return fmt.Errorf("gslistener server: %w", err)
		}
		return nil
	})

	// Wait for both servers to finish
	if err := g.Wait(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
