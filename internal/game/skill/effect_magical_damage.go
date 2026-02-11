package skill

import (
	"log/slog"
	"strconv"
)

// MagicalDamageEffect deals instant magical damage.
// Params: "power" (float64).
//
// Phase 5.9.3: Effect Framework.
// Java reference: MagicalDamage.java
type MagicalDamageEffect struct {
	power float64
}

func NewMagicalDamageEffect(params map[string]string) Effect {
	power, _ := strconv.ParseFloat(params["power"], 64)
	return &MagicalDamageEffect{power: power}
}

func (e *MagicalDamageEffect) Name() string    { return "MagicalDamage" }
func (e *MagicalDamageEffect) IsInstant() bool { return true }

func (e *MagicalDamageEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("magical damage", "power", e.power, "caster", casterObjID, "target", targetObjID)
}

func (e *MagicalDamageEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return false
}

func (e *MagicalDamageEffect) OnExit(casterObjID, targetObjID uint32) {}
