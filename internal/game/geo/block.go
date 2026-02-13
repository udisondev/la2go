package geo

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Block provides height and NSWE data for 8x8 cells.
type Block interface {
	// GetNearestZ returns the Z closest to worldZ at local cell (cellX, cellY).
	GetNearestZ(cellX, cellY int32, worldZ int32) int32
	// GetNextHigherZ returns the lowest Z that is >= worldZ at local cell.
	// Used for LOS when movement is blocked (wall top height).
	// If no layer is >= worldZ, returns the highest available layer.
	GetNextHigherZ(cellX, cellY int32, worldZ int32) int32
	// GetNSWE returns the NSWE bitmask for the layer closest to worldZ.
	GetNSWE(cellX, cellY int32, worldZ int32) byte
	// HasGeoData returns true if this block has per-cell geodata (not flat).
	HasGeoData() bool
}

// FlatBlock — all 64 cells share one height and allow all movement.
// Binary format: 1 byte type (0x00) + 2 bytes int16 height (LE).
type FlatBlock struct {
	height int16
}

func (b *FlatBlock) GetNearestZ(_, _ int32, _ int32) int32 {
	return int32(b.height)
}

func (b *FlatBlock) GetNextHigherZ(_, _ int32, _ int32) int32 {
	return int32(b.height)
}

func (b *FlatBlock) GetNSWE(_, _ int32, _ int32) byte {
	return NSWEAll
}

func (b *FlatBlock) HasGeoData() bool {
	return false
}

// ComplexBlock — each of 64 cells has its own height+NSWE packed into uint16.
// Binary format: 1 byte type (0x01) + 64×2 bytes (128 bytes).
// Bit packing: [15:4] = height*2 (signed), [3:0] = NSWE mask.
type ComplexBlock struct {
	data [BlockCells]uint16
}

func (b *ComplexBlock) cellData(cellX, cellY int32) uint16 {
	return b.data[cellX*BlockCellsY+cellY]
}

func (b *ComplexBlock) GetNearestZ(cellX, cellY int32, _ int32) int32 {
	d := b.cellData(cellX, cellY)
	return int32(int16(d&0xFFF0) >> 1)
}

func (b *ComplexBlock) GetNextHigherZ(cellX, cellY int32, _ int32) int32 {
	// Complex block has only one layer per cell — same as GetNearestZ.
	d := b.cellData(cellX, cellY)
	return int32(int16(d&0xFFF0) >> 1)
}

func (b *ComplexBlock) GetNSWE(cellX, cellY int32, _ int32) byte {
	return byte(b.cellData(cellX, cellY) & 0x000F)
}

func (b *ComplexBlock) HasGeoData() bool {
	return true
}

// MultilayerBlock — each cell may have multiple Z layers (bridges, floors).
// Binary format: 1 byte type (0x02) + variable-length data.
// Per cell: 1 byte nLayers + nLayers×2 bytes (height+NSWE per layer).
type MultilayerBlock struct {
	data        []byte
	cellOffsets [BlockCells]int // byte offset into data for each cell
}

func (b *MultilayerBlock) GetNearestZ(cellX, cellY int32, worldZ int32) int32 {
	idx := cellX*BlockCellsY + cellY
	offset := b.cellOffsets[idx]
	nLayers := int(b.data[offset])
	offset++

	bestZ := int32(math.MinInt32)
	bestDist := int32(math.MaxInt32)

	for range nLayers {
		raw := binary.LittleEndian.Uint16(b.data[offset:])
		layerZ := int32(int16(raw&0xFFF0) >> 1)
		offset += 2

		dist := layerZ - worldZ
		if dist < 0 {
			dist = -dist
		}
		if dist < bestDist {
			bestDist = dist
			bestZ = layerZ
		}
	}
	return bestZ
}

func (b *MultilayerBlock) GetNextHigherZ(cellX, cellY int32, worldZ int32) int32 {
	idx := cellX*BlockCellsY + cellY
	offset := b.cellOffsets[idx]
	nLayers := int(b.data[offset])
	offset++

	bestZ := int32(math.MaxInt32)  // Track lowest Z that is >= worldZ
	highestZ := int32(math.MinInt32) // Fallback: highest layer overall

	for range nLayers {
		raw := binary.LittleEndian.Uint16(b.data[offset:])
		layerZ := int32(int16(raw&0xFFF0) >> 1)
		offset += 2

		if layerZ > highestZ {
			highestZ = layerZ
		}
		if layerZ >= worldZ && layerZ < bestZ {
			bestZ = layerZ
		}
	}

	if bestZ == int32(math.MaxInt32) {
		return highestZ // No layer >= worldZ, return highest available
	}
	return bestZ
}

func (b *MultilayerBlock) GetNSWE(cellX, cellY int32, worldZ int32) byte {
	idx := cellX*BlockCellsY + cellY
	offset := b.cellOffsets[idx]
	nLayers := int(b.data[offset])
	offset++

	bestNSWE := NSWEAll
	bestDist := int32(math.MaxInt32)

	for range nLayers {
		raw := binary.LittleEndian.Uint16(b.data[offset:])
		layerZ := int32(int16(raw&0xFFF0) >> 1)
		nswe := byte(raw & 0x000F)
		offset += 2

		dist := layerZ - worldZ
		if dist < 0 {
			dist = -dist
		}
		if dist < bestDist {
			bestDist = dist
			bestNSWE = nswe
		}
	}
	return bestNSWE
}

func (b *MultilayerBlock) HasGeoData() bool {
	return true
}

// ParseBlock reads one block from data at the given offset.
// Returns the parsed Block and the number of bytes consumed.
func ParseBlock(data []byte, offset int) (Block, int, error) {
	if offset >= len(data) {
		return nil, 0, fmt.Errorf("parse block: offset %d beyond data length %d", offset, len(data))
	}

	blockType := data[offset]
	offset++

	switch blockType {
	case BlockTypeFlat:
		if offset+2 > len(data) {
			return nil, 0, fmt.Errorf("parse flat block: insufficient data at offset %d", offset)
		}
		height := int16(binary.LittleEndian.Uint16(data[offset:]))
		return &FlatBlock{height: height}, 3, nil // 1 type + 2 height

	case BlockTypeComplex:
		need := BlockCells * 2
		if offset+need > len(data) {
			return nil, 0, fmt.Errorf("parse complex block: insufficient data at offset %d", offset)
		}
		b := &ComplexBlock{}
		for i := range BlockCells {
			b.data[i] = binary.LittleEndian.Uint16(data[offset:])
			offset += 2
		}
		return b, 1 + need, nil

	case BlockTypeMultilayer:
		start := offset
		cellOffsets := [BlockCells]int{}

		for cellIdx := range BlockCells {
			if offset >= len(data) {
				return nil, 0, fmt.Errorf("parse multilayer block: unexpected end at cell %d", cellIdx)
			}
			cellOffsets[cellIdx] = offset - start
			nLayers := int(data[offset])
			if nLayers <= 0 || nLayers > 125 {
				return nil, 0, fmt.Errorf("parse multilayer block: invalid layer count %d at cell %d", nLayers, cellIdx)
			}
			offset++
			offset += nLayers * 2
		}
		if offset > len(data) {
			return nil, 0, fmt.Errorf("parse multilayer block: data overflow")
		}

		blockData := make([]byte, offset-start)
		copy(blockData, data[start:offset])

		return &MultilayerBlock{
			data:        blockData,
			cellOffsets: cellOffsets,
		}, 1 + (offset - start), nil

	default:
		return nil, 0, fmt.Errorf("parse block: unknown block type 0x%02X at offset %d", blockType, offset-1)
	}
}
