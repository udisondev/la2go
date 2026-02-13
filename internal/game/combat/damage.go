package combat

import (
	"math"
	"math/rand/v2"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

// Shield block result constants.
// Java: Formulas.java SHIELD_DEFENSE_* constants.
const (
	ShieldDefFailed       byte = 0 // Shield did not block
	ShieldDefSucceed      byte = 1 // Normal block: shieldDef added to defense
	ShieldDefPerfectBlock byte = 2 // Perfect block: damage = 1
)

// CalcShieldUse checks if target's shield blocks the attack.
// Interlude formula: shldRate = shield_pDef_proxy * DEX_bonus.
//
// Java reference: Formulas.calcShldUse() (line 1176-1239).
// Java uses Stat.SHIELD_RATE from item XML; our data doesn't store shield_rate directly,
// so we derive it from shield PDef: rate = clamp(pDef/3, 10, 50) — approximates
// the No-Grade(~30%) to S-Grade(~50%) range of Interlude shields.
func CalcShieldUse(target *model.Character) byte {
	shieldDef := getShieldDef(target)
	if shieldDef <= 0 {
		return ShieldDefFailed
	}

	// Derive shield block rate from PDef (proxy for shield_rate stat).
	// Typical Interlude shields: Bronze Shield pDef=36→12%, Eldarake pDef=118→39%.
	shldRate := float64(shieldDef) / 3.0
	if shldRate < 10 {
		shldRate = 10
	}
	if shldRate > 50 {
		shldRate = 50
	}

	// Apply DEX bonus of shield owner (Java: BaseStat.DEX.calcBonus(target))
	if player, ok := target.WorldObject.Data.(*model.Player); ok {
		shldRate *= data.GetDEXBonus(player.GetDEX())
	}

	// Perfect block: 2% chance (Java: PlayerConfig.ALT_PERFECT_SHLD_BLOCK = 2)
	roll := rand.IntN(100)
	if (100 - 2) < roll {
		return ShieldDefPerfectBlock
	}

	// Normal block: shldRate% chance
	if float64(rand.IntN(100)) < shldRate {
		return ShieldDefSucceed
	}

	return ShieldDefFailed
}

// getShieldDef extracts shield defense value from target.
// Returns 0 if target has no shield or is not a Player.
func getShieldDef(target *model.Character) int32 {
	if target == nil || target.WorldObject == nil || target.WorldObject.Data == nil {
		return 0
	}
	player, ok := target.WorldObject.Data.(*model.Player)
	if !ok {
		return 0
	}
	inv := player.Inventory()
	if inv == nil {
		return 0
	}
	shield := inv.GetPaperdollItem(model.PaperdollLHand)
	if shield == nil || shield.Template() == nil {
		return 0
	}
	return shield.Template().PDef
}

// CalcPhysicalDamage calculates physical damage for auto-attack.
// Interlude formula: (76 × pAtk × ssBoost × proximityBonus) / pDef × random × crit(×2).
//
// Parameters:
//   - attacker: player performing attack
//   - target: character being attacked
//   - isCrit: true if critical hit (×2 damage multiplier in Interlude)
//   - ss: true if soulshot active (×2 P.Atk boost)
//   - shieldResult: result of CalcShieldUse (0=none, 1=block, 2=perfect)
//   - targetPDef: target's physical defense
//
// Returns damage value (minimum 1).
//
// Java reference: Formulas.calcPhysDam() (line 655-801).
func CalcPhysicalDamage(attacker *model.Player, target *model.Character, isCrit bool, ss bool, shieldResult byte, targetPDef int32) int32 {
	// Perfect shield block → damage = 1
	if shieldResult == ShieldDefPerfectBlock {
		return 1
	}

	pAtk := float64(attacker.GetPAtk())

	// Soulshot: ×2 P.Atk (Java: Formulas.java line 686-688)
	if ss {
		pAtk *= 2.0
	}

	pDef := float64(targetPDef)

	// Shield block: add shieldDef to defense (Java: line 670-684)
	if shieldResult == ShieldDefSucceed {
		shieldDef := getShieldDef(target)
		pDef += float64(shieldDef)
	}

	if pDef < 1 {
		pDef = 1
	}

	// Proximity bonus: Behind ×1.2, Side ×1.1, Front ×1.0
	// Java reference: Formulas.calcPhysDam() line 667:
	//   proximityBonus = attacker.isBehind(target) ? 1.2 : attacker.isInFrontOf(target) ? 1 : 1.1
	proximityBonus := CalcProximityBonus(attacker, target)

	// Basic damage formula (Java line 721)
	damage := (76.0 * pAtk * proximityBonus) / pDef

	// Random variance — use weapon randomDamage if available
	randomDmg := int32(0)
	inv := attacker.Inventory()
	if inv != nil {
		weapon := inv.GetPaperdollItem(model.PaperdollRHand)
		if weapon != nil && weapon.Template() != nil {
			randomDmg = weapon.Template().RandomDamage
		}
	}
	damage *= getRandomDamageMultiplierWeapon(randomDmg)

	// Critical hit: base ×2 in Interlude (NOT weapon-dependent — that's Gracia+)
	// Java: Formulas.java line 689-718: base multiplier = 2
	// Final = base × (1 + CRITICAL_DAMAGE_bonus) × (1 - DEFENCE_CRITICAL_DAMAGE_bonus)
	if isCrit {
		critMul := 2.0

		// Apply attacker's critDamage stat bonus from buffs/items
		if em := attacker.EffectManager(); em != nil {
			if bonus := em.GetStatBonus("critDamage"); bonus != 0 {
				critMul *= (1.0 + bonus)
			}
		}

		// Apply target's defCritDamage stat (reduces crit damage)
		if target != nil && target.WorldObject != nil && target.WorldObject.Data != nil {
			if targetPlayer, ok := target.WorldObject.Data.(*model.Player); ok {
				if em := targetPlayer.EffectManager(); em != nil {
					if defBonus := em.GetStatBonus("defCritDamage"); defBonus != 0 {
						critMul *= (1.0 - defBonus)
					}
				}
			}
		}

		if critMul < 1.0 {
			critMul = 1.0
		}
		damage *= critMul
	}

	// Minimum 1 damage
	if damage < 1 {
		damage = 1
	}

	return int32(damage)
}

// CalcMagicDamage calculates magic skill damage.
// Interlude formula: (91 × sqrt(mAtk)) / mDef × skillPower × spiritShotBoost × random.
//
// Parameters:
//   - mAtk: attacker's magic attack
//   - mDef: target's magic defense
//   - skillPower: skill's base power value
//   - sps: true if Spirit Shot active (×2 M.Atk)
//   - bss: true if Blessed Spirit Shot active (×4 M.Atk)
//   - mcrit: true if magic critical hit
//   - isPvP: true if PvP combat (affects mcrit multiplier)
//   - attackerLevel: attacker's level (for random variance)
//
// Returns damage value (minimum 1).
//
// Java reference: Formulas.calcMagicDam() (line 803-924).
func CalcMagicDamage(mAtk, mDef, skillPower float64, sps, bss, mcrit, isPvP bool, attackerLevel int32) int32 {
	if mDef < 1 {
		mDef = 1
	}

	// Spirit shot boost (Java: line 838)
	if bss {
		mAtk *= 4.0
	} else if sps {
		mAtk *= 2.0
	}

	// Main formula: (91 × sqrt(mAtk)) / mDef × skillPower
	damage := (91.0 * math.Sqrt(mAtk)) / mDef * skillPower

	// Magic critical: ×2.5 PvP, ×3 PvE (Java: line 848-851)
	if mcrit {
		if isPvP {
			damage *= 2.5
		} else {
			damage *= 3.0
		}
	}

	// Random variance
	damage *= getRandomDamageMultiplier(attackerLevel)

	if damage < 1 {
		damage = 1
	}

	return int32(damage)
}

// getRandomDamageMultiplier returns random damage multiplier for variance.
// For unarmed: 1.0 ± (5 + √level) / 100.
//
// Java reference: Creature.getRandomDamageMultiplier() (line 6224-6239).
func getRandomDamageMultiplier(level int32) float64 {
	random := 5 + int(math.Sqrt(float64(level)))
	return float64(rand.IntN(2*random))/100.0 + 1.0 - float64(random)/100.0
}

// getRandomDamageMultiplierWeapon returns random damage multiplier using weapon data.
// If weapon has randomDamage, uses weapon.randomDamage/2 for variance range.
// Otherwise falls back to unarmed formula.
//
// Java reference: Creature.getRandomDamageMultiplier() (line 6224-6239).
func getRandomDamageMultiplierWeapon(weaponRandomDamage int32) float64 {
	if weaponRandomDamage <= 0 {
		// Unarmed fallback: fixed 10% variance
		random := 10
		return float64(rand.IntN(2*random))/100.0 + 1.0 - float64(random)/100.0
	}
	random := int(weaponRandomDamage) / 2
	if random < 1 {
		random = 1
	}
	return float64(rand.IntN(2*random))/100.0 + 1.0 - float64(random)/100.0
}

// CalcCrit checks if attack is critical hit using weapon crit rate × DEX bonus.
// Rate is per-mille: e.g. critRate=80 means 8% chance.
//
// Java reference: Formulas.calcCrit() (line 1018-1059).
func CalcCrit(attacker *model.Player, target *model.Character) bool {
	critRate := attacker.GetCriticalHit() // baseCritRate × DEXBonus (per-mille /1000)
	if critRate < 0 {
		critRate = 0
	}
	return rand.IntN(1000) < int(critRate)
}

// CalcHitMiss checks if attack misses.
// Interlude formula: hitRate = 80 + 2*(accuracy - evasion), clamped [20%, 98%].
//
// Java reference: Formulas.calcHitMiss() (line 1152-1163).
// Full Java: chance = (80 + 2*(accuracy - evasion)) * 10, clamp [200, 980], miss if < Rnd.get(1000)
func CalcHitMiss(attacker *model.Player, target *model.Character) bool {
	accuracy := getAccuracy(attacker)
	evasion := getEvasion(target)

	// Java: chance = (80 + 2*(accuracy - evasion)) * 10
	chance := (80 + 2*(accuracy-evasion)) * 10

	// Clamp: min 20% (200), max 98% (980)
	if chance < 200 {
		chance = 200
	}
	if chance > 980 {
		chance = 980
	}

	// Miss if roll >= chance
	return rand.IntN(1000) >= chance
}

// getAccuracy returns attacker's accuracy rating.
// Java reference: FuncAtkAccuracy.java:42-59
// Formula: sqrt(DEX) * 6 + level + high-level bonuses.
func getAccuracy(attacker *model.Player) int {
	dex := int(attacker.GetDEX())
	level := int(attacker.Level())

	value := int(math.Sqrt(float64(dex)))*6 + level

	// High-level bonuses (Java: FuncAtkAccuracy lines 52-57)
	if level > 77 {
		value += level - 76
	}
	if level > 69 {
		value += level - 69
	}

	return value
}

// getEvasion returns target's evasion rating.
// Java reference: FuncAtkEvasion.java:42-72
// Players: sqrt(DEX) * 6 + level + high-level bonuses (×1.2 at 78+).
// NPCs: sqrt(DEX) * 6 + level + (level-69)+2 bonus.
func getEvasion(target *model.Character) int {
	if target == nil || target.WorldObject == nil || target.WorldObject.Data == nil {
		return 0
	}

	// Player evasion with high-level scaling
	if player, ok := target.WorldObject.Data.(*model.Player); ok {
		dex := int(player.GetDEX())
		level := int(player.Level())
		value := int(math.Sqrt(float64(dex)))*6 + level

		if level >= 70 {
			diff := float64(level - 69)
			if level >= 78 {
				diff *= 1.2
			}
			value += int(diff)
		}
		return value
	}

	// NPC evasion (default DEX=30)
	level := int(target.Level())
	value := int(math.Sqrt(30.0))*6 + level
	if level > 69 {
		value += (level - 69) + 2
	}
	return value
}

// CalcCritGeneric checks if attack is critical hit (attacker-independent).
// MVP: base crit rate = 40 (4% chance).
// Phase 5.7: Generalized for both Player and NPC attacks.
var CalcCritGeneric = func() bool {
	baseCritRate := 40 // 4% base crit rate
	return rand.IntN(1000) < baseCritRate
}

// CalcHitMissGeneric checks if attack misses (attacker-independent).
// Used for NPC attacks where we don't have accuracy/evasion stats.
// MVP: 80% base hit chance (20% miss rate).
// Phase 5.7: Generalized for both Player and NPC attacks.
var CalcHitMissGeneric = func() bool {
	hitChance := 800 // 80% base hit chance (out of 1000)
	return rand.IntN(1000) >= hitChance
}

// --- Position / Proximity bonus ---

// Position represents relative position of attacker to target.
type Position int

const (
	PositionFront Position = iota
	PositionSide
	PositionBack
)

// CalcProximityBonus returns the proximity (position) damage multiplier.
// Behind: ×1.2, Side: ×1.1, Front: ×1.0.
// Java reference: Formulas.calcPhysDam() line 667.
func CalcProximityBonus(attacker *model.Player, target *model.Character) float64 {
	pos := GetPosition(attacker.Location(), target.Location())
	switch pos {
	case PositionBack:
		return 1.2
	case PositionSide:
		return 1.1
	default:
		return 1.0
	}
}

// GetPosition calculates relative position of from→to using heading.
// Java reference: Position.getPosition() in creature/Position.java.
//
// Algorithm:
//  1. Calculate heading from attacker toward target: atan2(dy, dx) mapped to uint16.
//  2. Compute heading difference: abs(targetHeading - headingToTarget).
//  3. Classify into FRONT/SIDE/BACK by heading sector.
func GetPosition(from, to model.Location) Position {
	// Calculate heading from 'from' toward 'to' (Java: calculateHeadingTo)
	dx := float64(to.X - from.X)
	dy := float64(to.Y - from.Y)
	headingTo := int(math.Floor(math.Atan2(dy, dx) * 65535.0 / (2.0 * math.Pi)))

	// Heading difference: how far off the target's facing direction is the attacker
	heading := int(to.Heading) - headingTo
	if heading < 0 {
		heading = -heading
	}
	// Normalize to uint16 range [0, 65535]
	heading &= 0xFFFF

	// Java Position.getPosition() sector classification:
	// SIDE: [0x2000, 0x6000] or [0xA000, 0xE000]
	// FRONT: [0x2000, 0xE000] excluding SIDE ranges
	// BACK: everything else (near 0 or near 0xFFFF)
	if (heading >= 0x2000 && heading <= 0x6000) || (heading >= 0xA000 && heading <= 0xE000) {
		return PositionSide
	}
	if heading >= 0x2000 && heading <= 0xE000 {
		return PositionFront
	}
	return PositionBack
}
