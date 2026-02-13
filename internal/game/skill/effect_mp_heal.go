package skill

import (
	"log/slog"
	"strconv"
)

// MpHealEffect restores MP instantly.
// Params: "power" (float64) — base MP restore amount.
//
// Phase 5.9.3+: Actual MP restoration with overheal clamp.
// Java reference: ManaHeal.java — onStart().
type MpHealEffect struct {
	power float64
}

func NewMpHealEffect(params map[string]string) Effect {
	power, _ := strconv.ParseFloat(params["power"], 64)
	return &MpHealEffect{power: power}
}

func (e *MpHealEffect) Name() string    { return "MpHeal" }
func (e *MpHealEffect) IsInstant() bool { return true }

func (e *MpHealEffect) OnStart(casterObjID, targetObjID uint32) {
	target := resolveCharacter(targetObjID)
	if target == nil || target.IsDead() {
		return
	}

	amount := e.power

	// Clamp to not exceed maxMP (overheal prevention).
	currentMP := target.CurrentMP()
	maxMP := target.MaxMP()
	maxRestore := float64(maxMP - currentMP)
	if amount > maxRestore {
		amount = maxRestore
	}
	if amount <= 0 {
		return
	}

	target.SetCurrentMP(currentMP + int32(amount))

	slog.Debug("mp heal applied",
		"power", e.power,
		"restored", int32(amount),
		"caster", casterObjID,
		"target", targetObjID)
}

func (e *MpHealEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return false
}

func (e *MpHealEffect) OnExit(casterObjID, targetObjID uint32) {}
