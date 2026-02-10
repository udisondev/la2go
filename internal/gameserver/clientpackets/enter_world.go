package clientpackets

const (
	// OpcodeEnterWorld is the opcode for EnterWorld packet (C2S 0x03)
	OpcodeEnterWorld = 0x03
)

// EnterWorld represents the EnterWorld packet sent by client.
// Client sends this after CharacterSelect to spawn in the world.
// Packet has no payload in Lineage 2 Interlude protocol.
type EnterWorld struct {
	// No fields — packet contains only opcode
}

// ParseEnterWorld parses EnterWorld packet from raw bytes.
// Packet structure: [opcode:1]
func ParseEnterWorld(data []byte) (*EnterWorld, error) {
	// EnterWorld has no payload — only opcode
	// Nothing to parse, just return empty struct
	return &EnterWorld{}, nil
}
