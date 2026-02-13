package zone

import "github.com/udisondev/la2go/internal/model"

// PvPZone represents a combat area where PvP is explicitly allowed (Siege, Arena).
// Java reference: SiegeZone.java, ArenaZone.java
type PvPZone struct {
	*BaseZone
}

// NewPvPZone creates a PvPZone with onEnter/onExit callbacks.
func NewPvPZone(base *BaseZone) *PvPZone {
	z := &PvPZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}

// IsPeace returns false — PvP zones are not peaceful.
func (z *PvPZone) IsPeace() bool { return false }

// AllowsPvP returns true — PvP is explicitly allowed.
func (z *PvPZone) AllowsPvP() bool { return true }

// onEnter sets PVP, SIEGE, and NO_SUMMON_FRIEND flags for siege zones.
// Java reference: SiegeZone.onEnter (PVP + SIEGE + NO_SUMMON_FRIEND when siege active)
func (z *PvPZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDPVP, true)

	if z.zoneType == TypeSiege {
		creature.SetInsideZone(model.ZoneIDSiege, true)
		creature.SetInsideZone(model.ZoneIDNoSummonFriend, true)
	}
}

// onExit clears PVP, SIEGE, and NO_SUMMON_FRIEND flags.
func (z *PvPZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDPVP, false)

	if z.zoneType == TypeSiege {
		creature.SetInsideZone(model.ZoneIDSiege, false)
		creature.SetInsideZone(model.ZoneIDNoSummonFriend, false)
	}
}
