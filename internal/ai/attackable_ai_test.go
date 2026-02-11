package ai

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func newTestMonster(objectID uint32, aggroRange int32) *model.Monster {
	template := model.NewNpcTemplate(
		1000, "TestMob", "Monster",
		10, 1000, 500,
		100, 50, 80, 40,
		aggroRange, 120, 253,
		30, 60, 0, 0,
	)
	m := model.NewMonster(objectID, 1000, template)
	m.SetLocation(model.NewLocation(17000, 170000, -3500, 0))
	return m
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

func TestAttackableAI_SpawnImmunity(t *testing.T) {
	monster := newTestMonster(100001, 300)

	var attackCalled bool
	attackFunc := func(m *model.Monster, target *model.WorldObject) {
		attackCalled = true
	}

	// Player within aggro range
	playerLoc := model.NewLocation(17050, 170050, -3500, 0)
	playerObj := model.NewWorldObject(0x10000001, "Player1", playerLoc)
	player := newTestPlayer(t,0x10000001, 17050, 170050, -3500)
	playerObj.Data = player

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {
		fn(playerObj)
	}

	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		if objectID == playerObj.ObjectID() {
			return playerObj, true
		}
		return nil, false
	}

	ai := NewAttackableAI(monster, attackFunc, scanFunc, getObjectFunc)
	ai.Start()

	// First 9 ticks should have spawn immunity — no aggro
	// globalAggro: -10 → -9 → ... → -1 (9 ticks)
	// tick 10: -1 → 0, NPC can aggro
	for range 9 {
		ai.Tick()
	}

	if ai.CurrentIntention() == model.IntentionAttack {
		t.Error("monster should NOT attack during spawn immunity (9 ticks)")
	}

	if attackCalled {
		t.Error("attackFunc should NOT be called during spawn immunity")
	}
}

func TestAttackableAI_DetectsPlayerAfterImmunity(t *testing.T) {
	monster := newTestMonster(100002, 300)

	var attackCalled bool
	attackFunc := func(m *model.Monster, target *model.WorldObject) {
		attackCalled = true
	}

	// Player within aggro range (distance < 300)
	playerLoc := model.NewLocation(17050, 170050, -3500, 0)
	playerObj := model.NewWorldObject(0x10000002, "Player1", playerLoc)
	player := newTestPlayer(t,0x10000002, 17050, 170050, -3500)
	playerObj.Data = player

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {
		fn(playerObj)
	}

	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		if objectID == playerObj.ObjectID() {
			return playerObj, true
		}
		return nil, false
	}

	ai := NewAttackableAI(monster, attackFunc, scanFunc, getObjectFunc)
	ai.Start()

	// Burn spawn immunity
	for range 11 {
		ai.Tick()
	}

	// After immunity, monster should detect player and switch to ATTACK
	if ai.CurrentIntention() != model.IntentionAttack {
		t.Errorf("intention = %v, want ATTACK", ai.CurrentIntention())
	}

	// One more tick should execute attack (player within attack range)
	ai.Tick()

	if !attackCalled {
		t.Error("attackFunc should be called when target is in attack range")
	}
}

func TestAttackableAI_NotifyDamage_CancelsImmunity(t *testing.T) {
	monster := newTestMonster(100003, 300)

	attackFunc := func(m *model.Monster, target *model.WorldObject) {}

	// Player within attack range
	playerLoc := model.NewLocation(17050, 170050, -3500, 0)
	playerObj := model.NewWorldObject(0x10000003, "Player1", playerLoc)
	player := newTestPlayer(t,0x10000003, 17050, 170050, -3500)
	playerObj.Data = player

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {
		fn(playerObj)
	}

	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		if objectID == playerObj.ObjectID() {
			return playerObj, true
		}
		return nil, false
	}

	ai := NewAttackableAI(monster, attackFunc, scanFunc, getObjectFunc)
	ai.Start()

	// NotifyDamage should cancel immunity immediately
	ai.NotifyDamage(playerObj.ObjectID(), 100)

	// After damage notification, monster should switch to attack
	if ai.CurrentIntention() != model.IntentionAttack {
		t.Errorf("after NotifyDamage, intention = %v, want ATTACK", ai.CurrentIntention())
	}

	// Attacker should be in hate list
	if monster.AggroList().IsEmpty() {
		t.Error("hate list should not be empty after NotifyDamage")
	}

	mostHated := monster.AggroList().GetMostHated()
	if mostHated != playerObj.ObjectID() {
		t.Errorf("most hated = %d, want %d", mostHated, playerObj.ObjectID())
	}
}

func TestAttackableAI_PlayerOutOfAggroRange(t *testing.T) {
	monster := newTestMonster(100004, 300)

	attackFunc := func(m *model.Monster, target *model.WorldObject) {}

	// Player far away (distance > 300)
	playerLoc := model.NewLocation(20000, 170000, -3500, 0) // 3000 units away
	playerObj := model.NewWorldObject(0x10000004, "FarPlayer", playerLoc)
	player := newTestPlayer(t,0x10000004, 20000, 170000, -3500)
	playerObj.Data = player

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {
		fn(playerObj)
	}

	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		return nil, false
	}

	ai := NewAttackableAI(monster, attackFunc, scanFunc, getObjectFunc)
	ai.Start()

	// Burn immunity + extra ticks
	for range 15 {
		ai.Tick()
	}

	// Monster should stay in ACTIVE — player is out of aggro range
	if ai.CurrentIntention() == model.IntentionAttack {
		t.Error("monster should NOT attack player out of aggro range")
	}
}

func TestAttackableAI_TargetDies_ReturnsToActive(t *testing.T) {
	monster := newTestMonster(100005, 300)

	attackFunc := func(m *model.Monster, target *model.WorldObject) {}

	playerLoc := model.NewLocation(17050, 170050, -3500, 0)
	playerObj := model.NewWorldObject(0x10000005, "Player1", playerLoc)
	player := newTestPlayer(t,0x10000005, 17050, 170050, -3500)
	playerObj.Data = player

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {
		fn(playerObj)
	}

	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		if objectID == playerObj.ObjectID() {
			return playerObj, true
		}
		return nil, false
	}

	ai := NewAttackableAI(monster, attackFunc, scanFunc, getObjectFunc)
	ai.Start()

	// Cancel immunity and add target
	ai.NotifyDamage(playerObj.ObjectID(), 100)

	// Kill the player
	player.SetCurrentHP(0)

	// Tick — AI should notice target is dead and return to ACTIVE
	ai.Tick()

	if ai.CurrentIntention() != model.IntentionActive {
		t.Errorf("after target death, intention = %v, want ACTIVE", ai.CurrentIntention())
	}
}

func TestAttackableAI_DeadNPC_NoTick(t *testing.T) {
	monster := newTestMonster(100006, 300)

	var tickExecuted bool
	attackFunc := func(m *model.Monster, target *model.WorldObject) {
		tickExecuted = true
	}

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {}
	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		return nil, false
	}

	ai := NewAttackableAI(monster, attackFunc, scanFunc, getObjectFunc)
	ai.Start()

	// Kill the monster
	monster.SetCurrentHP(0)

	// Tick should be no-op
	for range 20 {
		ai.Tick()
	}

	if tickExecuted {
		t.Error("dead NPC should not execute attacks")
	}
}

func TestAttackableAI_MostHatedTarget(t *testing.T) {
	monster := newTestMonster(100007, 300)

	var lastTargetID uint32
	attackFunc := func(m *model.Monster, target *model.WorldObject) {
		lastTargetID = target.ObjectID()
	}

	player1Loc := model.NewLocation(17050, 170050, -3500, 0)
	player1Obj := model.NewWorldObject(0x10000007, "Player1", player1Loc)
	player1 := newTestPlayer(t, 0x10000007, 17050, 170050, -3500)
	player1Obj.Data = player1

	player2Loc := model.NewLocation(17010, 170010, -3500, 0)
	player2Obj := model.NewWorldObject(0x10000008, "Player2", player2Loc)
	player2 := newTestPlayer(t, 0x10000008, 17010, 170010, -3500)
	player2Obj.Data = player2

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {
		fn(player1Obj)
		fn(player2Obj)
	}

	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		switch objectID {
		case player1Obj.ObjectID():
			return player1Obj, true
		case player2Obj.ObjectID():
			return player2Obj, true
		}
		return nil, false
	}

	ai := NewAttackableAI(monster, attackFunc, scanFunc, getObjectFunc)
	ai.Start()

	// Player2 does more damage — should be most hated
	ai.NotifyDamage(player1Obj.ObjectID(), 50)
	ai.NotifyDamage(player2Obj.ObjectID(), 200)

	// Tick to execute attack
	ai.Tick()

	if lastTargetID != player2Obj.ObjectID() {
		t.Errorf("should attack most hated (player2), but attacked %d", lastTargetID)
	}
}
