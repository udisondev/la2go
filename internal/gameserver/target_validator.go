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

	// MaxPhysicalAttackRange is the maximum physical attack range (units).
	// TODO Phase 5.3: Make this weapon-dependent (sword=40, bow=500, etc).
	MaxPhysicalAttackRange = 100

	// MaxPhysicalAttackRangeSquared is the squared attack range for performance.
	MaxPhysicalAttackRangeSquared = MaxPhysicalAttackRange * MaxPhysicalAttackRange
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
		// No visibility cache yet (player just logged in)
		// Fallback: assume visible if in same region
		// TODO Phase 5.3: More sophisticated fallback
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
// TODO Phase 5.3: Make this weapon-dependent (sword=40, bow=500, etc).
//
// Phase 5.2: Target System (simplified, fixed range).
func IsInAttackRange(attacker *model.Player, target *model.WorldObject) bool {
	attackerLoc := attacker.Location()
	targetLoc := target.Location()
	distSq := attackerLoc.DistanceSquared(targetLoc)

	return distSq <= MaxPhysicalAttackRangeSquared
}

// CanSeeTarget checks line of sight between player and target.
// Simplified implementation without geodata collision detection.
//
// TODO Phase 5.4: Integrate geodata for accurate line of sight.
// For now, always returns true if target is visible (visibility cache check).
//
// Phase 5.2: Target System (simplified).
func CanSeeTarget(player *model.Player, target *model.WorldObject) bool {
	// Phase 5.2 MVP: If target is visible in cache, line of sight is OK
	// TODO Phase 5.4: Add geodata raycasting for walls/obstacles
	return IsTargetVisible(player, target)
}
