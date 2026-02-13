package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeRequestDuelAnswerStart is the sub-opcode for duel answer (C2S 0xD0:0x28).
const SubOpcodeRequestDuelAnswerStart int16 = 0x28

// RequestDuelAnswerStart represents a client's response to a duel request.
//
// Packet structure (C2S 0xD0:0x1C):
//   - partyDuel int32  0 = 1v1, 1 = party duel
//   - response  int32  0 = decline, 1 = accept
//
// Java reference: RequestDuelAnswerStart.java.
type RequestDuelAnswerStart struct {
	PartyDuel bool
	Accepted  bool
}

// ParseRequestDuelAnswerStart parses RequestDuelAnswerStart from raw bytes.
// Sub-opcode already stripped by the extended opcode dispatcher.
func ParseRequestDuelAnswerStart(data []byte) (*RequestDuelAnswerStart, error) {
	r := packet.NewReader(data)

	partyDuelRaw, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading partyDuel: %w", err)
	}

	response, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return &RequestDuelAnswerStart{
		PartyDuel: partyDuelRaw != 0,
		Accepted:  response == 1,
	}, nil
}
