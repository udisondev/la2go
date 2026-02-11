package skill

import (
	"log/slog"
	"strconv"
	"strings"
)

// BuffEffect applies stat modifiers for a duration.
// Params: "stat" (e.g. "pAtk"), "type" ("ADD"/"MUL"), "value" (float64).
//
// Phase 5.9.3: Effect Framework.
// Java reference: Buff.java
type BuffEffect struct {
	stat  string
	mType StatModType
	value float64
}

func NewBuffEffect(params map[string]string) Effect {
	stat := params["stat"]
	value, _ := strconv.ParseFloat(params["value"], 64)

	mType := StatModAdd
	if strings.EqualFold(params["type"], "MUL") {
		mType = StatModMul
	}

	return &BuffEffect{stat: stat, mType: mType, value: value}
}

func (e *BuffEffect) Name() string    { return "Buff" }
func (e *BuffEffect) IsInstant() bool { return false }

func (e *BuffEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("buff applied", "stat", e.stat, "value", e.value, "target", targetObjID)
}

func (e *BuffEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return true // Buff just exists, no periodic action
}

func (e *BuffEffect) OnExit(casterObjID, targetObjID uint32) {
	slog.Debug("buff removed", "stat", e.stat, "target", targetObjID)
}

// StatModifiers returns the stat modifiers applied by this buff.
func (e *BuffEffect) StatModifiers() []StatModifier {
	return []StatModifier{{Stat: e.stat, Type: e.mType, Value: e.value}}
}
