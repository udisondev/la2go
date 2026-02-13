package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestJoinParty is the client packet opcode for party invite (C2S 0x29).
//
// Packet structure (C2S 0x29):
//   - objectID         int32  target objectID to invite
//   - itemDistribution int32  loot distribution rule (0-4)
//
// Java reference: RequestJoinParty.java (opcode 0x29).
const OpcodeRequestJoinParty = 0x29

// RequestJoinParty represents a client request to invite a player to party.
type RequestJoinParty struct {
	ObjectID         int32 // target player's objectID
	ItemDistribution int32 // loot distribution rule (LootRuleFinders..LootRuleOrderSpoil)
}

// ParseRequestJoinParty parses RequestJoinParty packet from raw bytes.
// Opcode already stripped by HandlePacket.
func ParseRequestJoinParty(data []byte) (*RequestJoinParty, error) {
	r := packet.NewReader(data)

	objectID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading objectID: %w", err)
	}

	itemDistribution, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading itemDistribution: %w", err)
	}

	return &RequestJoinParty{
		ObjectID:         objectID,
		ItemDistribution: itemDistribution,
	}, nil
}
