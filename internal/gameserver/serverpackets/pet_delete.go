package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

const (
	// OpcodePetDelete is the opcode for PetDelete (S2C 0xB6).
	OpcodePetDelete = 0xB6
)

// PetDelete tells the client to remove a pet/summon.
// Phase 19: Pets/Summons System.
// Java reference: PetDelete.java
type PetDelete struct {
	SummonType int32  // 2=pet, 1=servitor
	ObjectID   uint32 // pet objectID
}

// NewPetDelete creates a PetDelete packet.
func NewPetDelete(summonType int32, objectID uint32) PetDelete {
	return PetDelete{
		SummonType: summonType,
		ObjectID:   objectID,
	}
}

// Write serializes PetDelete packet.
func (p *PetDelete) Write() ([]byte, error) {
	w := packet.NewWriter(16)

	w.WriteByte(OpcodePetDelete)
	w.WriteInt(p.SummonType)
	w.WriteInt(int32(p.ObjectID))

	return w.Bytes(), nil
}
