package ai

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// --- helpers ---

func benchMonster(objectID uint32, aggroRange int32) *model.Monster {
	template := model.NewNpcTemplate(
		1000, "BenchMob", "Monster",
		40, 5000, 2000,
		200, 100, 150, 80,
		aggroRange, 120, 253,
		30, 60, 500, 100,
	)
	m := model.NewMonster(objectID, 1000, template)
	m.SetLocation(model.NewLocation(17000, 170000, -3500, 0))
	return m
}

func benchPlayerObj(objectID uint32, x, y, z int32) (*model.Player, *model.WorldObject) {
	p, err := model.NewPlayer(objectID, int64(objectID), 1, "BenchPlayer", 40, 0, 0)
	if err != nil {
		panic(err)
	}
	p.SetLocation(model.NewLocation(x, y, z, 0))
	obj := model.NewWorldObject(objectID, "BenchPlayer", model.NewLocation(x, y, z, 0))
	obj.Data = p
	p.WorldObject.Data = p
	return p, obj
}

func noopAttack(*model.Monster, *model.WorldObject) {}

func emptyScan(int32, int32, func(*model.WorldObject) bool) {}

func emptyGetObject(uint32) (*model.WorldObject, bool) { return nil, false }

// makePlayerScan creates a ScanFunc that returns N player objects within aggro range.
func makePlayerScan(count int, baseX, baseY, baseZ int32) (ScanFunc, []*model.WorldObject) {
	objs := make([]*model.WorldObject, count)
	for i := range count {
		_, obj := benchPlayerObj(uint32(0x10000000+i), baseX+int32(i*10), baseY+int32(i*10), baseZ)
		objs[i] = obj
	}
	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {
		for _, obj := range objs {
			if !fn(obj) {
				return
			}
		}
	}
	return scanFunc, objs
}

func setupAI(aggroRange int32, scanFunc ScanFunc, getObjectFunc GetObjectFunc) *AttackableAI {
	monster := benchMonster(100001, aggroRange)
	ai := NewAttackableAI(monster, noopAttack, scanFunc, getObjectFunc)
	ai.Start()
	// Burn spawn immunity
	for range 11 {
		ai.Tick()
	}
	return ai
}

// --- Tick benchmarks ---

// BenchmarkAttackableAI_Tick_Idle benchmarks idle tick (NPC running, no players).
// Expected: ~50-100ns (atomic loads + switch + thinkActive early return).
func BenchmarkAttackableAI_Tick_Idle(b *testing.B) {
	b.ReportAllocs()
	monster := benchMonster(100001, 300)
	ai := NewAttackableAI(monster, noopAttack, emptyScan, emptyGetObject)
	ai.Start()
	// Burn spawn immunity
	for range 11 {
		ai.Tick()
	}

	b.ResetTimer()
	for range b.N {
		ai.Tick()
	}
}

// BenchmarkAttackableAI_Tick_SpawnImmunity benchmarks tick during spawn immunity.
// Expected: ~20-50ns (atomic load + increment + early return).
func BenchmarkAttackableAI_Tick_SpawnImmunity(b *testing.B) {
	b.ReportAllocs()
	monster := benchMonster(100002, 300)
	ai := NewAttackableAI(monster, noopAttack, emptyScan, emptyGetObject)
	ai.Start()

	b.ResetTimer()
	for range b.N {
		// Reset globalAggro each iteration to stay in immunity
		ai.globalAggro.Store(-10)
		ai.Tick()
	}
}

// BenchmarkAttackableAI_Tick_DeadNPC benchmarks tick on dead NPC (early return).
// Expected: ~10-20ns (2x atomic load).
func BenchmarkAttackableAI_Tick_DeadNPC(b *testing.B) {
	b.ReportAllocs()
	monster := benchMonster(100003, 300)
	monster.SetCurrentHP(0)
	ai := NewAttackableAI(monster, noopAttack, emptyScan, emptyGetObject)
	ai.Start()

	b.ResetTimer()
	for range b.N {
		ai.Tick()
	}
}

// --- thinkActive benchmarks (via Tick in ACTIVE state) ---

// BenchmarkAttackableAI_ThinkActive_NoPlayers benchmarks scanning empty region.
// Expected: ~50-100ns (scan callback with 0 objects).
func BenchmarkAttackableAI_ThinkActive_NoPlayers(b *testing.B) {
	b.ReportAllocs()
	ai := setupAI(300, emptyScan, emptyGetObject)

	b.ResetTimer()
	for range b.N {
		ai.Tick()
	}
}

// BenchmarkAttackableAI_ThinkActive_10Players benchmarks scanning 10 players (all in range).
// Expected: ~500ns-1us (10 iterations: ObjectID + type assert + IsDead + DistanceSquared).
func BenchmarkAttackableAI_ThinkActive_10Players(b *testing.B) {
	b.ReportAllocs()
	scanFunc, _ := makePlayerScan(10, 17010, 170010, -3500)
	monster := benchMonster(100004, 300)
	ai := NewAttackableAI(monster, noopAttack, scanFunc, emptyGetObject)
	ai.Start()
	for range 11 {
		ai.Tick()
	}

	b.ResetTimer()
	for range b.N {
		// Clear aggro list so thinkActive runs fresh
		b.StopTimer()
		monster.AggroList().Clear()
		monster.ClearTarget()
		ai.SetIntention(model.IntentionActive)
		b.StartTimer()

		ai.Tick()
	}
}

// BenchmarkAttackableAI_ThinkActive_50Players benchmarks scanning 50 players in range.
// Expected: ~2-5us (50 iterations with distance checks).
func BenchmarkAttackableAI_ThinkActive_50Players(b *testing.B) {
	b.ReportAllocs()
	scanFunc, _ := makePlayerScan(50, 17010, 170010, -3500)
	monster := benchMonster(100005, 5000) // large aggro range to include all
	ai := NewAttackableAI(monster, noopAttack, scanFunc, emptyGetObject)
	ai.Start()
	for range 11 {
		ai.Tick()
	}

	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		monster.AggroList().Clear()
		monster.ClearTarget()
		ai.SetIntention(model.IntentionActive)
		b.StartTimer()

		ai.Tick()
	}
}

// --- thinkAttack benchmarks (via Tick in ATTACK state) ---

// BenchmarkAttackableAI_ThinkAttack_ValidTarget benchmarks attack with valid target.
// Expected: ~100-500ns (GetMostHated + GetObject + distance check + attackFunc).
func BenchmarkAttackableAI_ThinkAttack_ValidTarget(b *testing.B) {
	b.ReportAllocs()
	_, playerObj := benchPlayerObj(0x10000001, 17050, 170050, -3500)

	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		if objectID == playerObj.ObjectID() {
			return playerObj, true
		}
		return nil, false
	}

	monster := benchMonster(100006, 300)
	ai := NewAttackableAI(monster, noopAttack, emptyScan, getObjectFunc)
	ai.Start()
	ai.globalAggro.Store(0)
	ai.NotifyDamage(playerObj.ObjectID(), 100)
	// Now AI is in ATTACK state with target

	b.ResetTimer()
	for range b.N {
		ai.Tick()
	}
}

// BenchmarkAttackableAI_ThinkAttack_TargetGone benchmarks attack with missing target.
// Expected: ~50-100ns (GetMostHated + GetObject miss + return to ACTIVE).
func BenchmarkAttackableAI_ThinkAttack_TargetGone(b *testing.B) {
	b.ReportAllocs()
	_, playerObj := benchPlayerObj(0x10000002, 17050, 170050, -3500)

	monster := benchMonster(100007, 300)
	ai := NewAttackableAI(monster, noopAttack, emptyScan, emptyGetObject)
	ai.Start()
	ai.globalAggro.Store(0)

	b.ResetTimer()
	for range b.N {
		// Re-add hate and set attack state
		b.StopTimer()
		monster.AggroList().AddHate(playerObj.ObjectID(), 100)
		monster.SetTarget(playerObj.ObjectID())
		ai.SetIntention(model.IntentionAttack)
		b.StartTimer()

		ai.Tick()
	}
}

// BenchmarkAttackableAI_ThinkAttack_OutOfRange benchmarks target out of attack range.
// Expected: ~100-300ns (GetMostHated + GetObject + distance check > 100^2).
func BenchmarkAttackableAI_ThinkAttack_OutOfRange(b *testing.B) {
	b.ReportAllocs()
	_, playerObj := benchPlayerObj(0x10000003, 20000, 170000, -3500) // far away

	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		if objectID == playerObj.ObjectID() {
			return playerObj, true
		}
		return nil, false
	}

	monster := benchMonster(100008, 300)
	ai := NewAttackableAI(monster, noopAttack, emptyScan, getObjectFunc)
	ai.Start()
	ai.globalAggro.Store(0)
	ai.NotifyDamage(playerObj.ObjectID(), 100)

	b.ResetTimer()
	for range b.N {
		ai.Tick()
	}
}

// --- NotifyDamage benchmarks ---

// BenchmarkAttackableAI_NotifyDamage_FirstHit benchmarks first damage notification.
// Expected: ~50-100ns (CalcHateValue + AggroList.AddHate + intention switch).
func BenchmarkAttackableAI_NotifyDamage_FirstHit(b *testing.B) {
	b.ReportAllocs()
	monster := benchMonster(100009, 300)
	ai := NewAttackableAI(monster, noopAttack, emptyScan, emptyGetObject)
	ai.Start()
	ai.globalAggro.Store(0)

	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		monster.AggroList().Clear()
		monster.ClearTarget()
		ai.SetIntention(model.IntentionActive)
		b.StartTimer()

		ai.NotifyDamage(0x10000001, 100)
	}
}

// BenchmarkAttackableAI_NotifyDamage_SpawnImmunity benchmarks damage during immunity.
// Expected: ~50-100ns (cancel immunity + CalcHateValue + AddHate).
func BenchmarkAttackableAI_NotifyDamage_SpawnImmunity(b *testing.B) {
	b.ReportAllocs()
	monster := benchMonster(100010, 300)
	ai := NewAttackableAI(monster, noopAttack, emptyScan, emptyGetObject)
	ai.Start()

	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		ai.globalAggro.Store(-10) // Reset immunity
		monster.AggroList().Clear()
		monster.ClearTarget()
		ai.SetIntention(model.IntentionActive)
		b.StartTimer()

		ai.NotifyDamage(0x10000001, 100)
	}
}

// BenchmarkAttackableAI_NotifyDamage_AlreadyAttacking benchmarks damage while attacking.
// Expected: ~30-50ns (fast path: no intention switch needed).
func BenchmarkAttackableAI_NotifyDamage_AlreadyAttacking(b *testing.B) {
	b.ReportAllocs()
	monster := benchMonster(100011, 300)
	ai := NewAttackableAI(monster, noopAttack, emptyScan, emptyGetObject)
	ai.Start()
	ai.globalAggro.Store(0)
	ai.NotifyDamage(0x10000001, 100) // Set attack state

	b.ResetTimer()
	for range b.N {
		ai.NotifyDamage(0x10000002, 50)
	}
}
