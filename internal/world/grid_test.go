package world

import "testing"

func TestCoordToRegionIndex(t *testing.T) {
	tests := []struct {
		name       string
		x, y       int32
		wantRX, wantRY int32
	}{
		{
			name:   "origin (0,0)",
			x:      0,
			y:      0,
			wantRX: OffsetX,
			wantRY: OffsetY,
		},
		{
			name:   "min boundaries",
			x:      WorldXMin,
			y:      WorldYMin,
			wantRX: 0,
			wantRY: 0,
		},
		{
			name:   "max boundaries",
			x:      WorldXMax - 1,
			y:      WorldYMax - 1,
			wantRX: RegionsX - 1,  // 159
			wantRY: 239,            // (229375 >> 11) + 128 = 111 + 128 = 239 (NOT 240)
		},
		{
			name:   "Talking Island spawn (17000, 170000)",
			x:      17000,
			y:      170000,
			wantRX: 72,  // (17000 >> 11) + 64 = 8 + 64 = 72
			wantRY: 211, // (170000 >> 11) + 128 = 83 + 128 = 211
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rx, ry := CoordToRegionIndex(tt.x, tt.y)
			if rx != tt.wantRX || ry != tt.wantRY {
				t.Errorf("CoordToRegionIndex(%d, %d) = (%d, %d), want (%d, %d)",
					tt.x, tt.y, rx, ry, tt.wantRX, tt.wantRY)
			}
		})
	}
}

func TestIsValidRegionIndex(t *testing.T) {
	tests := []struct {
		name string
		rx, ry int32
		want bool
	}{
		{"valid center", OffsetX, OffsetY, true},
		{"valid min", 0, 0, true},
		{"valid max", RegionsX - 1, RegionsY - 1, true},
		{"invalid negative X", -1, 0, false},
		{"invalid negative Y", 0, -1, false},
		{"invalid out of bounds X", RegionsX, 0, false},
		{"invalid out of bounds Y", 0, RegionsY, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidRegionIndex(tt.rx, tt.ry)
			if got != tt.want {
				t.Errorf("IsValidRegionIndex(%d, %d) = %v, want %v", tt.rx, tt.ry, got, tt.want)
			}
		})
	}
}

func TestRegionIndexToCoord(t *testing.T) {
	tests := []struct {
		name string
		rx, ry int32
		wantX, wantY int32
	}{
		{
			name:  "region (0,0) center",
			rx:    0,
			ry:    0,
			wantX: WorldXMin + RegionSize/2,
			wantY: WorldYMin + RegionSize/2,
		},
		{
			name:  "region (OffsetX, OffsetY) center",
			rx:    OffsetX,
			ry:    OffsetY,
			wantX: RegionSize / 2,
			wantY: RegionSize / 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := RegionIndexToCoord(tt.rx, tt.ry)
			if x != tt.wantX || y != tt.wantY {
				t.Errorf("RegionIndexToCoord(%d, %d) = (%d, %d), want (%d, %d)",
					tt.rx, tt.ry, x, y, tt.wantX, tt.wantY)
			}
		})
	}
}

func BenchmarkCoordToRegionIndex(b *testing.B) {
	for range b.N {
		CoordToRegionIndex(17000, 170000)
	}
}
