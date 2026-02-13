package skill

import (
	"log/slog"
	"strconv"
)

// CubicEffect creates a floating cubic that auto-casts skills on the caster's target.
// Params: "cubicID" (int32), "cubicPower" (float64), "cubicDuration" (int32 seconds).
//
// Java reference: SignetMDam.java, CubicSkill.java — spawns cubic with periodic skill trigger.
type CubicEffect struct {
	cubicID       int32
	cubicPower    float64
	cubicDuration int32
}

func NewCubicEffect(params map[string]string) Effect {
	cid, _ := strconv.Atoi(params["cubicID"])
	power, _ := strconv.ParseFloat(params["cubicPower"], 64)
	dur, _ := strconv.Atoi(params["cubicDuration"])
	if dur <= 0 {
		dur = 900 // Default 15 minutes
	}
	return &CubicEffect{
		cubicID:       int32(cid),
		cubicPower:    power,
		cubicDuration: int32(dur),
	}
}

func (e *CubicEffect) Name() string    { return "Cubic" }
func (e *CubicEffect) IsInstant() bool { return true }

func (e *CubicEffect) OnStart(casterObjID, _ uint32) {
	slog.Debug("cubic spawned",
		"cubicID", e.cubicID,
		"power", e.cubicPower,
		"duration", e.cubicDuration,
		"owner", casterObjID)
	// Actual cubic creation handled by CastManager
}

func (e *CubicEffect) OnActionTime(_, _ uint32) bool { return false }
func (e *CubicEffect) OnExit(_, _ uint32)            {}

// CubicID returns the cubic template ID.
func (e *CubicEffect) CubicID() int32 { return e.cubicID }
