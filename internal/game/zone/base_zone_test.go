package zone

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNPolyContains(t *testing.T) {
	// Треугольник: (0,0), (100,0), (50,100) — Z от -1000 до 1000.
	z := &BaseZone{
		id:       1,
		name:     "triangle",
		zoneType: TypeTown,
		shape:    "NPoly",
		minZ:     -1000,
		maxZ:     1000,
		nodesX:   []int32{0, 100, 50},
		nodesY:   []int32{0, 0, 100},
	}

	tests := []struct {
		name     string
		x, y, zz int32
		want     bool
	}{
		{"center inside", 50, 30, 0, true},
		{"near apex inside", 50, 90, 0, true},
		{"outside left", -10, 50, 0, false},
		{"outside right", 110, 50, 0, false},
		{"outside above", 50, 110, 0, false},
		{"outside below", 50, -10, 0, false},
		{"on vertex", 0, 0, 0, true},
		{"on bottom edge", 50, 0, 0, true},
		{"inside XY but below Z range", 50, 30, -1500, false},
		{"inside XY but above Z range", 50, 30, 1500, false},
		{"inside at minZ boundary", 50, 30, -1000, true},
		{"inside at maxZ boundary", 50, 30, 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := z.Contains(tt.x, tt.y, tt.zz)
			assert.Equal(t, tt.want, got, "Contains(%d, %d, %d)", tt.x, tt.y, tt.zz)
		})
	}
}

func TestCuboidContains(t *testing.T) {
	// Прямоугольник: (0,0), (100,0), (100,200), (0,200) — Z от -500 до 500.
	z := &BaseZone{
		id:       2,
		name:     "cuboid",
		zoneType: TypeCastle,
		shape:    "Cuboid",
		minZ:     -500,
		maxZ:     500,
		nodesX:   []int32{0, 100, 100, 0},
		nodesY:   []int32{0, 0, 200, 200},
	}

	tests := []struct {
		name     string
		x, y, zz int32
		want     bool
	}{
		{"center inside", 50, 100, 0, true},
		{"corner inside", 0, 0, 0, true},
		{"opposite corner", 100, 200, 0, true},
		{"outside X", 150, 100, 0, false},
		{"outside Y", 50, 250, 0, false},
		{"negative outside", -1, 100, 0, false},
		{"outside Z range", 50, 100, 600, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := z.Contains(tt.x, tt.y, tt.zz)
			assert.Equal(t, tt.want, got, "Contains(%d, %d, %d)", tt.x, tt.y, tt.zz)
		})
	}
}

func TestEmptyPolygonContains(t *testing.T) {
	z := &BaseZone{
		id:       3,
		name:     "empty",
		zoneType: TypeEffect,
		shape:    "NPoly",
		minZ:     -1000,
		maxZ:     1000,
		nodesX:   []int32{},
		nodesY:   []int32{},
	}

	assert.False(t, z.Contains(0, 0, 0), "empty polygon should contain nothing")
}

func TestZRangeOnly(t *testing.T) {
	// Квадрат (0,0)-(100,100), но очень узкий Z диапазон.
	z := &BaseZone{
		id:       4,
		name:     "thin-z",
		zoneType: TypeDamage,
		shape:    "NPoly",
		minZ:     50,
		maxZ:     60,
		nodesX:   []int32{0, 100, 100, 0},
		nodesY:   []int32{0, 0, 100, 100},
	}

	assert.True(t, z.Contains(50, 50, 55), "inside XY and Z")
	assert.False(t, z.Contains(50, 50, 49), "inside XY, below Z")
	assert.False(t, z.Contains(50, 50, 61), "inside XY, above Z")
}

func TestNPolyNegativeCoordinates(t *testing.T) {
	// Квадрат в отрицательных координатах: (-100,-100) - (100,100).
	z := &BaseZone{
		id:       5,
		name:     "negative-coords",
		zoneType: TypeSiege,
		shape:    "NPoly",
		minZ:     -1000,
		maxZ:     1000,
		nodesX:   []int32{-100, 100, 100, -100},
		nodesY:   []int32{-100, -100, 100, 100},
	}

	assert.True(t, z.Contains(0, 0, 0), "center of negative zone")
	assert.True(t, z.Contains(-50, -50, 0), "negative quadrant inside")
	assert.False(t, z.Contains(-150, 0, 0), "outside left")
	assert.False(t, z.Contains(0, 150, 0), "outside top")
}
