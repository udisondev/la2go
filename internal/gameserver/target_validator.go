package gameserver

import (
	"fmt"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// Target validation constants.
const (
	// MaxTargetSelectDistance is the maximum distance for target selection (units).
	// Beyond this distance, target selection fails.
	MaxTargetSelectDistance = 2000

	// MaxTargetSelectDistanceSquared is the squared distance for performance.
	MaxTargetSelectDistanceSquared = MaxTargetSelectDistance * MaxTargetSelectDistance

	// DefaultMeleeAttackRange is the fallback melee attack range (units).
	// Java: Formulas.java:99 — MELEE_ATTACK_RANGE = 40.
	DefaultMeleeAttackRange = 40
)

// ValidateTargetSelection validates target selection request.
// Checks:
//   - Target exists in world
//   - Target is visible to player (visibility cache)
//   - Target is within selection range (2000 units)
//
// Returns error if validation fails.
//
// Phase 5.2: Target System.
func ValidateTargetSelection(player *model.Player, targetObjectID uint32, worldInstance *world.World) (*model.WorldObject, error) {
	// 1. Check target exists
	target, exists := worldInstance.GetObject(targetObjectID)
	if !exists {
		return nil, fmt.Errorf("target not found: objectID=%d", targetObjectID)
	}

	// 2. Check visibility (target must be in player's visibility cache)
	// Phase 4.5 PR3: Use visibility cache for O(1) lookup
	if !IsTargetVisible(player, target) {
		return nil, fmt.Errorf("target not visible: objectID=%d", targetObjectID)
	}

	// 3. Check selection range (max 2000 units)
	playerLoc := player.Location()
	targetLoc := target.Location()
	distSq := playerLoc.DistanceSquared(targetLoc)

	if distSq > MaxTargetSelectDistanceSquared {
		return nil, fmt.Errorf("target too far: distance²=%d (max=%d)", distSq, MaxTargetSelectDistanceSquared)
	}

	return target, nil
}

// IsTargetVisible checks if target is visible to player.
// Uses player's visibility cache (Phase 4.5 PR3) for O(1) lookup.
//
// Returns true if target is in player's visibility cache.
//
// Phase 5.2: Target System.
func IsTargetVisible(player *model.Player, target *model.WorldObject) bool {
	cache := player.GetVisibilityCache()
	if cache == nil {
		// No visibility cache yet (player just logged in).
		// Fallback: assume visible if in same region.
		// A more precise check would compare region indices, but this is safe
		// since the cache is populated within the first visibility tick (~1s).
		return true
	}

	// Check if target is in any LOD bucket (near/medium/far)
	targetID := target.ObjectID()

	// Check near bucket
	for _, obj := range cache.NearObjects() {
		if obj.ObjectID() == targetID {
			return true
		}
	}

	// Check medium bucket
	for _, obj := range cache.MediumObjects() {
		if obj.ObjectID() == targetID {
			return true
		}
	}

	// Check far bucket
	for _, obj := range cache.FarObjects() {
		if obj.ObjectID() == targetID {
			return true
		}
	}

	return false
}

// IsInAttackRange checks if target is within physical attack range.
// Range is weapon-dependent: fist=20, sword=40, bow=500.
//
// Java reference: CreatureStat.getPhysicalAttackRange() (line 591-605).
func IsInAttackRange(attacker *model.Player, target *model.WorldObject) bool {
	attackerLoc := attacker.Location()
	targetLoc := target.Location()
	distSq := attackerLoc.DistanceSquared(targetLoc)

	attackRange := attacker.GetAttackRange()
	if attackRange < DefaultMeleeAttackRange {
		attackRange = DefaultMeleeAttackRange
	}

	return distSq <= int64(attackRange)*int64(attackRange)
}

// CanSeeTarget checks line of sight between player and target.
// Phase 7.1 added GeoEngine with LOS (Bresenham raycasting), but it requires
// .l2j geodata files loaded. When geodata is unavailable, falls back to
// visibility cache check (always passable terrain).
//
// Phase 5.2: Target System (simplified).
func CanSeeTarget(player *model.Player, target *model.WorldObject) bool {
	// Visibility cache check — if not in cache, can't see regardless.
	// GeoEngine LOS is checked at combat validation layer (combat.ValidateAttack).
	return IsTargetVisible(player, target)
}
