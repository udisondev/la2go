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
	"github.com/udisondev/la2go/internal/game/itemhandler"
	"github.com/udisondev/la2go/internal/game/raid"
	"github.com/udisondev/la2go/internal/game/quest"
	"github.com/udisondev/la2go/internal/game/skill"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/gslistener"
	"github.com/udisondev/la2go/internal/html"
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

	// Load skill data (Phase 5.9.1: Skill Data Model)
	if err := data.LoadSkills(); err != nil {
		return fmt.Errorf("loading skills: %w", err)
	}
	if err := data.LoadSkillTrees(); err != nil {
		return fmt.Errorf("loading skill trees: %w", err)
	}

	// Load NPC templates from XML-generated data
	if err := data.LoadNpcTemplates(); err != nil {
		return fmt.Errorf("loading NPC templates: %w", err)
	}

	// Load item templates from XML-generated data
	if err := data.LoadItemTemplates(); err != nil {
		return fmt.Errorf("loading item templates: %w", err)
	}

	// Load henna templates (Phase 13)
	if err := data.LoadHennaTemplates(); err != nil {
		return fmt.Errorf("loading henna templates: %w", err)
	}

	// Load spawns from XML-generated data
	if err := data.LoadSpawns(); err != nil {
		return fmt.Errorf("loading spawns: %w", err)
	}

	// Load zones
	if err := data.LoadZones(); err != nil {
		return fmt.Errorf("loading zones: %w", err)
	}

	// Load misc data (buylists, teleporters, multisell, doors, etc.)
	if err := data.LoadBuylists(); err != nil {
		return fmt.Errorf("loading buylists: %w", err)
	}
	if err := data.LoadTeleporters(); err != nil {
		return fmt.Errorf("loading teleporters: %w", err)
	}
	if err := data.LoadMultisell(); err != nil {
		return fmt.Errorf("loading multisell: %w", err)
	}
	if err := data.LoadDoors(); err != nil {
		return fmt.Errorf("loading doors: %w", err)
	}
	if err := data.LoadArmorsets(); err != nil {
		return fmt.Errorf("loading armorsets: %w", err)
	}
	if err := data.LoadRecipes(); err != nil {
		return fmt.Errorf("loading recipes: %w", err)
	}
	if err := data.LoadAugmentations(); err != nil {
		return fmt.Errorf("loading augmentations: %w", err)
	}
	if err := data.LoadPlayerConfig(); err != nil {
		return fmt.Errorf("loading player config: %w", err)
	}
	if err := data.LoadCategoryData(); err != nil {
		return fmt.Errorf("loading category data: %w", err)
	}
	if err := data.LoadPetData(); err != nil {
		return fmt.Errorf("loading pet data: %w", err)
	}
	if err := data.LoadFishingData(); err != nil {
		return fmt.Errorf("loading fishing data: %w", err)
	}
	if err := data.LoadSeeds(); err != nil {
		return fmt.Errorf("loading seeds: %w", err)
	}

	// Phase 51: Initialize item handler registry
	itemhandler.Init()
	slog.Info("item handlers initialized")

	// Initialize World Grid
	worldInstance := world.Instance()
	slog.Info("world initialized", "regions", worldInstance.RegionCount())

	// Create repositories
	npcRepo := spawn.NewDataNpcRepo()   // data-backed NPC template lookup
	spawnRepo := spawn.NewDataSpawnRepo() // data-backed spawn list
	charRepo := db.NewCharacterRepository(database.Pool()) // Phase 4.6: character repository
	itemRepo := db.NewItemRepository(database.Pool())      // Phase 6.0: item repository
	skillRepo := db.NewSkillRepository(database.Pool())    // Phase 6.0: skill repository
	recipeRepo := db.NewRecipeRepository(database.Pool())    // Phase 15: recipe repository
	hennaRepo := db.NewHennaRepository(database.Pool())          // Phase 13: henna repository
	subclassRepo := db.NewSubClassRepository(database.Pool())  // Phase 14: subclass repository
	questRepo := db.NewQuestRepository(database.Pool())        // Phase 16: quest repository
	friendRepo := db.NewFriendRepository(database.Pool())      // Phase 35: friend/block repository
	persister := db.NewPlayerPersistenceService(database.Pool(), charRepo, itemRepo, skillRepo, recipeRepo, hennaRepo, subclassRepo, friendRepo)

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

	// Initialize HTML dialog system (Phase 11)
	htmlCache, err := html.NewCache(gameCfg.HtmlDir, gameCfg.LazyHtmlLoad)
	if err != nil {
		return fmt.Errorf("creating HTML cache: %w", err)
	}
	dialogMgr := html.NewDialogManager(htmlCache)

	// Phase 16: Quest system
	questMgr := quest.NewManager(questRepo)

	// Create game server (game clients on :7777)
	gameServer, err := gameserver.NewServer(gameCfg, loginServer.SessionManager(), charRepo, persister, dialogMgr, questMgr)
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
	attackStanceMgr := combat.NewAttackStanceManager(func(source *model.Player, data []byte, size int) {
		gameServer.ClientManager().BroadcastToVisibleNear(source, data, size)
	})
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

	// Phase 37: Wire player death callback — sends Die packet to victim's client
	combatMgr.SetPlayerDeathFunc(func(victim *model.Character, killer *model.Player) {
		victimClient := gameServer.ClientManager().GetClientByObjectID(uint32(victim.ObjectID()))
		if victimClient == nil {
			return
		}

		diePkt := &serverpackets.Die{
			ObjectID:    int32(victim.ObjectID()),
			CanTeleport: true,
		}
		dieData, dieErr := diePkt.Write()
		if dieErr != nil {
			slog.Error("serializing Die packet", "error", dieErr)
			return
		}
		victimClient.Send(dieData)

		// Broadcast Die to visible players
		gameServer.ClientManager().BroadcastFromPosition(victim.X(), victim.Y(), dieData, len(dieData))
	})

	slog.Info("combat manager initialized")

	// Create CastManager (Phase 5.9.4: Cast Flow & Packets)
	castMgr := skill.NewCastManager(sendToPlayerFunc, broadcastFunc, nil)
	skill.CastMgr = castMgr
	slog.Info("cast manager initialized")

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
				delay := spawn.CalculateRespawnDelay(npcSpawn)
				respawnMgr.ScheduleRespawn(npcSpawn, delay)
			}
		})
	})

	// Spawn all NPCs from data
	if err := spawnMgr.SpawnAll(ctx); err != nil {
		slog.Warn("failed to spawn all NPCs", "error", err)
	}

	slog.Info("spawn system initialized",
		"spawns_loaded", spawnMgr.SpawnCount(),
		"world_objects", worldInstance.ObjectCount())

	// Phase 23: Raid Boss System — managers for raid/grand boss respawns and points
	raidRepo := db.NewRaidRepository(database.Pool())

	raidSpawnStore := &raidSpawnStoreAdapter{repo: raidRepo}
	grandBossStore := &grandBossStoreAdapter{repo: raidRepo}
	raidPointsStore := &raidPointsStoreAdapter{repo: raidRepo}

	// Raid boss spawn manager — tracks respawn times for regular raid bosses
	var raidSpawnMgr *raid.SpawnManager
	raidSpawnMgr = raid.NewSpawnManager(raidSpawnStore, func(bossID int32) error {
		slog.Info("raid boss respawn triggered", "bossID", bossID)
		raidSpawnMgr.OnRaidBossSpawned(bossID)
		return nil
	})
	if err := raidSpawnMgr.Init(ctx); err != nil {
		slog.Warn("raid spawn manager init failed", "error", err)
	}

	// Grand boss manager — tracks states (ALIVE/DEAD/FIGHTING) for grand bosses
	grandBossMgr := raid.NewGrandBossManager(grandBossStore, func(bossID int32) (*model.GrandBoss, error) {
		slog.Info("grand boss respawn triggered", "bossID", bossID)
		return nil, nil // grand boss spawning handled by spawn system
	})
	if err := grandBossMgr.Init(ctx); err != nil {
		slog.Warn("grand boss manager init failed", "error", err)
	}

	// Raid points manager — tracks per-player kill points
	raidPointsMgr := raid.NewPointsManager(raidPointsStore)

	// Wire raid death callback: combat → raid managers
	combatMgr.SetRaidDeathFunc(func(killer *model.Player, npc *model.Npc, templateID int32) {
		// Award raid points to killer
		if err := raidPointsMgr.AddPoints(ctx, int32(killer.CharacterID()), templateID, npc.Level()); err != nil {
			slog.Error("add raid points", "charID", killer.CharacterID(), "bossID", templateID, "error", err)
		}

		// Notify raid spawn manager (regular raid boss death)
		if data.IsRaidBoss(templateID) {
			if err := raidSpawnMgr.OnRaidBossDeath(ctx, templateID, 0, 0); err != nil {
				slog.Error("raid boss death tracking", "bossID", templateID, "error", err)
			}
		}

		// Notify grand boss manager (grand boss death)
		if data.IsGrandBoss(templateID) {
			if err := grandBossMgr.OnGrandBossDeath(ctx, templateID, 172800); err != nil { // 48h default
				slog.Error("grand boss death tracking", "bossID", templateID, "error", err)
			}
		}
	})

	// Start raid respawn loops
	g.Go(func() error {
		slog.Info("starting raid boss respawn loop", "interval", "30s")
		if err := raidSpawnMgr.RunRespawnLoop(gctx); err != nil {
			return fmt.Errorf("raid respawn loop: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		slog.Info("starting grand boss respawn loop", "interval", "60s")
		if err := grandBossMgr.RunRespawnLoop(gctx); err != nil {
			return fmt.Errorf("grand boss respawn loop: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		slog.Info("starting grand boss save loop", "interval", "5m")
		if err := grandBossMgr.RunSaveLoop(gctx); err != nil {
			return fmt.Errorf("grand boss save loop: %w", err)
		}
		return nil
	})

	slog.Info("raid boss system initialized",
		"raidSpawns", raidSpawnMgr.EntryCount(),
		"grandBosses", grandBossMgr.EntryCount())

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
