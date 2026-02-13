package skill

import (
	"log/slog"
	"strconv"

	"github.com/udisondev/la2go/internal/game/combat"
)

// MagicalDamageEffect deals instant magical damage.
// Params: "power" (float64) — skill power (damage base).
//
// Phase 5.9.3+: Actual damage using CalcMagicDamage formula.
// Java reference: MagicalDamage.java — onStart().
type MagicalDamageEffect struct {
	power float64
}

func NewMagicalDamageEffect(params map[string]string) Effect {
	power, _ := strconv.ParseFloat(params["power"], 64)
	return &MagicalDamageEffect{power: power}
}

func (e *MagicalDamageEffect) Name() string    { return "MagicalDamage" }
func (e *MagicalDamageEffect) IsInstant() bool { return true }

func (e *MagicalDamageEffect) OnStart(casterObjID, targetObjID uint32) {
	caster := resolvePlayer(casterObjID)
	target := resolveCharacter(targetObjID)
	if caster == nil || target == nil || target.IsDead() {
		return
	}

	mAtk := float64(caster.GetMAtk())
	mDef := float64(targetMDef(targetObjID))

	// Calculate magic damage using Interlude formula.
	// No SPS/BSS/MCrit for now (would need shot state tracking).
	// Java: Formulas.calcMagicDam(effector, effected, skill, shld, sps, bss, mcrit)
	damage := combat.CalcMagicDamage(mAtk, mDef, e.power,
		false, // sps
		false, // bss
		false, // mcrit
		false, // isPvP (simplified)
		caster.Level())

	if damage <= 0 {
		return
	}

	target.ReduceCurrentHP(damage)

	slog.Debug("magical damage dealt",
		"power", e.power,
		"damage", damage,
		"caster", casterObjID,
		"target", targetObjID)
}

func (e *MagicalDamageEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return false
}

func (e *MagicalDamageEffect) OnExit(casterObjID, targetObjID uint32) {}
