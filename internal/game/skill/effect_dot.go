package skill

import (
	"log/slog"
	"strconv"
)

// DamageOverTimeEffect deals periodic damage (DOT).
// Params: "power" (float64 per tick), "canKill" (bool, default false).
//
// Phase 5.9.3+: Actual periodic damage with canKill protection.
// Java reference: DamOverTime.java — onActionTime().
// If canKill is false, damage cannot reduce HP below 1.
type DamageOverTimeEffect struct {
	power   float64
	canKill bool
}

func NewDamageOverTimeEffect(params map[string]string) Effect {
	power, _ := strconv.ParseFloat(params["power"], 64)
	canKill, _ := strconv.ParseBool(params["canKill"])
	return &DamageOverTimeEffect{power: power, canKill: canKill}
}

func (e *DamageOverTimeEffect) Name() string    { return "DamageOverTime" }
func (e *DamageOverTimeEffect) IsInstant() bool { return false }

func (e *DamageOverTimeEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("dot started", "power", e.power, "canKill", e.canKill, "target", targetObjID)
}

func (e *DamageOverTimeEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	target := resolveCharacter(targetObjID)
	if target == nil || target.IsDead() {
		return false // Stop ticking on dead target
	}

	damage := int32(e.power)
	if damage <= 0 {
		return true
	}

	currentHP := target.CurrentHP()

	// Kill protection: if canKill is false, don't reduce below 1 HP.
	// Java: if (!_canKill && damage >= currentHp - 1) damage = currentHp - 1
	if !e.canKill && damage >= currentHP {
		damage = currentHP - 1
		if damage <= 0 {
			return true // Can't deal any more damage
		}
	}

	target.ReduceCurrentHP(damage)

	slog.Debug("dot tick",
		"power", e.power,
		"damage", damage,
		"target", targetObjID)

	return true // Continue ticking
}

func (e *DamageOverTimeEffect) OnExit(casterObjID, targetObjID uint32) {
	slog.Debug("dot ended", "target", targetObjID)
}
