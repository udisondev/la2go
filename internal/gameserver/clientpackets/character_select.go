package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const (
	// OpcodeCharacterSelect is the opcode for CharacterSelect packet (C2S 0x0D)
	OpcodeCharacterSelect = 0x0D
)

// CharacterSelect represents the CharacterSelect packet sent by client.
// Client sends this when user selects a character from the character list.
type CharacterSelect struct {
	CharSlot int32 // Character slot index (0-7)
}

// ParseCharacterSelect parses CharacterSelect packet from raw bytes.
// Packet structure: [opcode:1] [charSlot:4]
func ParseCharacterSelect(data []byte) (*CharacterSelect, error) {
	r := packet.NewReader(data)

	charSlot, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading charSlot: %w", err)
	}

	return &CharacterSelect{
		CharSlot: charSlot,
	}, nil
}
