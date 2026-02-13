package skill

import (
	"log/slog"
	"strconv"
)

// HealOverTimeEffect periodically heals HP (HOT).
// Params: "power" (float64 per tick).
//
// Phase 5.9.3+: Actual periodic healing with dead/full HP stop.
// Java reference: HealOverTime.java — onActionTime().
type HealOverTimeEffect struct {
	power float64
}

func NewHealOverTimeEffect(params map[string]string) Effect {
	power, _ := strconv.ParseFloat(params["power"], 64)
	return &HealOverTimeEffect{power: power}
}

func (e *HealOverTimeEffect) Name() string    { return "HealOverTime" }
func (e *HealOverTimeEffect) IsInstant() bool { return false }

func (e *HealOverTimeEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("hot started", "power", e.power, "target", targetObjID)
}

func (e *HealOverTimeEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	target := resolveCharacter(targetObjID)
	if target == nil || target.IsDead() {
		return false // Stop ticking on dead target
	}

	currentHP := target.CurrentHP()
	maxHP := target.MaxHP()

	// Stop if already at max HP.
	// Java: if (hp >= maxRecoverableHp) return false
	if currentHP >= maxHP {
		return false
	}

	heal := int32(e.power)
	if heal <= 0 {
		return true
	}

	// Clamp to maxHP.
	if currentHP+heal > maxHP {
		heal = maxHP - currentHP
	}

	target.SetCurrentHP(currentHP + heal)

	slog.Debug("hot tick",
		"power", e.power,
		"healed", heal,
		"target", targetObjID)

	return true // Continue ticking
}

func (e *HealOverTimeEffect) OnExit(casterObjID, targetObjID uint32) {
	slog.Debug("hot ended", "target", targetObjID)
}
