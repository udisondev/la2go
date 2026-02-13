package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeRequestAutoSoulShot is the extended sub-opcode 0xD0:0x05.
// Client sends this to toggle auto-use of SoulShots/SpiritShots.
const SubOpcodeRequestAutoSoulShot int16 = 0x05

// RequestAutoSoulShot represents a toggle request for auto soulshot.
type RequestAutoSoulShot struct {
	ItemID int32 // SoulShot/SpiritShot item ID
	Type   int32 // 1 = enable, 0 = disable
}

// ParseRequestAutoSoulShot parses from sub-body (after 0xD0 sub-opcode stripped).
func ParseRequestAutoSoulShot(data []byte) (*RequestAutoSoulShot, error) {
	r := packet.NewReader(data)

	itemID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading ItemID: %w", err)
	}

	typ, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading Type: %w", err)
	}

	return &RequestAutoSoulShot{
		ItemID: itemID,
		Type:   typ,
	}, nil
}
