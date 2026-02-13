package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestOustPartyMember is the client packet opcode for kicking a party member (C2S 0x2C).
//
// Packet structure (C2S 0x2C):
//   - playerName string  name of the player to kick (UTF-16LE null-terminated)
//
// Java reference: RequestOustPartyMember.java (opcode 0x2C).
const OpcodeRequestOustPartyMember = 0x2C

// RequestOustPartyMember represents a client request to kick a player from party.
// Only the party leader can send this packet.
type RequestOustPartyMember struct {
	PlayerName string // name of the player to kick
}

// ParseRequestOustPartyMember parses RequestOustPartyMember packet from raw bytes.
// Opcode already stripped by HandlePacket.
func ParseRequestOustPartyMember(data []byte) (*RequestOustPartyMember, error) {
	r := packet.NewReader(data)

	playerName, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading playerName: %w", err)
	}

	return &RequestOustPartyMember{
		PlayerName: playerName,
	}, nil
}
