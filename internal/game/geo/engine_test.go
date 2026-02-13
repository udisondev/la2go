package geo

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildFlatRegionData creates a region with all flat blocks at the given height.
func buildFlatRegionData(height int16) []byte {
	data := make([]byte, 0, RegionBlocks*3)
	for range RegionBlocks {
		data = append(data, BlockTypeFlat)
		var buf [2]byte
		binary.LittleEndian.PutUint16(buf[:], uint16(height))
		data = append(data, buf[:]...)
	}
	return data
}

// buildMixedRegionData creates a region where block (0,0) is complex with a wall.
func buildMixedRegionData() []byte {
	var data []byte

	data = append(data, BlockTypeComplex)
	for cellIdx := range BlockCells {
		var nswe byte
		if cellIdx == 0 {
			nswe = NSWEEast | NSWESouth
		} else {
			nswe = NSWEAll
		}
		var buf [2]byte
		binary.LittleEndian.PutUint16(buf[:], packCell(96, nswe))
		data = append(data, buf[:]...)
	}

	for range RegionBlocks - 1 {
		data = append(data, BlockTypeFlat)
		var buf [2]byte
		binary.LittleEndian.PutUint16(buf[:], uint16(int16(96)))
		data = append(data, buf[:]...)
	}

	return data
}

func TestEngineNoGeodata(t *testing.T) {
	e := NewEngine()
	assert.False(t, e.IsLoaded())

	assert.True(t, e.CanSeeTarget(0, 0, 0, 100, 100, 0))
	assert.True(t, e.CanMoveToTarget(0, 0, 0, 100, 100, 0))
	assert.Equal(t, int32(50), e.GetHeight(100, 200, 50))
}

func TestEngineFlatRegion(t *testing.T) {
	data := buildFlatRegionData(96)
	region, err := LoadRegion(data)
	require.NoError(t, err)

	e := NewEngine()
	e.regions[0].Store(region)
	e.loaded.Store(1)

	assert.True(t, e.IsLoaded())

	worldX := int32(WorldMinX + CoordinateOffset)
	worldY := int32(WorldMinY + CoordinateOffset)
	h := e.GetHeight(worldX, worldY, 0)
	assert.Equal(t, int32(96), h)

	wx1 := int32(WorldMinX + 100)
	wy1 := int32(WorldMinY + 100)
	wx2 := int32(WorldMinX + 200)
	wy2 := int32(WorldMinY + 200)
	assert.True(t, e.CanSeeTarget(wx1, wy1, 96, wx2, wy2, 96))
	assert.True(t, e.CanMoveToTarget(wx1, wy1, 96, wx2, wy2, 96))
}

func TestEngineComplexBlockWall(t *testing.T) {
	data := buildMixedRegionData()
	region, err := LoadRegion(data)
	require.NoError(t, err)

	e := NewEngine()
	e.regions[0].Store(region)
	e.loaded.Store(1)

	// Cell (0,0) has restricted NSWE (only East and South)
	nswe := e.getNSWE(0, 0, 96)
	assert.Equal(t, NSWEEast|NSWESouth, nswe)

	// Cell (1,0) has NSWEAll
	nswe1 := e.getNSWE(1, 0, 96)
	assert.Equal(t, NSWEAll, nswe1)
}

func TestEngineHasGeoPos(t *testing.T) {
	e := NewEngine()
	assert.False(t, e.HasGeoPos(0, 0))

	data := buildFlatRegionData(96)
	region, err := LoadRegion(data)
	require.NoError(t, err)
	e.regions[0].Store(region)
	e.loaded.Store(1)

	worldX := int32(WorldMinX + CoordinateOffset)
	worldY := int32(WorldMinY + CoordinateOffset)
	assert.False(t, e.HasGeoPos(worldX, worldY))

	data2 := buildMixedRegionData()
	region2, err := LoadRegion(data2)
	require.NoError(t, err)
	e.regions[0].Store(region2)

	assert.True(t, e.HasGeoPos(worldX, worldY))
}
