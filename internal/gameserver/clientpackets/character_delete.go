package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeCharacterDelete is the opcode for CharacterDelete (C2S 0x0C).
// Java reference: ClientPackets.CHARACTER_DELETE(0x0C).
const OpcodeCharacterDelete = 0x0C

// CharacterDelete represents the client request to delete a character.
//
// Packet structure (body after opcode):
//   - charSlot (int32) — character slot index (0-6)
type CharacterDelete struct {
	CharSlot int32
}

// ParseCharacterDelete parses CharacterDelete packet.
func ParseCharacterDelete(data []byte) (*CharacterDelete, error) {
	r := packet.NewReader(data)

	charSlot, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading charSlot: %w", err)
	}

	return &CharacterDelete{CharSlot: charSlot}, nil
}
