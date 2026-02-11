package data

import (
	"testing"
)

// --- Skill Loading benchmarks ---

// BenchmarkLoadSkills benchmarks full skill XML loading (1000+ skills).
// Expected: ~50-200ms (XML parsing + struct allocation for 2694 skill IDs).
func BenchmarkLoadSkills(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		// Reset table to force full reload
		SkillTable = nil
		if err := LoadSkills(); err != nil {
			b.Fatalf("LoadSkills: %v", err)
		}
	}
}

// --- Item Loading benchmarks ---

// BenchmarkLoadItemTemplates benchmarks full item XML loading (10K+ items).
// Expected: ~100-500ms (XML parsing + struct allocation for 9208 items).
func BenchmarkLoadItemTemplates(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		// Reset table to force full reload
		ItemTable = nil
		if err := LoadItemTemplates(); err != nil {
			b.Fatalf("LoadItemTemplates: %v", err)
		}
	}
}

// --- NPC Loading benchmarks ---

// BenchmarkLoadNpcTemplates benchmarks full NPC XML loading (5K+ NPCs).
// Expected: ~100-300ms (XML parsing + struct allocation for 6519 NPCs).
func BenchmarkLoadNpcTemplates(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		// Reset table to force full reload
		NpcTable = nil
		if err := LoadNpcTemplates(); err != nil {
			b.Fatalf("LoadNpcTemplates: %v", err)
		}
	}
}

// --- Lookup benchmarks (after loading) ---

// BenchmarkGetItemDef_Hit benchmarks item definition lookup (existing item).
// Expected: ~10-30ns (map[int32] lookup).
func BenchmarkGetItemDef_Hit(b *testing.B) {
	b.ReportAllocs()
	if ItemTable == nil {
		if err := LoadItemTemplates(); err != nil {
			b.Fatalf("LoadItemTemplates: %v", err)
		}
	}

	var itemID int32
	for id := range ItemTable {
		itemID = id
		break
	}
	if itemID == 0 {
		b.Skip("no items loaded")
	}

	b.ResetTimer()
	for range b.N {
		_ = GetItemDef(itemID)
	}
}

// BenchmarkGetItemDef_Miss benchmarks item definition lookup (non-existing item).
// Expected: ~5-10ns (map miss).
func BenchmarkGetItemDef_Miss(b *testing.B) {
	b.ReportAllocs()
	if ItemTable == nil {
		if err := LoadItemTemplates(); err != nil {
			b.Fatalf("LoadItemTemplates: %v", err)
		}
	}

	b.ResetTimer()
	for range b.N {
		_ = GetItemDef(999999)
	}
}

// BenchmarkGetNpcDef_Hit benchmarks NPC definition lookup (existing NPC).
// Expected: ~10-30ns (map[int32] lookup).
func BenchmarkGetNpcDef_Hit(b *testing.B) {
	b.ReportAllocs()
	if NpcTable == nil {
		if err := LoadNpcTemplates(); err != nil {
			b.Fatalf("LoadNpcTemplates: %v", err)
		}
	}

	var npcID int32
	for id := range NpcTable {
		npcID = id
		break
	}
	if npcID == 0 {
		b.Skip("no NPCs loaded")
	}

	b.ResetTimer()
	for range b.N {
		_ = GetNpcDef(npcID)
	}
}

// BenchmarkGetNpcDef_Miss benchmarks NPC definition lookup (non-existing NPC).
// Expected: ~5-10ns (map miss).
func BenchmarkGetNpcDef_Miss(b *testing.B) {
	b.ReportAllocs()
	if NpcTable == nil {
		if err := LoadNpcTemplates(); err != nil {
			b.Fatalf("LoadNpcTemplates: %v", err)
		}
	}

	b.ResetTimer()
	for range b.N {
		_ = GetNpcDef(999999)
	}
}

// BenchmarkGetItemDef_Concurrent benchmarks concurrent item lookups.
// Expected: measures CPU cache contention on map reads.
func BenchmarkGetItemDef_Concurrent(b *testing.B) {
	b.ReportAllocs()
	if ItemTable == nil {
		if err := LoadItemTemplates(); err != nil {
			b.Fatalf("LoadItemTemplates: %v", err)
		}
	}

	var itemID int32
	for id := range ItemTable {
		itemID = id
		break
	}
	if itemID == 0 {
		b.Skip("no items loaded")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = GetItemDef(itemID)
		}
	})
}

// BenchmarkGetNpcDef_Concurrent benchmarks concurrent NPC lookups.
// Expected: measures CPU cache contention on map reads.
func BenchmarkGetNpcDef_Concurrent(b *testing.B) {
	b.ReportAllocs()
	if NpcTable == nil {
		if err := LoadNpcTemplates(); err != nil {
			b.Fatalf("LoadNpcTemplates: %v", err)
		}
	}

	var npcID int32
	for id := range NpcTable {
		npcID = id
		break
	}
	if npcID == 0 {
		b.Skip("no NPCs loaded")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = GetNpcDef(npcID)
		}
	})
}
