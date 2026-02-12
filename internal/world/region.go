package world

import (
	"sync"
	"sync/atomic"

	"github.com/udisondev/la2go/internal/model"
)

// Region represents a single world region (2048×2048 game units)
// Phase 4.11 Tier 2: Added snapshot cache for -70% visibility query latency.
type Region struct {
	rx, ry int32 // region coordinates

	mu             sync.RWMutex
	visibleObjects sync.Map // map[uint32]*model.WorldObject — objectID → object

	surroundingRegions []*Region // 3×3 window (9 regions max, excluding nil)

	// Phase 4.11 Tier 2: Snapshot cache (immutable slice)
	snapshotCache atomic.Value // []*model.WorldObject (immutable after rebuild)
	snapshotDirty atomic.Bool  // true if cache is stale (objects added/removed)

	// Phase 4.11 Tier 3: Version tracking for dirty region detection
	version atomic.Uint64 // incremented on Add/Remove (used for fingerprint)
}

// NewRegion creates a new region
func NewRegion(rx, ry int32) *Region {
	return &Region{
		rx: rx,
		ry: ry,
	}
}

// RX returns region X index
func (r *Region) RX() int32 {
	return r.rx
}

// RY returns region Y index
func (r *Region) RY() int32 {
	return r.ry
}

// Version returns current region version (incremented on Add/Remove).
// Phase 4.11 Tier 3: Used for fingerprint computation (skip cache update if unchanged).
func (r *Region) Version() uint64 {
	return r.version.Load()
}

// AddVisibleObject adds object to region's visible objects (concurrent-safe)
// Phase 4.11 Tier 2: Marks snapshot cache as dirty for lazy rebuild.
// Phase 4.11 Tier 3: Increments version for fingerprint tracking.
func (r *Region) AddVisibleObject(obj *model.WorldObject) {
	r.visibleObjects.Store(obj.ObjectID(), obj)
	r.version.Add(1)             // bump version (Tier 3)
	r.snapshotDirty.Store(true) // invalidate snapshot cache
}

// RemoveVisibleObject removes object from region's visible objects (concurrent-safe)
// Phase 4.11 Tier 2: Marks snapshot cache as dirty for lazy rebuild.
// Phase 4.11 Tier 3: Increments version for fingerprint tracking.
func (r *Region) RemoveVisibleObject(objectID uint32) {
	r.visibleObjects.Delete(objectID)
	r.version.Add(1)             // bump version (Tier 3)
	r.snapshotDirty.Store(true) // invalidate snapshot cache
}

// ForEachVisibleObject iterates over all visible objects in this region
// fn receives WorldObject pointer
// If fn returns false, iteration stops
func (r *Region) ForEachVisibleObject(fn func(*model.WorldObject) bool) {
	r.visibleObjects.Range(func(key, value any) bool {
		obj := value.(*model.WorldObject)
		return fn(obj)
	})
}

// SetSurroundingRegions sets surrounding regions (3×3 window)
// Called ONCE during world initialization
// IMPORTANT: After initialization, surroundingRegions is IMMUTABLE
func (r *Region) SetSurroundingRegions(regions []*Region) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.surroundingRegions = regions
}

// SurroundingRegions returns surrounding regions (ZERO-COPY immutable slice)
// IMPORTANT: Returned slice is immutable — DO NOT modify.
// surroundingRegions is set ONCE during World initialization and never changes.
func (r *Region) SurroundingRegions() []*Region {
	// Zero-copy: no mutex, no allocation
	// Safe because surroundingRegions is immutable after World.initialize()
	return r.surroundingRegions
}

// GetVisibleObjectsSnapshot returns cached snapshot of visible objects.
// Phase 4.11 Tier 2: Lazy rebuild if cache is dirty (-70% sync.Map overhead).
// IMPORTANT: Returned slice is immutable — DO NOT modify.
func (r *Region) GetVisibleObjectsSnapshot() []*model.WorldObject {
	// Fast path: cache is clean
	if !r.snapshotDirty.Load() {
		if cache := r.snapshotCache.Load(); cache != nil {
			return cache.([]*model.WorldObject)
		}
	}

	// Slow path: rebuild snapshot (cache is dirty or empty)
	return r.rebuildSnapshot()
}

// ClearVisibleObjects removes all visible objects from this region.
// Used for test isolation (World.Reset).
func (r *Region) ClearVisibleObjects() {
	r.visibleObjects.Range(func(key, _ any) bool {
		r.visibleObjects.Delete(key)
		return true
	})
	r.version.Add(1)
	r.snapshotDirty.Store(true)
	r.snapshotCache.Store(([]*model.WorldObject)(nil))
}

// rebuildSnapshot rebuilds snapshot cache from sync.Map.
// Phase 4.11 Tier 2: Called only when objects change (lazy rebuild).
func (r *Region) rebuildSnapshot() []*model.WorldObject {
	// Pre-allocate for typical case: 50 objects per region
	objects := make([]*model.WorldObject, 0, 64)

	// Collect all objects from sync.Map
	r.visibleObjects.Range(func(key, value any) bool {
		objects = append(objects, value.(*model.WorldObject))
		return true
	})

	// Store immutable snapshot
	r.snapshotCache.Store(objects)
	r.snapshotDirty.Store(false)

	return objects
}
