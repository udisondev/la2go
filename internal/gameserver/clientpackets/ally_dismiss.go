package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeAllyDismiss is the C2S opcode 0x85.
// Client sends this when the alliance leader dismisses a clan from the alliance.
const OpcodeAllyDismiss byte = 0x85

// AllyDismiss represents a request to dismiss a clan from the alliance.
//
// Packet structure:
//   - ClanName (string): UTF-16LE null-terminated name of the clan to dismiss
type AllyDismiss struct {
	ClanName string
}

// ParseAllyDismiss parses the C2S AllyDismiss packet.
func ParseAllyDismiss(data []byte) (*AllyDismiss, error) {
	r := packet.NewReader(data)

	clanName, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading clanName: %w", err)
	}

	return &AllyDismiss{ClanName: clanName}, nil
}
