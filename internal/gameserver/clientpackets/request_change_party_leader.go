package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeRequestChangePartyLeader is the extended C2S sub-opcode 0x04.
// Client requests to transfer party leadership to another member.
const SubOpcodeRequestChangePartyLeader int16 = 0x04

// RequestChangePartyLeader contains the new leader's name.
type RequestChangePartyLeader struct {
	Name string
}

// ParseRequestChangePartyLeader parses the C2S extended packet body (after sub-opcode).
func ParseRequestChangePartyLeader(data []byte) (*RequestChangePartyLeader, error) {
	r := packet.NewReader(data)
	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading name: %w", err)
	}
	return &RequestChangePartyLeader{Name: name}, nil
}
