package geo

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// packCell packs height and NSWE into the .l2j cell format.
// Height precision is quantized to 8 world units (12-bit field × 8).
// Format: bits[15:4] = (height/8) as 12-bit signed, bits[3:0] = NSWE.
// Extraction: (raw & 0xFFF0) >> 1 = height (rounded to multiple of 8).
func packCell(height int16, nswe byte) uint16 {
	h := height >> 3 // Quantize to 8-unit steps
	return uint16(h<<4) | uint16(nswe)
}

func TestFlatBlock(t *testing.T) {
	b := &FlatBlock{height: -500}

	assert.Equal(t, int32(-500), b.GetNearestZ(0, 0, 0))
	assert.Equal(t, int32(-500), b.GetNearestZ(7, 7, 1000))
	assert.Equal(t, NSWEAll, b.GetNSWE(3, 3, 0))
	assert.False(t, b.HasGeoData())
}

func TestComplexBlock(t *testing.T) {
	b := &ComplexBlock{}

	// Height 96 (multiple of 8), NSWE=ALL
	b.data[0] = packCell(96, NSWEAll)

	// Height 200 (multiple of 8), NSWE=North only
	b.data[1*BlockCellsY+0] = packCell(200, NSWENorth)

	z := b.GetNearestZ(0, 0, 0)
	assert.Equal(t, int32(96), z)

	nswe := b.GetNSWE(0, 0, 0)
	assert.Equal(t, NSWEAll, nswe)

	z1 := b.GetNearestZ(1, 0, 0)
	assert.Equal(t, int32(200), z1)

	nswe1 := b.GetNSWE(1, 0, 0)
	assert.Equal(t, NSWENorth, nswe1)

	assert.True(t, b.HasGeoData())
}

func TestComplexBlockNegativeHeight(t *testing.T) {
	b := &ComplexBlock{}

	// Negative height: -104 (multiple of 8)
	b.data[0] = packCell(-104, NSWEAll)

	z := b.GetNearestZ(0, 0, 0)
	assert.Equal(t, int32(-104), z)
}

func TestMultilayerBlock(t *testing.T) {
	var data []byte
	var buf [2]byte

	// Cell (0,0): 2 layers at Z=96 (NSWE_ALL) and Z=304 (NSWE_NORTH)
	data = append(data, 2)
	binary.LittleEndian.PutUint16(buf[:], packCell(96, NSWEAll))
	data = append(data, buf[:]...)
	binary.LittleEndian.PutUint16(buf[:], packCell(304, NSWENorth))
	data = append(data, buf[:]...)

	// Remaining 63 cells: 1 layer each at Z=0, NSWE_ALL
	for range 63 {
		data = append(data, 1)
		binary.LittleEndian.PutUint16(buf[:], packCell(0, NSWEAll))
		data = append(data, buf[:]...)
	}

	cellOffsets := [BlockCells]int{}
	offset := 0
	for cellIdx := range BlockCells {
		cellOffsets[cellIdx] = offset
		nLayers := int(data[offset])
		offset++
		offset += nLayers * 2
	}

	b := &MultilayerBlock{data: data, cellOffsets: cellOffsets}

	// Query Z=90 → should find Z=96 (closest)
	z := b.GetNearestZ(0, 0, 90)
	assert.Equal(t, int32(96), z)

	// Query Z=280 → should find Z=304 (closest)
	z = b.GetNearestZ(0, 0, 280)
	assert.Equal(t, int32(304), z)

	// NSWE at Z=90 → layer at Z=96 → NSWE_ALL
	nswe := b.GetNSWE(0, 0, 90)
	assert.Equal(t, NSWEAll, nswe)

	// NSWE at Z=290 → layer at Z=304 → NSWE_NORTH
	nswe = b.GetNSWE(0, 0, 290)
	assert.Equal(t, NSWENorth, nswe)

	assert.True(t, b.HasGeoData())
}

func TestParseBlockFlat(t *testing.T) {
	data := make([]byte, 3)
	data[0] = BlockTypeFlat
	binary.LittleEndian.PutUint16(data[1:], uint16(int16(150)))

	block, consumed, err := ParseBlock(data, 0)
	require.NoError(t, err)
	assert.Equal(t, 3, consumed)

	z := block.GetNearestZ(0, 0, 0)
	assert.Equal(t, int32(150), z)
	assert.False(t, block.HasGeoData())
}

func TestParseBlockComplex(t *testing.T) {
	data := make([]byte, 1+BlockCells*2)
	data[0] = BlockTypeComplex

	cellValue := packCell(48, NSWEAll) // 48 is multiple of 8
	for i := range BlockCells {
		binary.LittleEndian.PutUint16(data[1+i*2:], cellValue)
	}

	block, consumed, err := ParseBlock(data, 0)
	require.NoError(t, err)
	assert.Equal(t, 1+BlockCells*2, consumed)

	z := block.GetNearestZ(3, 3, 0)
	assert.Equal(t, int32(48), z)
	assert.Equal(t, NSWEAll, block.GetNSWE(3, 3, 0))
	assert.True(t, block.HasGeoData())
}

func TestParseBlockMultilayer(t *testing.T) {
	var data []byte
	data = append(data, BlockTypeMultilayer)

	var buf [2]byte
	for range BlockCells {
		data = append(data, 1)
		binary.LittleEndian.PutUint16(buf[:], packCell(0, NSWEAll))
		data = append(data, buf[:]...)
	}

	block, _, err := ParseBlock(data, 0)
	require.NoError(t, err)
	assert.True(t, block.HasGeoData())
}

func TestParseBlockUnknownType(t *testing.T) {
	data := []byte{0xFF}
	_, _, err := ParseBlock(data, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown block type")
}
