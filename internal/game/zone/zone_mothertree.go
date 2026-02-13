package zone

import (
	"strconv"

	"github.com/udisondev/la2go/internal/model"
)

// MotherTreeZone represents elf village Mother Tree areas with HP/MP regen bonus.
// Java reference: MotherTreeZone.java — sets ZoneId.MOTHER_TREE.
type MotherTreeZone struct {
	*BaseZone
	hpRegen int32
	mpRegen int32
}

// NewMotherTreeZone creates a MotherTreeZone with onEnter/onExit callbacks.
func NewMotherTreeZone(base *BaseZone) *MotherTreeZone {
	z := &MotherTreeZone{BaseZone: base}
	if v, ok := base.params["HpRegenBonus"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			z.hpRegen = int32(n)
		}
	}
	if v, ok := base.params["MpRegenBonus"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			z.mpRegen = int32(n)
		}
	}
	z.onEnterFn = z.onEnter
	z.onExitFn = z.onExit
	return z
}

// IsPeace returns false.
func (z *MotherTreeZone) IsPeace() bool { return false }

// AllowsPvP returns true.
func (z *MotherTreeZone) AllowsPvP() bool { return true }

// HpRegenBonus returns the HP regen bonus for this zone.
func (z *MotherTreeZone) HpRegenBonus() int32 { return z.hpRegen }

// MpRegenBonus returns the MP regen bonus for this zone.
func (z *MotherTreeZone) MpRegenBonus() int32 { return z.mpRegen }

func (z *MotherTreeZone) onEnter(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDMotherTree, true)
}

func (z *MotherTreeZone) onExit(creature *model.Character) {
	creature.SetInsideZone(model.ZoneIDMotherTree, false)
}
