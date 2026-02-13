package zone

import "github.com/udisondev/la2go/internal/model"

// JailZone represents the prison area.
// Java reference: JailZone.java — sets ZoneId.JAIL + NO_SUMMON_FRIEND.
type JailZone struct {
	*BaseZone
}

// NewJailZone creates a JailZone with onEnter/onExit callbacks.
func NewJailZone(base *BaseZone) *JailZone {
	z := &JailZone{BaseZone: base}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}

// IsPeace returns false.
func (z *JailZone) IsPeace() bool { return false }

// AllowsPvP returns false.
func (z *JailZone) AllowsPvP() bool { return false }

func (z *JailZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDJail, true)
	creature.SetInsideZone(model.ZoneIDNoSummonFriend, true)
}

func (z *JailZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDJail, false)
	creature.SetInsideZone(model.ZoneIDNoSummonFriend, false)
}
