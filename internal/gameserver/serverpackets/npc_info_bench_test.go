package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// BenchmarkNpcInfo_Write measures NpcInfo packet serialization performance.
// Phase 4.18 Sprint 3: Hot path benchmark for sendVisibleObjectsInfo().
//
// NpcInfo sent for each visible NPC during EnterWorld (up to ~200 NPCs per player).
// Expected: <1µs per packet (similar to CharInfo).
func BenchmarkNpcInfo_Write(b *testing.B) {
	// Create NPC template (realistic stats for level 50 guard)
	template := model.NewNpcTemplate(
		30001,      // templateID
		"Guard",    // name
		"Warrior",  // title
		50,         // level
		5000,       // maxHP
		1000,       // maxMP
		100,        // pAtk
		80,         // pDef
		50,         // mAtk
		40,         // mDef
		120,        // aggroRange
		253,        // moveSpeed
		30,         // atkSpeed
		60,         // respawnMin
		120,        // respawnMax
		0,          // baseExp
		0,          // baseSP
	)

	// Create NPC instance
	npc := model.NewNpc(
		0x20000001, // objectID (NPC range)
		30001,      // templateID
		template,   // template
	)
	npc.SetLocation(model.NewLocation(10000, 20000, 1500, 0))

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		packet := NewNpcInfo(npc)
		_, err := packet.Write()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNpcInfo_Write_Batch simulates sendVisibleObjectsInfo() workload.
// Creates 200 NpcInfo packets (typical for EnterWorld with many visible NPCs).
// Expected: <200µs total (1µs × 200 packets).
func BenchmarkNpcInfo_Write_Batch(b *testing.B) {
	// Create shared template
	template := model.NewNpcTemplate(
		30001, "Guard", "Warrior", 50, 5000, 1000,
		100, 80, 50, 40, 120, 253, 30, 60, 120, 0, 0,
	)

	// Create 200 NPC instances
	npcs := make([]*model.Npc, 200)
	for i := range 200 {
		npcs[i] = model.NewNpc(
			uint32(0x20000001+i),
			30001,
			template,
		)
		npcs[i].SetLocation(model.NewLocation(
			int32(10000+i*100),
			int32(20000+i*100),
			1500,
			0,
		))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		for _, npc := range npcs {
			packet := NewNpcInfo(npc)
			_, err := packet.Write()
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
