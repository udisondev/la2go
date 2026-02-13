package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeDlgAnswer is the C2S opcode 0xC5 — dialog confirmation answer.
const OpcodeDlgAnswer byte = 0xC5

// DlgAnswer represents a player's response to a yes/no dialog.
type DlgAnswer struct {
	MessageID   int32 // SystemMessageId that triggered the dialog
	Answer      int32 // 0 = No, 1 = Yes
	RequesterID int32 // Object ID of the requester (NPC or player)
}

// ParseDlgAnswer parses the packet from raw bytes.
func ParseDlgAnswer(data []byte) (*DlgAnswer, error) {
	r := packet.NewReader(data)

	messageID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading MessageID: %w", err)
	}

	answer, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading Answer: %w", err)
	}

	requesterID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading RequesterID: %w", err)
	}

	return &DlgAnswer{
		MessageID:   messageID,
		Answer:      answer,
		RequesterID: requesterID,
	}, nil
}
