package geo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineIterator3DHorizontal(t *testing.T) {
	it := NewLineIterator3D(0, 0, 0, 5, 0, 0)

	var points []Point3D
	for it.Next() {
		points = append(points, Point3D{it.X(), it.Y(), it.Z()})
	}

	assert.Equal(t, 6, len(points), "should visit 6 points (0..5)")
	assert.Equal(t, int32(0), points[0].X)
	assert.Equal(t, int32(5), points[5].X)

	// All Y and Z should be 0
	for _, p := range points {
		assert.Equal(t, int32(0), p.Y)
		assert.Equal(t, int32(0), p.Z)
	}
}

func TestLineIterator3DVertical(t *testing.T) {
	it := NewLineIterator3D(0, 0, 0, 0, 3, 0)

	var points []Point3D
	for it.Next() {
		points = append(points, Point3D{it.X(), it.Y(), it.Z()})
	}

	assert.Equal(t, 4, len(points))
	assert.Equal(t, int32(0), points[0].Y)
	assert.Equal(t, int32(3), points[3].Y)
}

func TestLineIterator3DDiagonal(t *testing.T) {
	it := NewLineIterator3D(0, 0, 0, 3, 3, 0)

	var points []Point3D
	for it.Next() {
		points = append(points, Point3D{it.X(), it.Y(), it.Z()})
	}

	// Start and end should match
	assert.Equal(t, int32(0), points[0].X)
	assert.Equal(t, int32(0), points[0].Y)
	assert.Equal(t, int32(3), points[len(points)-1].X)
	assert.Equal(t, int32(3), points[len(points)-1].Y)
}

func TestLineIterator3DNegative(t *testing.T) {
	it := NewLineIterator3D(5, 5, 100, 2, 2, 50)

	var points []Point3D
	for it.Next() {
		points = append(points, Point3D{it.X(), it.Y(), it.Z()})
	}

	// Start at (5,5,100), end at (2,2,50)
	assert.Equal(t, int32(5), points[0].X)
	assert.Equal(t, int32(2), points[len(points)-1].X)
	assert.Equal(t, int32(5), points[0].Y)
	assert.Equal(t, int32(2), points[len(points)-1].Y)
}

func TestLineIterator3DSamePoint(t *testing.T) {
	it := NewLineIterator3D(3, 3, 100, 3, 3, 100)

	count := 0
	for it.Next() {
		count++
	}
	// Only start point
	assert.Equal(t, 1, count)
}

func TestLineIterator3DZDominant(t *testing.T) {
	// Large Z difference, same X/Y
	it := NewLineIterator3D(0, 0, 0, 0, 0, 10)

	var zValues []int32
	for it.Next() {
		zValues = append(zValues, it.Z())
	}

	assert.Equal(t, 11, len(zValues))
	assert.Equal(t, int32(0), zValues[0])
	assert.Equal(t, int32(10), zValues[10])
}
