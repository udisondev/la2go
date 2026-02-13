package skill

import (
	"log/slog"
	"strconv"
)

// ResurrectEffect revives a dead player/NPC with a percentage of HP/MP.
// Params: "power" (float64, % of max HP to restore, 0.0-1.0).
//
// Java reference: Resurrection.java — revives target with power% HP.
type ResurrectEffect struct {
	power float64 // HP restore percentage (0.0-1.0)
}

func NewResurrectEffect(params map[string]string) Effect {
	power, _ := strconv.ParseFloat(params["power"], 64)
	if power <= 0 {
		power = 0.2 // Default 20% HP restore
	}
	if power > 1.0 {
		power = 1.0
	}
	return &ResurrectEffect{power: power}
}

func (e *ResurrectEffect) Name() string    { return "Resurrect" }
func (e *ResurrectEffect) IsInstant() bool { return true }

func (e *ResurrectEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("resurrect",
		"power", e.power,
		"caster", casterObjID,
		"target", targetObjID)
	// Actual resurrection handled by CastManager (needs Player access to setDead(false))
}

func (e *ResurrectEffect) OnActionTime(_, _ uint32) bool { return false }
func (e *ResurrectEffect) OnExit(_, _ uint32)            {}

// Power returns the HP restore percentage.
func (e *ResurrectEffect) Power() float64 { return e.power }
