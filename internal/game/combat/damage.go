package combat

import (
	"math"
	"math/rand"

	"github.com/udisondev/la2go/internal/model"
)

// CalcPhysicalDamage calculates physical damage for auto-attack.
// MVP simplified formula: (76 × pAtk × proximityBonus) / pDef × random × crit(×2).
//
// Parameters:
//   - attacker: player performing attack
//   - target: character being attacked
//   - isCrit: true if critical hit (×2 damage multiplier)
//
// Returns damage value (minimum 1).
//
// MVP simplifications:
//   - No buffs/debuffs
//   - No weapon stats (uses base pAtk/pDef from level)
//   - Proximity bonus always 1.0 (front attack)
//   - No shield block calculation
//
// Phase 5.3: Basic Combat System.
// Java reference: Formulas.calcPhysDam() (line 655-801).
func CalcPhysicalDamage(attacker *model.Player, target *model.Character, isCrit bool) int32 {
	// Base stats (TODO Phase 5.4: load from templates)
	pAtk := float64(attacker.GetBasePAtk())
	pDef := float64(target.GetBasePDef())

	// Proximity bonus (MVP: always 1.0 — front attack)
	// TODO Phase 5.5: implement isBehind/isSide checks
	// Behind: 1.2, Side: 1.1, Front: 1.0
	proximityBonus := 1.0

	// Basic damage formula (Java line 721)
	damage := (76.0 * pAtk * proximityBonus) / pDef

	// Random variance (weapon or unarmed)
	randomMult := getRandomDamageMultiplier(attacker.Level())
	damage *= randomMult

	// Critical hit (×2)
	if isCrit {
		damage *= 2.0
	}

	// Minimum 1 damage
	if damage < 1 {
		damage = 1
	}

	return int32(damage)
}

// getRandomDamageMultiplier returns random damage multiplier for variance.
// For unarmed: 1.0 ± (5 + √level) / 100.
//
// Example values:
//   - Level 1: random=5-6 → range [0.94, 1.06]
//   - Level 10: random=8-9 → range [0.91, 1.09]
//   - Level 80: random=13-14 → range [0.87, 1.13]
//
// Phase 5.3: Basic Combat System.
// Java reference: Creature.getRandomDamageMultiplier() (line 6224-6239).
func getRandomDamageMultiplier(level int32) float64 {
	// Calculate random range (weapon-dependent in full implementation)
	random := 5 + int(math.Sqrt(float64(level)))

	// Returns 1.0 ± (random/100)
	// Example: level 80, random=13 → Rnd.get(26)/100 + 0.87 → [0.87, 1.13]
	return float64(rand.Intn(2*random))/100.0 + 1.0 - float64(random)/100.0
}

// CalcCrit checks if attack is critical hit.
// MVP: base crit rate = 40 (4% chance).
//
// Returns true if random roll < crit rate.
//
// MVP simplifications:
//   - Fixed 4% crit rate
//   - No DEX influence
//   - No buffs/weapon modifiers
//
// Phase 5.3: Basic Combat System.
// Java reference: Formulas.calcCrit() (line 1018-1059).
func CalcCrit(attacker *model.Player, target *model.Character) bool {
	return CalcCritGeneric()
}

// CalcHitMiss checks if attack misses.
// MVP: 80% base hit chance (20% miss rate).
//
// Returns true if attack misses.
//
// MVP simplifications:
//   - Fixed 80% hit chance
//   - No accuracy/evasion stats
//   - No level difference modifier
//   - No position modifiers (behind/side bonus)
//
// Phase 5.3: Basic Combat System.
// Java reference: Formulas.calcHitMiss() (line 1152-1173).
func CalcHitMiss(attacker *model.Player, target *model.Character) bool {
	return CalcHitMissGeneric()
}

// CalcCritGeneric checks if attack is critical hit (attacker-independent).
// MVP: base crit rate = 40 (4% chance).
// Phase 5.7: Generalized for both Player and NPC attacks.
func CalcCritGeneric() bool {
	baseCritRate := 40 // 4% base crit rate
	return rand.Intn(1000) < baseCritRate
}

// CalcHitMissGeneric checks if attack misses (attacker-independent).
// MVP: 80% base hit chance (20% miss rate).
// Phase 5.7: Generalized for both Player and NPC attacks.
func CalcHitMissGeneric() bool {
	hitChance := 800 // 80% base hit chance (out of 1000)
	return rand.Intn(1000) >= hitChance
}
