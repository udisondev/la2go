package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeRequestShortCutDel is the C2S opcode for deleting a shortcut (0x35).
const OpcodeRequestShortCutDel = 0x35

// RequestShortCutDel represents a client request to delete a shortcut.
//
// Packet structure:
//
//	slotPage (int32) -> slot = slotPage % 12, page = slotPage / 12
//
// Reference: L2J_Mobius RequestShortcutDel.java
type RequestShortCutDel struct {
	Slot int8
	Page int8
}

// ParseRequestShortCutDel parses a RequestShortCutDel packet from raw data.
func ParseRequestShortCutDel(data []byte) (*RequestShortCutDel, error) {
	r := packet.NewReader(data)

	slotPage, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading slot/page: %w", err)
	}

	return &RequestShortCutDel{
		Slot: int8(slotPage % model.MaxShortcutsPerBar),
		Page: int8(slotPage / model.MaxShortcutsPerBar),
	}, nil
}
