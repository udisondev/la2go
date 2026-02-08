package world

import (
	"sync"

	"github.com/udisondev/la2go/internal/model"
)

// Region represents a single world region (2048×2048 game units)
type Region struct {
	rx, ry int32 // region coordinates

	mu             sync.RWMutex
	visibleObjects sync.Map // map[uint32]*model.WorldObject — objectID → object

	surroundingRegions []*Region // 3×3 window (9 regions max, excluding nil)
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

// AddVisibleObject adds object to region's visible objects (concurrent-safe)
func (r *Region) AddVisibleObject(obj *model.WorldObject) {
	r.visibleObjects.Store(obj.ObjectID(), obj)
}

// RemoveVisibleObject removes object from region's visible objects (concurrent-safe)
func (r *Region) RemoveVisibleObject(objectID uint32) {
	r.visibleObjects.Delete(objectID)
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
