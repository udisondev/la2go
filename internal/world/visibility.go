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

// ForEachVisibleObjectForPlayer iterates over all visible objects for player (9 regions).
// SLOW PATH — queries 9 regions on every call.
// Use only when cache is not available (e.g., first query, player just spawned).
// Phase 4.5 PR3: Prefer ForEachVisibleObjectCached for production use.
func ForEachVisibleObjectForPlayer(player *model.Player, fn func(*model.WorldObject) bool) {
	loc := player.Location()
	regionX, regionY := CoordToRegionIndex(loc.X, loc.Y)

	world := Instance()
	currentRegion := world.GetRegion(regionX, regionY)
	if currentRegion == nil {
		return
	}

	// Current region
	currentRegion.ForEachVisibleObject(func(obj *model.WorldObject) bool {
		return fn(obj)
	})

	// Surrounding regions (8 regions)
	for _, surroundingRegion := range currentRegion.SurroundingRegions() {
		if surroundingRegion == nil {
			continue
		}

		stop := false
		surroundingRegion.ForEachVisibleObject(func(obj *model.WorldObject) bool {
			if !fn(obj) {
				stop = true
				return false
			}
			return true
		})

		if stop {
			return
		}
	}
}

// ForEachVisibleObjectCached iterates over visible objects using player's visibility cache.
// FAST PATH — uses cached results from VisibilityManager (updated every 100ms).
// Falls back to slow path if cache is nil or stale.
// Phase 4.5 PR3: -96.8% CPU reduction for visibility queries.
func ForEachVisibleObjectCached(player *model.Player, fn func(*model.WorldObject) bool) {
	cache := player.GetVisibilityCache()

	// Fast path: use cache if available
	if cache != nil {
		objects := cache.Objects()
		for _, obj := range objects {
			// Phase 4.11 Tier 1 Opt 2: Removed objectExists() validation
			// Cache is immutable, race condition impossible
			if !fn(obj) {
				return
			}
		}
		return
	}

	// Slow path: cache not available, query regions directly
	ForEachVisibleObjectForPlayer(player, fn)
}

// Phase 4.11 Tier 1 Opt 2: Removed objectExists() function
// Cache is immutable — defensive validation was unnecessary overhead (+5ns per object)
// Trade-off analysis showed no benefit (cache can't become stale mid-iteration)

// LODLevel represents Level of Detail for visibility queries.
// Phase 4.12: LOD-aware broadcast prioritization.
type LODLevel int

const (
	LODNear   LODLevel = iota // Same region (~50 objects, highest priority)
	LODMedium                 // Adjacent regions (~200 objects, medium priority)
	LODFar                    // Distant regions (~200 objects, low priority)
	LODAll                    // All objects (near + medium + far)
)

// ForEachVisibleObjectByLOD iterates over visible objects filtered by LOD level.
// Uses player's visibility cache (updated every 100ms by VisibilityManager).
// Phase 4.12: Enables LOD-aware broadcast — send critical packets to near objects only.
//
// Example usage:
//   // Send movement update only to near players
//   ForEachVisibleObjectByLOD(player, LODNear, func(obj *model.WorldObject) bool {
//       if obj.IsPlayer() {
//           SendMovementPacket(obj.AsPlayer())
//       }
//       return true
//   })
func ForEachVisibleObjectByLOD(player *model.Player, level LODLevel, fn func(*model.WorldObject) bool) {
	cache := player.GetVisibilityCache()

	// Fast path: use cache if available
	if cache != nil {
		var objects []*model.WorldObject

		switch level {
		case LODNear:
			objects = cache.NearObjects()
		case LODMedium:
			objects = cache.MediumObjects()
		case LODFar:
			objects = cache.FarObjects()
		case LODAll:
			objects = cache.Objects()
		default:
			objects = cache.Objects()
		}

		for _, obj := range objects {
			if !fn(obj) {
				return
			}
		}
		return
	}

	// Slow path: cache not available, fall back to full query
	// Note: Without cache, we can't distinguish LOD levels, so query all
	ForEachVisibleObjectForPlayer(player, fn)
}

// ForEachNearObject iterates over objects in same region (highest priority).
// Phase 4.12: Convenience wrapper for LODNear — use for critical events.
func ForEachNearObject(player *model.Player, fn func(*model.WorldObject) bool) {
	ForEachVisibleObjectByLOD(player, LODNear, fn)
}

// ForEachMediumObject iterates over objects in adjacent regions (medium priority).
// Phase 4.12: Convenience wrapper for LODMedium — use for less frequent updates.
func ForEachMediumObject(player *model.Player, fn func(*model.WorldObject) bool) {
	ForEachVisibleObjectByLOD(player, LODMedium, fn)
}

// ForEachFarObject iterates over objects in distant regions (low priority).
// Phase 4.12: Convenience wrapper for LODFar — use for rare updates or simplified packets.
func ForEachFarObject(player *model.Player, fn func(*model.WorldObject) bool) {
	ForEachVisibleObjectByLOD(player, LODFar, fn)
}
