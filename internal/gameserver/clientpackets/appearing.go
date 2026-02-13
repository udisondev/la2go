package clientpackets

// OpcodeAppearing is the opcode for Appearing (C2S 0x30).
// Java reference: ClientPackets.APPEARING(0x30).
// Sent by client after teleport or zone transition.
const OpcodeAppearing = 0x30

// Appearing has no body fields.
type Appearing struct{}

// ParseAppearing parses Appearing packet (no fields).
func ParseAppearing(_ []byte) (*Appearing, error) {
	return &Appearing{}, nil
}
