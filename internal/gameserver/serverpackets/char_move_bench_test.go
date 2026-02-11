package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// BenchmarkCharMoveToLocation_Write measures CharMoveToLocation packet serialization.
// CRITICAL hot path: 10-50 times/sec per player (every movement → broadcast).
// Expected: <200ns, 1 alloc/op (writer buffer).
func BenchmarkCharMoveToLocation_Write(b *testing.B) {
	player, err := model.NewPlayer(1, 1, 1, "Test", 10, 0, 1)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.Location{X: 10000, Y: 20000, Z: 1500, Heading: 0})

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		pkt := NewCharMoveToLocation(player, 11000, 21000, 1500)
		_, err := pkt.Write()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCharMoveToLocation_Write_Batch simulates movement broadcast workload.
// 200 players moving simultaneously (typical busy area).
// Expected: <40µs total (200ns × 200 packets).
func BenchmarkCharMoveToLocation_Write_Batch(b *testing.B) {
	players := make([]*model.Player, 200)
	for i := range 200 {
		p, err := model.NewPlayer(uint32(i+1), int64(i+1), 1, "Player", 10, 0, 1)
		if err != nil {
			b.Fatalf("NewPlayer: %v", err)
		}
		p.SetLocation(model.Location{X: int32(10000 + i*100), Y: int32(20000 + i*100), Z: 1500, Heading: 0})
		players[i] = p
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, p := range players {
			pkt := NewCharMoveToLocation(p, p.Location().X+1000, p.Location().Y+1000, 1500)
			_, err := pkt.Write()
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkCharMoveToLocation_Write_Parallel measures concurrent serialization.
// Expected: ~200ns per op (no contention — Player data is read-only during Write).
func BenchmarkCharMoveToLocation_Write_Parallel(b *testing.B) {
	player, err := model.NewPlayer(1, 1, 1, "Test", 10, 0, 1)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.Location{X: 10000, Y: 20000, Z: 1500, Heading: 0})

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pkt := NewCharMoveToLocation(player, 11000, 21000, 1500)
			_, err := pkt.Write()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
