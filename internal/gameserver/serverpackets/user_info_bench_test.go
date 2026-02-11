package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// BenchmarkUserInfo_Write measures UserInfo packet serialization.
// HIGH priority: largest packet (~500B), sent on login + equipment change.
// Expected: <1µs, 1 alloc/op (writer buffer).
func BenchmarkUserInfo_Write(b *testing.B) {
	player, err := model.NewPlayer(1, 1, 1, "TestPlayer", 80, 0, 1)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.Location{X: 10000, Y: 20000, Z: 1500, Heading: 0})

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		pkt := NewUserInfo(player)
		_, err := pkt.Write()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUserInfo_Write_Parallel measures concurrent UserInfo serialization.
// Expected: ~1µs per op (no contention — Player data is read-only during Write).
func BenchmarkUserInfo_Write_Parallel(b *testing.B) {
	player, err := model.NewPlayer(1, 1, 1, "TestPlayer", 80, 0, 1)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.Location{X: 10000, Y: 20000, Z: 1500, Heading: 0})

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pkt := NewUserInfo(player)
			_, err := pkt.Write()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
