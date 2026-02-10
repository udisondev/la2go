package gameserver

import (
	"fmt"

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
