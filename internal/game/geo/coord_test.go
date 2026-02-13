package geo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeoXConversion(t *testing.T) {
	tests := []struct {
		name   string
		worldX int32
		wantGX int32
	}{
		{"world min", WorldMinX, 0},
		{"origin", 0, -WorldMinX / CoordinateScale},
		{"positive", 10000, (10000 - WorldMinX) / CoordinateScale},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gx := GeoX(tt.worldX)
			assert.Equal(t, tt.wantGX, gx)
		})
	}
}

func TestGeoYConversion(t *testing.T) {
	tests := []struct {
		name   string
		worldY int32
		wantGY int32
	}{
		{"world min", WorldMinY, 0},
		{"origin", 0, -WorldMinY / CoordinateScale},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gy := GeoY(tt.worldY)
			assert.Equal(t, tt.wantGY, gy)
		})
	}
}

func TestWorldXRoundTrip(t *testing.T) {
	// World → Geo → World should land in the same cell (±offset)
	worldX := int32(50000)
	gx := GeoX(worldX)
	backX := WorldX(gx)

	// backX should be within CoordinateScale of worldX
	diff := worldX - backX
	if diff < 0 {
		diff = -diff
	}
	assert.LessOrEqual(t, diff, int32(CoordinateScale))
}

func TestComputeNSWE(t *testing.T) {
	tests := []struct {
		name string
		fx   int32
		fy   int32
		tx   int32
		ty   int32
		want byte
	}{
		{"east", 0, 0, 1, 0, NSWEEast},
		{"west", 1, 0, 0, 0, NSWEWest},
		{"south", 0, 0, 0, 1, NSWESouth},
		{"north", 0, 1, 0, 0, NSWENorth},
		{"north-east", 0, 1, 1, 0, NSWENorthEast},
		{"south-west", 1, 0, 0, 1, NSWESouthWest},
		{"same", 5, 5, 5, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeNSWE(tt.fx, tt.fy, tt.tx, tt.ty)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegionXY(t *testing.T) {
	rx, ry := RegionXY(0, 0)
	assert.Equal(t, int32(0), rx)
	assert.Equal(t, int32(0), ry)

	// Cell in second region
	rx, ry = RegionXY(RegionCellsX, RegionCellsY)
	assert.Equal(t, int32(1), rx)
	assert.Equal(t, int32(1), ry)
}

func TestBlockXY(t *testing.T) {
	// First cell → block 0
	idx := BlockXY(0, 0)
	assert.Equal(t, int32(0), idx)

	// Cell (8, 0) → block (1, 0)
	idx = BlockXY(8, 0)
	assert.Equal(t, int32(1*RegionBlocksY+0), idx)
}

func TestCellXY(t *testing.T) {
	cx, cy := CellXY(0, 0)
	assert.Equal(t, int32(0), cx)
	assert.Equal(t, int32(0), cy)

	cx, cy = CellXY(10, 13)
	assert.Equal(t, int32(2), cx) // 10 % 8 = 2
	assert.Equal(t, int32(5), cy) // 13 % 8 = 5
}
