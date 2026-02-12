package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const (
	// OpcodeShortCutInit is the opcode for ShortCutInit packet (S2C 0x44)
	OpcodeShortCutInit = 0x44
)

// ShortCutInit packet (S2C 0x44) sends shortcuts for action bar (F1-F12, etc.).
// Sent after UserInfo during spawn.
type ShortCutInit struct {
	// Shortcuts []Shortcut // TODO Phase 4.8: implement shortcut system
}

// NewShortCutInit creates empty ShortCutInit packet.
// TODO Phase 4.8: Load shortcuts from database.
func NewShortCutInit() ShortCutInit {
	return ShortCutInit{}
}

// Write serializes ShortCutInit packet to binary format.
func (p *ShortCutInit) Write() ([]byte, error) {
	// Empty shortcuts for now
	w := packet.NewWriter(16)

	w.WriteByte(OpcodeShortCutInit)
	w.WriteInt(0) // Shortcut count = 0

	return w.Bytes(), nil
}
