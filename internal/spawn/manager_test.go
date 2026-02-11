package spawn

import (
	"context"
	"testing"

	"github.com/udisondev/la2go/internal/ai"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// mockNpcRepository для тестов
type mockNpcRepository struct {
	templates map[int32]*model.NpcTemplate
}

func newMockNpcRepository() *mockNpcRepository {
	return &mockNpcRepository{
		templates: make(map[int32]*model.NpcTemplate),
	}
}

func (r *mockNpcRepository) LoadTemplate(ctx context.Context, templateID int32) (*model.NpcTemplate, error) {
	template, ok := r.templates[templateID]
	if !ok {
		return nil, context.DeadlineExceeded // placeholder error
	}
	return template, nil
}

func (r *mockNpcRepository) AddTemplate(template *model.NpcTemplate) {
	r.templates[template.TemplateID()] = template
}

// mockSpawnRepository для тестов
type mockSpawnRepository struct {
	spawns map[int64]*model.Spawn
}

func newMockSpawnRepository() *mockSpawnRepository {
	return &mockSpawnRepository{
		spawns: make(map[int64]*model.Spawn),
	}
}

func (r *mockSpawnRepository) LoadAll(ctx context.Context) ([]*model.Spawn, error) {
	spawns := make([]*model.Spawn, 0, len(r.spawns))
	for _, spawn := range r.spawns {
		spawns = append(spawns, spawn)
	}
	return spawns, nil
}

func (r *mockSpawnRepository) AddSpawn(spawn *model.Spawn) {
	r.spawns[spawn.SpawnID()] = spawn
}

func TestManager_DoSpawn(t *testing.T) {
	// Setup
	npcRepo := newMockNpcRepository()
	spawnRepo := newMockSpawnRepository()
	w := world.Instance()
	aiMgr := ai.NewTickManager()

	mgr := NewManager(
		npcRepo,   // mock implements NpcRepository interface
		spawnRepo, // mock implements SpawnRepository interface
		w,
		aiMgr,
		nil,
	)

	// Create test template
	template := model.NewNpcTemplate(
		1000, "Wolf", "", 5, 1500, 800,
		100, 50, 80, 40, 0, 120, 253, 30, 60, 0, 0,
	)
	npcRepo.AddTemplate(template)

	// Create test spawn
	spawn := model.NewSpawn(1, 1000, 17000, 170000, -3500, 0, 1, true)

	// DoSpawn
	ctx := context.Background()
	npc, err := mgr.DoSpawn(ctx, spawn)
	if err != nil {
		t.Fatalf("DoSpawn() error = %v", err)
	}

	if npc == nil {
		t.Fatal("DoSpawn() returned nil NPC")
	}

	if npc.Name() != "Wolf" {
		t.Errorf("NPC name = %q, want Wolf", npc.Name())
	}

	if npc.Spawn() != spawn {
		t.Error("NPC spawn != expected spawn")
	}

	if spawn.CurrentCount() != 1 {
		t.Errorf("spawn.CurrentCount() = %d, want 1", spawn.CurrentCount())
	}

	// Verify NPC is in world
	_, ok := w.GetObject(npc.ObjectID())
	if !ok {
		t.Error("NPC not found in world")
	}

	// Cleanup
	mgr.DespawnNpc(npc)
}

func TestManager_DespawnNpc(t *testing.T) {
	// Setup
	npcRepo := newMockNpcRepository()
	spawnRepo := newMockSpawnRepository()
	w := world.Instance()
	aiMgr := ai.NewTickManager()

	mgr := NewManager(
		npcRepo,
		spawnRepo,
		w,
		aiMgr,
		nil,
	)

	template := model.NewNpcTemplate(
		1001, "Orc", "", 10, 2000, 1000,
		150, 75, 100, 50, 0, 100, 273, 60, 120, 0, 0,
	)
	npcRepo.AddTemplate(template)

	spawn := model.NewSpawn(2, 1001, 17000, 170000, -3500, 0, 1, true)

	ctx := context.Background()
	npc, err := mgr.DoSpawn(ctx, spawn)
	if err != nil {
		t.Fatalf("DoSpawn() error = %v", err)
	}

	// Verify NPC is spawned
	if spawn.CurrentCount() != 1 {
		t.Errorf("spawn.CurrentCount() before despawn = %d, want 1", spawn.CurrentCount())
	}

	// Despawn
	mgr.DespawnNpc(npc)

	// Verify NPC is removed
	if spawn.CurrentCount() != 0 {
		t.Errorf("spawn.CurrentCount() after despawn = %d, want 0", spawn.CurrentCount())
	}

	// Verify NPC is not in world
	_, ok := w.GetObject(npc.ObjectID())
	if ok {
		t.Error("NPC still in world after despawn")
	}
}

func TestManager_DoSpawn_SpawnFull(t *testing.T) {
	// Setup
	npcRepo := newMockNpcRepository()
	spawnRepo := newMockSpawnRepository()
	w := world.Instance()
	aiMgr := ai.NewTickManager()

	mgr := NewManager(
		npcRepo,
		spawnRepo,
		w,
		aiMgr,
		nil,
	)

	template := model.NewNpcTemplate(
		1002, "Rabbit", "", 1, 500, 100,
		10, 5, 5, 5, 0, 100, 253, 10, 20, 0, 0,
	)
	npcRepo.AddTemplate(template)

	// Spawn with maximumCount = 1
	spawn := model.NewSpawn(3, 1002, 17000, 170000, -3500, 0, 1, true)

	ctx := context.Background()

	// Spawn first NPC (should succeed)
	npc1, err := mgr.DoSpawn(ctx, spawn)
	if err != nil {
		t.Fatalf("first DoSpawn() error = %v", err)
	}

	// Try to spawn second NPC (should fail — spawn full)
	_, err = mgr.DoSpawn(ctx, spawn)
	if err == nil {
		t.Error("second DoSpawn() should fail (spawn full)")
	}

	// Cleanup
	mgr.DespawnNpc(npc1)
}

func TestCalculateRespawnDelay(t *testing.T) {
	template := model.NewNpcTemplate(
		1003, "Test", "", 1, 1000, 500,
		0, 0, 0, 0, 0, 80, 253, 30, 60, 0, 0,
	)

	// Run multiple times to verify randomness
	for range 10 {
		delay := CalculateRespawnDelay(template)

		if delay < 30 || delay > 60 {
			t.Errorf("CalculateRespawnDelay() = %d, want between 30 and 60", delay)
		}
	}
}

func TestCalculateRespawnDelay_SameMinMax(t *testing.T) {
	template := model.NewNpcTemplate(
		1004, "Test", "", 1, 1000, 500,
		0, 0, 0, 0, 0, 80, 253, 45, 45, 0, 0, // same min/max
	)

	delay := CalculateRespawnDelay(template)

	if delay != 45 {
		t.Errorf("CalculateRespawnDelay() with same min/max = %d, want 45", delay)
	}
}
