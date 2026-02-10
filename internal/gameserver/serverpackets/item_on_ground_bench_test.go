package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// BenchmarkItemOnGround_Write measures ItemOnGround packet serialization performance.
// Phase 4.18 Sprint 3: Hot path benchmark for sendVisibleObjectsInfo().
//
// ItemOnGround sent for each visible dropped item during EnterWorld (up to ~50 items per player).
// Expected: <500ns per packet (smaller than CharInfo/NpcInfo).
func BenchmarkItemOnGround_Write(b *testing.B) {
	// Create dropped item (Adena - most common drop)
	item, err := model.NewItem(
		0,    // ownerID (0 = dropped)
		57,   // itemType (Adena)
		1000, // count
	)
	if err != nil {
		b.Fatal(err)
	}

	droppedItem := model.NewDroppedItem(
		0x30000001,                               // objectID (Item range)
		item,                                     // item
		model.NewLocation(10000, 20000, 1500, 0), // location
		0, // dropperID (0 = monster drop)
	)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		packet := NewItemOnGround(droppedItem)
		_, err := packet.Write()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkItemOnGround_Write_Batch simulates sendVisibleObjectsInfo() workload.
// Creates 50 ItemOnGround packets (typical for EnterWorld with visible drops).
// Expected: <25µs total (500ns × 50 packets).
func BenchmarkItemOnGround_Write_Batch(b *testing.B) {
	// Create 50 dropped items
	droppedItems := make([]*model.DroppedItem, 50)
	for i := range 50 {
		item, err := model.NewItem(
			0,
			57+int32(i%10), // Variety of items
			1000,
		)
		if err != nil {
			b.Fatal(err)
		}

		droppedItems[i] = model.NewDroppedItem(
			uint32(0x30000001+i),
			item,
			model.NewLocation(
				int32(10000+i*100),
				int32(20000+i*100),
				1500,
				0,
			),
			0,
		)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		for _, droppedItem := range droppedItems {
			packet := NewItemOnGround(droppedItem)
			_, err := packet.Write()
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
