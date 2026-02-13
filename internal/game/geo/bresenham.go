package geo

// LineIterator3D implements 3D Bresenham line algorithm for LOS checks.
// Steps through geo cells along a 3D line from start to end.
// Java reference: GridLineIterator3D.java:68-199.
type LineIterator3D struct {
	currentX, currentY, currentZ int32
	targetX, targetY, targetZ   int32
	deltaX, deltaY, deltaZ      int32
	stepX, stepY, stepZ         int32
	errorXY, errorXZ            int32
	dominant                    int // 0=X, 1=Y, 2=Z
	started                     bool
}

// NewLineIterator3D creates a 3D Bresenham line iterator.
func NewLineIterator3D(sx, sy, sz, ex, ey, ez int32) *LineIterator3D {
	it := &LineIterator3D{
		currentX: sx, currentY: sy, currentZ: sz,
		targetX: ex, targetY: ey, targetZ: ez,
	}

	it.deltaX = abs32(ex - sx)
	it.deltaY = abs32(ey - sy)
	it.deltaZ = abs32(ez - sz)

	if sx < ex {
		it.stepX = 1
	} else {
		it.stepX = -1
	}
	if sy < ey {
		it.stepY = 1
	} else {
		it.stepY = -1
	}
	if sz < ez {
		it.stepZ = 1
	} else {
		it.stepZ = -1
	}

	// Determine dominant axis and init error terms.
	if it.deltaX >= it.deltaY && it.deltaX >= it.deltaZ {
		it.dominant = 0
		it.errorXY = it.deltaX / 2
		it.errorXZ = it.deltaX / 2
	} else if it.deltaY >= it.deltaX && it.deltaY >= it.deltaZ {
		it.dominant = 1
		it.errorXY = it.deltaY / 2
		it.errorXZ = it.deltaY / 2
	} else {
		it.dominant = 2
		it.errorXY = it.deltaZ / 2
		it.errorXZ = it.deltaZ / 2
	}

	return it
}

// Next advances the iterator to the next cell.
// Returns false when the target is reached.
func (it *LineIterator3D) Next() bool {
	if !it.started {
		it.started = true
		return true // Return start point
	}

	if it.currentX == it.targetX && it.currentY == it.targetY && it.currentZ == it.targetZ {
		return false
	}

	switch it.dominant {
	case 0: // X-dominant
		it.currentX += it.stepX
		it.errorXY += it.deltaY
		if it.errorXY >= it.deltaX {
			it.currentY += it.stepY
			it.errorXY -= it.deltaX
		}
		it.errorXZ += it.deltaZ
		if it.errorXZ >= it.deltaX {
			it.currentZ += it.stepZ
			it.errorXZ -= it.deltaX
		}

	case 1: // Y-dominant
		it.currentY += it.stepY
		it.errorXY += it.deltaX
		if it.errorXY >= it.deltaY {
			it.currentX += it.stepX
			it.errorXY -= it.deltaY
		}
		it.errorXZ += it.deltaZ
		if it.errorXZ >= it.deltaY {
			it.currentZ += it.stepZ
			it.errorXZ -= it.deltaY
		}

	case 2: // Z-dominant
		it.currentZ += it.stepZ
		it.errorXY += it.deltaX
		if it.errorXY >= it.deltaZ {
			it.currentX += it.stepX
			it.errorXY -= it.deltaZ
		}
		it.errorXZ += it.deltaY
		if it.errorXZ >= it.deltaZ {
			it.currentY += it.stepY
			it.errorXZ -= it.deltaZ
		}
	}

	return true
}

// X returns current X position.
func (it *LineIterator3D) X() int32 { return it.currentX }

// Y returns current Y position.
func (it *LineIterator3D) Y() int32 { return it.currentY }

// Z returns current Z position.
func (it *LineIterator3D) Z() int32 { return it.currentZ }

func abs32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}
