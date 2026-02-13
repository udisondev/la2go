package geo

// World coordinate boundaries.
const (
	WorldMinX = -655360
	WorldMinY = -589824
)

// GeoEngine grid dimensions.
const (
	GeoRegionsX    = 32
	GeoRegionsY    = 32
	RegionBlocksX  = 256
	RegionBlocksY  = 256
	RegionBlocks   = RegionBlocksX * RegionBlocksY // 65536
	BlockCellsX    = 8
	BlockCellsY    = 8
	BlockCells     = BlockCellsX * BlockCellsY // 64
	RegionCellsX   = RegionBlocksX * BlockCellsX // 2048
	RegionCellsY   = RegionBlocksY * BlockCellsY // 2048
	CoordinateScale  = 16 // 1 geo cell = 16 world units
	CoordinateOffset = 8
)

// NSWE direction bitmask constants.
// 4-bit mask for cell movement permissions.
const (
	NSWEEast  byte = 1 << 0 // 0x01
	NSWEWest  byte = 1 << 1 // 0x02
	NSWESouth byte = 1 << 2 // 0x04
	NSWENorth byte = 1 << 3 // 0x08
	NSWEAll   byte = 0x0F
)

// Composite NSWE directions.
const (
	NSWENorthEast = NSWENorth | NSWEEast // 0x09
	NSWENorthWest = NSWENorth | NSWEWest // 0x0A
	NSWESouthEast = NSWESouth | NSWEEast // 0x05
	NSWESouthWest = NSWESouth | NSWEWest // 0x06
)

// Block type identifiers in .l2j binary format.
const (
	BlockTypeFlat       byte = 0x00
	BlockTypeComplex    byte = 0x01
	BlockTypeMultilayer byte = 0x02
)

// Pathfinding configuration.
const (
	MaxPathfindIterations = 7000
	MaxSeeOverHeight      = 48
	ElevatedSeeOverDist   = 2
	HeightIncrLimit       = 40

	// A* weights.
	WeightLow      = 0.5
	WeightMedium   = 2.0
	WeightHigh     = 3.0
	WeightDiagonal = 0.707 // sqrt(2)/2
)
