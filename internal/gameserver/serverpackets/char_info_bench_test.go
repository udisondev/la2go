package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// BenchmarkServerPackets_CharInfo_Write measures CharInfo serialization overhead.
// Phase 4.11 Priority 1: Called for EACH visible player in sendVisibleObjectsInfo.
// Baseline expectation: <500ns per packet (~512 bytes).
func BenchmarkServerPackets_CharInfo_Write(b *testing.B) {
	// Create realistic player for benchmark
	player, err := model.NewPlayer(
		int64(0x10000001), // characterID
		int64(1000),        // accountID
		"BenchmarkPlayer",
		int32(76),          // level
		0,                  // raceID (Human)
		88,                 // classID (Paladin example)
	)
	if err != nil {
		b.Fatalf("failed to create player: %v", err)
	}

	// Set location
	player.SetLocation(model.NewLocation(100000, 200000, -3000, 0))

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		pkt := NewCharInfo(player)
		data, err := pkt.Write()
		if err != nil {
			b.Fatalf("failed to write CharInfo: %v", err)
		}
		_ = data // Prevent compiler optimization
	}
}

// BenchmarkServerPackets_CharInfo_Write_Parallel measures concurrent serialization.
// Expected: ~500ns per op (no contention — Player data is immutable during Write).
func BenchmarkServerPackets_CharInfo_Write_Parallel(b *testing.B) {
	player, err := model.NewPlayer(
		int64(0x10000001),
		int64(1000),
		"BenchmarkPlayer",
		int32(76),
		0,
		88,
	)
	if err != nil {
		b.Fatalf("failed to create player: %v", err)
	}

	player.SetLocation(model.NewLocation(100000, 200000, -3000, 0))

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pkt := NewCharInfo(player)
			data, err := pkt.Write()
			if err != nil {
				b.Fatalf("failed to write CharInfo: %v", err)
			}
			_ = data
		}
	})
}
