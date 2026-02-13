package zone

import "github.com/udisondev/la2go/internal/model"

// WaterZone represents a body of water in the game world.
// Java reference: WaterZone.java
type WaterZone struct {
	*BaseZone
}

// NewWaterZone creates a WaterZone with onEnter/onExit callbacks.
func NewWaterZone(base *BaseZone) *WaterZone {
	z := &WaterZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}

// IsPeace returns false — water zones are not peace zones.
func (z *WaterZone) IsPeace() bool { return false }

// AllowsPvP returns true — PvP is allowed in water.
func (z *WaterZone) AllowsPvP() bool { return true }

// IsWater returns true — this is a water zone.
func (z *WaterZone) IsWater() bool { return true }

// onEnter sets WATER flag.
// Java reference: WaterZone.onEnter
func (z *WaterZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDWater, true)
}

// onExit clears WATER flag.
func (z *WaterZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDWater, false)
}
