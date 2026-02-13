package clientpackets

// SubOpcodeRequestDuelSurrender is the sub-opcode for duel surrender (C2S 0xD0:0x30).
const SubOpcodeRequestDuelSurrender int16 = 0x30

// RequestDuelSurrender represents a client's request to surrender in a duel.
//
// Packet structure (C2S 0xD0:0x1D): no fields.
//
// Java reference: RequestDuelSurrender.java.
type RequestDuelSurrender struct{}

// ParseRequestDuelSurrender parses RequestDuelSurrender from raw bytes.
// No fields to parse — surrender has no parameters.
func ParseRequestDuelSurrender(_ []byte) (*RequestDuelSurrender, error) {
	return &RequestDuelSurrender{}, nil
}
