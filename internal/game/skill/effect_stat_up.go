package skill

import (
	"log/slog"
	"strconv"
	"strings"
)

// StatUpEffect is a generic stat increase effect.
// Params: "stat" (e.g. "pAtk"), "type" ("ADD"/"MUL"), "value" (float64).
//
// Phase 5.9.3: Effect Framework.
// Java reference: StatUp.java
type StatUpEffect struct {
	stat  string
	mType StatModType
	value float64
}

func NewStatUpEffect(params map[string]string) Effect {
	stat := params["stat"]
	value, _ := strconv.ParseFloat(params["value"], 64)
	mType := StatModAdd
	if strings.EqualFold(params["type"], "MUL") {
		mType = StatModMul
	}
	return &StatUpEffect{stat: stat, mType: mType, value: value}
}

func (e *StatUpEffect) Name() string    { return "StatUp" }
func (e *StatUpEffect) IsInstant() bool { return false }

func (e *StatUpEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("stat up applied", "stat", e.stat, "value", e.value, "target", targetObjID)
}

func (e *StatUpEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return true
}

func (e *StatUpEffect) OnExit(casterObjID, targetObjID uint32) {
	slog.Debug("stat up removed", "stat", e.stat, "target", targetObjID)
}

// StatModifiers returns the stat modification.
func (e *StatUpEffect) StatModifiers() []StatModifier {
	return []StatModifier{{Stat: e.stat, Type: e.mType, Value: e.value}}
}
