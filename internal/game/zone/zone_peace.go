package zone

import "github.com/udisondev/la2go/internal/model"

// PeaceZone represents a safe area where PvP is not allowed.
// Java reference: PeaceZone.java — sets ZoneId.PEACE only.
// Note: TownZone and CastleZone are separate types that set their own flags.
type PeaceZone struct {
	*BaseZone
}

// NewPeaceZone creates a PeaceZone with onEnter/onExit callbacks.
func NewPeaceZone(base *BaseZone) *PeaceZone {
	z := &PeaceZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}

// IsPeace returns true — peace zones prohibit combat.
func (z *PeaceZone) IsPeace() bool { return true }

// AllowsPvP returns false — PvP is forbidden in peace zones.
func (z *PeaceZone) AllowsPvP() bool { return false }

// onEnter sets PEACE flag.
// Java reference: PeaceZone.onEnter
func (z *PeaceZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDPeace, true)
}

// onExit clears PEACE flag.
func (z *PeaceZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDPeace, false)
}
