package skill

import (
	"log/slog"
	"strconv"
)

// HealEffect restores HP instantly.
// Params: "power" (float64).
//
// Phase 5.9.3: Effect Framework.
// Java reference: Heal.java
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
	slog.Debug("heal", "power", e.power, "caster", casterObjID, "target", targetObjID)
}

func (e *HealEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return false
}

func (e *HealEffect) OnExit(casterObjID, targetObjID uint32) {}
