package zone

import "github.com/udisondev/la2go/internal/model"

// ClanHallZone represents a clan hall area.
// Java reference: ClanHallZone.java — sets ZoneId.CLAN_HALL.
type ClanHallZone struct {
	*BaseZone
	residenceID int32
}

// NewClanHallZone creates a ClanHallZone with onEnter/onExit callbacks.
func NewClanHallZone(base *BaseZone) *ClanHallZone {
	z := &ClanHallZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}

// IsPeace returns false.
func (z *ClanHallZone) IsPeace() bool { return false }

// AllowsPvP returns true.
func (z *ClanHallZone) AllowsPvP() bool { return true }

func (z *ClanHallZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDClanHall, true)
}

func (z *ClanHallZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDClanHall, false)
}
