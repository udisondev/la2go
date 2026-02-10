package combat

import (
	"fmt"

	"github.com/udisondev/la2go/internal/model"
)

// MaxPhysicalAttackRange is the maximum physical attack range (units).
// TODO Phase 5.3: Make this weapon-dependent (sword=40, bow=500, etc).
const MaxPhysicalAttackRange = 100

// MaxPhysicalAttackRangeSquared is the squared attack range for performance.
const MaxPhysicalAttackRangeSquared = MaxPhysicalAttackRange * MaxPhysicalAttackRange

// ValidateAttack validates attack request before initiating combat.
// Returns error if validation fails (attack should not proceed).
//
// Checks:
//   - Target exists (not nil)
//   - Attacker alive
//   - Target alive (for creatures)
//   - Target in attack range (MaxPhysicalAttackRange from Phase 5.2)
//   - Target is targetable (TODO Phase 5.4)
//   - Peace zone check (TODO Phase 5.5)
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

	// 3. Target alive (only for creatures)
	// MVP: assume all WorldObjects except items are Characters
	// TODO Phase 5.4: implement proper type checking (target.IsCreature())
	character := getCharacterFromObject(target)
	if character != nil && character.IsDead() {
		return fmt.Errorf("target is dead")
	}

	// 4. Range check
	// MaxPhysicalAttackRange = 100 units
	if !IsInAttackRange(attacker, target) {
		return fmt.Errorf("target out of attack range")
	}

	// 5. Target is targetable (not hidden/GM/etc)
	// TODO Phase 5.4: implement IsTargetable() method
	// if !target.IsTargetable() {
	//     return fmt.Errorf("target not targetable")
	// }

	// 6. Peace zone check
	// TODO Phase 5.5: implement peace zone system
	// if attacker.IsInsidePeaceZone() || target.IsInsidePeaceZone() {
	//     return fmt.Errorf("cannot attack in peace zone")
	// }

	return nil
}

// IsInAttackRange checks if target is within physical attack range.
// Returns true if distance <= MaxPhysicalAttackRange (100 units).
//
// TODO Phase 5.4: Make this weapon-dependent (sword=40, bow=500, etc).
//
// Phase 5.3: Basic Combat System (simplified, fixed range).
// Moved from gameserver package to avoid import cycle.
func IsInAttackRange(attacker *model.Player, target *model.WorldObject) bool {
	attackerLoc := attacker.Location()
	targetLoc := target.Location()
	distSq := attackerLoc.DistanceSquared(targetLoc)

	return distSq <= MaxPhysicalAttackRangeSquared
}

// getCharacterFromObject attempts to extract Character from WorldObject.
// Returns nil if object is not a Character (e.g., item, door).
//
// MVP: simplified implementation без proper type checking.
// TODO Phase 5.4: implement WorldObject.AsCharacter() method.
//
// Phase 5.3: Basic Combat System.
func getCharacterFromObject(obj *model.WorldObject) *model.Character {
	// TODO Phase 5.4: proper type checking
	// For MVP, return nil (assume all objects can be attacked)
	// This prevents crashes but doesn't provide full validation
	return nil
}
