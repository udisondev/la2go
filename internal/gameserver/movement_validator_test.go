package gameserver

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// TestValidateMoveToLocation_ZBounds tests Z-coordinate boundary validation.
func TestValidateMoveToLocation_ZBounds(t *testing.T) {
	tests := []struct {
		name    string
		z       int32
		wantErr bool
	}{
		{"Valid Z (zero)", 0, false},
		{"Valid Z (positive)", 1000, false},
		{"Valid Z (negative)", -1000, false},
		{"Valid Z (min boundary)", MinZCoordinate, false},
		{"Valid Z (max boundary)", MaxZCoordinate, false},
		{"Invalid Z (below min)", MinZCoordinate - 1, true},
		{"Invalid Z (above max)", MaxZCoordinate + 1, true},
		{"Invalid Z (far below)", -50000, true},
		{"Invalid Z (far above)", 50000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create player at origin
			player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
			if err != nil {
				t.Fatalf("NewPlayer failed: %v", err)
			}

			// Set player location to origin
			player.SetLocation(model.NewLocation(0, 0, 0, 0))

			// Test movement to (0, 0, Z)
			err = ValidateMoveToLocation(player, 0, 0, tt.z)

			if tt.wantErr && err == nil {
				t.Errorf("expected error for Z=%d, got nil", tt.z)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for Z=%d: %v", tt.z, err)
			}
		})
	}
}

// TestValidateMoveToLocation_MaxDistance tests maximum movement distance validation.
func TestValidateMoveToLocation_MaxDistance(t *testing.T) {
	tests := []struct {
		name    string
		dx      int32 // distance X from origin
		dy      int32 // distance Y from origin
		wantErr bool
	}{
		{"Zero distance", 0, 0, false},
		{"Valid short distance", 100, 0, false},
		{"Valid medium distance", 1000, 1000, false},
		{"Valid max distance (exactly 9900)", 9900, 0, false},
		{"Valid max distance (diagonal)", 7000, 7000, false}, // sqrt(7000²+7000²) ≈ 9899
		{"Invalid (9901 units)", 9901, 0, true},
		{"Invalid (far away)", 20000, 0, true},
		{"Invalid (diagonal too far)", 10000, 10000, true}, // sqrt(10000²+10000²) ≈ 14142
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
			if err != nil {
				t.Fatalf("NewPlayer failed: %v", err)
			}

			// Set player at origin
			player.SetLocation(model.NewLocation(0, 0, 0, 0))

			// Test movement to (dx, dy, 0)
			err = ValidateMoveToLocation(player, tt.dx, tt.dy, 0)

			if tt.wantErr && err == nil {
				t.Errorf("expected error for distance (%d,%d), got nil", tt.dx, tt.dy)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for distance (%d,%d): %v", tt.dx, tt.dy, err)
			}
		})
	}
}

// TestValidateMoveToLocation_MinDistance tests minimum movement distance validation.
func TestValidateMoveToLocation_MinDistance(t *testing.T) {
	tests := []struct {
		name    string
		dx      int32
		dy      int32
		wantErr bool
	}{
		{"Zero distance (allowed)", 0, 0, false}, // Click same position
		{"Valid (exactly 17 units)", 17, 0, false},
		{"Valid (diagonal 17)", 13, 13, false}, // sqrt(13²+13²) = sqrt(338) ≈ 18.4
		{"Invalid (16 units)", 16, 0, true},
		{"Invalid (1 unit)", 1, 0, true},
		{"Invalid (diagonal 16)", 11, 11, true}, // sqrt(11²+11²) = sqrt(242) ≈ 15.6
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
			if err != nil {
				t.Fatalf("NewPlayer failed: %v", err)
			}

			// Set player at origin
			player.SetLocation(model.NewLocation(0, 0, 0, 0))

			// Test movement to (dx, dy, 0)
			err = ValidateMoveToLocation(player, tt.dx, tt.dy, 0)

			if tt.wantErr && err == nil {
				t.Errorf("expected error for distance (%d,%d), got nil", tt.dx, tt.dy)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for distance (%d,%d): %v", tt.dx, tt.dy, err)
			}
		})
	}
}

// TestValidatePositionDesync tests client-server position desync detection.
func TestValidatePositionDesync(t *testing.T) {
	tests := []struct {
		name             string
		serverX, serverY int32 // server position
		clientX, clientY int32 // client position
		wantCorrection   bool  // expect correction needed
		minDiffSq        int64 // minimum expected difference squared
		maxDiffSq        int64 // maximum expected difference squared
	}{
		{
			name:           "No desync (identical)",
			serverX:        0,
			serverY:        0,
			clientX:        0,
			clientY:        0,
			wantCorrection: false,
			minDiffSq:      0,
			maxDiffSq:      0,
		},
		{
			name:           "Small desync (499 units, no correction)",
			serverX:        0,
			serverY:        0,
			clientX:        499,
			clientY:        0,
			wantCorrection: false,
			minDiffSq:      249000,  // 499²
			maxDiffSq:      250000,  // threshold
		},
		{
			name:           "Desync warning (501 units, needs correction)",
			serverX:        0,
			serverY:        0,
			clientX:        501,
			clientY:        0,
			wantCorrection: true,
			minDiffSq:      250000,  // threshold
			maxDiffSq:      252000,  // 501²
		},
		{
			name:           "Critical desync (600 units)",
			serverX:        0,
			serverY:        0,
			clientX:        600,
			clientY:        0,
			wantCorrection: true,
			minDiffSq:      360000,  // 600²
			maxDiffSq:      360000,
		},
		{
			name:           "Large desync (1000 units)",
			serverX:        0,
			serverY:        0,
			clientX:        1000,
			clientY:        0,
			wantCorrection: true,
			minDiffSq:      1000000, // 1000²
			maxDiffSq:      1000000,
		},
		{
			name:           "Diagonal desync (400,400 = 565 units)",
			serverX:        0,
			serverY:        0,
			clientX:        400,
			clientY:        400,
			wantCorrection: true,
			minDiffSq:      250000,  // threshold
			maxDiffSq:      320000,  // 400²+400² = 320000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
			if err != nil {
				t.Fatalf("NewPlayer failed: %v", err)
			}

			// Set server position
			player.SetLocation(model.NewLocation(tt.serverX, tt.serverY, 0, 0))

			// Check desync with client position
			needsCorrection, diffSq := ValidatePositionDesync(player, tt.clientX, tt.clientY, 0)

			if needsCorrection != tt.wantCorrection {
				t.Errorf("ValidatePositionDesync() needsCorrection = %v, want %v (diffSq=%d)",
					needsCorrection, tt.wantCorrection, diffSq)
			}

			if diffSq < tt.minDiffSq || diffSq > tt.maxDiffSq {
				t.Errorf("ValidatePositionDesync() diffSq = %d, want range [%d..%d]",
					diffSq, tt.minDiffSq, tt.maxDiffSq)
			}
		})
	}
}

// TestValidateMoveToLocation_NormalFlow tests valid movement scenarios.
func TestValidateMoveToLocation_NormalFlow(t *testing.T) {
	tests := []struct {
		name                 string
		startX, startY, startZ int32
		endX, endY, endZ       int32
	}{
		{"Short move forward", 0, 0, 0, 100, 0, 0},
		{"Medium move diagonal", 1000, 1000, 0, 2000, 2000, 0},
		{"Long move (max distance)", 0, 0, 0, 9900, 0, 0},
		{"Move with Z change", 0, 0, 0, 1000, 1000, 500},
		{"Move in negative direction", 1000, 1000, 0, 0, 0, 0},
		{"Move at high Z", 0, 0, 10000, 1000, 1000, 10500},
		{"Move at low Z", 0, 0, -10000, 1000, 1000, -9500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player, err := model.NewPlayer(1, 100, 200, "TestPlayer", 10, 0, 0)
			if err != nil {
				t.Fatalf("NewPlayer failed: %v", err)
			}

			// Set start position
			player.SetLocation(model.NewLocation(tt.startX, tt.startY, tt.startZ, 0))

			// Validate movement to end position
			err = ValidateMoveToLocation(player, tt.endX, tt.endY, tt.endZ)

			if err != nil {
				t.Errorf("ValidateMoveToLocation() failed for valid move: %v", err)
			}
		})
	}
}
