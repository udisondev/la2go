package skill

import (
	"log/slog"
	"strconv"
)

// NegateEffect cancels an entire category of buffs/debuffs on the target.
// Params: "negateType" (string: "BUFF", "DEBUFF", "ALL"), "count" (int, max to remove).
//
// Java reference: Negate.java — mass buff cancellation.
type NegateEffect struct {
	negateType string
	count      int
}

func NewNegateEffect(params map[string]string) Effect {
	negateType := params["negateType"]
	if negateType == "" {
		negateType = "BUFF"
	}
	count, _ := strconv.Atoi(params["count"])
	if count <= 0 {
		count = 5
	}
	return &NegateEffect{
		negateType: negateType,
		count:      count,
	}
}

func (e *NegateEffect) Name() string    { return "Negate" }
func (e *NegateEffect) IsInstant() bool { return true }

func (e *NegateEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("negate effect",
		"negateType", e.negateType,
		"count", e.count,
		"caster", casterObjID,
		"target", targetObjID)
	// Actual removal done by CastManager via EffectManager
}

func (e *NegateEffect) OnActionTime(_, _ uint32) bool { return false }
func (e *NegateEffect) OnExit(_, _ uint32)            {}
