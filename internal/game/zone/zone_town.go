package zone

import "github.com/udisondev/la2go/internal/model"

// TownZone represents a town area.
// Java reference: TownZone.java — sets ZoneId.TOWN only (NOT PEACE).
// In the game world, towns have separate overlapping PeaceZone + TownZone.
type TownZone struct {
	*BaseZone
	townID int32
	taxByID int32
}

// NewTownZone creates a TownZone with onEnter/onExit callbacks.
func NewTownZone(base *BaseZone) *TownZone {
	z := &TownZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}

// IsPeace returns false — TownZone itself is not a peace zone (separate PeaceZone handles that).
func (z *TownZone) IsPeace() bool { return false }

// AllowsPvP returns true — PvP control is via PeaceZone, not TownZone.
func (z *TownZone) AllowsPvP() bool { return true }

func (z *TownZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDTown, true)
}

func (z *TownZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDTown, false)
}
