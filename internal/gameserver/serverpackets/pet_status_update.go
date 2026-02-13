package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodePetStatusUpdate is the opcode for PetStatusUpdate (S2C 0xB5).
	OpcodePetStatusUpdate = 0xB5
)

// PetStatusUpdate sends pet HP/MP/exp/feed status to owner.
// Phase 19: Pets/Summons System.
// Java reference: PetStatusUpdate.java
type PetStatusUpdate struct {
	Summon *model.Summon

	// Pet-specific data (set for pets, zero for servitors)
	CurrentFed int32
	MaxFed     int32
	Exp        int64
	ExpMax     int64
}

// NewPetStatusUpdate creates a PetStatusUpdate for a servitor (no feed).
func NewPetStatusUpdate(summon *model.Summon) PetStatusUpdate {
	return PetStatusUpdate{Summon: summon}
}

// NewPetStatusUpdateWithFeed creates PetStatusUpdate with pet feed data.
func NewPetStatusUpdateWithFeed(summon *model.Summon, currentFed, maxFed int32, exp, expMax int64) PetStatusUpdate {
	return PetStatusUpdate{
		Summon:     summon,
		CurrentFed: currentFed,
		MaxFed:     maxFed,
		Exp:        exp,
		ExpMax:     expMax,
	}
}

// Write serializes PetStatusUpdate packet.
func (p *PetStatusUpdate) Write() ([]byte, error) {
	w := packet.NewWriter(64)

	w.WriteByte(OpcodePetStatusUpdate)

	// Summon type
	w.WriteInt(int32(p.Summon.Type()))

	// ObjectID
	w.WriteInt(int32(p.Summon.ObjectID()))

	// Position
	loc := p.Summon.Location()
	w.WriteInt(loc.X)
	w.WriteInt(loc.Y)
	w.WriteInt(loc.Z)

	// Title (owner name)
	w.WriteString("")

	// Feed
	w.WriteInt(p.CurrentFed)
	w.WriteInt(p.MaxFed)

	// HP / MP
	w.WriteInt(p.Summon.CurrentHP())
	w.WriteInt(p.Summon.MaxHP())
	w.WriteInt(p.Summon.CurrentMP())
	w.WriteInt(p.Summon.MaxMP())

	// Level
	w.WriteInt(p.Summon.Level())

	// Experience
	w.WriteLong(p.Exp)
	w.WriteLong(p.ExpMax)

	return w.Bytes(), nil
}
