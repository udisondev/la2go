package gameserver

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// TestValidateTargetSelection_TargetNotFound tests target validation when target doesn't exist.
func TestValidateTargetSelection_TargetNotFound(t *testing.T) {
	// Create player
	player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Get world instance
	worldInst := world.Instance()

	// Try to target non-existent object
	_, err = ValidateTargetSelection(player, 999999, worldInst)
	if err == nil {
		t.Error("Expected error for non-existent target, got nil")
	}
}

// TestValidateTargetSelection_TooFar tests target validation when target is too far.
func TestValidateTargetSelection_TooFar(t *testing.T) {
	// Create player at origin
	player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Create target far away (3000 units, max is 2000)
	targetObj := model.NewWorldObject(2, "FarTarget", model.NewLocation(3000, 0, 0, 0))

	// Get world instance and add target
	worldInst := world.Instance()
	if err := worldInst.AddObject(targetObj); err != nil {
		t.Fatalf("AddObject failed: %v", err)
	}

	// Try to target (should fail — too far)
	_, err = ValidateTargetSelection(player, 2, worldInst)
	if err == nil {
		t.Error("Expected error for target too far, got nil")
	}
}

// TestValidateTargetSelection_Success tests successful target selection.
func TestValidateTargetSelection_Success(t *testing.T) {
	// Create player at origin
	player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Create target nearby (500 units, within 2000 limit)
	targetObj := model.NewWorldObject(2, "NearTarget", model.NewLocation(500, 0, 0, 0))

	// Get world instance and add both objects
	worldInst := world.Instance()
	if err := worldInst.AddObject(player.WorldObject); err != nil {
		t.Fatalf("AddObject player failed: %v", err)
	}
	if err := worldInst.AddObject(targetObj); err != nil {
		t.Fatalf("AddObject target failed: %v", err)
	}

	// Set up visibility cache manually (normally done by VisibilityManager).
	// Include targetObj in the near bucket so IsTargetVisible returns true.
	cache := model.NewVisibilityCache(
		[]*model.WorldObject{targetObj}, // near: target is nearby
		nil,                             // medium
		nil,                             // far
		0, 0, 0,
	)
	player.SetVisibilityCache(cache)

	// Try to target — should succeed with visibility cache set
	target, err := ValidateTargetSelection(player, 2, worldInst)
	if err != nil {
		t.Fatalf("ValidateTargetSelection failed: %v", err)
	}

	if target.ObjectID() != 2 {
		t.Errorf("Expected target objectID=2, got %d", target.ObjectID())
	}
}

// TestIsInAttackRange tests physical attack range validation.
func TestIsInAttackRange(t *testing.T) {
	// DefaultMeleeAttackRange = 40 (unarmed player, fists)
	tests := []struct {
		name        string
		attackerX   int32
		attackerY   int32
		targetX     int32
		targetY     int32
		wantInRange bool
	}{
		{"Same position", 0, 0, 0, 0, true},
		{"Within range (20 units)", 0, 0, 20, 0, true},
		{"Within range (39 units)", 0, 0, 39, 0, true},
		{"At range boundary (40 units)", 0, 0, 40, 0, true},
		{"Just outside range (41 units)", 0, 0, 41, 0, false},
		{"Far outside range (500 units)", 0, 0, 500, 0, false},
		{"Diagonal outside range", 0, 0, 30, 30, false}, // sqrt(30²+30²) ≈ 42
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create attacker
			attacker, err := model.NewPlayer(1, 100, 200, "Attacker", 10, 0, 0)
			if err != nil {
				t.Fatalf("NewPlayer failed: %v", err)
			}
			attacker.SetLocation(model.NewLocation(tt.attackerX, tt.attackerY, 0, 0))

			// Create target
			target := model.NewWorldObject(2, "Target", model.NewLocation(tt.targetX, tt.targetY, 0, 0))

			// Test attack range
			inRange := IsInAttackRange(attacker, target)
			if inRange != tt.wantInRange {
				t.Errorf("IsInAttackRange() = %v, want %v (distance from (%d,%d) to (%d,%d))",
					inRange, tt.wantInRange, tt.attackerX, tt.attackerY, tt.targetX, tt.targetY)
			}
		})
	}
}

// TestIsTargetVisible_NoCache tests visibility check when cache is nil.
func TestIsTargetVisible_NoCache(t *testing.T) {
	// Create player without visibility cache
	player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))

	// Create target
	target := model.NewWorldObject(2, "Target", model.NewLocation(500, 0, 0, 0))

	// Test visibility (should return true as fallback when cache is nil)
	visible := IsTargetVisible(player, target)
	if !visible {
		t.Error("Expected visible=true (fallback when cache is nil), got false")
	}
}

// TestPlayerTargetMethods tests target getter/setter methods.
func TestPlayerTargetMethods(t *testing.T) {
	// Create player
	player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	// Initially no target
	if player.HasTarget() {
		t.Error("Expected HasTarget()=false initially, got true")
	}
	if player.Target() != nil {
		t.Error("Expected Target()=nil initially, got non-nil")
	}

	// Set target
	target := model.NewWorldObject(2, "Target", model.NewLocation(0, 0, 0, 0))
	player.SetTarget(target)

	// Verify target set
	if !player.HasTarget() {
		t.Error("Expected HasTarget()=true after SetTarget, got false")
	}
	if player.Target() != target {
		t.Error("Expected Target() to return set target")
	}
	if player.Target().ObjectID() != 2 {
		t.Errorf("Expected target objectID=2, got %d", player.Target().ObjectID())
	}

	// Clear target
	player.ClearTarget()

	// Verify target cleared
	if player.HasTarget() {
		t.Error("Expected HasTarget()=false after ClearTarget, got true")
	}
	if player.Target() != nil {
		t.Error("Expected Target()=nil after ClearTarget, got non-nil")
	}
}
