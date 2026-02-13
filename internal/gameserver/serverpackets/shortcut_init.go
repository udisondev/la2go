package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeShortCutInit is the opcode for ShortCutInit packet (S2C 0x45).
	// Reference: L2J_Mobius ServerPackets.java — SHORT_CUT_INIT(0x45)
	OpcodeShortCutInit = 0x45
)

// ShortCutInit packet (S2C 0x45) sends all shortcuts for action bar (F1-F12, pages 0-9).
// Sent after UserInfo during spawn, and after shortcut deletion.
//
// Reference: L2J_Mobius ShortcutInit.java
type ShortCutInit struct {
	Shortcuts []*model.Shortcut
}

// NewShortCutInit creates a ShortCutInit packet with the given shortcuts.
// Pass nil or empty slice for an empty action bar.
func NewShortCutInit(shortcuts []*model.Shortcut) ShortCutInit {
	return ShortCutInit{Shortcuts: shortcuts}
}

// Write serializes ShortCutInit packet to binary format.
func (p *ShortCutInit) Write() ([]byte, error) {
	// Estimate: 5 (header+count) + ~30 bytes per shortcut
	w := packet.NewWriter(5 + len(p.Shortcuts)*30)

	_ = w.WriteByte(OpcodeShortCutInit)
	w.WriteInt(int32(len(p.Shortcuts)))

	for _, sc := range p.Shortcuts {
		w.WriteInt(int32(sc.Type))
		w.WriteInt(int32(sc.Slot) + int32(sc.Page)*model.MaxShortcutsPerBar)
		writeShortcutBody(w, sc)
	}

	return w.Bytes(), nil
}
