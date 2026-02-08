package world

import "github.com/udisondev/la2go/internal/model"

// ForEachVisibleObject iterates over all objects visible from location (x, y)
// Visibility = current region + 8 surrounding regions (3×3 window)
// fn receives WorldObject pointer
// If fn returns false, iteration stops early
func ForEachVisibleObject(world *World, x, y int32, fn func(*model.WorldObject) bool) {
	region := world.GetRegion(x, y)
	if region == nil {
		return // invalid coordinates
	}

	// Iterate over current region + surrounding regions
	surrounding := region.SurroundingRegions()
	for _, r := range surrounding {
		if r == nil {
			continue
		}

		continueIterating := true
		r.ForEachVisibleObject(func(obj *model.WorldObject) bool {
			if !fn(obj) {
				continueIterating = false
				return false
			}
			return true
		})

		if !continueIterating {
			break
		}
	}
}

// CountVisibleObjects counts objects visible from location (x, y)
// Convenience function for testing
func CountVisibleObjects(world *World, x, y int32) int {
	count := 0
	ForEachVisibleObject(world, x, y, func(obj *model.WorldObject) bool {
		count++
		return true
	})
	return count
}
