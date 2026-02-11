package skill

import (
	"log/slog"
	"strconv"
)

// DamageOverTimeEffect deals periodic damage (DOT).
// Params: "power" (float64 per tick).
//
// Phase 5.9.3: Effect Framework.
// Java reference: DamOverTime.java
type DamageOverTimeEffect struct {
	power float64
}

func NewDamageOverTimeEffect(params map[string]string) Effect {
	power, _ := strconv.ParseFloat(params["power"], 64)
	return &DamageOverTimeEffect{power: power}
}

func (e *DamageOverTimeEffect) Name() string    { return "DamageOverTime" }
func (e *DamageOverTimeEffect) IsInstant() bool { return false }

func (e *DamageOverTimeEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("dot started", "power", e.power, "target", targetObjID)
}

func (e *DamageOverTimeEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	slog.Debug("dot tick", "power", e.power, "target", targetObjID)
	return true
}

func (e *DamageOverTimeEffect) OnExit(casterObjID, targetObjID uint32) {
	slog.Debug("dot ended", "target", targetObjID)
}
