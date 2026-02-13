package geo

import "fmt"

// Region represents a loaded geodata region file (.l2j).
// Each region contains 256×256 blocks, each block has 8×8 cells.
// Total: 2048×2048 cells per region.
type Region struct {
	blocks [RegionBlocks]Block
}

// LoadRegion parses a .l2j file's raw bytes into a Region.
func LoadRegion(data []byte) (*Region, error) {
	r := &Region{}
	offset := 0

	for i := range RegionBlocks {
		block, consumed, err := ParseBlock(data, offset)
		if err != nil {
			return nil, fmt.Errorf("load region block %d: %w", i, err)
		}
		r.blocks[i] = block
		offset += consumed
	}

	return r, nil
}

// GetBlock returns the block at (blockX, blockY) indices within the region.
func (r *Region) GetBlock(blockX, blockY int32) Block {
	return r.blocks[blockX*RegionBlocksY+blockY]
}

// GetNearestZ returns the nearest Z for geo coordinates local to this region.
func (r *Region) GetNearestZ(localGeoX, localGeoY int32, worldZ int32) int32 {
	blockX := localGeoX / BlockCellsX
	blockY := localGeoY / BlockCellsY
	cellX := localGeoX % BlockCellsX
	cellY := localGeoY % BlockCellsY

	block := r.blocks[blockX*RegionBlocksY+blockY]
	return block.GetNearestZ(cellX, cellY, worldZ)
}

// GetNextHigherZ returns the next higher Z for geo coordinates local to this region.
func (r *Region) GetNextHigherZ(localGeoX, localGeoY int32, worldZ int32) int32 {
	blockX := localGeoX / BlockCellsX
	blockY := localGeoY / BlockCellsY
	cellX := localGeoX % BlockCellsX
	cellY := localGeoY % BlockCellsY

	block := r.blocks[blockX*RegionBlocksY+blockY]
	return block.GetNextHigherZ(cellX, cellY, worldZ)
}

// GetNSWE returns the NSWE mask for geo coordinates local to this region.
func (r *Region) GetNSWE(localGeoX, localGeoY int32, worldZ int32) byte {
	blockX := localGeoX / BlockCellsX
	blockY := localGeoY / BlockCellsY
	cellX := localGeoX % BlockCellsX
	cellY := localGeoY % BlockCellsY

	block := r.blocks[blockX*RegionBlocksY+blockY]
	return block.GetNSWE(cellX, cellY, worldZ)
}

// HasGeoData returns true if the given position has per-cell geodata.
func (r *Region) HasGeoData(localGeoX, localGeoY int32) bool {
	blockX := localGeoX / BlockCellsX
	blockY := localGeoY / BlockCellsY

	block := r.blocks[blockX*RegionBlocksY+blockY]
	return block.HasGeoData()
}
