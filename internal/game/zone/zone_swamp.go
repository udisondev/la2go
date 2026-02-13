package zone

import (
	"strconv"

	"github.com/udisondev/la2go/internal/model"
)

// SwampZone represents a swamp area that reduces movement speed.
// Java reference: SwampZone.java — sets ZoneId.SWAMP.
type SwampZone struct {
	*BaseZone
	moveBonus float64 // speed multiplier (default 0.5 = 50%)
	castleID  int32
}

// NewSwampZone creates a SwampZone with onEnter/onExit callbacks.
func NewSwampZone(base *BaseZone) *SwampZone {
	z := &SwampZone{BaseZone: base, moveBonus: 0.5}
	if v, ok := base.params["move_bonus"]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			z.moveBonus = f
		}
	}
	if v, ok := base.params["castleId"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			z.castleID = int32(n)
		}
	}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}

// IsPeace returns false.
func (z *SwampZone) IsPeace() bool { return false }

// AllowsPvP returns true.
func (z *SwampZone) AllowsPvP() bool { return true }

// MoveBonus returns the speed multiplier (e.g., 0.5 for 50% speed).
func (z *SwampZone) MoveBonus() float64 { return z.moveBonus }

func (z *SwampZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDSwamp, true)
}

func (z *SwampZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDSwamp, false)
}
