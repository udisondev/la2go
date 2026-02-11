package skill

import (
	"log/slog"
	"strconv"
	"strings"
)

// SpeedChangeEffect modifies movement speed for a duration.
// Params: "value" (float64), "type" ("ADD"/"MUL", default "ADD").
//
// Phase 5.9.3: Effect Framework.
// Java reference: Speed.java
type SpeedChangeEffect struct {
	mType StatModType
	value float64
}

func NewSpeedChangeEffect(params map[string]string) Effect {
	value, _ := strconv.ParseFloat(params["value"], 64)
	mType := StatModAdd
	if strings.EqualFold(params["type"], "MUL") {
		mType = StatModMul
	}
	return &SpeedChangeEffect{mType: mType, value: value}
}

func (e *SpeedChangeEffect) Name() string    { return "SpeedChange" }
func (e *SpeedChangeEffect) IsInstant() bool { return false }

func (e *SpeedChangeEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("speed change applied", "value", e.value, "target", targetObjID)
}

func (e *SpeedChangeEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return true
}

func (e *SpeedChangeEffect) OnExit(casterObjID, targetObjID uint32) {
	slog.Debug("speed change removed", "target", targetObjID)
}

// StatModifiers returns speed modification.
func (e *SpeedChangeEffect) StatModifiers() []StatModifier {
	return []StatModifier{{Stat: "speed", Type: e.mType, Value: e.value}}
}
