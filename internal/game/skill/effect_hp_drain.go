package skill

import (
	"log/slog"
	"strconv"
)

// HpDrainEffect deals damage and heals caster for a percentage.
// Params: "power" (float64), "absorbPercent" (float64, 0.0-1.0).
//
// Phase 5.9.3: Effect Framework.
// Java reference: HpDrain.java
type HpDrainEffect struct {
	power          float64
	absorbPercent  float64
}

func NewHpDrainEffect(params map[string]string) Effect {
	power, _ := strconv.ParseFloat(params["power"], 64)
	absorbPercent, _ := strconv.ParseFloat(params["absorbPercent"], 64)
	if absorbPercent == 0 {
		absorbPercent = 0.5 // Default 50% absorb
	}
	return &HpDrainEffect{power: power, absorbPercent: absorbPercent}
}

func (e *HpDrainEffect) Name() string    { return "HpDrain" }
func (e *HpDrainEffect) IsInstant() bool { return true }

func (e *HpDrainEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("hp drain", "power", e.power, "absorb", e.absorbPercent, "caster", casterObjID, "target", targetObjID)
}

func (e *HpDrainEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return false
}

func (e *HpDrainEffect) OnExit(casterObjID, targetObjID uint32) {}
