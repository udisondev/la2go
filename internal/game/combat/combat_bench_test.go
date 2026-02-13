package combat

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// --- helpers ---

func benchPlayer(objectID uint32, x, y, z int32) *model.Player {
	p, err := model.NewPlayer(objectID, int64(objectID), 1, "BenchPlayer", 40, 0, 0)
	if err != nil {
		panic(err)
	}
	p.SetLocation(model.NewLocation(x, y, z, 0))
	p.WorldObject.Data = p
	return p
}

func benchNpcTemplate() *model.NpcTemplate {
	return model.NewNpcTemplate(
		1000, "BenchMob", "Monster",
		40, 5000, 2000,
		200, 100, 150, 80,
		300, 120, 253,
		30, 60, 500, 100,
	)
}

func benchNpc(objectID uint32, x, y, z int32) *model.Npc {
	tmpl := benchNpcTemplate()
	npc := model.NewNpc(objectID, 1000, tmpl)
	npc.SetLocation(model.NewLocation(x, y, z, 0))
	return npc
}

func benchMonster(objectID uint32, x, y, z int32) *model.Monster {
	tmpl := benchNpcTemplate()
	m := model.NewMonster(objectID, 1000, tmpl)
	m.SetLocation(model.NewLocation(x, y, z, 0))
	m.SetAggressive(true)
	return m
}

func benchWorldObject(objectID uint32, x, y, z int32, data any) *model.WorldObject {
	obj := model.NewWorldObject(objectID, "Target", model.NewLocation(x, y, z, 0))
	obj.Data = data
	return obj
}

// --- CalcCritGeneric / CalcHitMissGeneric benchmarks ---

// BenchmarkCalcCritGeneric benchmarks crit calculation (rand.Intn(1000)).
// Expected: ~10-20ns (single RNG call).
func BenchmarkCalcCritGeneric(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = CalcCritGeneric()
	}
}

// BenchmarkCalcHitMissGeneric benchmarks hit/miss calculation (rand.Intn(1000)).
// Expected: ~10-20ns (single RNG call).
func BenchmarkCalcHitMissGeneric(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = CalcHitMissGeneric()
	}
}

// --- CalcCrit / CalcHitMiss with Player/Character ---

// BenchmarkCalcCrit benchmarks crit calculation with player/target context.
// Expected: ~10-20ns (delegates to CalcCritGeneric).
func BenchmarkCalcCrit(b *testing.B) {
	b.ReportAllocs()
	attacker := benchPlayer(1, 0, 0, 0)
	target := benchPlayer(2, 50, 0, 0)

	b.ResetTimer()
	for range b.N {
		_ = CalcCrit(attacker, target.Character)
	}
}

// BenchmarkCalcHitMiss benchmarks hit/miss with player/target context.
// Expected: ~10-20ns (delegates to CalcHitMissGeneric).
func BenchmarkCalcHitMiss(b *testing.B) {
	b.ReportAllocs()
	attacker := benchPlayer(1, 0, 0, 0)
	target := benchPlayer(2, 50, 0, 0)

	b.ResetTimer()
	for range b.N {
		_ = CalcHitMiss(attacker, target.Character)
	}
}

// --- Damage Formula benchmarks ---

// BenchmarkCalcPhysicalDamage_NoCrit benchmarks damage formula without crit.
// Expected: ~50-100ns (float64 arithmetic + RNG).
func BenchmarkCalcPhysicalDamage_NoCrit(b *testing.B) {
	b.ReportAllocs()
	attacker := benchPlayer(1, 0, 0, 0)
	target := benchPlayer(2, 50, 0, 0)
	targetPDef := target.GetPDef()

	b.ResetTimer()
	for range b.N {
		_ = CalcPhysicalDamage(attacker, target.Character, false, false, ShieldDefFailed, targetPDef)
	}
}

// BenchmarkCalcPhysicalDamage_WithCrit benchmarks damage formula with crit multiplier.
// Expected: ~50-100ns (same as NoCrit + one float64 multiplication).
func BenchmarkCalcPhysicalDamage_WithCrit(b *testing.B) {
	b.ReportAllocs()
	attacker := benchPlayer(1, 0, 0, 0)
	target := benchPlayer(2, 50, 0, 0)
	targetPDef := target.GetPDef()

	b.ResetTimer()
	for range b.N {
		_ = CalcPhysicalDamage(attacker, target.Character, true, false, ShieldDefFailed, targetPDef)
	}
}

// BenchmarkGetRandomDamageMultiplier benchmarks damage variance calculation.
// Expected: ~10-20ns (sqrt + rand.Intn).
func BenchmarkGetRandomDamageMultiplier(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = getRandomDamageMultiplier(40)
	}
}

// --- Type assertion benchmarks ---

// BenchmarkTypeAssertion_Player benchmarks Player type assertion from WorldObject.Data.
// Expected: ~5-10ns (interface cast).
func BenchmarkTypeAssertion_Player(b *testing.B) {
	b.ReportAllocs()
	p := benchPlayer(1, 0, 0, 0)
	obj := p.WorldObject

	b.ResetTimer()
	for range b.N {
		if _, ok := obj.Data.(*model.Player); !ok {
			b.Fatal("type assertion failed")
		}
	}
}

// BenchmarkTypeAssertion_Monster benchmarks Monster type assertion from WorldObject.Data.
// Expected: ~5-10ns (interface cast).
func BenchmarkTypeAssertion_Monster(b *testing.B) {
	b.ReportAllocs()
	m := benchMonster(100, 0, 0, 0)
	obj := m.WorldObject
	obj.Data = m

	b.ResetTimer()
	for range b.N {
		if _, ok := obj.Data.(*model.Monster); !ok {
			b.Fatal("type assertion failed")
		}
	}
}

// BenchmarkTypeAssertion_Npc benchmarks Npc type assertion from WorldObject.Data.
// Expected: ~5-10ns (interface cast).
func BenchmarkTypeAssertion_Npc(b *testing.B) {
	b.ReportAllocs()
	n := benchNpc(100, 0, 0, 0)
	obj := n.WorldObject
	obj.Data = n

	b.ResetTimer()
	for range b.N {
		if _, ok := obj.Data.(*model.Npc); !ok {
			b.Fatal("type assertion failed")
		}
	}
}

// BenchmarkTypeAssertion_ThreeWay benchmarks the 3-way type assertion chain
// used in ExecuteAttack (Player → Monster → Npc).
// Expected: ~15-30ns (3 interface casts).
func BenchmarkTypeAssertion_ThreeWay(b *testing.B) {
	b.ReportAllocs()
	m := benchMonster(100, 0, 0, 0)
	obj := m.WorldObject
	obj.Data = m

	b.ResetTimer()
	for range b.N {
		if _, ok := obj.Data.(*model.Player); ok {
			continue
		}
		if _, ok := obj.Data.(*model.Monster); ok {
			continue
		}
		if _, ok := obj.Data.(*model.Npc); ok {
			continue
		}
	}
}

// --- Validation benchmarks ---

// BenchmarkValidateAttack_Valid benchmarks successful attack validation.
// Expected: ~30-50ns (nil check + IsDead + range check).
func BenchmarkValidateAttack_Valid(b *testing.B) {
	b.ReportAllocs()
	attacker := benchPlayer(1, 0, 0, 0)
	target := benchWorldObject(2, 50, 0, 0, benchPlayer(2, 50, 0, 0))

	b.ResetTimer()
	for range b.N {
		_ = ValidateAttack(attacker, target)
	}
}

// BenchmarkIsInAttackRange_InRange benchmarks range check (in range).
// Expected: ~10-20ns (2x Location() + DistanceSquared).
func BenchmarkIsInAttackRange_InRange(b *testing.B) {
	b.ReportAllocs()
	attacker := benchPlayer(1, 0, 0, 0)
	target := benchWorldObject(2, 50, 50, 0, nil)

	b.ResetTimer()
	for range b.N {
		_ = IsInAttackRange(attacker, target)
	}
}

// BenchmarkIsInAttackRange_OutOfRange benchmarks range check (out of range).
// Expected: ~10-20ns (same calc, different result).
func BenchmarkIsInAttackRange_OutOfRange(b *testing.B) {
	b.ReportAllocs()
	attacker := benchPlayer(1, 0, 0, 0)
	target := benchWorldObject(2, 5000, 5000, 0, nil)

	b.ResetTimer()
	for range b.N {
		_ = IsInAttackRange(attacker, target)
	}
}
