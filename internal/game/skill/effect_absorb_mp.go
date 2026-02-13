package skill

import (
	"log/slog"
	"strconv"
)

// AbsorbMPEffect drains MP from target and gives it to the caster.
// Params: "power" (float64 — base MP drain amount).
//
// Java reference: ManaHeal.java / ManaDam.java — MP drain/transfer.
type AbsorbMPEffect struct {
	power float64
}

func NewAbsorbMPEffect(params map[string]string) Effect {
	power, _ := strconv.ParseFloat(params["power"], 64)
	return &AbsorbMPEffect{power: power}
}

func (e *AbsorbMPEffect) Name() string    { return "AbsorbMP" }
func (e *AbsorbMPEffect) IsInstant() bool { return true }

func (e *AbsorbMPEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("absorb MP",
		"power", e.power,
		"caster", casterObjID,
		"target", targetObjID)
	// Actual MP transfer handled by CastManager (needs Player access)
}

func (e *AbsorbMPEffect) OnActionTime(_, _ uint32) bool { return false }
func (e *AbsorbMPEffect) OnExit(_, _ uint32)            {}
