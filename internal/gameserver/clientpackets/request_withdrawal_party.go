package clientpackets

// OpcodeRequestWithdrawalParty is the client packet opcode for leaving party (C2S 0x2B).
//
// Packet structure (C2S 0x2B):
//   - no payload (opcode only)
//
// Java reference: RequestWithDrawalParty.java (opcode 0x2B).
const OpcodeRequestWithdrawalParty = 0x2B

// RequestWithdrawalParty represents a client request to leave the party.
// Packet has no payload.
type RequestWithdrawalParty struct{}

// ParseRequestWithdrawalParty parses RequestWithdrawalParty packet from raw bytes.
// Opcode already stripped by HandlePacket.
// No payload to parse.
func ParseRequestWithdrawalParty(data []byte) (*RequestWithdrawalParty, error) {
	return &RequestWithdrawalParty{}, nil
}
