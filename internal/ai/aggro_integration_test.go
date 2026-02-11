package ai

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// Integration tests for NPC aggro system.
// Setup: Real World grid + Monster + Player(s).
// Tests full flow: scan → detect → attack → target death → return to idle.
//
// Phase 5.7: NPC Aggro & Basic AI.

// setupAggroWorld creates a Monster in the real World grid.
func setupAggroWorld(t *testing.T, monsterID uint32, aggroRange int32, x, y, z int32) (*model.Monster, *world.World) {
	t.Helper()

	template := model.NewNpcTemplate(
		1000, "AggroMob", "Monster",
		10, 1000, 500,
		100, 50, 80, 40,
		aggroRange, 120, 253,
		30, 60, 0, 0,
	)

	monster := model.NewMonster(monsterID, 1000, template)
	monster.SetLocation(model.NewLocation(x, y, z, 0))

	w := world.Instance()
	if err := w.AddNpc(monster.Npc); err != nil {
		t.Fatalf("failed to add monster to world: %v", err)
	}

	t.Cleanup(func() {
		w.RemoveObject(monsterID)
	})

	return monster, w
}

// setupPlayer creates a Player WorldObject in the real World grid.
func setupPlayer(t *testing.T, objectID uint32, x, y, z int32) (*model.Player, *model.WorldObject) {
	t.Helper()

	player, err := model.NewPlayer(objectID, int64(objectID), 1, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("failed to create player: %v", err)
	}
	player.SetLocation(model.NewLocation(x, y, z, 0))
	player.WorldObject.Data = player

	w := world.Instance()
	if err := w.AddObject(player.WorldObject); err != nil {
		t.Fatalf("failed to add player to world: %v", err)
	}

	t.Cleanup(func() {
		w.RemoveObject(objectID)
	})

	return player, player.WorldObject
}

// TestAggroIntegration_PlayerEntersRange verifies:
// 1. Monster spawns with 10-tick immunity
// 2. After immunity, monster detects player in aggroRange
// 3. Monster switches to ATTACK and calls attackFunc
func TestAggroIntegration_PlayerEntersRange(t *testing.T) {
	// Monster at valid L2 coordinates
	monster, w := setupAggroWorld(t, 0x20050001, 300, 17000, 170000, -3500)

	// Player nearby (within aggroRange=300, distance ~70)
	_, playerObj := setupPlayer(t, 0x10050001, 17050, 170050, -3500)

	var attacked bool
	var attackedTargetID uint32
	attackFunc := func(m *model.Monster, target *model.WorldObject) {
		attacked = true
		attackedTargetID = target.ObjectID()
	}

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {
		world.ForEachVisibleObject(w, x, y, fn)
	}
	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		return w.GetObject(objectID)
	}

	ai := NewAttackableAI(monster, attackFunc, scanFunc, getObjectFunc)
	ai.Start()

	// Burn immunity (10 ticks: -10 → 0)
	for range 10 {
		ai.Tick()
	}

	// Tick 11: should detect player and switch to ATTACK
	ai.Tick()

	if ai.CurrentIntention() != model.IntentionAttack {
		t.Errorf("intention = %v, want ATTACK", ai.CurrentIntention())
	}

	if monster.Target() != playerObj.ObjectID() {
		t.Errorf("target = %d, want %d", monster.Target(), playerObj.ObjectID())
	}

	// Tick 12: should execute attack (player within attack range ~70 < 100)
	ai.Tick()

	if !attacked {
		t.Error("expected attackFunc to be called")
	}
	if attackedTargetID != playerObj.ObjectID() {
		t.Errorf("attacked target = %d, want %d", attackedTargetID, playerObj.ObjectID())
	}
}

// TestAggroIntegration_DamageRetaliates verifies:
// 1. Player attacks Monster during spawn immunity
// 2. NotifyDamage cancels immunity
// 3. Monster immediately retaliates against attacker
func TestAggroIntegration_DamageRetaliates(t *testing.T) {
	monster, w := setupAggroWorld(t, 0x20050002, 300, 17000, 170000, -3500)
	_, playerObj := setupPlayer(t, 0x10050002, 17050, 170050, -3500)

	var attacked bool
	attackFunc := func(m *model.Monster, target *model.WorldObject) {
		attacked = true
	}

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {
		world.ForEachVisibleObject(w, x, y, fn)
	}
	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		return w.GetObject(objectID)
	}

	ai := NewAttackableAI(monster, attackFunc, scanFunc, getObjectFunc)
	ai.Start()

	// Monster has spawn immunity — verify it
	ai.Tick() // tick 1 (globalAggro = -9)
	if ai.CurrentIntention() == model.IntentionAttack {
		t.Fatal("monster should not attack during immunity")
	}

	// Player attacks monster → damage notification cancels immunity
	ai.NotifyDamage(playerObj.ObjectID(), 100)

	// Monster should immediately switch to ATTACK
	if ai.CurrentIntention() != model.IntentionAttack {
		t.Errorf("after damage, intention = %v, want ATTACK", ai.CurrentIntention())
	}

	// Next tick should execute attack
	ai.Tick()

	if !attacked {
		t.Error("monster should retaliate after receiving damage")
	}
}

// TestAggroIntegration_TargetDies_ReturnToIdle verifies:
// 1. Monster attacks target
// 2. Target dies
// 3. Monster returns to ACTIVE (no more targets)
func TestAggroIntegration_TargetDies_ReturnToIdle(t *testing.T) {
	monster, w := setupAggroWorld(t, 0x20050003, 300, 17000, 170000, -3500)
	player, playerObj := setupPlayer(t, 0x10050003, 17050, 170050, -3500)

	attackFunc := func(m *model.Monster, target *model.WorldObject) {}

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {
		world.ForEachVisibleObject(w, x, y, fn)
	}
	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		return w.GetObject(objectID)
	}

	ai := NewAttackableAI(monster, attackFunc, scanFunc, getObjectFunc)
	ai.Start()

	// Skip immunity, add target via damage
	ai.NotifyDamage(playerObj.ObjectID(), 100)

	// Kill the player
	player.SetCurrentHP(0)

	// Tick should detect dead target and return to ACTIVE
	ai.Tick()

	if ai.CurrentIntention() != model.IntentionActive {
		t.Errorf("after target death, intention = %v, want ACTIVE", ai.CurrentIntention())
	}

	// Hate list should be cleared for the dead target
	if monster.AggroList().Get(playerObj.ObjectID()) != nil {
		t.Error("dead target should be removed from hate list")
	}
}

// TestAggroIntegration_MultipleTargets_AttacksMostHated verifies:
// 1. Two players attack monster
// 2. Player2 deals more damage → higher hate
// 3. Monster targets player2 (most hated)
func TestAggroIntegration_MultipleTargets_AttacksMostHated(t *testing.T) {
	monster, w := setupAggroWorld(t, 0x20050004, 300, 17000, 170000, -3500)

	_, player1Obj := setupPlayer(t, 0x10050004, 17050, 170050, -3500)
	_, player2Obj := setupPlayer(t, 0x10050005, 17010, 170010, -3500)

	var lastTargetID uint32
	attackFunc := func(m *model.Monster, target *model.WorldObject) {
		lastTargetID = target.ObjectID()
	}

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {
		world.ForEachVisibleObject(w, x, y, fn)
	}
	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		return w.GetObject(objectID)
	}

	ai := NewAttackableAI(monster, attackFunc, scanFunc, getObjectFunc)
	ai.Start()

	// Player1 does little damage
	ai.NotifyDamage(player1Obj.ObjectID(), 50)

	// Player2 does more damage → higher hate
	ai.NotifyDamage(player2Obj.ObjectID(), 300)

	// Tick to execute attack
	ai.Tick()

	if lastTargetID != player2Obj.ObjectID() {
		t.Errorf("should attack most hated (player2=%d), but attacked %d",
			player2Obj.ObjectID(), lastTargetID)
	}

	// Verify target is set to most hated
	if monster.Target() != player2Obj.ObjectID() {
		t.Errorf("monster target = %d, want %d (most hated)",
			monster.Target(), player2Obj.ObjectID())
	}
}
