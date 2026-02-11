package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// BenchmarkAttack_Write_SingleHit measures Attack packet serialization with single hit.
// HIGH priority hot path: 1-3 times/sec in combat, broadcast to visible.
// Expected: <300ns, 1 alloc/op (writer buffer).
func BenchmarkAttack_Write_SingleHit(b *testing.B) {
	player, err := model.NewPlayer(1, 1, 1, "Attacker", 50, 0, 1)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.Location{X: 10000, Y: 20000, Z: 1500, Heading: 0})

	targetObj := model.NewWorldObject(2, "Target", model.Location{X: 10100, Y: 20100, Z: 1500, Heading: 0})

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		atk := NewAttack(player, targetObj)
		atk.AddHit(2, 150, false, false)
		_, err := atk.Write()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAttack_Write_CritHit measures Attack packet with critical hit flag.
func BenchmarkAttack_Write_CritHit(b *testing.B) {
	player, err := model.NewPlayer(1, 1, 1, "Attacker", 50, 0, 1)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.Location{X: 10000, Y: 20000, Z: 1500, Heading: 0})

	targetObj := model.NewWorldObject(2, "Target", model.Location{X: 10100, Y: 20100, Z: 1500, Heading: 0})

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		atk := NewAttack(player, targetObj)
		atk.AddHit(2, 300, false, true)
		_, err := atk.Write()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAttack_Write_DualHit measures Attack packet with 2 hits (dual weapon).
// Worst case for Phase 5.4+ dual weapon combat.
func BenchmarkAttack_Write_DualHit(b *testing.B) {
	player, err := model.NewPlayer(1, 1, 1, "Attacker", 50, 0, 1)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.Location{X: 10000, Y: 20000, Z: 1500, Heading: 0})

	targetObj := model.NewWorldObject(2, "Target", model.Location{X: 10100, Y: 20100, Z: 1500, Heading: 0})

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		atk := NewAttack(player, targetObj)
		atk.AddHit(2, 150, false, false)
		atk.AddHit(2, 120, false, false)
		_, err := atk.Write()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAttack_Write_NpcAttack measures NPC attack packet serialization.
// Phase 5.7: NPC aggro auto-attack hot path.
func BenchmarkAttack_Write_NpcAttack(b *testing.B) {
	attackerLoc := model.Location{X: 10000, Y: 20000, Z: 1500, Heading: 0}
	targetObj := model.NewWorldObject(100, "Player", model.Location{X: 10100, Y: 20100, Z: 1500, Heading: 0})

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		atk := NewNpcAttack(0x20000001, attackerLoc, targetObj)
		atk.AddHit(100, 80, false, false)
		_, err := atk.Write()
		if err != nil {
			b.Fatal(err)
		}
	}
}
