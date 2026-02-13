package zone

import "github.com/udisondev/la2go/internal/model"

// CastleZone represents a castle area.
// Java reference: CastleZone.java — sets ZoneId.CASTLE only.
type CastleZone struct {
	*BaseZone
	residenceID int32
}

// NewCastleZone creates a CastleZone with onEnter/onExit callbacks.
func NewCastleZone(base *BaseZone) *CastleZone {
	z := &CastleZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}

// IsPeace returns false — CastleZone itself is not a peace zone.
func (z *CastleZone) IsPeace() bool { return false }

// AllowsPvP returns true — PvP control is via separate zones.
func (z *CastleZone) AllowsPvP() bool { return true }

func (z *CastleZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDCastle, true)
}

func (z *CastleZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDCastle, false)
}

// ResidenceID returns the castle ID for this zone.
func (z *CastleZone) ResidenceID() int32 { return z.residenceID }
