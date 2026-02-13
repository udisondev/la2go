package gameserver

import (
	"fmt"

	"github.com/udisondev/la2go/internal/game/geo"
	"github.com/udisondev/la2go/internal/model"
)

// Movement validation constants.
// These limits prevent common exploits: teleportation, speed hacking, and Z-axis exploits.
const (
	// Z-coordinate boundaries
	MinZCoordinate = -20000 // Minimum allowed Z coordinate (underwater limit)
	MaxZCoordinate = 20000  // Maximum allowed Z coordinate (sky limit)

	// Movement distance limits (squared for performance)
	MaxMoveDistanceSquared = 9900 * 9900 // 98,010,000 - max single move distance
	MinMoveDistanceSquared = 17 * 17     // 289 - min distance to prevent spam

	// Desync detection thresholds (squared)
	// Client position vs server position difference thresholds
	DesyncWarningSquared  = 250000 // 500 units squared - trigger correction
	DesyncCriticalSquared = 360000 // 600 units squared - critical desync (potential hack)
)

// ValidateMoveToLocation validates a MoveToLocation packet from the client.
// This prevents common movement exploits:
//   - Teleportation (distance > 9900 units)
//   - Movement spam (distance < 17 units)
//   - Z-axis exploits (Z outside -20000..20000 range)
//
// Returns an error if validation fails (movement should be rejected).
//
// Reference: L2J_Mobius ValidatePosition.java (lines 76-82, Z-bounds check)
func ValidateMoveToLocation(player *model.Player, targetX, targetY, targetZ int32) error {
	// 1. Z-bounds check (prevent flying/underground exploits)
	if targetZ < MinZCoordinate || targetZ > MaxZCoordinate {
		return fmt.Errorf("invalid Z coordinate: %d (allowed range: %d..%d)",
			targetZ, MinZCoordinate, MaxZCoordinate)
	}

	// 2. Distance checks (prevent teleportation and spam)
	currentLoc := player.Location()

	// Calculate distance squared (avoid sqrt for performance)
	dx := int64(targetX - currentLoc.X)
	dy := int64(targetY - currentLoc.Y)
	distSq := dx*dx + dy*dy

	// 3. Max distance check (anti-teleport)
	if distSq > MaxMoveDistanceSquared {
		return fmt.Errorf("movement distance too large: %d (max: %d)",
			distSq, MaxMoveDistanceSquared)
	}

	// 4. Min distance check (anti-spam)
	// Allow zero-distance moves (player clicks same position)
	if distSq > 0 && distSq < MinMoveDistanceSquared {
		return fmt.Errorf("movement distance too small: %d (min: %d, allowed: 0)",
			distSq, MinMoveDistanceSquared)
	}

	return nil
}

// GeoMoveResult holds the result of geodata-validated movement.
type GeoMoveResult struct {
	// Blocked is true if direct movement is blocked by geodata (walls/obstacles).
	Blocked bool
	// Path contains waypoints from A* pathfinding if direct movement was blocked.
	// nil if movement is direct or no path was found.
	Path []geo.Point3D
	// CorrectedZ is the geodata-corrected Z coordinate at the target position.
	CorrectedZ int32
}

// ValidateMoveWithGeo checks movement against geodata (walls, obstacles, height).
// If geodata is not loaded, returns an unblocked result with original targetZ.
// If direct movement is possible, returns unblocked with corrected Z.
// If blocked, attempts A* pathfinding and returns the path.
//
// Java reference: GeoEngine.canMoveToTarget() + GeoEngine.findPath().
func ValidateMoveWithGeo(geoEng *geo.Engine, fromX, fromY, fromZ, toX, toY, toZ int32) GeoMoveResult {
	if geoEng == nil || !geoEng.IsLoaded() {
		return GeoMoveResult{CorrectedZ: toZ}
	}

	// Correct target Z to geodata height
	correctedZ := geoEng.GetHeight(toX, toY, toZ)

	// Check direct movement
	if geoEng.CanMoveToTarget(fromX, fromY, fromZ, toX, toY, correctedZ) {
		return GeoMoveResult{CorrectedZ: correctedZ}
	}

	// Direct movement blocked — try A* pathfinding
	path := geoEng.FindPath(fromX, fromY, fromZ, toX, toY, correctedZ)
	return GeoMoveResult{
		Blocked:    true,
		Path:       path,
		CorrectedZ: correctedZ,
	}
}

// ValidatePositionDesync checks if client position differs significantly from server position.
// This detects when client and server are out of sync (due to lag, packet loss, or hacks).
//
// Returns:
//   - needsCorrection: true if desync exceeds warning threshold (send ValidateLocation)
//   - diffSquared: the squared distance difference (for logging/metrics)
//
// Reference: L2J_Mobius ValidatePosition.java (position comparison logic)
func ValidatePositionDesync(player *model.Player, clientX, clientY, clientZ int32) (needsCorrection bool, diffSquared int64) {
	serverLoc := player.Location()

	// Calculate 2D distance difference (Z is validated separately)
	dx := int64(clientX - serverLoc.X)
	dy := int64(clientY - serverLoc.Y)
	diffSq := dx*dx + dy*dy

	// Check against warning threshold
	needsCorrection = diffSq > DesyncWarningSquared

	return needsCorrection, diffSq
}
