package clientpackets

const (
	// OpcodeLogout is the opcode for Logout packet (C2S 0x09)
	OpcodeLogout = 0x09
)

// Logout represents the Logout packet sent by client.
// Client sends this when user clicks Exit button.
// Packet has no payload in Lineage 2 Interlude protocol.
//
// Reference: L2J_Mobius Logout.java
type Logout struct {
	// No fields — packet contains only opcode
}

// ParseLogout parses Logout packet from raw bytes.
// Packet structure: [opcode:1]
// No payload to parse.
func ParseLogout(data []byte) (*Logout, error) {
	// Logout has no payload — only opcode
	// Nothing to parse, just return empty struct
	return &Logout{}, nil
}
