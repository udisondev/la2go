package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

const (
	// OpcodePetStatusShow is the opcode for PetStatusShow (S2C 0xB0).
	OpcodePetStatusShow = 0xB0
)

// PetStatusShow toggles the pet stat window on the client.
// Phase 19: Pets/Summons System.
// Java reference: PetStatusShow.java
type PetStatusShow struct {
	SummonType int32 // 2=pet, 1=servitor
}

// NewPetStatusShow creates a PetStatusShow packet.
func NewPetStatusShow(summonType int32) PetStatusShow {
	return PetStatusShow{SummonType: summonType}
}

// Write serializes PetStatusShow packet.
func (p *PetStatusShow) Write() ([]byte, error) {
	w := packet.NewWriter(8)

	w.WriteByte(OpcodePetStatusShow)
	w.WriteInt(p.SummonType)

	return w.Bytes(), nil
}
