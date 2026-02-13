package geo

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindPathNoGeodata(t *testing.T) {
	e := NewEngine()

	path := e.FindPath(0, 0, 0, 100, 100, 0)
	require.NotNil(t, path)
	assert.Equal(t, 1, len(path))
	assert.Equal(t, int32(100), path[0].X)
	assert.Equal(t, int32(100), path[0].Y)
}

func TestFindPathSameCell(t *testing.T) {
	e := setupFlatEngine(t, 0)

	wx := int32(WorldMinX + CoordinateOffset)
	wy := int32(WorldMinY + CoordinateOffset)
	path := e.FindPath(wx, wy, 0, wx, wy, 0)
	require.NotNil(t, path)
	assert.Equal(t, 1, len(path))
}

func TestFindPathFlatTerrain(t *testing.T) {
	e := setupFlatEngine(t, 100)

	wx1 := int32(WorldMinX + 16)
	wy1 := int32(WorldMinY + 16)
	wx2 := int32(WorldMinX + 80)
	wy2 := int32(WorldMinY + 80)

	path := e.FindPath(wx1, wy1, 100, wx2, wy2, 100)
	require.NotNil(t, path, "should find path on flat terrain")
	assert.GreaterOrEqual(t, len(path), 2, "path should have at least start and end")

	last := path[len(path)-1]
	dx := abs32(last.X - wx2)
	dy := abs32(last.Y - wy2)
	assert.LessOrEqual(t, dx, int32(CoordinateScale), "X should be near dest")
	assert.LessOrEqual(t, dy, int32(CoordinateScale), "Y should be near dest")
}

func TestFindPathWithWall(t *testing.T) {
	e := setupWallEngine(t)

	wx1 := int32(WorldMinX + CoordinateOffset)
	wy1 := int32(WorldMinY + CoordinateOffset)
	wx2 := int32(WorldMinX + CoordinateOffset)
	wy2 := int32(WorldMinY + CoordinateOffset + 2*CoordinateScale)

	path := e.FindPath(wx1, wy1, 100, wx2, wy2, 100)
	if path != nil {
		assert.GreaterOrEqual(t, len(path), 3, "path should go around wall")
	}
}

func TestHeuristic(t *testing.T) {
	h := heuristic(0, 0, 0, 0, 0, 0)
	assert.Equal(t, 0.0, h)

	h = heuristic(0, 0, 0, 10, 0, 0)
	assert.InDelta(t, 10.0, h, 0.001)

	h = heuristic(0, 0, 0, 10, 10, 0)
	assert.InDelta(t, 14.142, h, 0.01)
}

func TestNodeHeap(t *testing.T) {
	h := &nodeHeap{}

	n1 := &geoNode{x: 1, fCost: 10.0}
	n2 := &geoNode{x: 2, fCost: 5.0}
	n3 := &geoNode{x: 3, fCost: 15.0}

	heapPush(h, n1)
	heapPush(h, n2)
	heapPush(h, n3)

	assert.Equal(t, 3, h.Len())

	min := heapPop(h)
	assert.Equal(t, int32(2), min.x)
	assert.Equal(t, 5.0, min.fCost)
}

// Manual heap operations to avoid container/heap import in test.
func heapPush(h *nodeHeap, n *geoNode) {
	n.index = len(*h)
	*h = append(*h, n)
	i := len(*h) - 1
	for i > 0 {
		parent := (i - 1) / 2
		if (*h)[parent].fCost <= (*h)[i].fCost {
			break
		}
		(*h)[parent], (*h)[i] = (*h)[i], (*h)[parent]
		(*h)[parent].index = parent
		(*h)[i].index = i
		i = parent
	}
}

func heapPop(h *nodeHeap) *geoNode {
	old := *h
	n := len(old)
	if n == 0 {
		return nil
	}
	old[0], old[n-1] = old[n-1], old[0]
	old[0].index = 0
	node := old[n-1]
	*h = old[:n-1]
	i := 0
	for {
		left := 2*i + 1
		right := 2*i + 2
		smallest := i
		if left < len(*h) && (*h)[left].fCost < (*h)[smallest].fCost {
			smallest = left
		}
		if right < len(*h) && (*h)[right].fCost < (*h)[smallest].fCost {
			smallest = right
		}
		if smallest == i {
			break
		}
		(*h)[i], (*h)[smallest] = (*h)[smallest], (*h)[i]
		(*h)[i].index = i
		(*h)[smallest].index = smallest
		i = smallest
	}
	return node
}

// Setup helpers

func setupFlatEngine(t *testing.T, height int16) *Engine {
	t.Helper()
	data := buildFlatRegionData(height)
	region, err := LoadRegion(data)
	require.NoError(t, err)

	e := NewEngine()
	e.regions[0].Store(region)
	e.loaded.Store(1)
	return e
}

func setupWallEngine(t *testing.T) *Engine {
	t.Helper()
	var data []byte

	data = append(data, BlockTypeComplex)
	for cellIdx := range BlockCells {
		cellX := int32(cellIdx) / BlockCellsY
		cellY := int32(cellIdx) % BlockCellsY

		var nswe byte
		if cellX == 0 && cellY == 0 {
			nswe = NSWEEast | NSWESouth
		} else if cellX == 0 && cellY == 1 {
			nswe = 0
		} else {
			nswe = NSWEAll
		}

		height := int16(100)
		cellVal := uint16(height<<1) | uint16(nswe)
		var buf [2]byte
		binary.LittleEndian.PutUint16(buf[:], cellVal)
		data = append(data, buf[:]...)
	}

	for range RegionBlocks - 1 {
		data = append(data, BlockTypeFlat)
		var buf [2]byte
		binary.LittleEndian.PutUint16(buf[:], uint16(int16(100)))
		data = append(data, buf[:]...)
	}

	region, err := LoadRegion(data)
	require.NoError(t, err)

	e := NewEngine()
	e.regions[0].Store(region)
	e.loaded.Store(1)
	return e
}
