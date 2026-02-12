package combat

import (
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/udisondev/la2go/internal/model"
)

func newTestNpcTemplate() *model.NpcTemplate {
	return model.NewNpcTemplate(
		1000, "TestMob", "Monster",
		10, 1000, 500,
		100, 50, 80, 40,
		300, 120, 253,
		30, 60, 0, 0,
	)
}

func newTestPlayer(t *testing.T, objectID uint32, x, y, z int32) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(objectID, int64(objectID), 1, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	p.SetLocation(model.NewLocation(x, y, z, 0))
	p.WorldObject.Data = p
	return p
}

func TestExecuteNpcAttack_DamageApplied(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		template := newTestNpcTemplate()
		npc := model.NewNpc(100001, 1000, template)
		npc.SetLocation(model.NewLocation(17000, 170000, -3500, 0))

		player := newTestPlayer(t, 0x10000001, 17050, 170050, -3500)
		playerObj := player.WorldObject
		playerObj.Data = player

		initialHP := player.CurrentHP()

		var broadcastCount atomic.Int32
		npcBroadcast := func(x, y int32, data []byte, size int) {
			broadcastCount.Add(1)
		}

		mgr := NewCombatManager(nil, npcBroadcast, nil)

		// Use NpcAttackUntilHit to guarantee a non-miss hit
		result, attempts := NpcAttackUntilHit(mgr, npc, playerObj, 20)
		if result.Miss {
			t.Fatalf("could not land a hit in %d attempts", attempts)
		}

		if broadcastCount.Load() == 0 {
			t.Error("expected at least one broadcast (attack packet)")
		}

		if result.Damage <= 0 {
			t.Errorf("expected positive damage, got %d", result.Damage)
		}

		finalHP := player.CurrentHP()
		if finalHP >= initialHP {
			t.Errorf("HP should decrease: initial=%d, final=%d", initialHP, finalHP)
		}
		t.Logf("HP: %d -> %d (delta: %d, attempts: %d)", initialHP, finalHP, initialHP-finalHP, attempts)
	})
}

func TestExecuteNpcAttack_DeadTargetSkipped(t *testing.T) {
	template := newTestNpcTemplate()
	npc := model.NewNpc(100002, 1000, template)
	npc.SetLocation(model.NewLocation(17000, 170000, -3500, 0))

	player := newTestPlayer(t,0x10000002, 17050, 170050, -3500)
	playerObj := player.WorldObject
	playerObj.Data = player

	// Kill the player first
	player.SetCurrentHP(0)

	var broadcastCount int
	npcBroadcast := func(x, y int32, data []byte, size int) {
		broadcastCount++
	}

	mgr := NewCombatManager(nil, npcBroadcast, nil)
	mgr.ExecuteNpcAttack(npc, playerObj)

	// Should not broadcast anything for dead target
	if broadcastCount != 0 {
		t.Errorf("broadcast count = %d, want 0 for dead target", broadcastCount)
	}
}

func TestExecuteNpcAttack_NonPlayerTarget(t *testing.T) {
	template := newTestNpcTemplate()
	npc := model.NewNpc(100003, 1000, template)
	npc.SetLocation(model.NewLocation(17000, 170000, -3500, 0))

	// Target is another NPC (not Player)
	targetTemplate := newTestNpcTemplate()
	targetNpc := model.NewNpc(100004, 1001, targetTemplate)
	targetObj := targetNpc.WorldObject

	var broadcastCount int
	npcBroadcast := func(x, y int32, data []byte, size int) {
		broadcastCount++
	}

	mgr := NewCombatManager(nil, npcBroadcast, nil)
	mgr.ExecuteNpcAttack(npc, targetObj) // Should log warning, not crash

	// No broadcast for non-player target
	if broadcastCount != 0 {
		t.Errorf("broadcast count = %d, want 0 for non-player target", broadcastCount)
	}
}

func TestCalcCritGeneric(t *testing.T) {
	// Run 100000 times for statistical stability (4% ± 1.5%)
	crits := 0
	const total = 100000
	for range total {
		if CalcCritGeneric() {
			crits++
		}
	}

	rate := float64(crits) / float64(total) * 100.0
	if rate < 2.5 || rate > 5.5 {
		t.Errorf("crit rate = %.1f%%, want ~4%% (range 2.5-5.5%%)", rate)
	}
}

func TestCalcHitMissGeneric(t *testing.T) {
	// Run 100000 times for statistical stability (20% ± 2.5%)
	misses := 0
	const total = 100000
	for range total {
		if CalcHitMissGeneric() {
			misses++
		}
	}

	rate := float64(misses) / float64(total) * 100.0
	if rate < 17.5 || rate > 22.5 {
		t.Errorf("miss rate = %.1f%%, want ~20%% (range 17.5-22.5%%)", rate)
	}
}
