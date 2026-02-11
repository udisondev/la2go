package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// BenchmarkStatusUpdate_Write measures StatusUpdate serialization with default 6 attributes.
// CRITICAL hot path: 5-20 times/sec per player in combat (HP/MP/CP changes).
// Expected: <300ns, 1 alloc/op (writer buffer).
func BenchmarkStatusUpdate_Write(b *testing.B) {
	player, err := model.NewPlayer(1, 1, 1, "Test", 50, 0, 1)
	if err != nil {
		b.Fatalf("NewPlayer: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		pkt := NewStatusUpdate(player)
		_, err := pkt.Write()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStatusUpdate_Write_MinimalAttrs measures StatusUpdate with 1 attribute (HP only).
// Best case — single stat update (e.g., HP regen tick).
func BenchmarkStatusUpdate_Write_MinimalAttrs(b *testing.B) {
	pkt := &StatusUpdate{
		ObjectID: 1,
		Attributes: []StatusAttribute{
			{ID: AttrCurrentHP, Value: 500},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := pkt.Write()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStatusUpdate_Write_AllAttrs measures StatusUpdate with all 19 attributes.
// Worst case — full stats resync (rare, e.g., buff/debuff apply).
func BenchmarkStatusUpdate_Write_AllAttrs(b *testing.B) {
	pkt := &StatusUpdate{
		ObjectID: 1,
		Attributes: []StatusAttribute{
			{ID: AttrLevel, Value: 80},
			{ID: AttrExp, Value: 100000000},
			{ID: AttrSTR, Value: 88},
			{ID: AttrDEX, Value: 59},
			{ID: AttrCON, Value: 73},
			{ID: AttrINT, Value: 25},
			{ID: AttrWIT, Value: 16},
			{ID: AttrMEN, Value: 27},
			{ID: AttrCurrentHP, Value: 5000},
			{ID: AttrMaxHP, Value: 6000},
			{ID: AttrCurrentMP, Value: 1000},
			{ID: AttrMaxMP, Value: 1500},
			{ID: AttrSP, Value: 99999},
			{ID: AttrCurrentCP, Value: 800},
			{ID: AttrMaxCP, Value: 1000},
			{ID: AttrKarma, Value: 0},
			{ID: AttrPvPFlag, Value: 0},
			{ID: AttrLoad, Value: 5000},
			{ID: AttrMaxLoad, Value: 80000},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := pkt.Write()
		if err != nil {
			b.Fatal(err)
		}
	}
}
