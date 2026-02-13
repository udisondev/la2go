package geo

// GeoX converts world X coordinate to geodata X.
func GeoX(worldX int32) int32 {
	return (worldX - WorldMinX) / CoordinateScale
}

// GeoY converts world Y coordinate to geodata Y.
func GeoY(worldY int32) int32 {
	return (worldY - WorldMinY) / CoordinateScale
}

// WorldX converts geodata X to world X (centered in cell).
func WorldX(geoX int32) int32 {
	return geoX*CoordinateScale + WorldMinX + CoordinateOffset
}

// WorldY converts geodata Y to world Y (centered in cell).
func WorldY(geoY int32) int32 {
	return geoY*CoordinateScale + WorldMinY + CoordinateOffset
}

// RegionXY returns region indices from geo coordinates.
func RegionXY(geoX, geoY int32) (int32, int32) {
	return geoX / RegionCellsX, geoY / RegionCellsY
}

// BlockXY returns block index within region from geo coordinates.
func BlockXY(geoX, geoY int32) int32 {
	localX := (geoX % RegionCellsX) / BlockCellsX
	localY := (geoY % RegionCellsY) / BlockCellsY
	return localX*RegionBlocksY + localY
}

// CellXY returns cell index within block from geo coordinates.
func CellXY(geoX, geoY int32) (int32, int32) {
	return geoX % BlockCellsX, geoY % BlockCellsY
}

// ComputeNSWE computes the NSWE direction from (fromX,fromY) to (toX,toY).
func ComputeNSWE(fromX, fromY, toX, toY int32) byte {
	var nswe byte
	if toX > fromX {
		nswe |= NSWEEast
	} else if toX < fromX {
		nswe |= NSWEWest
	}
	if toY > fromY {
		nswe |= NSWESouth
	} else if toY < fromY {
		nswe |= NSWENorth
	}
	return nswe
}
