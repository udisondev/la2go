package world

import (
	"fmt"
	"sync"

	"github.com/udisondev/la2go/internal/model"
)

// World represents the game world with 2D region grid
// Singleton pattern — use Instance() to access
type World struct {
	regions [][]*Region // 2D array [RegionsX][RegionsY]
	objects sync.Map    // map[uint32]*model.WorldObject — objectID → object
	npcs    sync.Map    // map[uint32]*model.Npc — objectID → npc (Phase 4.10 Part 2)
}

var (
	instance *World
	once     sync.Once
)

// Instance returns singleton World instance
func Instance() *World {
	once.Do(func() {
		instance = &World{}
		instance.initialize()
	})
	return instance
}

// initialize creates 2D region grid and sets up surrounding regions
func (w *World) initialize() {
	// Allocate 2D array
	w.regions = make([][]*Region, RegionsX)
	for rx := range RegionsX {
		w.regions[rx] = make([]*Region, RegionsY)
		for ry := range RegionsY {
			w.regions[rx][ry] = NewRegion(int32(rx), int32(ry))
		}
	}

	// Set surrounding regions for each region (3×3 window)
	for rx := range RegionsX {
		for ry := range RegionsY {
			region := w.regions[rx][ry]
			surrounding := w.getSurroundingRegions(int32(rx), int32(ry))
			region.SetSurroundingRegions(surrounding)
		}
	}
}

// getSurroundingRegions returns 3×3 window of regions around (rx, ry)
// Returns slice of valid regions (excluding out-of-bounds)
func (w *World) getSurroundingRegions(rx, ry int32) []*Region {
	surrounding := make([]*Region, 0, 9) // max 9 regions (3×3)

	for dx := int32(-1); dx <= 1; dx++ {
		for dy := int32(-1); dy <= 1; dy++ {
			nx := rx + dx
			ny := ry + dy

			if IsValidRegionIndex(nx, ny) {
				surrounding = append(surrounding, w.regions[nx][ny])
			}
		}
	}

	return surrounding
}

// GetRegion returns region at world coordinates (x, y)
// Returns nil if coordinates are out of bounds
func (w *World) GetRegion(x, y int32) *Region {
	rx, ry := CoordToRegionIndex(x, y)
	if !IsValidRegionIndex(rx, ry) {
		return nil
	}
	return w.regions[rx][ry]
}

// GetRegionByIndex returns region at region index (rx, ry)
// Returns nil if index is out of bounds
func (w *World) GetRegionByIndex(rx, ry int32) *Region {
	if !IsValidRegionIndex(rx, ry) {
		return nil
	}
	return w.regions[rx][ry]
}

// AddObject adds object to world and its region
// Returns error if region is invalid
func (w *World) AddObject(obj *model.WorldObject) error {
	loc := obj.Location()
	region := w.GetRegion(loc.X, loc.Y)
	if region == nil {
		return fmt.Errorf("invalid coordinates for object %d: (%d, %d)", obj.ObjectID(), loc.X, loc.Y)
	}

	w.objects.Store(obj.ObjectID(), obj)
	region.AddVisibleObject(obj)
	return nil
}

// AddNpc adds NPC to world and registers it in npcs map.
// Convenience method for adding NPCs (calls AddObject internally).
// Phase 4.10 Part 2: Separate NPC tracking for efficient GetNpc lookups.
func (w *World) AddNpc(npc *model.Npc) error {
	// Add to world grid (via WorldObject)
	if err := w.AddObject(npc.WorldObject); err != nil {
		return fmt.Errorf("adding npc to world: %w", err)
	}

	// Register in npcs map for fast lookup
	w.npcs.Store(npc.ObjectID(), npc)
	return nil
}

// RemoveObject removes object from world and its region
// Also removes from npcs map if object is an NPC (Phase 4.10 Part 2)
func (w *World) RemoveObject(objectID uint32) {
	value, ok := w.objects.LoadAndDelete(objectID)
	if !ok {
		return
	}

	obj := value.(*model.WorldObject)
	loc := obj.Location()
	region := w.GetRegion(loc.X, loc.Y)
	if region != nil {
		region.RemoveVisibleObject(objectID)
	}

	// Remove from npcs map if this is an NPC (Phase 4.10 Part 2)
	// Check ObjectID range instead of type assertion (more efficient)
	if objectID >= 0x20000000 { // NPC range
		w.npcs.Delete(objectID)
	}
}

// GetObject returns object by ID
func (w *World) GetObject(objectID uint32) (*model.WorldObject, bool) {
	value, ok := w.objects.Load(objectID)
	if !ok {
		return nil, false
	}
	return value.(*model.WorldObject), true
}

// GetNpc returns NPC by ObjectID.
// Returns nil, false if NPC not found or objectID is not in NPC range.
// Phase 4.10 Part 2: Efficient NPC lookup for sendVisibleObjectsInfo.
func (w *World) GetNpc(objectID uint32) (*model.Npc, bool) {
	value, ok := w.npcs.Load(objectID)
	if !ok {
		return nil, false
	}
	return value.(*model.Npc), true
}

// RegionCount returns total number of regions
func (w *World) RegionCount() int {
	return RegionsX * RegionsY
}

// ObjectCount returns total number of objects in world (O(N) — expensive!)
func (w *World) ObjectCount() int {
	count := 0
	w.objects.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}
