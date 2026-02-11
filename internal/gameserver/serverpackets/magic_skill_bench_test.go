package serverpackets

import (
	"testing"
)

// BenchmarkMagicSkillUse_Write measures MagicSkillUse packet serialization.
// HIGH priority: broadcast to visible on every skill cast.
// Expected: <200ns, 1 alloc/op (writer buffer).
func BenchmarkMagicSkillUse_Write(b *testing.B) {
	pkt := NewMagicSkillUse(1, 2, 1001, 1, 1500, 2000, 10000, 20000, 1500)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := pkt.Write()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMagicSkillUse_Write_Batch simulates multiple skill casts in a busy area.
// 50 players casting skills simultaneously (raid boss scenario).
func BenchmarkMagicSkillUse_Write_Batch(b *testing.B) {
	pkts := make([]*MagicSkillUse, 50)
	for i := range 50 {
		pkts[i] = NewMagicSkillUse(
			int32(i+1),      // casterID
			int32(1000+i),   // targetID
			int32(1001+i%5), // skillID (5 different skills)
			1,               // skillLevel
			1500,            // hitTime
			2000,            // reuseDelay
			int32(10000+i*100), // x
			int32(20000+i*100), // y
			1500,               // z
		)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, pkt := range pkts {
			_, err := pkt.Write()
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkMagicSkillUse_Write_Parallel measures concurrent skill packet serialization.
func BenchmarkMagicSkillUse_Write_Parallel(b *testing.B) {
	pkt := NewMagicSkillUse(1, 2, 1001, 1, 1500, 2000, 10000, 20000, 1500)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := pkt.Write()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
