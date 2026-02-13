package skill

import (
	"log/slog"
	"strconv"

	"github.com/udisondev/la2go/internal/game/combat"
)

// HpDrainEffect deals magical damage and heals caster for a percentage of damage dealt.
// Params: "power" (float64 — skill power), "absorbPercent" (float64, 0.0-1.0 — drain ratio).
//
// Phase 5.9.3+: Actual drain logic — deal damage AND heal caster.
// Java reference: HpDrain.java — onStart().
// Java's "_power" field is absorb ratio (0.0-1.0), named "absorbPercent" in Go.
type HpDrainEffect struct {
	power         float64
	absorbPercent float64
}

func NewHpDrainEffect(params map[string]string) Effect {
	power, _ := strconv.ParseFloat(params["power"], 64)
	absorbPercent, _ := strconv.ParseFloat(params["absorbPercent"], 64)
	if absorbPercent == 0 {
		absorbPercent = 0.5 // Default 50% absorb
	}
	return &HpDrainEffect{power: power, absorbPercent: absorbPercent}
}

func (e *HpDrainEffect) Name() string    { return "HpDrain" }
func (e *HpDrainEffect) IsInstant() bool { return true }

func (e *HpDrainEffect) OnStart(casterObjID, targetObjID uint32) {
	caster := resolvePlayer(casterObjID)
	casterChar := resolveCharacter(casterObjID)
	target := resolveCharacter(targetObjID)
	if caster == nil || casterChar == nil || target == nil || target.IsDead() {
		return
	}

	mAtk := float64(caster.GetMAtk())
	mDef := float64(targetMDef(targetObjID))

	// Calculate magic damage (same as MagicalDamage).
	damage := combat.CalcMagicDamage(mAtk, mDef, e.power,
		false, false, false, false, caster.Level())

	if damage <= 0 {
		return
	}

	// Deal damage to target.
	target.ReduceCurrentHP(damage)

	// Heal caster by absorbPercent of damage dealt.
	// Java: hpAdd = _power * drain (where drain = damage dealt)
	// Capped at caster's max HP.
	hpAdd := int32(float64(damage) * e.absorbPercent)
	if hpAdd > 0 {
		currentHP := casterChar.CurrentHP()
		maxHP := casterChar.MaxHP()
		if currentHP+hpAdd > maxHP {
			hpAdd = maxHP - currentHP
		}
		if hpAdd > 0 {
			casterChar.SetCurrentHP(currentHP + hpAdd)
		}
	}

	slog.Debug("hp drain applied",
		"damage", damage,
		"healed", hpAdd,
		"absorbPercent", e.absorbPercent,
		"caster", casterObjID,
		"target", targetObjID)
}

func (e *HpDrainEffect) OnActionTime(casterObjID, targetObjID uint32) bool {
	return false
}

func (e *HpDrainEffect) OnExit(casterObjID, targetObjID uint32) {}
