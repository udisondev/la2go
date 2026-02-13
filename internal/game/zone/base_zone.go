package zone

import (
	"sync"

	"github.com/udisondev/la2go/internal/model"
)

// onEnterExitFunc is the callback signature for zone enter/exit events.
type onEnterExitFunc func(creature *model.Character)

// BaseZone holds common zone geometry, metadata, and character tracking.
// Specific zone types embed BaseZone and register onEnter/onExit callbacks.
// Java reference: model/zone/ZoneType.java
type BaseZone struct {
	id       int32
	name     string
	zoneType string
	shape    string // "NPoly", "Cuboid", or "Cylinder"
	minZ     int32
	maxZ     int32
	nodesX   []int32
	nodesY   []int32
	rad      int32 // radius for Cylinder shape
	params   map[string]string

	// Character tracking (Phase 11).
	// Stores all creatures currently inside this zone.
	// Key: ObjectID (uint32), Value: *model.Character.
	// Java reference: ZoneType._characterList (ConcurrentHashMap<Integer, Creature>)
	characters sync.Map

	// Callbacks set by concrete zone types.
	onEnterFn onEnterExitFunc
	onExitFn  onEnterExitFunc
}

// ID returns the zone identifier.
func (z *BaseZone) ID() int32 { return z.id }

// Name returns the zone display name.
func (z *BaseZone) Name() string { return z.name }

// ZoneType returns the zone type string (e.g. "TownZone", "SiegeZone").
func (z *BaseZone) ZoneType() string { return z.zoneType }

// Params returns the zone parameter map.
func (z *BaseZone) Params() map[string]string { return z.params }

// Contains checks if point (x, y, zCoord) is inside the zone geometry.
// For "NPoly" shape — uses ray casting (point-in-polygon) algorithm.
// For "Cuboid" shape — uses axis-aligned bounding box check.
// For "Cylinder" shape — uses center + radius circle check.
func (z *BaseZone) Contains(x, y, zCoord int32) bool {
	if zCoord < z.minZ || zCoord > z.maxZ {
		return false
	}

	n := len(z.nodesX)
	if n == 0 {
		return false
	}

	switch z.shape {
	case "Cuboid":
		return z.containsCuboid(x, y)
	case "Cylinder":
		return z.containsCylinder(x, y)
	default:
		return z.containsNPoly(x, y)
	}
}

// containsCuboid проверяет попадание в AABB (axis-aligned bounding box).
func (z *BaseZone) containsCuboid(x, y int32) bool {
	if len(z.nodesX) < 2 {
		return false
	}

	minX, maxX := z.nodesX[0], z.nodesX[0]
	minY, maxY := z.nodesY[0], z.nodesY[0]

	for i := 1; i < len(z.nodesX); i++ {
		if z.nodesX[i] < minX {
			minX = z.nodesX[i]
		}
		if z.nodesX[i] > maxX {
			maxX = z.nodesX[i]
		}
		if z.nodesY[i] < minY {
			minY = z.nodesY[i]
		}
		if z.nodesY[i] > maxY {
			maxY = z.nodesY[i]
		}
	}

	return x >= minX && x <= maxX && y >= minY && y <= maxY
}

// containsCylinder проверяет попадание точки в цилиндр (center + radius).
// Java reference: ZoneCylinder.isInsideZone
func (z *BaseZone) containsCylinder(x, y int32) bool {
	if len(z.nodesX) == 0 || z.rad <= 0 {
		return false
	}
	dx := int64(x - z.nodesX[0])
	dy := int64(y - z.nodesY[0])
	r := int64(z.rad)
	return dx*dx+dy*dy <= r*r
}

// containsNPoly проверяет попадание точки в полигон алгоритмом ray casting.
func (z *BaseZone) containsNPoly(x, y int32) bool {
	n := len(z.nodesX)
	count := 0
	j := n - 1

	for i := range n {
		if (z.nodesY[i] > y) != (z.nodesY[j] > y) {
			slope := int64(x-z.nodesX[i])*int64(z.nodesY[j]-z.nodesY[i]) -
				int64(z.nodesX[j]-z.nodesX[i])*int64(y-z.nodesY[i])

			if slope == 0 {
				// Точка лежит на границе полигона.
				return true
			}

			if (slope < 0) != (int64(z.nodesY[j]-z.nodesY[i]) < 0) {
				count++
			}
		}
		j = i
	}

	return count%2 == 1
}

// --- Character tracking (Phase 11) ---

// RevalidateInZone checks if creature is inside or outside the zone.
// If inside and not yet tracked — adds to character list and calls onEnter.
// If outside and currently tracked — removes from list and calls onExit.
// Java reference: ZoneType.revalidateInZone(Creature)
func (z *BaseZone) RevalidateInZone(creature *model.Character) {
	loc := creature.Location()
	if z.Contains(loc.X, loc.Y, loc.Z) {
		// Creature is inside — add if not already tracked
		_, loaded := z.characters.LoadOrStore(creature.ObjectID(), creature)
		if !loaded && z.onEnterFn != nil {
			z.onEnterFn(creature)
		}
	} else {
		z.RemoveCharacter(creature)
	}
}

// RemoveCharacter removes a creature from the zone's tracking list.
// Calls onExit if the creature was tracked.
// Java reference: ZoneType.removeCharacter(Creature)
func (z *BaseZone) RemoveCharacter(creature *model.Character) {
	_, loaded := z.characters.LoadAndDelete(creature.ObjectID())
	if loaded && z.onExitFn != nil {
		z.onExitFn(creature)
	}
}

// GetCharactersInside returns all characters currently tracked in this zone.
// Java reference: ZoneType.getCharactersInside()
func (z *BaseZone) GetCharactersInside() []*model.Character {
	var result []*model.Character
	z.characters.Range(func(_, value any) bool {
		if ch, ok := value.(*model.Character); ok {
			result = append(result, ch)
		}
		return true
	})
	return result
}

// GetPlayersInside returns only Player characters tracked in this zone.
// Java reference: ZoneType.getPlayersInside()
func (z *BaseZone) GetPlayersInside() []*model.Player {
	var result []*model.Player
	z.characters.Range(func(_, value any) bool {
		if ch, ok := value.(*model.Character); ok {
			if ch.WorldObject != nil && ch.WorldObject.Data != nil {
				if p, ok := ch.WorldObject.Data.(*model.Player); ok {
					result = append(result, p)
				}
			}
		}
		return true
	})
	return result
}

// CharacterCount returns the number of characters currently tracked in this zone.
func (z *BaseZone) CharacterCount() int {
	count := 0
	z.characters.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}
