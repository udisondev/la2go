package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeShortCutRegister is the opcode for ShortCutRegister packet (S2C 0x44).
	// Reference: L2J_Mobius ServerPackets.java — SHORT_CUT_REGISTER(0x44)
	OpcodeShortCutRegister = 0x44
)

// ShortCutRegister packet (S2C 0x44) confirms registration of a single shortcut.
// Sent in response to RequestShortCutReg (C2S 0x33).
//
// Reference: L2J_Mobius ShortcutRegister.java
type ShortCutRegister struct {
	Shortcut *model.Shortcut
}

// NewShortCutRegister creates a ShortCutRegister packet for the given shortcut.
func NewShortCutRegister(sc *model.Shortcut) *ShortCutRegister {
	return &ShortCutRegister{Shortcut: sc}
}

// Write serializes ShortCutRegister packet to binary format.
func (p *ShortCutRegister) Write() ([]byte, error) {
	w := packet.NewWriter(32)

	_ = w.WriteByte(OpcodeShortCutRegister)

	sc := p.Shortcut
	w.WriteInt(int32(sc.Type))
	w.WriteInt(int32(sc.Slot) + int32(sc.Page)*model.MaxShortcutsPerBar)

	// Write type-specific body
	switch sc.Type {
	case model.ShortcutTypeItem:
		w.WriteInt(sc.ID)
	case model.ShortcutTypeSkill:
		w.WriteInt(sc.ID)
		w.WriteInt(sc.Level)
		_ = w.WriteByte(0) // C5
	case model.ShortcutTypeAction, model.ShortcutTypeMacro, model.ShortcutTypeRecipe:
		w.WriteInt(sc.ID)
	}

	// Trailing int (always 1) — present in Java for all types.
	w.WriteInt(1)

	return w.Bytes(), nil
}

// writeShortcutBody writes the type-specific body of a shortcut for ShortCutInit.
// ShortCutInit has different (larger) body format than ShortCutRegister.
//
// Reference: L2J_Mobius ShortcutInit.java (writeImpl)
func writeShortcutBody(w *packet.Writer, sc *model.Shortcut) {
	switch sc.Type {
	case model.ShortcutTypeItem:
		w.WriteInt(sc.ID)    // objectID
		w.WriteInt(1)        // equipped (placeholder)
		w.WriteInt(-1)       // bodyPart
		w.WriteInt(0)        // enchantLevel
		w.WriteInt(0)        // augmentationID
		w.WriteShort(0)      // mana
		w.WriteShort(0)      // unknown
	case model.ShortcutTypeSkill:
		w.WriteInt(sc.ID)    // skillID
		w.WriteInt(sc.Level) // skillLevel
		_ = w.WriteByte(0)   // C5
		w.WriteInt(1)        // unknown
	case model.ShortcutTypeAction, model.ShortcutTypeMacro, model.ShortcutTypeRecipe:
		w.WriteInt(sc.ID)
		w.WriteInt(1) // unknown
	}
}
