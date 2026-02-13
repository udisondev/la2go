package skill

import (
	"log/slog"
	"strconv"
)

// HealEffect restores HP instantly.
// Params: "power" (float64) — base heal amount.
//
// Phase 5.9.3+: Actual HP restoration with overheal clamp.
// Java reference: Heal.java — onStart().
type HealEffect struct {
	power float64
}

func NewHealEffect(params map[string]string) Effect {
	power, _ := strconv.ParseFloat(params["power"], 64)
	return &HealEffect{power: power}
}

func (e *HealEffect) Name() string    { return "Heal" }
func (e *HealEffect) IsInstant() bool { return true }

func (e *HealEffect) OnStart(casterObjID, targetObjID uint32) {
	target := resolveCharacter(targetObjID)
	if target == nil || target.IsDead() {
		return
	}

	amount := e.power

	// Clamp to not exceed maxHP (overheal prevention).
	// Java: Math.max(Math.min(amount, maxRecoverableHp - currentHp), 0)
	currentHP := target.CurrentHP()
	maxHP := target.MaxHP()
	maxHeal := float64(maxHP - currentHP)
	if amount > maxHeal {
		amount = maxHeal
	}
	if amount <= 0 {
		return
	}

	target.SetCurrentHP(currentHP + int32(amount))

	slog.Debug("heal applied",
		"power", e.power,
		"healed", int32(amount),
		"caster", casterObjID,
		"target", targetObjID)
}

func (e *HealEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return false
}

func (e *HealEffect) OnExit(casterObjID, targetObjID uint32) {}
