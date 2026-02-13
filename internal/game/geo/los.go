package geo

// CanSeeTarget checks line of sight between two world positions.
// Uses 3D Bresenham to trace through geodata cells and check obstacles.
// Returns true if no obstacle blocks the view.
//
// Java reference: GeoEngine.canSeeTarget() (lines 513-629).
func (e *Engine) CanSeeTarget(x1, y1, z1, x2, y2, z2 int32) bool {
	if !e.IsLoaded() {
		return true // No geodata — assume clear LOS
	}

	gx1 := GeoX(x1)
	gy1 := GeoY(y1)
	gx2 := GeoX(x2)
	gy2 := GeoY(y2)

	nearestFromZ := e.getNearestZ(gx1, gy1, z1)
	nearestToZ := e.getNearestZ(gx2, gy2, z2)

	// Same cell — check if same Z layer
	if gx1 == gx2 && gy1 == gy2 {
		if !e.hasGeoData(gx2, gy2) {
			return true
		}
		return nearestFromZ == nearestToZ
	}

	// Swap so we always go from higher to lower Z (better see-over logic)
	startGX, startGY, startZ := gx1, gy1, nearestFromZ
	endGX, endGY, endZ := gx2, gy2, nearestToZ
	if nearestToZ > nearestFromZ {
		startGX, startGY, startZ = gx2, gy2, nearestToZ
		endGX, endGY, endZ = gx1, gy1, nearestFromZ
	}

	it := NewLineIterator3D(startGX, startGY, startZ, endGX, endGY, endZ)
	it.Next() // Skip start point

	prevX := it.X()
	prevY := it.Y()
	prevGeoZ := startZ
	pointIndex := 0

	for it.Next() {
		curX := it.X()
		curY := it.Y()

		// Skip duplicates (same cell)
		if curX == prevX && curY == prevY {
			continue
		}

		beeZ := it.Z() // Expected Z along the straight line

		curGeoZ := prevGeoZ

		if e.hasGeoData(curX, curY) {
			nswe := ComputeNSWE(prevX, prevY, curX, curY)

			// Check NSWE movement permission and get adjusted Z
			curGeoZ = e.getLosGeoZ(prevX, prevY, prevGeoZ, curX, curY, nswe)

			// Calculate max allowed height for see-over
			var maxHeight int32
			if pointIndex < ElevatedSeeOverDist {
				maxHeight = startZ + MaxSeeOverHeight
			} else {
				maxHeight = beeZ + MaxSeeOverHeight
			}

			if curGeoZ > maxHeight {
				return false // Obstacle too high
			}

			// For diagonal movement, check corner-cut
			if nswe == NSWENorthEast || nswe == NSWENorthWest ||
				nswe == NSWESouthEast || nswe == NSWESouthWest {
				if !e.checkDiagonalLOS(prevX, prevY, prevGeoZ, curX, curY, nswe, maxHeight) {
					return false
				}
			}
		}

		prevX = curX
		prevY = curY
		prevGeoZ = curGeoZ
		pointIndex++
	}

	return true
}

// CanMoveToTarget checks if direct movement from (x1,y1,z1) to (x2,y2,z2)
// is possible (no walls blocking path).
// Uses 2D Bresenham (ignoring Z for wall check, but validating NSWE).
func (e *Engine) CanMoveToTarget(x1, y1, z1, x2, y2, z2 int32) bool {
	if !e.IsLoaded() {
		return true
	}

	gx1 := GeoX(x1)
	gy1 := GeoY(y1)
	gx2 := GeoX(x2)
	gy2 := GeoY(y2)

	curZ := e.getNearestZ(gx1, gy1, z1)

	it := NewLineIterator3D(gx1, gy1, curZ, gx2, gy2, e.getNearestZ(gx2, gy2, z2))
	it.Next() // Skip start

	prevX := gx1
	prevY := gy1

	for it.Next() {
		cx := it.X()
		cy := it.Y()

		if cx == prevX && cy == prevY {
			continue
		}

		if !e.hasGeoData(cx, cy) {
			prevX, prevY = cx, cy
			continue
		}

		nswe := ComputeNSWE(prevX, prevY, cx, cy)
		prevNSWE := e.getNSWE(prevX, prevY, curZ)

		// Check if movement in this direction is allowed from previous cell
		if prevNSWE&nswe == 0 {
			return false
		}

		// Get height at new cell
		newZ := e.getNearestZ(cx, cy, curZ)
		if abs32(newZ-curZ) > HeightIncrLimit {
			return false // Height difference too large
		}

		curZ = newZ
		prevX = cx
		prevY = cy
	}

	return true
}

// getLosGeoZ gets the Z considering NSWE movement permission for LOS.
// When movement is allowed, returns nearest Z at target cell.
// When blocked, returns next higher Z (wall top height) for accurate obstacle detection.
func (e *Engine) getLosGeoZ(prevX, prevY, prevZ, curX, curY int32, nswe byte) int32 {
	prevNSWE := e.getNSWE(prevX, prevY, prevZ)

	if prevNSWE&nswe != 0 {
		// Movement allowed — return nearest Z at target cell
		return e.getNearestZ(curX, curY, prevZ)
	}

	// Movement blocked — return the wall top height (next higher Z)
	return e.getNextHigherZ(curX, curY, prevZ)
}

// checkDiagonalLOS verifies that a diagonal LOS step doesn't cut through walls.
// For diagonal movement NE/NW/SE/SW, both adjacent cardinal cells must be passable.
func (e *Engine) checkDiagonalLOS(prevX, prevY, prevZ, _, _ int32, nswe byte, maxHeight int32) bool {
	switch nswe {
	case NSWENorthEast:
		// Check North cell and East cell
		nZ := e.getLosGeoZ(prevX, prevY, prevZ, prevX, prevY-1, NSWENorth)
		eZ := e.getLosGeoZ(prevX, prevY, prevZ, prevX+1, prevY, NSWEEast)
		return nZ <= maxHeight && eZ <= maxHeight

	case NSWENorthWest:
		nZ := e.getLosGeoZ(prevX, prevY, prevZ, prevX, prevY-1, NSWENorth)
		wZ := e.getLosGeoZ(prevX, prevY, prevZ, prevX-1, prevY, NSWEWest)
		return nZ <= maxHeight && wZ <= maxHeight

	case NSWESouthEast:
		sZ := e.getLosGeoZ(prevX, prevY, prevZ, prevX, prevY+1, NSWESouth)
		eZ := e.getLosGeoZ(prevX, prevY, prevZ, prevX+1, prevY, NSWEEast)
		return sZ <= maxHeight && eZ <= maxHeight

	case NSWESouthWest:
		sZ := e.getLosGeoZ(prevX, prevY, prevZ, prevX, prevY+1, NSWESouth)
		wZ := e.getLosGeoZ(prevX, prevY, prevZ, prevX-1, prevY, NSWEWest)
		return sZ <= maxHeight && wZ <= maxHeight
	}

	return true
}
