package skill

import (
	"log/slog"
	"strconv"
)

// DispelEffect removes specific buff(s) from target by AbnormalType.
// Params: "abnormalType" (string, e.g. "SPEED_UP"), "count" (int, max buffs to remove).
//
// Java reference: Dispel.java — removes buffs matching condition.
type DispelEffect struct {
	abnormalType string
	count        int
}

func NewDispelEffect(params map[string]string) Effect {
	count, _ := strconv.Atoi(params["count"])
	if count <= 0 {
		count = 1
	}
	return &DispelEffect{
		abnormalType: params["abnormalType"],
		count:        count,
	}
}

func (e *DispelEffect) Name() string    { return "Dispel" }
func (e *DispelEffect) IsInstant() bool { return true }

func (e *DispelEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("dispel effect",
		"abnormalType", e.abnormalType,
		"count", e.count,
		"caster", casterObjID,
		"target", targetObjID)
	// Actual removal done by CastManager via EffectManager.RemoveEffect()
}

func (e *DispelEffect) OnActionTime(_, _ uint32) bool { return false }
func (e *DispelEffect) OnExit(_, _ uint32)            {}
