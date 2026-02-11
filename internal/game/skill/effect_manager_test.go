package skill

import "testing"

// testEffect is a simple Effect for testing.
type testEffect struct {
	name      string
	instant   bool
	modifiers []StatModifier
	started   bool
	exited    bool
}

func newTestEffect(name string, instant bool, mods ...StatModifier) *testEffect {
	return &testEffect{name: name, instant: instant, modifiers: mods}
}

func (e *testEffect) Name() string    { return e.name }
func (e *testEffect) IsInstant() bool { return e.instant }

func (e *testEffect) OnStart(casterObjID, targetObjID uint32)  { e.started = true }
func (e *testEffect) OnExit(casterObjID, targetObjID uint32)   { e.exited = true }
func (e *testEffect) OnActionTime(_, _ uint32) bool             { return true }
func (e *testEffect) StatModifiers() []StatModifier             { return e.modifiers }

func makeActiveEffect(skillID int32, abnormalType string, abnormalLevel int32, remainingMs int32, effect Effect) *ActiveEffect {
	return &ActiveEffect{
		CasterObjID:   1,
		TargetObjID:   2,
		SkillID:       skillID,
		SkillLevel:    1,
		Effect:        effect,
		RemainingMs:   remainingMs,
		AbnormalType:  abnormalType,
		AbnormalLevel: abnormalLevel,
	}
}

func TestAddBuff_Stacking_HigherReplaces(t *testing.T) {
	m := NewEffectManager()

	eff1 := newTestEffect("buff1", false, StatModifier{Stat: "pAtk", Type: StatModAdd, Value: 10})
	ae1 := makeActiveEffect(100, "MIGHT", 1, 60000, eff1)
	m.AddBuff(ae1)

	if m.BuffCount() != 1 {
		t.Fatalf("expected 1 buff, got %d", m.BuffCount())
	}

	// Higher level replaces
	eff2 := newTestEffect("buff2", false, StatModifier{Stat: "pAtk", Type: StatModAdd, Value: 20})
	ae2 := makeActiveEffect(101, "MIGHT", 2, 60000, eff2)
	replaced := m.AddBuff(ae2)

	if !replaced {
		t.Fatal("higher level should replace")
	}
	if m.BuffCount() != 1 {
		t.Fatalf("expected 1 buff after replace, got %d", m.BuffCount())
	}
	if !eff1.exited {
		t.Error("old effect OnExit should have been called")
	}
	if !eff2.started {
		t.Error("new effect OnStart should have been called")
	}
}

func TestAddBuff_Stacking_SameRefreshes(t *testing.T) {
	m := NewEffectManager()

	eff1 := newTestEffect("buff1", false)
	ae1 := makeActiveEffect(100, "MIGHT", 1, 30000, eff1)
	m.AddBuff(ae1)

	// Same level refreshes duration
	eff2 := newTestEffect("buff2", false)
	ae2 := makeActiveEffect(100, "MIGHT", 1, 60000, eff2)
	m.AddBuff(ae2)

	if m.BuffCount() != 1 {
		t.Fatalf("expected 1 buff, got %d", m.BuffCount())
	}

	// Original effect should have refreshed duration
	buffs := m.ActiveBuffs()
	if buffs[0].RemainingMs != 60000 {
		t.Errorf("expected refreshed duration 60000, got %d", buffs[0].RemainingMs)
	}
}

func TestAddBuff_Stacking_LowerRejected(t *testing.T) {
	m := NewEffectManager()

	eff1 := newTestEffect("buff1", false)
	ae1 := makeActiveEffect(100, "MIGHT", 2, 60000, eff1)
	m.AddBuff(ae1)

	// Lower level rejected
	eff2 := newTestEffect("buff2", false)
	ae2 := makeActiveEffect(101, "MIGHT", 1, 60000, eff2)
	added := m.AddBuff(ae2)

	if added {
		t.Fatal("lower level should be rejected")
	}
	if m.BuffCount() != 1 {
		t.Fatalf("expected 1 buff, got %d", m.BuffCount())
	}
}

func TestRemoveEffect(t *testing.T) {
	m := NewEffectManager()

	eff := newTestEffect("buff", false)
	ae := makeActiveEffect(100, "MIGHT", 1, 60000, eff)
	m.AddBuff(ae)

	m.RemoveEffect("MIGHT")

	if m.BuffCount() != 0 {
		t.Fatalf("expected 0 buffs after remove, got %d", m.BuffCount())
	}
	if !eff.exited {
		t.Error("OnExit should have been called on remove")
	}
}

func TestGetStatBonus(t *testing.T) {
	m := NewEffectManager()

	// Add buff with +100 pAtk
	eff1 := newTestEffect("might", false, StatModifier{Stat: "pAtk", Type: StatModAdd, Value: 100})
	ae1 := makeActiveEffect(100, "MIGHT", 1, 60000, eff1)
	m.AddBuff(ae1)

	// Add buff with +50 pAtk
	eff2 := newTestEffect("focus", false, StatModifier{Stat: "pAtk", Type: StatModAdd, Value: 50})
	ae2 := makeActiveEffect(101, "FOCUS", 1, 60000, eff2)
	m.AddBuff(ae2)

	bonus := m.GetStatBonus("pAtk")
	if bonus != 150 {
		t.Errorf("expected 150 pAtk bonus, got %.1f", bonus)
	}

	// Different stat should return 0
	bonus = m.GetStatBonus("mAtk")
	if bonus != 0 {
		t.Errorf("expected 0 mAtk bonus, got %.1f", bonus)
	}
}

func TestGetStatBonus_Multiplicative(t *testing.T) {
	m := NewEffectManager()

	// +100 pAtk additive
	eff1 := newTestEffect("add", false, StatModifier{Stat: "pAtk", Type: StatModAdd, Value: 100})
	ae1 := makeActiveEffect(100, "ADD", 1, 60000, eff1)
	m.AddBuff(ae1)

	// x1.5 pAtk multiplicative
	eff2 := newTestEffect("mul", false, StatModifier{Stat: "pAtk", Type: StatModMul, Value: 1.5})
	ae2 := makeActiveEffect(101, "MUL", 1, 60000, eff2)
	m.AddBuff(ae2)

	bonus := m.GetStatBonus("pAtk")
	// 100 * 1.5 = 150
	if bonus != 150 {
		t.Errorf("expected 150, got %.1f", bonus)
	}
}

func TestBuffLimit(t *testing.T) {
	m := NewEffectManager()

	// Add maxBuffs + 1 buffs (each with unique abnormal type)
	for i := range int32(maxBuffs + 1) {
		eff := newTestEffect("buff", false)
		ae := makeActiveEffect(i, "", 0, 60000, eff)
		m.AddBuff(ae)
	}

	if m.BuffCount() != maxBuffs {
		t.Errorf("expected %d buffs after limit, got %d", maxBuffs, m.BuffCount())
	}
}

func TestDebuffLimit(t *testing.T) {
	m := NewEffectManager()

	for i := range int32(maxDebuffs + 1) {
		eff := newTestEffect("debuff", false)
		ae := makeActiveEffect(i, "", 0, 60000, eff)
		m.AddDebuff(ae)
	}

	if m.DebuffCount() != maxDebuffs {
		t.Errorf("expected %d debuffs after limit, got %d", maxDebuffs, m.DebuffCount())
	}
}

func TestTick_ExpiresEffects(t *testing.T) {
	m := NewEffectManager()

	eff1 := newTestEffect("short", false, StatModifier{Stat: "pAtk", Type: StatModAdd, Value: 50})
	ae1 := makeActiveEffect(100, "SHORT", 1, 1000, eff1) // 1 second
	m.AddBuff(ae1)

	eff2 := newTestEffect("long", false, StatModifier{Stat: "pDef", Type: StatModAdd, Value: 30})
	ae2 := makeActiveEffect(101, "LONG", 1, 5000, eff2) // 5 seconds
	m.AddBuff(ae2)

	if m.BuffCount() != 2 {
		t.Fatalf("expected 2 buffs, got %d", m.BuffCount())
	}

	// Tick 2 seconds — short buff expires
	m.Tick(2000)

	if m.BuffCount() != 1 {
		t.Fatalf("expected 1 buff after tick, got %d", m.BuffCount())
	}
	if !eff1.exited {
		t.Error("expired effect OnExit should have been called")
	}

	// pAtk bonus should be gone
	bonus := m.GetStatBonus("pAtk")
	if bonus != 0 {
		t.Errorf("expected 0 pAtk after expiry, got %.1f", bonus)
	}

	// pDef bonus should remain
	bonus = m.GetStatBonus("pDef")
	if bonus != 30 {
		t.Errorf("expected 30 pDef, got %.1f", bonus)
	}
}

func TestEffectRegistry(t *testing.T) {
	// Test all 15 registered effects can be created
	effectNames := []string{
		"Buff", "PhysicalDamage", "MagicalDamage", "Heal", "MpHeal",
		"HpDrain", "DamageOverTime", "HealOverTime", "Stun", "Root",
		"Paralyze", "Sleep", "CancelTarget", "SpeedChange", "StatUp",
	}

	for _, name := range effectNames {
		eff, err := CreateEffect(name, map[string]string{"power": "100", "stat": "pAtk", "value": "50", "type": "ADD"})
		if err != nil {
			t.Errorf("CreateEffect(%q) failed: %v", name, err)
			continue
		}
		if eff.Name() != name {
			t.Errorf("expected Name()=%q, got %q", name, eff.Name())
		}
	}

	// Unknown effect should error
	_, err := CreateEffect("NonExistent", nil)
	if err == nil {
		t.Error("expected error for unknown effect")
	}
}

func TestAddPassive_ReplacesExisting(t *testing.T) {
	m := NewEffectManager()

	eff1 := newTestEffect("passive1", false, StatModifier{Stat: "pDef", Type: StatModAdd, Value: 10})
	ae1 := makeActiveEffect(228, "", 0, 0, eff1)
	m.AddPassive(ae1)

	eff2 := newTestEffect("passive2", false, StatModifier{Stat: "pDef", Type: StatModAdd, Value: 20})
	ae2 := makeActiveEffect(228, "", 0, 0, eff2) // same skill ID
	m.AddPassive(ae2)

	bonus := m.GetStatBonus("pDef")
	if bonus != 20 {
		t.Errorf("expected 20 pDef after passive replace, got %.1f", bonus)
	}
	if !eff1.exited {
		t.Error("old passive OnExit should have been called")
	}
}
