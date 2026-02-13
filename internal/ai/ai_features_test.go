package ai

import (
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

// --- Feature 1: Skill Casting ---

func TestSkillCast_NpcCastsSkillWhenAvailable(t *testing.T) {
	// Set up NPC data with a skill
	setupTestNpcData(t, 1000, []data.TestNpcSkill{{SkillID: 1001, Level: 1}}, nil)

	// Set up skill template with range 500
	setupTestSkillData(t, 1001, 1, 500, 0, 5000) // range=500, mp=0, reuse=5000ms

	monster := newTestMonster(200001, 300)
	monster.SetCurrentMP(500)

	var castCalled bool
	var castSkillID int32
	castFunc := func(m *model.Monster, target *model.WorldObject, skillID, skillLevel int32) {
		castCalled = true
		castSkillID = skillID
	}

	playerObj := makeTestWorldPlayer(t, 0x20000001, 17050, 170050, -3500)

	ai := NewAttackableAI(monster, nil, nil, nil)
	ai.SetCastFunc(castFunc)
	ai.Start()

	// Try skill cast at distance ~70 (within range 500)
	if !ai.trySkillCast(playerObj, 70) {
		t.Fatal("expected skill cast to succeed")
	}

	if !castCalled {
		t.Fatal("castFunc should have been called")
	}
	if castSkillID != 1001 {
		t.Errorf("cast skill ID = %d, want 1001", castSkillID)
	}
}

func TestSkillCast_NoCastWhenNoCastFunc(t *testing.T) {
	monster := newTestMonster(200002, 300)
	ai := NewAttackableAI(monster, nil, nil, nil)
	ai.Start()

	playerObj := makeTestWorldPlayer(t, 0x20000002, 17050, 170050, -3500)

	if ai.trySkillCast(playerObj, 70) {
		t.Fatal("should not cast without castFunc")
	}
}

func TestSkillCast_SkillOnCooldown(t *testing.T) {
	setupTestNpcData(t, 1000, []data.TestNpcSkill{{SkillID: 1002, Level: 1}}, nil)
	setupTestSkillData(t, 1002, 1, 500, 0, 60000) // reuse=60s

	monster := newTestMonster(200003, 300)
	monster.SetCurrentMP(500)

	var castCount int
	castFunc := func(m *model.Monster, target *model.WorldObject, skillID, skillLevel int32) {
		castCount++
	}

	playerObj := makeTestWorldPlayer(t, 0x20000003, 17050, 170050, -3500)

	ai := NewAttackableAI(monster, nil, nil, nil)
	ai.SetCastFunc(castFunc)
	ai.Start()

	// First cast should succeed
	ai.trySkillCast(playerObj, 70)
	if castCount != 1 {
		t.Fatalf("first cast: count = %d, want 1", castCount)
	}

	// Second cast should fail (on cooldown)
	if ai.trySkillCast(playerObj, 70) {
		t.Fatal("second cast should fail (cooldown)")
	}
	if castCount != 1 {
		t.Fatalf("after cooldown: count = %d, want 1", castCount)
	}
}

func TestSkillCast_OutOfRange(t *testing.T) {
	setupTestNpcData(t, 1000, []data.TestNpcSkill{{SkillID: 1003, Level: 1}}, nil)
	setupTestSkillData(t, 1003, 1, 100, 0, 1000) // range=100

	monster := newTestMonster(200004, 300)
	monster.SetCurrentMP(500)

	castFunc := func(m *model.Monster, target *model.WorldObject, skillID, skillLevel int32) {
		t.Fatal("should not cast when out of range")
	}

	playerObj := makeTestWorldPlayer(t, 0x20000004, 17050, 170050, -3500)

	ai := NewAttackableAI(monster, nil, nil, nil)
	ai.SetCastFunc(castFunc)
	ai.Start()

	// Distance 200 > range 100
	if ai.trySkillCast(playerObj, 200) {
		t.Fatal("should not cast at distance 200 with range 100")
	}
}

func TestSkillCast_NotEnoughMP(t *testing.T) {
	setupTestNpcData(t, 1000, []data.TestNpcSkill{{SkillID: 1004, Level: 1}}, nil)
	setupTestSkillData(t, 1004, 1, 500, 100, 1000) // mp=100

	monster := newTestMonster(200005, 300)
	monster.SetCurrentMP(50) // Only 50 MP, need 100

	castFunc := func(m *model.Monster, target *model.WorldObject, skillID, skillLevel int32) {
		t.Fatal("should not cast when not enough MP")
	}

	playerObj := makeTestWorldPlayer(t, 0x20000005, 17050, 170050, -3500)

	ai := NewAttackableAI(monster, nil, nil, nil)
	ai.SetCastFunc(castFunc)
	ai.Start()

	if ai.trySkillCast(playerObj, 70) {
		t.Fatal("should not cast with insufficient MP")
	}
}

// --- Feature 2: Chase ---

func TestChase_MovesTowardTarget(t *testing.T) {
	monster := newTestMonster(200010, 300)

	var moveCalled bool
	var moveX, moveY, moveZ int32
	moveFunc := func(npc *model.Npc, x, y, z int32) {
		moveCalled = true
		moveX = x
		moveY = y
		moveZ = z
	}

	targetLoc := model.NewLocation(18000, 171000, -3500, 0)
	playerObj := model.NewWorldObject(0x20000010, "FarPlayer", targetLoc)
	player := newTestPlayer(t, 0x20000010, 18000, 171000, -3500)
	playerObj.Data = player

	ai := NewAttackableAI(monster, nil, nil, nil)
	ai.SetMoveFunc(moveFunc)
	ai.Start()

	ai.chaseTarget(playerObj)

	if !moveCalled {
		t.Fatal("moveFunc should be called during chase")
	}
	if moveX != 18000 || moveY != 171000 || moveZ != -3500 {
		t.Errorf("move target = (%d,%d,%d), want (18000,171000,-3500)", moveX, moveY, moveZ)
	}
}

func TestChase_NoMoveWithoutMoveFunc(t *testing.T) {
	monster := newTestMonster(200011, 300)
	playerObj := makeTestWorldPlayer(t, 0x20000011, 18000, 171000, -3500)

	ai := NewAttackableAI(monster, nil, nil, nil)
	ai.Start()

	// Should not panic without moveFunc
	ai.chaseTarget(playerObj)
}

// --- Feature 3: Faction Call ---

func TestFactionCall_AlertsNearbyClanMembers(t *testing.T) {
	// Setup: two monsters of same clan
	setupTestNpcData(t, 1000,
		nil,
		[]string{"orc_clan"},
	)

	monster1 := newTestMonster(200020, 300) // The attacked one
	monster2 := newTestMonster(200021, 300) // The helper

	monster2.SetLocation(model.NewLocation(17100, 170100, -3500, 0)) // 141 units away

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {
		// Return monster2's WorldObject
		fn(monster2.WorldObject)
	}

	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		return nil, false
	}

	attackerID := uint32(0x30000001)
	ai := NewAttackableAI(monster1, nil, scanFunc, getObjectFunc)
	ai.Start()

	ai.callFaction(attackerID)

	// Monster2 should now have hate for the attacker
	info := monster2.AggroList().Get(attackerID)
	if info == nil {
		t.Fatal("helper monster should have attacker in aggro list")
	}
	if info.Hate() < 1 {
		t.Errorf("helper hate = %d, want >= 1", info.Hate())
	}
}

func TestFactionCall_IgnoresNonClanMembers(t *testing.T) {
	// Setup: two monsters of different clans
	data.ClearTestNpcTable()

	// Use template IDs 1000 and 1001 for different clans
	setupTestNpcDataByTemplate(t, 1000, nil, []string{"orc_clan"})
	setupTestNpcDataByTemplate(t, 1001, nil, []string{"elf_clan"})

	template1 := model.NewNpcTemplate(1000, "Mob1", "", 10, 1000, 500, 100, 50, 80, 40, 300, 120, 253, 30, 60, 0, 0)
	template2 := model.NewNpcTemplate(1001, "Mob2", "", 10, 1000, 500, 100, 50, 80, 40, 300, 120, 253, 30, 60, 0, 0)

	monster1 := model.NewMonster(200030, 1000, template1)
	monster1.SetLocation(model.NewLocation(17000, 170000, -3500, 0))

	monster2 := model.NewMonster(200031, 1001, template2) // Different template = different clan
	monster2.SetLocation(model.NewLocation(17100, 170100, -3500, 0))

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {
		fn(monster2.WorldObject)
	}

	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		return nil, false
	}

	attackerID := uint32(0x30000002)
	ai := NewAttackableAI(monster1, nil, scanFunc, getObjectFunc)
	ai.Start()

	ai.callFaction(attackerID)

	// Monster2 should NOT have hate (different clan)
	if !monster2.AggroList().IsEmpty() {
		t.Fatal("non-clan monster should not receive faction call")
	}
}

// --- Feature 4: Random Walk ---

func TestRandomWalk_MovesNearSpawn(t *testing.T) {
	monster := newTestMonster(200040, 300)

	spawn := model.NewSpawn(1, 1000, 17000, 170000, -3500, 0, 1, true)
	monster.SetSpawn(spawn)

	var moveCount int
	moveFunc := func(npc *model.Npc, x, y, z int32) {
		moveCount++
		// Verify within maxDriftRange of spawn
		dx := x - 17000
		dy := y - 170000
		if dx < -300 || dx > 300 || dy < -300 || dy > 300 {
			t.Errorf("random walk out of range: dx=%d, dy=%d", dx, dy)
		}
	}

	ai := NewAttackableAI(monster, nil, nil, nil)
	ai.SetMoveFunc(moveFunc)
	ai.Start()

	// Run many ticks — statistically should walk at least once (1/30 chance each)
	for range 300 {
		ai.tryRandomWalk()
	}

	if moveCount == 0 {
		t.Fatal("expected at least one random walk in 300 attempts")
	}
}

func TestRandomWalk_NoWalkWithoutSpawn(t *testing.T) {
	monster := newTestMonster(200041, 300)
	// No spawn set

	moveFunc := func(npc *model.Npc, x, y, z int32) {
		t.Fatal("should not walk without spawn point")
	}

	ai := NewAttackableAI(monster, nil, nil, nil)
	ai.SetMoveFunc(moveFunc)
	ai.Start()

	for range 100 {
		ai.tryRandomWalk()
	}
}

// --- Feature 5: Hate Decay ---

func TestHateDecay_ClearsAtFullHP(t *testing.T) {
	monster := newTestMonster(200050, 300)
	// Monster is at full HP/MP by default after creation

	// Add some hate
	monster.AggroList().AddHate(0x40000001, 100)

	ai := NewAttackableAI(monster, nil, nil, nil)
	ai.Start()

	// Run enough ticks — with 1/500 chance, ~500 attempts should trigger
	cleared := false
	for range 5000 {
		ai.checkHateDecay()
		if monster.AggroList().IsEmpty() {
			cleared = true
			break
		}
	}

	if !cleared {
		t.Fatal("hate should eventually decay at full HP/MP")
	}
}

func TestHateDecay_NoClearWhenDamaged(t *testing.T) {
	monster := newTestMonster(200051, 300)
	monster.SetCurrentHP(500) // Not full HP (max is 1000)

	monster.AggroList().AddHate(0x40000002, 100)

	ai := NewAttackableAI(monster, nil, nil, nil)
	ai.Start()

	for range 1000 {
		ai.checkHateDecay()
	}

	// Should NOT clear because HP is not full
	if monster.AggroList().IsEmpty() {
		t.Fatal("hate should not decay when HP is not full")
	}
}

// --- Attack Timeout / Return Home ---

func TestAttackTimeout_ReturnsHome(t *testing.T) {
	monster := newTestMonster(200060, 300)
	spawn := model.NewSpawn(1, 1000, 17000, 170000, -3500, 0, 1, true)
	monster.SetSpawn(spawn)

	// Move monster away from spawn
	monster.SetLocation(model.NewLocation(18000, 171000, -3500, 0))
	monster.SetCurrentHP(500)

	var moveCalled bool
	moveFunc := func(npc *model.Npc, x, y, z int32) {
		moveCalled = true
	}

	playerObj := makeTestWorldPlayer(t, 0x50000001, 18050, 171050, -3500)

	scanFunc := func(x, y int32, fn func(*model.WorldObject) bool) {}
	getObjectFunc := func(objectID uint32) (*model.WorldObject, bool) {
		if objectID == playerObj.ObjectID() {
			return playerObj, true
		}
		return nil, false
	}

	ai := NewAttackableAI(monster, nil, scanFunc, getObjectFunc)
	ai.SetMoveFunc(moveFunc)
	ai.Start()

	// Set attack timeout to past
	ai.attackTimeout.Store(time.Now().Add(-1 * time.Second).UnixMilli())

	// Force attack intention
	ai.NotifyDamage(playerObj.ObjectID(), 100)
	ai.attackTimeout.Store(time.Now().Add(-1 * time.Second).UnixMilli())

	// Tick should trigger returnHome
	ai.Tick()

	if ai.CurrentIntention() != model.IntentionActive {
		t.Errorf("after timeout, intention = %v, want ACTIVE", ai.CurrentIntention())
	}

	// HP should be restored
	if monster.CurrentHP() != monster.MaxHP() {
		t.Errorf("HP = %d, want %d (full)", monster.CurrentHP(), monster.MaxHP())
	}

	if !moveCalled {
		t.Error("moveFunc should be called to return to spawn")
	}
}

// --- Return to Spawn (Idle Drift) ---

func TestReturnToSpawn_WhenTooFarFromSpawn(t *testing.T) {
	monster := newTestMonster(200070, 300)
	spawn := model.NewSpawn(1, 1000, 17000, 170000, -3500, 0, 1, true)
	monster.SetSpawn(spawn)

	// Move monster 400 units from spawn (> maxDriftRange=300)
	monster.SetLocation(model.NewLocation(17400, 170000, -3500, 0))

	var moveX int32
	moveFunc := func(npc *model.Npc, x, y, z int32) {
		moveX = x
	}

	ai := NewAttackableAI(monster, nil, nil, nil)
	ai.SetMoveFunc(moveFunc)
	ai.Start()

	ai.checkReturnToSpawn()

	if moveX != 17000 {
		t.Errorf("should return to spawn X=17000, got %d", moveX)
	}
}

func TestReturnToSpawn_StaysIfCloseToSpawn(t *testing.T) {
	monster := newTestMonster(200071, 300)
	spawn := model.NewSpawn(1, 1000, 17000, 170000, -3500, 0, 1, true)
	monster.SetSpawn(spawn)

	// Monster only 100 units from spawn (< maxDriftRange=300)
	monster.SetLocation(model.NewLocation(17100, 170000, -3500, 0))

	moveFunc := func(npc *model.Npc, x, y, z int32) {
		t.Fatal("should not return to spawn when close")
	}

	ai := NewAttackableAI(monster, nil, nil, nil)
	ai.SetMoveFunc(moveFunc)
	ai.Start()

	ai.checkReturnToSpawn()
}

// --- Too Far From Spawn (Combat Leash) ---

func TestTooFarFromSpawn_ReturnsHome(t *testing.T) {
	monster := newTestMonster(200080, 300)
	spawn := model.NewSpawn(1, 1000, 17000, 170000, -3500, 0, 1, true)
	monster.SetSpawn(spawn)

	// Move monster 2000 units from spawn (> chaseRangeNormal=1500)
	monster.SetLocation(model.NewLocation(19000, 170000, -3500, 0))

	if !(&AttackableAI{monster: monster}).isTooFarFromSpawn() {
		t.Fatal("NPC at 2000 units should be too far from spawn (limit 1500)")
	}

	// Within range
	monster.SetLocation(model.NewLocation(17500, 170000, -3500, 0))
	if (&AttackableAI{monster: monster}).isTooFarFromSpawn() {
		t.Fatal("NPC at 500 units should NOT be too far from spawn")
	}
}

// --- Helpers ---

func makeTestWorldPlayer(t *testing.T, objectID uint32, x, y, z int32) *model.WorldObject {
	t.Helper()
	playerLoc := model.NewLocation(x, y, z, 0)
	playerObj := model.NewWorldObject(objectID, "TestPlayer", playerLoc)
	player := newTestPlayer(t, objectID, x, y, z)
	playerObj.Data = player
	return playerObj
}

// setupTestNpcData populates data.NpcTable with a test NPC at templateID 1000.
func setupTestNpcData(t *testing.T, _ int32, skills []data.TestNpcSkill, clans []string) {
	t.Helper()
	setupTestNpcDataByTemplate(t, 1000, skills, clans)
}

// setupTestNpcDataByTemplate populates data.NpcTable for a specific templateID.
func setupTestNpcDataByTemplate(t *testing.T, templateID int32, skills []data.TestNpcSkill, clans []string) {
	t.Helper()
	data.SetTestNpcDef(templateID, skills, clans)
	t.Cleanup(func() {
		data.DeleteTestNpcDef(templateID)
	})
}

// setupTestSkillData populates data.SkillTable with a test skill.
func setupTestSkillData(t *testing.T, skillID, level, castRange, mpConsume, reuseDelay int32) {
	t.Helper()
	if data.SkillTable == nil {
		data.SkillTable = make(map[int32]map[int32]*data.SkillTemplate)
	}
	if data.SkillTable[skillID] == nil {
		data.SkillTable[skillID] = make(map[int32]*data.SkillTemplate)
	}
	data.SkillTable[skillID][level] = &data.SkillTemplate{
		ID:         skillID,
		Level:      level,
		Name:       "TestSkill",
		CastRange:  castRange,
		MpConsume:  mpConsume,
		ReuseDelay: reuseDelay,
	}
	t.Cleanup(func() {
		delete(data.SkillTable[skillID], level)
	})
}
