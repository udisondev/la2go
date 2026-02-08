package spawn

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"sync/atomic"

	"github.com/udisondev/la2go/internal/ai"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// NpcRepository interface for loading NPC templates
type NpcRepository interface {
	LoadTemplate(ctx context.Context, templateID int32) (*model.NpcTemplate, error)
}

// SpawnRepository interface for loading spawns
type SpawnRepository interface {
	LoadAll(ctx context.Context) ([]*model.Spawn, error)
}

// Manager manages NPC spawns and respawns
type Manager struct {
	spawns    sync.Map // map[int64]*model.Spawn — spawnID → spawn
	npcRepo   NpcRepository
	spawnRepo SpawnRepository
	world     *world.World
	aiManager *ai.TickManager

	objectIDCounter atomic.Uint32 // for generating unique objectIDs
	spawnCount      atomic.Int32  // cached count of spawns (O(1) access)
}

// NewManager creates new spawn manager
func NewManager(
	npcRepo NpcRepository,
	spawnRepo SpawnRepository,
	world *world.World,
	aiManager *ai.TickManager,
) *Manager {
	mgr := &Manager{
		npcRepo:   npcRepo,
		spawnRepo: spawnRepo,
		world:     world,
		aiManager: aiManager,
	}

	// Start objectID counter from 100000 (players use lower IDs)
	mgr.objectIDCounter.Store(100000)

	return mgr
}

// LoadSpawns loads all spawns from database
func (m *Manager) LoadSpawns(ctx context.Context) error {
	spawns, err := m.spawnRepo.LoadAll(ctx)
	if err != nil {
		return fmt.Errorf("loading spawns from database: %w", err)
	}

	count := 0
	for _, spawn := range spawns {
		m.spawns.Store(spawn.SpawnID(), spawn)
		count++
	}

	// Update cached count
	m.spawnCount.Store(int32(count))

	slog.Info("spawns loaded from database", "count", count)
	return nil
}

// DoSpawn spawns NPC at spawn point
// Returns spawned NPC or error
func (m *Manager) DoSpawn(ctx context.Context, spawn *model.Spawn) (*model.Npc, error) {
	// Check if spawn is full
	if spawn.CurrentCount() >= spawn.MaximumCount() {
		return nil, fmt.Errorf("spawn %d is full (%d/%d)", spawn.SpawnID(), spawn.CurrentCount(), spawn.MaximumCount())
	}

	// Load NPC template
	template, err := m.npcRepo.LoadTemplate(ctx, spawn.TemplateID())
	if err != nil {
		return nil, fmt.Errorf("loading template %d for spawn %d: %w", spawn.TemplateID(), spawn.SpawnID(), err)
	}

	// Generate unique objectID
	objectID := m.objectIDCounter.Add(1)

	// Create NPC
	npc := model.NewNpc(objectID, spawn.TemplateID(), template)

	// Set spawn reference
	npc.SetSpawn(spawn)

	// Set location from spawn
	npc.SetLocation(spawn.Location())

	// Increase spawn count
	spawn.IncreaseCount()

	// Add NPC to spawn's NPC list
	spawn.AddNpc(npc)

	// Add NPC to world
	if err := m.world.AddObject(npc.WorldObject); err != nil {
		// Rollback
		spawn.DecreaseCount()
		spawn.RemoveNpc(npc)
		return nil, fmt.Errorf("adding NPC to world: %w", err)
	}

	// Create and register AI
	npcAI := ai.NewBasicNpcAI(npc)
	m.aiManager.Register(objectID, npcAI)

	slog.Info("NPC spawned",
		"objectID", objectID,
		"name", npc.Name(),
		"templateID", template.TemplateID(),
		"spawnID", spawn.SpawnID(),
		"location", spawn.Location())

	return npc, nil
}

// DespawnNpc despawns NPC (removes from world)
func (m *Manager) DespawnNpc(npc *model.Npc) {
	spawn := npc.Spawn()
	if spawn == nil {
		slog.Warn("despawning NPC without spawn", "objectID", npc.ObjectID())
		return
	}

	// Unregister AI
	m.aiManager.Unregister(npc.ObjectID())

	// Remove from world
	m.world.RemoveObject(npc.ObjectID())

	// Remove from spawn's NPC list
	spawn.RemoveNpc(npc)

	// Decrease spawn count
	spawn.DecreaseCount()

	slog.Info("NPC despawned",
		"objectID", npc.ObjectID(),
		"name", npc.Name(),
		"spawnID", spawn.SpawnID())
}

// ScheduleRespawn schedules NPC respawn after delay
// Used by RespawnTaskManager
func (m *Manager) ScheduleRespawn(ctx context.Context, spawn *model.Spawn) (*model.Npc, error) {
	return m.DoSpawn(ctx, spawn)
}

// GetSpawn returns spawn by ID
func (m *Manager) GetSpawn(spawnID int64) (*model.Spawn, bool) {
	value, ok := m.spawns.Load(spawnID)
	if !ok {
		return nil, false
	}
	return value.(*model.Spawn), true
}

// SpawnCount returns total number of spawns (O(1) cached count)
// IMPORTANT: Count is cached atomically and updated when spawns are loaded.
// This is a performance optimization to avoid O(N) Range() on sync.Map.
func (m *Manager) SpawnCount() int {
	return int(m.spawnCount.Load())
}

// SpawnAll spawns all NPCs for all loaded spawns
func (m *Manager) SpawnAll(ctx context.Context) error {
	count := 0
	var firstErr error

	m.spawns.Range(func(key, value any) bool {
		spawn := value.(*model.Spawn)

		// Spawn up to maximumCount NPCs
		for range spawn.MaximumCount() {
			if _, err := m.DoSpawn(ctx, spawn); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				slog.Error("failed to spawn NPC",
					"spawnID", spawn.SpawnID(),
					"templateID", spawn.TemplateID(),
					"error", err)
				return true // continue with next spawn
			}
			count++
		}

		return true
	})

	if firstErr != nil {
		slog.Warn("SpawnAll completed with errors", "spawned", count, "error", firstErr)
		return fmt.Errorf("spawning all NPCs: %w", firstErr)
	}

	slog.Info("all NPCs spawned", "count", count)
	return nil
}

// CalculateRespawnDelay calculates respawn delay for NPC template
// Returns random delay between respawnMin and respawnMax (in seconds)
func CalculateRespawnDelay(template *model.NpcTemplate) int32 {
	min := template.RespawnMin()
	max := template.RespawnMax()

	if min == max {
		return min
	}

	// Random delay between min and max
	return min + rand.Int32N(max-min+1)
}
