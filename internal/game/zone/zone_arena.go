package zone

import "github.com/udisondev/la2go/internal/model"

// ArenaZone represents a PvP arena.
// Java reference: ArenaZone.java — sets ZoneId.PVP.
type ArenaZone struct {
	*BaseZone
}

// NewArenaZone creates an ArenaZone with onEnter/onExit callbacks.
func NewArenaZone(base *BaseZone) *ArenaZone {
	z := &ArenaZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}

// IsPeace returns false — arena is a combat zone.
func (z *ArenaZone) IsPeace() bool { return false }

// AllowsPvP returns true — PvP is allowed in arenas.
func (z *ArenaZone) AllowsPvP() bool { return true }

func (z *ArenaZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDPVP, true)
}

func (z *ArenaZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDPVP, false)
}
