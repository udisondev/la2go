package skill

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// setupWorldResolver sets up the package-level resolver with an in-memory object map.
// Returns the map so tests can add/remove objects.
func setupWorldResolver(t *testing.T) map[uint32]*model.WorldObject {
	t.Helper()
	objects := make(map[uint32]*model.WorldObject)
	SetWorldResolver(func(objectID uint32) (*model.WorldObject, bool) {
		obj, ok := objects[objectID]
		return obj, ok
	})
	t.Cleanup(func() { SetWorldResolver(nil) })
	return objects
}

// makeEffectTestPlayer creates a Player with a WorldObject wrapping it, and registers it.
func makeEffectTestPlayer(t *testing.T, objects map[uint32]*model.WorldObject, objectID uint32, hp, mp int32) *model.Player {
	t.Helper()
	player, err := model.NewPlayer(objectID, int64(objectID), 1, "Test", 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	player.Character.SetMaxHP(hp)
	player.Character.SetCurrentHP(hp)
	player.Character.SetMaxMP(mp)
	player.Character.SetCurrentMP(mp)

	wo := model.NewWorldObject(objectID, "Test", model.Location{})
	wo.Data = player
	player.Character.WorldObject = wo
	objects[objectID] = wo
	return player
}

// makeTestCharacter creates a basic Character wrapped in WorldObject and registers it.
func makeTestCharacter(t *testing.T, objects map[uint32]*model.WorldObject, objectID uint32, hp, mp int32) *model.Character {
	t.Helper()
	char := model.NewCharacter(objectID, "NPC", model.Location{}, 40, hp, mp, 0)
	wo := char.WorldObject
	objects[objectID] = wo
	return char
}

// --- Heal Effect Tests ---

func TestHealEffect_RestoresHP(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 100, 1000, 500)
	player.Character.SetCurrentHP(600) // Damaged

	heal := NewHealEffect(map[string]string{"power": "200"})
	heal.OnStart(100, 100)

	if hp := player.Character.CurrentHP(); hp != 800 {
		t.Errorf("expected HP=800 after heal, got %d", hp)
	}
}

func TestHealEffect_ClampsToMaxHP(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 101, 1000, 500)
	player.Character.SetCurrentHP(900) // Missing only 100 HP

	heal := NewHealEffect(map[string]string{"power": "500"})
	heal.OnStart(101, 101)

	if hp := player.Character.CurrentHP(); hp != 1000 {
		t.Errorf("expected HP=1000 (clamped), got %d", hp)
	}
}

func TestHealEffect_NoHealOnDead(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 102, 1000, 500)
	player.Character.SetCurrentHP(0) // Dead

	heal := NewHealEffect(map[string]string{"power": "500"})
	heal.OnStart(102, 102)

	if hp := player.Character.CurrentHP(); hp != 0 {
		t.Errorf("expected HP=0 (dead, no heal), got %d", hp)
	}
}

func TestHealEffect_NoResolver(t *testing.T) {
	SetWorldResolver(nil)
	// Should not panic
	heal := NewHealEffect(map[string]string{"power": "100"})
	heal.OnStart(1, 2) // No resolver — should silently do nothing
}

// --- MpHeal Effect Tests ---

func TestMpHealEffect_RestoresMP(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 200, 1000, 500)
	player.Character.SetCurrentMP(200) // Low MP

	mpHeal := NewMpHealEffect(map[string]string{"power": "150"})
	mpHeal.OnStart(200, 200)

	if mp := player.Character.CurrentMP(); mp != 350 {
		t.Errorf("expected MP=350, got %d", mp)
	}
}

func TestMpHealEffect_ClampsToMaxMP(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 201, 1000, 500)
	player.Character.SetCurrentMP(450) // Missing only 50

	mpHeal := NewMpHealEffect(map[string]string{"power": "200"})
	mpHeal.OnStart(201, 201)

	if mp := player.Character.CurrentMP(); mp != 500 {
		t.Errorf("expected MP=500 (clamped), got %d", mp)
	}
}

// --- DOT Effect Tests ---

func TestDOT_DealsDamage(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 300, 1000, 500)

	dot := NewDamageOverTimeEffect(map[string]string{"power": "50"})
	dot.OnStart(1, 300)

	cont := dot.OnActionTime(1, 300)
	if !cont {
		t.Error("DOT should continue ticking")
	}
	if hp := player.Character.CurrentHP(); hp != 950 {
		t.Errorf("expected HP=950, got %d", hp)
	}
}

func TestDOT_CanKillFalse_PreventsKill(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 301, 1000, 500)
	player.Character.SetCurrentHP(30)

	dot := NewDamageOverTimeEffect(map[string]string{"power": "100", "canKill": "false"})
	dot.OnStart(1, 301)
	dot.OnActionTime(1, 301)

	if hp := player.Character.CurrentHP(); hp != 1 {
		t.Errorf("expected HP=1 (canKill=false), got %d", hp)
	}
}

func TestDOT_CanKillTrue_CanKill(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 302, 1000, 500)
	player.Character.SetCurrentHP(30)

	dot := NewDamageOverTimeEffect(map[string]string{"power": "100", "canKill": "true"})
	dot.OnStart(1, 302)
	dot.OnActionTime(1, 302)

	if hp := player.Character.CurrentHP(); hp != 0 {
		t.Errorf("expected HP=0 (canKill=true), got %d", hp)
	}
}

func TestDOT_StopsOnDead(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 303, 1000, 500)
	player.Character.SetCurrentHP(0)

	dot := NewDamageOverTimeEffect(map[string]string{"power": "50"})
	cont := dot.OnActionTime(1, 303)

	if cont {
		t.Error("DOT should stop on dead target")
	}
}

// --- HOT Effect Tests ---

func TestHOT_HealsOverTime(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 400, 1000, 500)
	player.Character.SetCurrentHP(700)

	hot := NewHealOverTimeEffect(map[string]string{"power": "100"})
	hot.OnStart(1, 400)

	cont := hot.OnActionTime(1, 400)
	if !cont {
		t.Error("HOT should continue ticking")
	}
	if hp := player.Character.CurrentHP(); hp != 800 {
		t.Errorf("expected HP=800, got %d", hp)
	}
}

func TestHOT_StopsAtMaxHP(t *testing.T) {
	objects := setupWorldResolver(t)
	_ = makeEffectTestPlayer(t, objects, 401, 1000, 500)
	// Already at full HP

	hot := NewHealOverTimeEffect(map[string]string{"power": "100"})
	cont := hot.OnActionTime(1, 401)

	if cont {
		t.Error("HOT should stop when target is at max HP")
	}
}

func TestHOT_StopsOnDead(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 402, 1000, 500)
	player.Character.SetCurrentHP(0)

	hot := NewHealOverTimeEffect(map[string]string{"power": "100"})
	cont := hot.OnActionTime(1, 402)

	if cont {
		t.Error("HOT should stop on dead target")
	}
}

// --- Stun Effect Tests ---

func TestStunEffect_SetsFlag(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 500, 1000, 500)

	stun := NewStunEffect(nil)
	stun.OnStart(1, 500)

	if !player.Character.IsStunned() {
		t.Error("expected stunned=true after OnStart")
	}

	stun.OnExit(1, 500)

	if player.Character.IsStunned() {
		t.Error("expected stunned=false after OnExit")
	}
}

// --- Root Effect Tests ---

func TestRootEffect_SetsFlag(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 600, 1000, 500)

	root := NewRootEffect(nil)
	root.OnStart(1, 600)

	if !player.Character.IsRooted() {
		t.Error("expected rooted=true after OnStart")
	}
	if player.Character.IsStunned() {
		t.Error("root should not set stun flag")
	}

	root.OnExit(1, 600)

	if player.Character.IsRooted() {
		t.Error("expected rooted=false after OnExit")
	}
}

// --- Sleep Effect Tests ---

func TestSleepEffect_SetsFlag(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 700, 1000, 500)

	sleep := NewSleepEffect(nil)
	sleep.OnStart(1, 700)

	if !player.Character.IsSleeping() {
		t.Error("expected sleeping=true after OnStart")
	}

	sleep.OnExit(1, 700)

	if player.Character.IsSleeping() {
		t.Error("expected sleeping=false after OnExit")
	}
}

// --- Paralyze Effect Tests ---

func TestParalyzeEffect_SetsFlag(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 800, 1000, 500)

	para := NewParalyzeEffect(nil)
	para.OnStart(1, 800)

	if !player.Character.IsParalyzed() {
		t.Error("expected paralyzed=true after OnStart")
	}

	para.OnExit(1, 800)

	if player.Character.IsParalyzed() {
		t.Error("expected paralyzed=false after OnExit")
	}
}

// --- Hold/Fear Effect Tests ---

func TestHoldEffect_SetsFlag(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 900, 1000, 500)

	hold := NewHoldEffect(nil)
	hold.OnStart(1, 900)

	if !player.Character.IsFeared() {
		t.Error("expected feared=true after OnStart")
	}

	hold.OnExit(1, 900)

	if player.Character.IsFeared() {
		t.Error("expected feared=false after OnExit")
	}
}

// --- CancelTarget Effect Tests ---

func TestCancelTargetEffect_ClearsTarget(t *testing.T) {
	objects := setupWorldResolver(t)
	player := makeEffectTestPlayer(t, objects, 1000, 1000, 500)

	// Set a target
	targetObj := model.NewWorldObject(9999, "Enemy", model.Location{})
	player.SetTarget(targetObj)

	if player.Target() == nil {
		t.Fatal("target should be set before test")
	}

	cancel := NewCancelTargetEffect(nil)
	cancel.OnStart(1, 1000)

	if player.Target() != nil {
		t.Error("expected target=nil after CancelTarget")
	}
}

// --- MagicalDamage Effect Tests ---

func TestMagicalDamageEffect_DealsDamage(t *testing.T) {
	objects := setupWorldResolver(t)
	caster := makeEffectTestPlayer(t, objects, 1100, 1000, 500)
	target := makeEffectTestPlayer(t, objects, 1101, 1000, 500)
	_ = caster // Caster is resolved internally

	initialHP := target.Character.CurrentHP()

	effect := NewMagicalDamageEffect(map[string]string{"power": "50"})
	effect.OnStart(1100, 1101)

	if hp := target.Character.CurrentHP(); hp >= initialHP {
		t.Errorf("expected HP < %d after magic damage, got %d", initialHP, hp)
	}
}

// --- PhysicalDamage Effect Tests ---

func TestPhysicalDamageEffect_DealsDamage(t *testing.T) {
	objects := setupWorldResolver(t)
	caster := makeEffectTestPlayer(t, objects, 1200, 1000, 500)
	target := makeEffectTestPlayer(t, objects, 1201, 1000, 500)
	_ = caster

	initialHP := target.Character.CurrentHP()

	effect := NewPhysicalDamageEffect(map[string]string{"power": "2"})
	effect.OnStart(1200, 1201)

	if hp := target.Character.CurrentHP(); hp >= initialHP {
		t.Errorf("expected HP < %d after physical damage, got %d", initialHP, hp)
	}
}

// --- HpDrain Effect Tests ---

func TestHpDrainEffect_DamagesAndHeals(t *testing.T) {
	objects := setupWorldResolver(t)
	caster := makeEffectTestPlayer(t, objects, 1300, 1000, 500)
	target := makeEffectTestPlayer(t, objects, 1301, 1000, 500)
	caster.Character.SetCurrentHP(500) // Caster is damaged

	initialCasterHP := caster.Character.CurrentHP()
	initialTargetHP := target.Character.CurrentHP()

	effect := NewHpDrainEffect(map[string]string{"power": "50", "absorbPercent": "0.5"})
	effect.OnStart(1300, 1301)

	// Target should take damage
	if hp := target.Character.CurrentHP(); hp >= initialTargetHP {
		t.Errorf("expected target HP < %d, got %d", initialTargetHP, hp)
	}

	// Caster should be healed
	if hp := caster.Character.CurrentHP(); hp <= initialCasterHP {
		t.Errorf("expected caster HP > %d (drain heal), got %d", initialCasterHP, hp)
	}
}

// --- IsImmobilized / IsDisabled Tests ---

func TestCharacter_IsImmobilized(t *testing.T) {
	char := model.NewCharacter(9000, "Test", model.Location{}, 1, 100, 100, 0)

	if char.IsImmobilized() {
		t.Error("should not be immobilized initially")
	}

	char.SetRooted(true)
	if !char.IsImmobilized() {
		t.Error("rooted should make immobilized")
	}
	char.SetRooted(false)

	char.SetStunned(true)
	if !char.IsImmobilized() {
		t.Error("stunned should make immobilized")
	}
	if !char.IsDisabled() {
		t.Error("stunned should make disabled")
	}
	char.SetStunned(false)
}

func TestCharacter_ClearAllCCFlags(t *testing.T) {
	char := model.NewCharacter(9001, "Test", model.Location{}, 1, 100, 100, 0)

	char.SetStunned(true)
	char.SetRooted(true)
	char.SetSleeping(true)
	char.SetParalyzed(true)
	char.SetFeared(true)

	char.ClearAllCCFlags()

	if char.IsStunned() || char.IsRooted() || char.IsSleeping() || char.IsParalyzed() || char.IsFeared() {
		t.Error("ClearAllCCFlags should clear all flags")
	}
}
