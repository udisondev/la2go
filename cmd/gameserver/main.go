package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/udisondev/la2go/internal/ai"
	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/gslistener"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/spawn"
	"github.com/udisondev/la2go/internal/world"
)

const (
	LoginConfigPath = "config/loginserver.yaml"
	GameConfigPath  = "config/gameserver.yaml"
)

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

	slog.Info("la2go server starting")

	// Load configs
	loginCfgPath := LoginConfigPath
	if p := os.Getenv("LA2GO_LOGIN_CONFIG"); p != "" {
		loginCfgPath = p
	}
	loginCfg, err := config.LoadLoginServer(loginCfgPath)
	if err != nil {
		return fmt.Errorf("loading login config: %w", err)
	}

	gameCfgPath := GameConfigPath
	if p := os.Getenv("LA2GO_GAME_CONFIG"); p != "" {
		gameCfgPath = p
	}
	gameCfg, err := config.LoadGameServer(gameCfgPath)
	if err != nil {
		return fmt.Errorf("loading game config: %w", err)
	}

	slog.Info("configs loaded",
		"login_bind", loginCfg.BindAddress,
		"login_port", loginCfg.Port,
		"game_bind", gameCfg.BindAddress,
		"game_port", gameCfg.Port)

	// Connect to database
	database, err := db.New(ctx, loginCfg.Database.DSN())
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer database.Close()
	slog.Info("database connected")

	// Run migrations
	if err := db.RunMigrations(ctx, loginCfg.Database.DSN()); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	slog.Info("database migrations applied")

	// Initialize World Grid
	worldInstance := world.Instance()
	slog.Info("world initialized", "regions", worldInstance.RegionCount())

	// Create repositories
	npcRepo := db.NewNpcRepository(database.Pool())
	spawnRepo := db.NewSpawnRepository(database.Pool())

	// Create GameServer table
	gsTable := gameserver.NewGameServerTable(database)
	slog.Info("GameServer table initialized")

	// Create login server (clients on :2106)
	loginServer, err := login.NewServer(loginCfg, database)
	if err != nil {
		return fmt.Errorf("creating login server: %w", err)
	}

	// Create gslistener server (GameServers on :9013)
	gsListener, err := gslistener.NewServer(loginCfg, database, gsTable, loginServer.SessionManager())
	if err != nil {
		return fmt.Errorf("creating gslistener server: %w", err)
	}

	// Create game server (game clients on :7777)
	gameServer, err := gameserver.NewServer(gameCfg, loginServer.SessionManager())
	if err != nil {
		return fmt.Errorf("creating game server: %w", err)
	}

	// Run all three servers + AI/Respawn managers in parallel
	g, gctx := errgroup.WithContext(ctx)

	// Create AI tick manager
	aiMgr := ai.NewTickManager()
	g.Go(func() error {
		slog.Info("starting AI tick manager", "interval", "1s")
		if err := aiMgr.Start(gctx); err != nil {
			return fmt.Errorf("AI tick manager: %w", err)
		}
		return nil
	})

	// Create Spawn manager
	spawnMgr := spawn.NewManager(npcRepo, spawnRepo, worldInstance, aiMgr)
	if err := spawnMgr.LoadSpawns(ctx); err != nil {
		return fmt.Errorf("loading spawns: %w", err)
	}

	// Create Respawn task manager
	respawnMgr := spawn.NewRespawnTaskManager(spawnMgr)
	g.Go(func() error {
		slog.Info("starting respawn task manager", "interval", "1s")
		if err := respawnMgr.Start(gctx); err != nil {
			return fmt.Errorf("respawn task manager: %w", err)
		}
		return nil
	})

	// Spawn all NPCs from database
	if err := spawnMgr.SpawnAll(ctx); err != nil {
		slog.Warn("failed to spawn all NPCs", "error", err)
	}

	slog.Info("spawn system initialized",
		"spawns_loaded", spawnMgr.SpawnCount(),
		"world_objects", worldInstance.ObjectCount())

	// DEMO: Spawn test NPC if no spawns in database
	if spawnMgr.SpawnCount() == 0 {
		simpleSpawner := spawn.NewSimpleSpawner(spawnMgr)
		testNpc, err := simpleSpawner.SpawnTestNpc(ctx)
		if err != nil {
			slog.Warn("failed to spawn test NPC", "error", err)
		} else {
			slog.Info("demo: test NPC spawned",
				"name", testNpc.Name(),
				"objectID", testNpc.ObjectID(),
				"location", testNpc.Location())
		}
	}

	g.Go(func() error {
		slog.Info("starting login server", "port", loginCfg.Port)
		if err := loginServer.Run(gctx); err != nil {
			return fmt.Errorf("login server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Info("starting gslistener server", "port", loginCfg.GSListenPort)
		if err := gsListener.Run(gctx); err != nil {
			return fmt.Errorf("gslistener server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		slog.Info("starting game server", "port", gameCfg.Port)
		if err := gameServer.Run(gctx); err != nil {
			return fmt.Errorf("game server: %w", err)
		}
		return nil
	})

	// Wait for all servers to finish
	if err := g.Wait(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
