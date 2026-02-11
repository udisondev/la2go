package skill

import (
	"log/slog"
	"strconv"
)

// HealOverTimeEffect periodically heals HP (HOT).
// Params: "power" (float64 per tick).
//
// Phase 5.9.3: Effect Framework.
// Java reference: HealOverTime.java
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
	slog.Debug("hot tick", "power", e.power, "target", targetObjID)
	return true
}

func (e *HealOverTimeEffect) OnExit(casterObjID, targetObjID uint32) {
	slog.Debug("hot ended", "target", targetObjID)
}
