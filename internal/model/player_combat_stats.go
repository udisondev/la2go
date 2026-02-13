package model

import (
	"math"

	"github.com/udisondev/la2go/internal/data"
)

// GetMAtk returns magic attack power.
// Formula: baseMAtk × INTBonus × levelMod
//
// Java reference: CreatureStat.getMAtk(), FuncMAtkMod.java
func (p *Player) GetMAtk() int32 {
	tmpl := data.GetTemplate(uint8(p.ClassID()))
	if tmpl == nil {
		return 5 // Fallback
	}

	baseMAtk := float64(tmpl.BaseMAtk)
	intBonus := data.GetINTBonus(p.GetINT())
	levelMod := p.GetLevelMod()

	return int32(baseMAtk * intBonus * levelMod)
}

// GetMAtkSpd returns magic attack speed.
// Formula: baseMAtkSpd × WITBonus
//
// Java reference: CreatureStat.getMAtkSpd(), FuncMAtkSpeed.java
func (p *Player) GetMAtkSpd() int32 {
	tmpl := data.GetTemplate(uint8(p.ClassID()))
	if tmpl == nil {
		return 333 // Fallback (default Interlude MAtkSpd)
	}

	baseSpd := float64(tmpl.BaseMAtkSpd)
	witBonus := data.GetWITBonus(p.GetWIT())

	return int32(baseSpd * witBonus)
}

// GetMDef returns magic defense.
// Formula: baseMDef × MENBonus × levelMod
//
// Java reference: CreatureStat.getMDef(), FuncMDefMod.java
func (p *Player) GetMDef() int32 {
	tmpl := data.GetTemplate(uint8(p.ClassID()))
	if tmpl == nil {
		return 100 // Fallback
	}

	baseMDef := float64(tmpl.BaseMDef)
	menBonus := data.GetMENBonus(p.GetMEN())
	levelMod := p.GetLevelMod()

	return int32(baseMDef * menBonus * levelMod)
}

// GetEvasionRate returns evasion rate.
// Formula: sqrt(DEX) × 6 + level
//
// Java reference: FuncAtkEvasion.java
func (p *Player) GetEvasionRate() int32 {
	dex := float64(p.GetDEX())
	return int32(math.Sqrt(dex)*6) + p.Level()
}

// GetAccuracy returns accuracy.
// Formula: sqrt(DEX) × 6 + level + 35
//
// Java reference: FuncAtkAccuracy.java
func (p *Player) GetAccuracy() int32 {
	dex := float64(p.GetDEX())
	return int32(math.Sqrt(dex)*6) + p.Level() + 35
}

// GetCriticalHit returns critical hit rate (promille, e.g. 80 = 8%).
// Uses weapon critRate if equipped, otherwise class template baseCritRate.
// Formula: critRate × DEXBonus
//
// Java reference: FuncAtkCritical.java — uses weapon.getCriticalHit().
func (p *Player) GetCriticalHit() int32 {
	// Try weapon crit rate first (daggers=12, swords=8, bows=12)
	var baseCrit float64
	inv := p.Inventory()
	if inv != nil {
		weapon := inv.GetPaperdollItem(PaperdollRHand)
		if weapon != nil && weapon.Template() != nil && weapon.Template().CritRate > 0 {
			baseCrit = float64(weapon.Template().CritRate)
		}
	}

	// Fallback to class template base crit rate
	if baseCrit == 0 {
		tmpl := data.GetTemplate(uint8(p.ClassID()))
		if tmpl != nil {
			baseCrit = float64(tmpl.BaseCritRate)
		} else {
			baseCrit = 4 // Ultimate fallback
		}
	}

	dexBonus := data.GetDEXBonus(p.GetDEX())
	return int32(baseCrit * dexBonus)
}

// GetAttackSpeedMultiplier returns the attack speed multiplier for animation.
// Formula: 1.1 × PAtkSpd / basePAtkSpd
//
// Java reference: Creature.getAttackSpeedMultiplier()
func (p *Player) GetAttackSpeedMultiplier() float64 {
	tmpl := data.GetTemplate(uint8(p.ClassID()))
	if tmpl == nil || tmpl.BasePAtkSpd == 0 {
		return 1.0
	}
	return 1.1 * p.GetPAtkSpd() / float64(tmpl.BasePAtkSpd)
}

// HasDwarvenCraft returns true if the player can use dwarven craft.
// True for Dwarf race (raceID == 4).
//
// Java reference: Player.hasDwarvenCraft()
func (p *Player) HasDwarvenCraft() bool {
	return p.RaceID() == 4 // Dwarf
}

// GetEnchantEffect returns the visual enchant glow level from equipped weapon.
// Returns 0 if no weapon equipped or if mounted.
//
// Java reference: Player.getEnchantEffect()
func (p *Player) GetEnchantEffect() int32 {
	if p.IsMounted() {
		return 0
	}
	weapon := p.GetEquippedWeapon()
	if weapon == nil {
		return 0
	}
	return weapon.Enchant()
}

// GetInventoryLimit returns the maximum number of inventory slots.
// Default: 80 for all races, 100 for Dwarfs.
//
// Java reference: Player.getInventoryLimit()
func (p *Player) GetInventoryLimit() int32 {
	if p.RaceID() == 4 { // Dwarf
		return 100
	}
	return 80
}

// GetCurrentLoad returns total weight of all items in inventory.
// Inventory weight tracking not yet implemented — returns 0.
func (p *Player) GetCurrentLoad() int32 {
	return 0
}

// GetMaxLoad returns maximum carry weight based on CON stat.
// Formula: 69000 × CONBonus
//
// Java reference: PlayerStat.getMaxLoad()
func (p *Player) GetMaxLoad() int32 {
	conBonus := data.GetCONBonus(p.GetCON())
	return int32(69000.0 * conBonus)
}
