package world

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// BenchmarkWorld_GetNpc measures sync.Map O(1) lookup for NPCs.
// Phase 4.11 Priority 1: Verify atomic load performance.
// Baseline expectation: ~10ns (sync.Map atomic load).
func BenchmarkWorld_GetNpc(b *testing.B) {
	w := Instance()

	// Add NPC to world
	npcID := uint32(0x20000001)
	template := model.NewNpcTemplate(
		18342,        // templateID
		"TestNpc",    // name
		"Monster",    // title
		10,           // level
		1000, 500,    // maxHP, maxMP
		100, 50,      // pAtk, pDef
		80, 40,       // mAtk, mDef
		300,          // aggroRange
		120,          // moveSpeed
		253,          // atkSpeed
		30, 60,       // respawnMin, respawnMax
	)
	npc := model.NewNpc(npcID, int32(18342), template)

	w.AddNpc(npc)

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_, _ = w.GetNpc(npcID)
	}
}

// BenchmarkWorld_GetNpc_Miss measures lookup when NPC not found.
// Expected: ~10ns (same as hit — sync.Map.Load is O(1)).
func BenchmarkWorld_GetNpc_Miss(b *testing.B) {
	w := Instance()

	// Add some NPCs (but not the one we're looking for)
	for i := range 100 {
		npcID := uint32(0x20000000 + i)
		template := model.NewNpcTemplate(
			int32(18342+i), "TestNpc", "Monster",
			10, 1000, 500, 100, 50, 80, 40, 300, 120, 253, 30, 60,
		)
		npc := model.NewNpc(npcID, int32(18342+i), template)
		w.AddNpc(npc)
	}

	// Lookup non-existent NPC
	nonExistentID := uint32(0x2FFFFFFF)

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_, _ = w.GetNpc(nonExistentID)
	}
}

// BenchmarkWorld_GetNpc_Parallel measures concurrent reads (should be fast — no contention).
// Expected: ~10ns per op (sync.Map uses atomic.Value for lock-free reads).
func BenchmarkWorld_GetNpc_Parallel(b *testing.B) {
	w := Instance()

	// Add NPC
	npcID := uint32(0x20000001)
	template := model.NewNpcTemplate(
		18342, "TestNpc", "Monster",
		10, 1000, 500, 100, 50, 80, 40, 300, 120, 253, 30, 60,
	)
	npc := model.NewNpc(npcID, int32(18342), template)

	w.AddNpc(npc)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = w.GetNpc(npcID)
		}
	})
}
