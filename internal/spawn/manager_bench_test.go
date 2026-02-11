package spawn

import (
	"context"
	"testing"

	"github.com/udisondev/la2go/internal/ai"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// mockNpcRepoBench для бенчмарков
type mockNpcRepoBench struct{}

func (m *mockNpcRepoBench) LoadTemplate(ctx context.Context, templateID int32) (*model.NpcTemplate, error) {
	return model.NewNpcTemplate(
		templateID,
		"TestNPC", "Title", // name, title
		80,   // level
		5000, // maxHP
		1000, // maxMP
		500,  // pAtk
		200,  // pDef
		300,  // mAtk
		150,  // mDef
		50,   // aggroRange
		100,  // moveSpeed
		50,   // atkSpeed
		60,   // respawnMin
		120,  // respawnMax
		0,    // baseExp
		0,    // baseSP
	), nil
}

// mockSpawnRepoBench для бенчмарков
type mockSpawnRepoBench struct {
	spawns []*model.Spawn
}

func (m *mockSpawnRepoBench) LoadAll(ctx context.Context) ([]*model.Spawn, error) {
	return m.spawns, nil
}

// BenchmarkSpawnManager_Count — atomic cache (O(1))
func BenchmarkSpawnManager_Count(b *testing.B) {
	// Setup manager with 1000 spawns
	world := world.Instance()
	aiManager := ai.NewTickManager()
	npcRepo := &mockNpcRepoBench{}

	spawns := make([]*model.Spawn, 1000)
	for i := range spawns {
		spawns[i] = model.NewSpawn(
			int64(i),
			1000+int32(i), // templateID
			150000+int32(i)*100, 150000, 0, // x, y, z
			0,     // heading
			1,     // maximumCount
			false, // randomSpawn
		)
	}

	spawnRepo := &mockSpawnRepoBench{spawns: spawns}
	mgr := NewManager(npcRepo, spawnRepo, world, aiManager, nil)

	// Load spawns
	if err := mgr.LoadSpawns(context.Background()); err != nil {
		b.Fatal(err)
	}

	b.Run("Atomic_O1", func(b *testing.B) {
		b.ReportAllocs()

		b.ResetTimer()
		for range b.N {
			_ = mgr.SpawnCount()
		}
	})
}

