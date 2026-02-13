package zone

import (
	"strconv"

	"github.com/udisondev/la2go/internal/model"
)

// OlympiadStadiumZone represents an Olympiad arena.
// Java reference: OlympiadStadiumZone.java — sets 5 zone flags.
type OlympiadStadiumZone struct {
	*BaseZone
	stadiumID int32
}

// NewOlympiadStadiumZone creates an OlympiadStadiumZone with onEnter/onExit callbacks.
func NewOlympiadStadiumZone(base *BaseZone) *OlympiadStadiumZone {
	z := &OlympiadStadiumZone{BaseZone: base}
	if v, ok := base.params["stadiumId"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			z.stadiumID = int32(n)
		}
	}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}

// IsPeace returns false — olympiad is a combat zone.
func (z *OlympiadStadiumZone) IsPeace() bool { return false }

// AllowsPvP returns true.
func (z *OlympiadStadiumZone) AllowsPvP() bool { return true }

// StadiumID returns the stadium ID.
func (z *OlympiadStadiumZone) StadiumID() int32 { return z.stadiumID }

func (z *OlympiadStadiumZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDPVP, true)
	creature.SetInsideZone(model.ZoneIDNoSummonFriend, true)
	creature.SetInsideZone(model.ZoneIDNoLanding, true)
	creature.SetInsideZone(model.ZoneIDNoRestart, true)
	creature.SetInsideZone(model.ZoneIDNoBookmark, true)
}

func (z *OlympiadStadiumZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDPVP, false)
	creature.SetInsideZone(model.ZoneIDNoSummonFriend, false)
	creature.SetInsideZone(model.ZoneIDNoLanding, false)
	creature.SetInsideZone(model.ZoneIDNoRestart, false)
	creature.SetInsideZone(model.ZoneIDNoBookmark, false)
}
