package skill

import (
	"log/slog"
	"math"
	"math/rand/v2"
	"strconv"
)

// PhysicalDamageEffect deals instant physical skill damage.
// Params: "power" (float64) — skill power multiplier.
//
// Phase 5.9.3+: Actual damage using physical skill formula.
// Java reference: PhysicalDamage.java — onStart().
// Formula: (power × pAtk × 70) / pDef × random variance.
type PhysicalDamageEffect struct {
	power float64
}

func NewPhysicalDamageEffect(params map[string]string) Effect {
	power, _ := strconv.ParseFloat(params["power"], 64)
	return &PhysicalDamageEffect{power: power}
}

func (e *PhysicalDamageEffect) Name() string    { return "PhysicalDamage" }
func (e *PhysicalDamageEffect) IsInstant() bool { return true }

func (e *PhysicalDamageEffect) OnStart(casterObjID, targetObjID uint32) {
	caster := resolvePlayer(casterObjID)
	target := resolveCharacter(targetObjID)
	if caster == nil || target == nil || target.IsDead() {
		return
	}

	pAtk := float64(caster.GetPAtk())
	pDef := float64(targetPDef(targetObjID))
	if pDef < 1 {
		pDef = 1
	}

	// Physical skill damage formula (Interlude).
	// Java: Formulas.calcPhysDam — skill variant uses skill power as multiplier.
	// Simplified: damage = power × pAtk × 70 / pDef × random
	damage := e.power * pAtk * 70.0 / pDef

	// Random variance ±10%
	variance := 0.9 + rand.Float64()*0.2
	damage *= variance

	dmg := int32(math.Max(damage, 1))
	target.ReduceCurrentHP(dmg)

	slog.Debug("physical damage dealt",
		"power", e.power,
		"damage", dmg,
		"caster", casterObjID,
		"target", targetObjID)
}

func (e *PhysicalDamageEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return false
}

func (e *PhysicalDamageEffect) OnExit(casterObjID, targetObjID uint32) {}
