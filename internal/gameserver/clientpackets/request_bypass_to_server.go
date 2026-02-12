package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestBypassToServer is the client packet opcode for RequestBypassToServer.
// Client sends this when player clicks a bypass link in an NPC dialog.
//
// Phase 8.2: NPC Dialogues.
// Java reference: RequestBypassToServer.java
const OpcodeRequestBypassToServer = 0x21

// RequestBypassToServer represents the client's bypass request.
// Sent when player clicks a link like:
//
//	<a action="bypass -h npc_%objectId%_Shop">Shop</a>
//
// Packet structure:
//   - bypass (string): bypass command (e.g., "npc_12345_Shop", "_bbshome")
//
// Phase 8.2: NPC Dialogues.
type RequestBypassToServer struct {
	Bypass string
}

// ParseRequestBypassToServer parses a RequestBypassToServer packet from raw bytes.
func ParseRequestBypassToServer(data []byte) (*RequestBypassToServer, error) {
	r := packet.NewReader(data)

	bypass, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading bypass: %w", err)
	}

	return &RequestBypassToServer{
		Bypass: bypass,
	}, nil
}
