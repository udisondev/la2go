package combat

import (
	"fmt"

	"github.com/udisondev/la2go/internal/model"
)

// DefaultMeleeRange is the fallback melee attack range when no weapon is equipped.
// Java: Formulas.java:99 — MELEE_ATTACK_RANGE = 40
const DefaultMeleeRange = 40

// ValidateAttack validates attack request before initiating combat.
// Returns error if validation fails (attack should not proceed).
//
// Checks:
//   - Target exists (not nil)
//   - Attacker alive
//   - Attacker not casting
//   - Peace zone check (attacker or target in peace zone)
//   - Target alive (for creatures)
//   - Target in attack range (weapon-dependent: fist=20, sword=40, bow=500)
//
// Phase 5.3: Basic Combat System (MVP validation).
// Java reference: Creature.doAttack() (line 1011-1070), onForcedAttack() (line 5291-5341).
func ValidateAttack(attacker *model.Player, target *model.WorldObject) error {
	// 1. Target exists
	if target == nil {
		return fmt.Errorf("target is nil")
	}

	// 2. Attacker alive
	if attacker.IsDead() {
		return fmt.Errorf("attacker is dead")
	}

	// 3. Attacker not casting
	if attacker.IsCasting() {
		return fmt.Errorf("attacker is casting")
	}

	// 4. Peace zone check: no PvP in peace zones
	// Java: Creature.doAttack — isInsideZone(ZoneId.PEACE) check
	character := getCharacterFromObject(target)
	if attacker.IsInsideZone(model.ZoneIDPeace) {
		return fmt.Errorf("attacker is in peace zone")
	}
	if character != nil && character.IsInsideZone(model.ZoneIDPeace) {
		return fmt.Errorf("target is in peace zone")
	}

	// 5. Target alive (only for creatures)
	if character != nil && character.IsDead() {
		return fmt.Errorf("target is dead")
	}

	// 6. Range check (weapon-dependent)
	if !IsInAttackRange(attacker, target) {
		return fmt.Errorf("target out of attack range")
	}

	return nil
}

// IsInAttackRange checks if target is within physical attack range.
// Range is weapon-dependent: fist=20, sword=40, bow=500 (from Player.GetAttackRange).
//
// Java reference: CreatureStat.getPhysicalAttackRange() (line 591-605).
func IsInAttackRange(attacker *model.Player, target *model.WorldObject) bool {
	attackerLoc := attacker.Location()
	targetLoc := target.Location()
	distSq := attackerLoc.DistanceSquared(targetLoc)

	attackRange := attacker.GetAttackRange()
	if attackRange < DefaultMeleeRange {
		attackRange = DefaultMeleeRange
	}

	return distSq <= int64(attackRange)*int64(attackRange)
}

// getCharacterFromObject attempts to extract Character from WorldObject via type assertion.
// Returns nil if object is not a Character (e.g., dropped item, door).
//
// Type assertion order: RaidBoss → GrandBoss → Monster → Npc → Player.
func getCharacterFromObject(obj *model.WorldObject) *model.Character {
	if obj == nil || obj.Data == nil {
		return nil
	}

	if rb, ok := obj.Data.(*model.RaidBoss); ok {
		return rb.Monster.Character
	}
	if gb, ok := obj.Data.(*model.GrandBoss); ok {
		return gb.Monster.Character
	}
	if monster, ok := obj.Data.(*model.Monster); ok {
		return monster.Character
	}
	if npc, ok := obj.Data.(*model.Npc); ok {
		return npc.Character
	}
	if player, ok := obj.Data.(*model.Player); ok {
		return player.Character
	}

	return nil
}
