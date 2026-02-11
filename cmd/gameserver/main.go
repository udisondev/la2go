package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/udisondev/la2go/internal/ai"
	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/game/combat"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/gslistener"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/model"
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
	// Load configs FIRST to determine log level
	loginCfgPath := LoginConfigPath
	if p := os.Getenv("LA2GO_LOGIN_CONFIG"); p != "" {
		loginCfgPath = p
	}
	loginCfg, err := config.LoadLoginServer(loginCfgPath)
	if err != nil {
		return fmt.Errorf("loading login config: %w", err)
	}

	// Configure slog based on config.LogLevel
	logLevel := parseLogLevel(loginCfg.LogLevel)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})))

	// Enable AI debug logging if log level is debug
	ai.EnableDebugLogging(logLevel == slog.LevelDebug)

	slog.Info("la2go server starting", "log_level", loginCfg.LogLevel)

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

	// Load player templates (Phase 5.4: Character Templates & Stats System)
	slog.Info("loading player templates")
	data.InitStatBonuses() // Initialize stat bonus tables
	if err := data.LoadPlayerTemplates(); err != nil {
		return fmt.Errorf("loading player templates: %w", err)
	}

	// Initialize World Grid
	worldInstance := world.Instance()
	slog.Info("world initialized", "regions", worldInstance.RegionCount())

	// Create repositories
	npcRepo := db.NewNpcRepository(database.Pool())
	spawnRepo := db.NewSpawnRepository(database.Pool())
	charRepo := db.NewCharacterRepository(database.Pool()) // Phase 4.6: character repository

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
	gameServer, err := gameserver.NewServer(gameCfg, loginServer.SessionManager(), charRepo)
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

	// Create Visibility manager (Phase 4.5 PR3)
	visibilityMgr := world.NewVisibilityManager(worldInstance, 100*time.Millisecond, 200*time.Millisecond)

	// Phase 4.18 Optimization 1: Link VisibilityManager to ClientManager for reverse cache
	// Enables O(1) broadcast queries via GetObservers() instead of O(N×M) iteration
	gameServer.ClientManager().SetVisibilityManager(visibilityMgr)

	g.Go(func() error {
		slog.Info("starting visibility manager", "interval", "100ms", "maxAge", "200ms")
		if err := visibilityMgr.Start(gctx); err != nil {
			return fmt.Errorf("visibility manager: %w", err)
		}
		return nil
	})

	// Create AttackStanceManager (Phase 5.3: Basic Combat System)
	attackStanceMgr := combat.NewAttackStanceManager()
	combat.AttackStanceMgr = attackStanceMgr

	g.Go(func() error {
		slog.Info("starting attack stance manager", "interval", "1s", "combatTime", "15s")
		attackStanceMgr.Start()
		<-gctx.Done()
		attackStanceMgr.Stop()
		return nil
	})

	// Create CombatManager (Phase 5.3: Basic Combat System)
	// Phase 5.7: Added NPC broadcast and AI manager for aggro
	broadcastFunc := func(source *model.Player, data []byte, size int) {
		gameServer.ClientManager().BroadcastToVisibleNear(source, data, size)
	}
	npcBroadcastFunc := func(x, y int32, data []byte, size int) {
		gameServer.ClientManager().BroadcastFromPosition(x, y, data, size)
	}
	combatMgr := combat.NewCombatManager(broadcastFunc, npcBroadcastFunc, &aiManagerAdapter{aiMgr})
	combat.CombatMgr = combatMgr

	// Phase 5.8: Wire experience reward callback
	sendToPlayerFunc := func(objectID uint32, pktData []byte, size int) {
		if err := gameServer.ClientManager().SendToPlayer(objectID, pktData, size); err != nil {
			slog.Warn("failed to send packet to player", "objectID", objectID, "error", err)
		}
	}
	combatMgr.SetRewardFunc(func(killer *model.Player, npc *model.Npc) {
		combat.RewardExpAndSp(killer, npc, sendToPlayerFunc, broadcastFunc)
	})

	slog.Info("combat manager initialized")

	// Create Spawn manager (Phase 5.7: inject attack callback for AttackableAI)
	attackFunc := func(monster *model.Monster, target *model.WorldObject) {
		combatMgr.ExecuteNpcAttack(monster.Npc, target)
	}
	spawnMgr := spawn.NewManager(npcRepo, spawnRepo, worldInstance, aiMgr, attackFunc)
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

	// Wire NPC death → despawn → respawn flow
	combatMgr.SetNpcDeathFunc(func(npc *model.Npc) {
		npcSpawn := npc.Spawn()

		// Stop AI immediately
		aiMgr.Unregister(npc.ObjectID())

		// Schedule corpse despawn after 8 seconds
		time.AfterFunc(8*time.Second, func() {
			// Remove corpse from world + broadcast DeleteObject
			spawnMgr.DespawnNpc(npc)

			deleteObj := serverpackets.NewDeleteObject(int32(npc.ObjectID()))
			deleteData, err := deleteObj.Write()
			if err != nil {
				slog.Error("failed to write DeleteObject for NPC corpse",
					"npc", npc.Name(),
					"error", err)
				return
			}

			loc := npc.Location()
			npcBroadcastFunc(loc.X, loc.Y, deleteData, len(deleteData))

			slog.Info("NPC corpse despawned",
				"objectID", npc.ObjectID(),
				"name", npc.Name())

			// Schedule respawn
			if npcSpawn != nil {
				delay := spawn.CalculateRespawnDelay(npc.Template())
				respawnMgr.ScheduleRespawn(npcSpawn, delay)
			}
		})
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

// aiManagerAdapter adapts ai.TickManager to combat.AIManagerInterface.
// Phase 5.7: ai.Controller → combat.AIController covariant return type.
type aiManagerAdapter struct {
	mgr *ai.TickManager
}

func (a *aiManagerAdapter) GetController(objectID uint32) (combat.AIController, error) {
	ctrl, err := a.mgr.GetController(objectID)
	if err != nil {
		return nil, err
	}
	return ctrl, nil
}

// parseLogLevel converts string log level to slog.Level.
// Defaults to Info if invalid or empty.
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
