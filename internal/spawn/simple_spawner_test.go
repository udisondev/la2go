package spawn

import (
	"context"
	"testing"

	"github.com/udisondev/la2go/internal/ai"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

func TestSimpleSpawner_SpawnTestNpc(t *testing.T) {
	// Setup
	npcRepo := newMockNpcRepository()
	spawnRepo := newMockSpawnRepository()
	w := world.Instance()
	aiMgr := ai.NewTickManager()

	mgr := NewManager(npcRepo, spawnRepo, w, aiMgr, nil)
	spawner := NewSimpleSpawner(mgr)

	// Spawn test NPC
	ctx := context.Background()
	npc, err := spawner.SpawnTestNpc(ctx)
	if err != nil {
		t.Fatalf("SpawnTestNpc() error = %v", err)
	}

	if npc == nil {
		t.Fatal("SpawnTestNpc() returned nil NPC")
	}

	// Verify NPC properties
	if npc.Name() != "Wolf" {
		t.Errorf("NPC name = %q, want Wolf", npc.Name())
	}

	if npc.Level() != 5 {
		t.Errorf("NPC level = %d, want 5", npc.Level())
	}

	loc := npc.Location()
	if loc.X != 17000 || loc.Y != 170000 || loc.Z != -3500 {
		t.Errorf("NPC location = (%d, %d, %d), want (17000, 170000, -3500)", loc.X, loc.Y, loc.Z)
	}

	// Verify NPC is in world
	_, ok := w.GetObject(npc.ObjectID())
	if !ok {
		t.Error("NPC not found in world")
	}

	// Verify AI is registered
	_, err = aiMgr.GetController(npc.ObjectID())
	if err != nil {
		t.Errorf("AI controller not registered: %v", err)
	}

	// Cleanup
	w.RemoveObject(npc.ObjectID())
	aiMgr.Unregister(npc.ObjectID())
}

func TestSimpleSpawner_SpawnTestNpcAt(t *testing.T) {
	// Setup
	npcRepo := newMockNpcRepository()
	spawnRepo := newMockSpawnRepository()
	w := world.Instance()
	aiMgr := ai.NewTickManager()

	mgr := NewManager(npcRepo, spawnRepo, w, aiMgr, nil)
	spawner := NewSimpleSpawner(mgr)

	// Spawn test NPC at custom coordinates
	ctx := context.Background()
	testX, testY, testZ := int32(50000), int32(100000), int32(-2000)

	npc, err := spawner.SpawnTestNpcAt(ctx, testX, testY, testZ)
	if err != nil {
		t.Fatalf("SpawnTestNpcAt() error = %v", err)
	}

	if npc == nil {
		t.Fatal("SpawnTestNpcAt() returned nil NPC")
	}

	// Verify location
	loc := npc.Location()
	if loc.X != testX || loc.Y != testY || loc.Z != testZ {
		t.Errorf("NPC location = (%d, %d, %d), want (%d, %d, %d)",
			loc.X, loc.Y, loc.Z, testX, testY, testZ)
	}

	// Verify NPC is in world
	_, ok := w.GetObject(npc.ObjectID())
	if !ok {
		t.Error("NPC not found in world")
	}

	// Cleanup
	w.RemoveObject(npc.ObjectID())
	aiMgr.Unregister(npc.ObjectID())
}

func TestSimpleSpawner_MultipleSpawns(t *testing.T) {
	// Setup
	npcRepo := newMockNpcRepository()
	spawnRepo := newMockSpawnRepository()
	w := world.Instance()
	aiMgr := ai.NewTickManager()

	mgr := NewManager(npcRepo, spawnRepo, w, aiMgr, nil)
	spawner := NewSimpleSpawner(mgr)

	ctx := context.Background()

	// Spawn multiple NPCs
	npcs := make([]*model.Npc, 0, 5)
	for i := range 5 {
		npc, err := spawner.SpawnTestNpcAt(ctx, int32(17000+i*1000), 170000, -3500)
		if err != nil {
			t.Fatalf("SpawnTestNpcAt(%d) error = %v", i, err)
		}
		npcs = append(npcs, npc)
	}

	// Verify all NPCs have unique objectIDs
	seen := make(map[uint32]bool)
	for i, npc := range npcs {
		if seen[npc.ObjectID()] {
			t.Errorf("duplicate objectID %d at index %d", npc.ObjectID(), i)
		}
		seen[npc.ObjectID()] = true
	}

	// Cleanup
	for _, npc := range npcs {
		w.RemoveObject(npc.ObjectID())
		aiMgr.Unregister(npc.ObjectID())
	}
}
