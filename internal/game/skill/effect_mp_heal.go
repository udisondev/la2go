package skill

import (
	"log/slog"
	"strconv"
)

// MpHealEffect restores MP instantly.
// Params: "power" (float64).
//
// Phase 5.9.3: Effect Framework.
// Java reference: ManaHeal.java
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
	slog.Debug("mp heal", "power", e.power, "caster", casterObjID, "target", targetObjID)
}

func (e *MpHealEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return false
}

func (e *MpHealEffect) OnExit(casterObjID, targetObjID uint32) {}
