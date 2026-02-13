package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeCharacterRestore is the opcode for CharacterRestore (C2S 0x62).
// Java reference: ClientPackets.CHARACTER_RESTORE(0x62).
const OpcodeCharacterRestore = 0x62

// CharacterRestore represents the client request to restore a deleted character.
//
// Packet structure (body after opcode):
//   - charSlot (int32) — character slot index (0-6)
type CharacterRestore struct {
	CharSlot int32
}

// ParseCharacterRestore parses CharacterRestore packet.
func ParseCharacterRestore(data []byte) (*CharacterRestore, error) {
	r := packet.NewReader(data)

	charSlot, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading charSlot: %w", err)
	}

	return &CharacterRestore{CharSlot: charSlot}, nil
}
