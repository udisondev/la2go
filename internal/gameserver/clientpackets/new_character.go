package clientpackets

// OpcodeNewCharacter is the opcode for NewCharacter (C2S 0x0E).
// Same opcode as ProtocolVersion, but handled in AUTHENTICATED state.
// Java reference: ClientPackets.NEW_CHARACTER(0x0E).
const OpcodeNewCharacter = 0x0E

// NewCharacter is an empty C2S packet requesting character creation templates.
// The client sends this when the user clicks "Create" on the character selection screen.
// Server responds with CharTemplates S2C packet (0x17).
type NewCharacter struct{}

// ParseNewCharacter parses NewCharacter packet (empty — no fields to read).
func ParseNewCharacter(_ []byte) (*NewCharacter, error) {
	return &NewCharacter{}, nil
}
