package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestExPacket is the extended client opcode (C2S 0xD0).
// Extended packets carry a 2-byte sub-opcode after the main opcode.
const OpcodeRequestExPacket = 0xD0

// SubOpcodeRequestDuelStart is the sub-opcode for duel request (C2S 0xD0:0x27).
const SubOpcodeRequestDuelStart int16 = 0x27

// RequestDuelStart represents a client request to challenge another player to a duel.
//
// Packet structure (C2S 0xD0:0x1B):
//   - name      string  target player name
//   - partyDuel int32   0 = 1v1, 1 = party duel
//
// Java reference: RequestDuelStart.java.
type RequestDuelStart struct {
	Name      string
	PartyDuel bool
}

// ParseRequestDuelStart parses RequestDuelStart from raw bytes.
// Sub-opcode already stripped by the extended opcode dispatcher.
func ParseRequestDuelStart(data []byte) (*RequestDuelStart, error) {
	r := packet.NewReader(data)

	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading name: %w", err)
	}

	partyDuelRaw, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading partyDuel: %w", err)
	}

	return &RequestDuelStart{
		Name:      name,
		PartyDuel: partyDuelRaw != 0,
	}, nil
}
