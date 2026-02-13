package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestQuestAbort is the opcode for quest abort request (C2S 0x64).
const OpcodeRequestQuestAbort = 0x64

// RequestQuestAbort packet (C2S 0x64) sent when player abandons a quest.
//
// Packet structure:
//   - questID (int32) — quest to abandon
//
// Phase 16: Quest System Framework.
type RequestQuestAbort struct {
	QuestID int32
}

// ParseRequestQuestAbort parses the quest abort packet.
func ParseRequestQuestAbort(data []byte) (*RequestQuestAbort, error) {
	r := packet.NewReader(data)

	questID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading questID: %w", err)
	}

	return &RequestQuestAbort{
		QuestID: questID,
	}, nil
}
