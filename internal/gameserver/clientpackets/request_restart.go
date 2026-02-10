package clientpackets

const (
	// OpcodeRequestRestart is the opcode for RequestRestart packet (C2S 0x46)
	OpcodeRequestRestart = 0x46
)

// RequestRestart represents the RequestRestart packet sent by client.
// Client sends this when user clicks "Restart" to return to character selection.
// Packet has no payload in Lineage 2 Interlude protocol.
//
// Reference: L2J_Mobius RequestRestart.java
type RequestRestart struct {
	// No fields — packet contains only opcode
}

// ParseRequestRestart parses RequestRestart packet from raw bytes.
// Packet structure: [opcode:1]
// No payload to parse.
func ParseRequestRestart(data []byte) (*RequestRestart, error) {
	// RequestRestart has no payload — only opcode
	// Nothing to parse, just return empty struct
	return &RequestRestart{}, nil
}
