package skill

import (
	"log/slog"
	"strconv"
)

// ReflectEffect reflects a percentage of damage back to the attacker.
// Params: "reflectPercent" (float64, 0.0-1.0), "reflectType" (string: "PHYSICAL", "MAGICAL", "ALL").
//
// Java reference: Reflect.java — applies damage reflection modifier.
type ReflectEffect struct {
	reflectPercent float64
	reflectType    string
}

func NewReflectEffect(params map[string]string) Effect {
	pct, _ := strconv.ParseFloat(params["reflectPercent"], 64)
	if pct == 0 {
		pct = 0.2 // Default 20% reflect
	}
	rtype := params["reflectType"]
	if rtype == "" {
		rtype = "PHYSICAL"
	}
	return &ReflectEffect{
		reflectPercent: pct,
		reflectType:    rtype,
	}
}

func (e *ReflectEffect) Name() string    { return "Reflect" }
func (e *ReflectEffect) IsInstant() bool { return false }

func (e *ReflectEffect) OnStart(casterObjID, targetObjID uint32) {
	slog.Debug("reflect started",
		"percent", e.reflectPercent,
		"type", e.reflectType,
		"target", targetObjID)
}

func (e *ReflectEffect) OnActionTime(_, _ uint32) bool {
	return true // Continues until duration expires
}

func (e *ReflectEffect) OnExit(_, targetObjID uint32) {
	slog.Debug("reflect ended", "target", targetObjID)
}

// StatModifiers implements StatModifierProvider for the damage reflection modifier.
func (e *ReflectEffect) StatModifiers() []StatModifier {
	stat := "reflectDamagePhysical"
	if e.reflectType == "MAGICAL" {
		stat = "reflectDamageMagical"
	} else if e.reflectType == "ALL" {
		return []StatModifier{
			{Stat: "reflectDamagePhysical", Type: StatModAdd, Value: e.reflectPercent},
			{Stat: "reflectDamageMagical", Type: StatModAdd, Value: e.reflectPercent},
		}
	}
	return []StatModifier{{Stat: stat, Type: StatModAdd, Value: e.reflectPercent}}
}
