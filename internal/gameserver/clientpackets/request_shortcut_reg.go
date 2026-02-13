package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeRequestShortCutReg is the C2S opcode for registering a shortcut (0x33).
const OpcodeRequestShortCutReg = 0x33

// RequestShortCutReg represents a client request to register (add/replace) a shortcut.
//
// Packet structure:
//
//	typeID   (int32) -> ShortcutType
//	slotPage (int32) -> slot = slotPage % 12, page = slotPage / 12
//	id       (int32) -> item objectID / skill ID / action ID / macro ID / recipe ID
//	level    (int32) -> skill level (read but unused by Java; server uses actual level)
//
// Reference: L2J_Mobius RequestShortcutReg.java
type RequestShortCutReg struct {
	Type  model.ShortcutType
	Slot  int8
	Page  int8
	ID    int32
	Level int32
}

// ParseRequestShortCutReg parses a RequestShortCutReg packet from raw data.
func ParseRequestShortCutReg(data []byte) (*RequestShortCutReg, error) {
	r := packet.NewReader(data)

	typeID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading type: %w", err)
	}

	slotPage, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading slot/page: %w", err)
	}

	id, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading id: %w", err)
	}

	// Level is always read from the packet but Java ignores the value for skills —
	// it overrides with the actual player skill level. We read it anyway.
	level, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading level: %w", err)
	}

	// Clamp typeID to valid range (Java: values()[(typeId < 1) || (typeId > 6) ? 0 : typeId])
	scType := model.ShortcutType(typeID)
	if scType < model.ShortcutTypeItem || scType > model.ShortcutTypeRecipe {
		scType = model.ShortcutTypeNone
	}

	return &RequestShortCutReg{
		Type:  scType,
		Slot:  int8(slotPage % model.MaxShortcutsPerBar),
		Page:  int8(slotPage / model.MaxShortcutsPerBar),
		ID:    id,
		Level: level,
	}, nil
}
