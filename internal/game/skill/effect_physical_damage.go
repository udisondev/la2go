package skill

import (
	"log/slog"
	"strconv"
)

// PhysicalDamageEffect deals instant physical damage.
// Params: "power" (float64).
//
// Phase 5.9.3: Effect Framework.
// Java reference: PhysicalDamage.java
type PhysicalDamageEffect struct {
	power float64
}

func NewPhysicalDamageEffect(params map[string]string) Effect {
	power, _ := strconv.ParseFloat(params["power"], 64)
	return &PhysicalDamageEffect{power: power}
}

func (e *PhysicalDamageEffect) Name() string    { return "PhysicalDamage" }
func (e *PhysicalDamageEffect) IsInstant() bool { return true }

func (e *PhysicalDamageEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("physical damage", "power", e.power, "caster", casterObjID, "target", targetObjID)
	// Actual damage calculation is done in CastManager using combat formulas
}

func (e *PhysicalDamageEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return false // Instant
}

func (e *PhysicalDamageEffect) OnExit(casterObjID, targetObjID uint32) {}
