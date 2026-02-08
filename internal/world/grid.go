package world

// Grid constants from Java World.java
const (
	// ShiftBy - shift by N bits for 2^N units per region (2^11 = 2048)
	ShiftBy = 11

	// World boundaries (game coordinates)
	WorldXMin = -131072
	WorldYMin = -262144
	WorldXMax = 196608
	WorldYMax = 229376

	// Offsets for array indexing
	// OffsetX = abs(WorldXMin >> ShiftBy) = abs(-131072 >> 11) = 64
	// OffsetY = abs(WorldYMin >> ShiftBy) = abs(-262144 >> 11) = 128
	OffsetX = 64
	OffsetY = 128

	// Grid size (regions count)
	// RegionsX = (WorldXMax >> ShiftBy) + OffsetX = (196608 >> 11) + 64 = 96 + 64 = 160
	// RegionsY = (WorldYMax >> ShiftBy) + OffsetY = (229376 >> 11) + 128 = 112 + 128 = 240 (но в Java 241)
	RegionsX = 160
	RegionsY = 241

	// Region size in game units
	RegionSize = 1 << ShiftBy // 2^11 = 2048
)

// CoordToRegionIndex converts world coordinate to region index
// Formula: (worldCoord >> ShiftBy) + Offset
func CoordToRegionIndex(x, y int32) (rx, ry int32) {
	rx = (x >> ShiftBy) + OffsetX
	ry = (y >> ShiftBy) + OffsetY
	return rx, ry
}

// IsValidRegionIndex checks if region index is within valid bounds
func IsValidRegionIndex(rx, ry int32) bool {
	return rx >= 0 && rx < RegionsX && ry >= 0 && ry < RegionsY
}

// RegionIndexToCoord converts region index to world coordinate (center of region)
func RegionIndexToCoord(rx, ry int32) (x, y int32) {
	// Reverse formula: worldCoord = (regionIndex - Offset) << ShiftBy
	// Return center of region (+ RegionSize/2)
	x = ((rx - OffsetX) << ShiftBy) + (RegionSize / 2)
	y = ((ry - OffsetY) << ShiftBy) + (RegionSize / 2)
	return x, y
}
