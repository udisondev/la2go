package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/udisondev/la2go/internal/ai"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/spawn"
	"github.com/udisondev/la2go/internal/world"
)

type SpawnIntegrationSuite struct {
	IntegrationSuite
}

func TestSpawnIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}

	suite.Run(t, new(SpawnIntegrationSuite))
}

// TestSpawnManager_LoadAndSpawnAll tests full spawn flow with database
func (s *SpawnIntegrationSuite) TestSpawnManager_LoadAndSpawnAll() {
	// Setup: Insert test template and spawn
	templateID := int32(1000)
	template := model.NewNpcTemplate(
		templateID, "Wolf", "Wild Beast", 5, 1500, 800,
		100, 50, 80, 40, 0, 120, 253, 30, 60,
	)

	npcRepo := db.NewNpcRepository(s.db.Pool())
	err := npcRepo.Create(s.ctx, template)
	s.Require().NoError(err)

	spawnRepo := db.NewSpawnRepository(s.db.Pool())
	spawnID, err := spawnRepo.Create(s.ctx,
		model.NewSpawn(0, templateID, 17000, 170000, -3500, 0, 3, true))
	s.Require().NoError(err)

	// Create managers
	w := world.Instance()
	aiMgr := ai.NewTickManager()
	spawnMgr := spawn.NewManager(npcRepo, spawnRepo, w, aiMgr)

	// Load spawns from DB
	err = spawnMgr.LoadSpawns(s.ctx)
	s.Require().NoError(err)
	s.Equal(1, spawnMgr.SpawnCount())

	// Spawn all
	err = spawnMgr.SpawnAll(s.ctx)
	s.Require().NoError(err)

	// Verify 3 NPCs spawned (maximumCount=3)
	s.Equal(3, w.ObjectCount())

	// Verify NPCs in world
	spawnObj, ok := spawnMgr.GetSpawn(spawnID)
	s.Require().True(ok)
	s.Equal(int32(3), spawnObj.CurrentCount())

	// Cleanup
	for _, npc := range spawnObj.NPCs() {
		spawnMgr.DespawnNpc(npc)
	}
	s.Equal(0, w.ObjectCount())
}

// TestRespawnTaskManager_FullFlow tests respawn scheduling and execution
func (s *SpawnIntegrationSuite) TestRespawnTaskManager_FullFlow() {
	// Setup template with short respawn (5 seconds)
	template := model.NewNpcTemplate(
		2000, "Rabbit", "", 1, 500, 100,
		10, 5, 5, 5, 0, 100, 253, 5, 5, // respawnMin=5, respawnMax=5
	)

	npcRepo := db.NewNpcRepository(s.db.Pool())
	err := npcRepo.Create(s.ctx, template)
	s.Require().NoError(err)

	spawnRepo := db.NewSpawnRepository(s.db.Pool())
	spawnID, err := spawnRepo.Create(s.ctx,
		model.NewSpawn(0, 2000, 20000, 180000, -3500, 0, 1, true))
	s.Require().NoError(err)

	// Create managers
	w := world.Instance()
	aiMgr := ai.NewTickManager()
	spawnMgr := spawn.NewManager(npcRepo, spawnRepo, w, aiMgr)

	err = spawnMgr.LoadSpawns(s.ctx)
	s.Require().NoError(err)

	// Spawn NPC
	err = spawnMgr.SpawnAll(s.ctx)
	s.Require().NoError(err)

	spawnObj, _ := spawnMgr.GetSpawn(spawnID)
	npcs := spawnObj.NPCs()
	s.Require().Len(npcs, 1)
	npc := npcs[0]

	// Start respawn manager with longer timeout (need time for respawn to complete)
	respawnMgr := spawn.NewRespawnTaskManager(spawnMgr)
	ctx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancel()

	go func() {
		_ = respawnMgr.Start(ctx)
	}()

	// Despawn NPC
	spawnMgr.DespawnNpc(npc)
	s.Equal(int32(0), spawnObj.CurrentCount())

	// Schedule respawn
	delay := spawn.CalculateRespawnDelay(template)
	respawnMgr.ScheduleRespawn(spawnObj, delay)

	// Wait for respawn (5s delay + 1s buffer + tick interval)
	time.Sleep(7 * time.Second)

	// Verify NPC respawned
	s.Equal(int32(1), spawnObj.CurrentCount(), "NPC should respawn after delay")

	// Cleanup
	cancel()
	for _, npc := range spawnObj.NPCs() {
		spawnMgr.DespawnNpc(npc)
	}
}

// TestWorld_NPCVisibility tests world grid visibility
func (s *SpawnIntegrationSuite) TestWorld_NPCVisibility() {
	// Setup
	template := model.NewNpcTemplate(
		3000, "Orc", "", 10, 2000, 1000,
		150, 75, 100, 50, 0, 100, 273, 60, 120,
	)

	npcRepo := db.NewNpcRepository(s.db.Pool())
	err := npcRepo.Create(s.ctx, template)
	s.Require().NoError(err)

	spawnRepo := db.NewSpawnRepository(s.db.Pool())
	w := world.Instance()
	aiMgr := ai.NewTickManager()
	spawnMgr := spawn.NewManager(npcRepo, spawnRepo, w, aiMgr)

	// Create spawn at specific coordinates
	baseX, baseY := int32(17000), int32(170000)
	spawnObj := model.NewSpawn(0, 3000, baseX, baseY, -3500, 0, 1, true)

	// Manually spawn (bypass DB)
	npc, err := spawnMgr.DoSpawn(s.ctx, spawnObj)
	s.Require().NoError(err)

	// Verify NPC is in world
	obj, ok := w.GetObject(npc.ObjectID())
	s.Require().True(ok)
	s.Equal(npc.ObjectID(), obj.ObjectID())

	// Verify NPC is visible from same region
	count := world.CountVisibleObjects(w, baseX, baseY)
	s.Equal(1, count)

	// Verify NPC is visible from neighboring region
	neighborX := baseX + world.RegionSize
	neighborY := baseY
	count = world.CountVisibleObjects(w, neighborX, neighborY)
	s.Equal(1, count) // Still visible (3×3 window)

	// Verify NPC is NOT visible from distant region
	distantX := baseX + 3*world.RegionSize
	distantY := baseY + 3*world.RegionSize
	count = world.CountVisibleObjects(w, distantX, distantY)
	s.Equal(0, count)

	// Cleanup
	spawnMgr.DespawnNpc(npc)
}

// TestAI_TickingNPCs tests AI state transitions
func (s *SpawnIntegrationSuite) TestAI_TickingNPCs() {
	// Setup
	template := model.NewNpcTemplate(
		4000, "Goblin", "", 8, 1200, 600,
		80, 40, 60, 30, 0, 110, 263, 40, 80,
	)

	npcRepo := db.NewNpcRepository(s.db.Pool())
	err := npcRepo.Create(s.ctx, template)
	s.Require().NoError(err)

	spawnRepo := db.NewSpawnRepository(s.db.Pool())
	w := world.Instance()

	// Start AI manager
	aiMgr := ai.NewTickManager()
	ctx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancel()

	go func() {
		_ = aiMgr.Start(ctx)
	}()

	spawnMgr := spawn.NewManager(npcRepo, spawnRepo, w, aiMgr)

	// Spawn NPC
	spawnObj := model.NewSpawn(0, 4000, 18000, 171000, -3500, 0, 1, true)
	npc, err := spawnMgr.DoSpawn(s.ctx, spawnObj)
	s.Require().NoError(err)

	// Verify initial intention = ACTIVE (from BasicNpcAI.Start)
	s.Equal(model.IntentionActive, npc.Intention())

	// Wait for ~5 ticks (BasicNpcAI toggles every 5 ticks)
	time.Sleep(6 * time.Second)

	// Verify intention changed to IDLE
	s.Equal(model.IntentionIdle, npc.Intention())

	// Wait another ~5 ticks
	time.Sleep(6 * time.Second)

	// Verify intention changed back to ACTIVE
	s.Equal(model.IntentionActive, npc.Intention())

	// Cleanup
	cancel()
	spawnMgr.DespawnNpc(npc)
}

// TestSpawnManager_ConcurrentDoSpawn tests concurrent spawning
func (s *SpawnIntegrationSuite) TestSpawnManager_ConcurrentDoSpawn() {
	// Setup
	template := model.NewNpcTemplate(
		5000, "Spider", "", 6, 800, 400,
		60, 30, 40, 20, 0, 95, 248, 25, 50,
	)

	npcRepo := db.NewNpcRepository(s.db.Pool())
	err := npcRepo.Create(s.ctx, template)
	s.Require().NoError(err)

	spawnRepo := db.NewSpawnRepository(s.db.Pool())
	w := world.Instance()
	aiMgr := ai.NewTickManager()
	spawnMgr := spawn.NewManager(npcRepo, spawnRepo, w, aiMgr)

	// Create spawn with maximumCount=10
	spawnObj := model.NewSpawn(0, 5000, 19000, 172000, -3500, 0, 10, true)

	// Spawn 10 NPCs concurrently
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := spawnMgr.DoSpawn(s.ctx, spawnObj)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Verify no errors
	for err := range errors {
		s.Require().NoError(err)
	}

	// Verify exactly 10 NPCs spawned
	s.Equal(int32(10), spawnObj.CurrentCount())
	s.Len(spawnObj.NPCs(), 10)

	// Verify unique objectIDs
	objectIDs := make(map[uint32]bool)
	for _, npc := range spawnObj.NPCs() {
		s.False(objectIDs[npc.ObjectID()], "duplicate objectID")
		objectIDs[npc.ObjectID()] = true
	}

	// Cleanup
	for _, npc := range spawnObj.NPCs() {
		spawnMgr.DespawnNpc(npc)
	}
}
