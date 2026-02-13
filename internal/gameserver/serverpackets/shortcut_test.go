package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

func TestShortCutInitEmpty(t *testing.T) {
	pkt := NewShortCutInit(nil)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(data)

	opcode, err := r.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte: %v", err)
	}
	if opcode != OpcodeShortCutInit {
		t.Errorf("opcode = 0x%02X; want 0x%02X", opcode, OpcodeShortCutInit)
	}

	count := readInt32(t, r)
	if count != 0 {
		t.Errorf("shortcut count = %d; want 0", count)
	}
}

func TestShortCutInitOpcode(t *testing.T) {
	// Verify opcode matches Java reference: SHORT_CUT_INIT(0x45)
	if OpcodeShortCutInit != 0x45 {
		t.Errorf("OpcodeShortCutInit = 0x%02X; want 0x45", OpcodeShortCutInit)
	}
}

func TestShortCutRegisterOpcode(t *testing.T) {
	// Verify opcode matches Java reference: SHORT_CUT_REGISTER(0x44)
	if OpcodeShortCutRegister != 0x44 {
		t.Errorf("OpcodeShortCutRegister = 0x%02X; want 0x44", OpcodeShortCutRegister)
	}
}

func TestShortCutInitWithSkill(t *testing.T) {
	sc := &model.Shortcut{
		Slot:  3,
		Page:  1,
		Type:  model.ShortcutTypeSkill,
		ID:    1001,
		Level: 5,
	}

	pkt := NewShortCutInit([]*model.Shortcut{sc})
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(data)

	opcode, err := r.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte: %v", err)
	}
	if opcode != OpcodeShortCutInit {
		t.Errorf("opcode = 0x%02X; want 0x%02X", opcode, OpcodeShortCutInit)
	}

	count := readInt32(t, r)
	if count != 1 {
		t.Errorf("count = %d; want 1", count)
	}

	// type (2 = skill)
	scType := readInt32(t, r)
	if scType != int32(model.ShortcutTypeSkill) {
		t.Errorf("type = %d; want %d", scType, model.ShortcutTypeSkill)
	}

	// absolute slot = 3 + 1*12 = 15
	absSlot := readInt32(t, r)
	if absSlot != 15 {
		t.Errorf("absSlot = %d; want 15", absSlot)
	}

	// skillID
	skillID := readInt32(t, r)
	if skillID != 1001 {
		t.Errorf("skillID = %d; want 1001", skillID)
	}

	// skillLevel
	skillLevel := readInt32(t, r)
	if skillLevel != 5 {
		t.Errorf("skillLevel = %d; want 5", skillLevel)
	}

	// C5 byte
	c5, err := r.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte (C5): %v", err)
	}
	if c5 != 0 {
		t.Errorf("C5 byte = %d; want 0", c5)
	}

	// unknown (1)
	unk := readInt32(t, r)
	if unk != 1 {
		t.Errorf("unknown = %d; want 1", unk)
	}
}

func TestShortCutInitWithItem(t *testing.T) {
	sc := &model.Shortcut{
		Slot: 0,
		Page: 0,
		Type: model.ShortcutTypeItem,
		ID:   500,
	}

	pkt := NewShortCutInit([]*model.Shortcut{sc})
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(data)

	// opcode
	if _, err := r.ReadByte(); err != nil {
		t.Fatalf("ReadByte: %v", err)
	}

	count := readInt32(t, r)
	if count != 1 {
		t.Errorf("count = %d; want 1", count)
	}

	scType := readInt32(t, r)
	if scType != int32(model.ShortcutTypeItem) {
		t.Errorf("type = %d; want %d", scType, model.ShortcutTypeItem)
	}

	absSlot := readInt32(t, r)
	if absSlot != 0 {
		t.Errorf("absSlot = %d; want 0", absSlot)
	}

	objectID := readInt32(t, r)
	if objectID != 500 {
		t.Errorf("objectID = %d; want 500", objectID)
	}

	equipped := readInt32(t, r)
	if equipped != 1 {
		t.Errorf("equipped = %d; want 1", equipped)
	}

	bodyPart := readInt32(t, r)
	if bodyPart != -1 {
		t.Errorf("bodyPart = %d; want -1", bodyPart)
	}

	enchantLevel := readInt32(t, r)
	if enchantLevel != 0 {
		t.Errorf("enchantLevel = %d; want 0", enchantLevel)
	}

	augID := readInt32(t, r)
	if augID != 0 {
		t.Errorf("augmentationID = %d; want 0", augID)
	}

	mana := readInt16(t, r)
	if mana != 0 {
		t.Errorf("mana = %d; want 0", mana)
	}

	unkShort := readInt16(t, r)
	if unkShort != 0 {
		t.Errorf("unknown short = %d; want 0", unkShort)
	}
}

func TestShortCutInitWithAction(t *testing.T) {
	sc := &model.Shortcut{
		Slot: 5,
		Page: 3,
		Type: model.ShortcutTypeAction,
		ID:   42,
	}

	pkt := NewShortCutInit([]*model.Shortcut{sc})
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(data)

	if _, err := r.ReadByte(); err != nil {
		t.Fatalf("ReadByte: %v", err)
	}

	count := readInt32(t, r)
	if count != 1 {
		t.Errorf("count = %d; want 1", count)
	}

	scType := readInt32(t, r)
	if scType != int32(model.ShortcutTypeAction) {
		t.Errorf("type = %d; want %d", scType, model.ShortcutTypeAction)
	}

	// absSlot = 5 + 3*12 = 41
	absSlot := readInt32(t, r)
	if absSlot != 41 {
		t.Errorf("absSlot = %d; want 41", absSlot)
	}

	actionID := readInt32(t, r)
	if actionID != 42 {
		t.Errorf("actionID = %d; want 42", actionID)
	}

	unk := readInt32(t, r)
	if unk != 1 {
		t.Errorf("unknown = %d; want 1", unk)
	}
}

func TestShortCutRegisterSkill(t *testing.T) {
	sc := &model.Shortcut{
		Slot:  7,
		Page:  2,
		Type:  model.ShortcutTypeSkill,
		ID:    2000,
		Level: 10,
	}

	pkt := NewShortCutRegister(sc)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(data)

	opcode, err := r.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte: %v", err)
	}
	if opcode != OpcodeShortCutRegister {
		t.Errorf("opcode = 0x%02X; want 0x%02X", opcode, OpcodeShortCutRegister)
	}

	scType := readInt32(t, r)
	if scType != int32(model.ShortcutTypeSkill) {
		t.Errorf("type = %d; want %d", scType, model.ShortcutTypeSkill)
	}

	// absSlot = 7 + 2*12 = 31
	absSlot := readInt32(t, r)
	if absSlot != 31 {
		t.Errorf("absSlot = %d; want 31", absSlot)
	}

	skillID := readInt32(t, r)
	if skillID != 2000 {
		t.Errorf("skillID = %d; want 2000", skillID)
	}

	skillLevel := readInt32(t, r)
	if skillLevel != 10 {
		t.Errorf("skillLevel = %d; want 10", skillLevel)
	}

	c5, err := r.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte (C5): %v", err)
	}
	if c5 != 0 {
		t.Errorf("C5 byte = %d; want 0", c5)
	}

	// trailing int = 1
	trailing := readInt32(t, r)
	if trailing != 1 {
		t.Errorf("trailing = %d; want 1", trailing)
	}
}

func TestShortCutRegisterItem(t *testing.T) {
	sc := &model.Shortcut{
		Slot: 0,
		Page: 0,
		Type: model.ShortcutTypeItem,
		ID:   700,
	}

	pkt := NewShortCutRegister(sc)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(data)

	opcode, err := r.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte: %v", err)
	}
	if opcode != OpcodeShortCutRegister {
		t.Errorf("opcode = 0x%02X; want 0x%02X", opcode, OpcodeShortCutRegister)
	}

	scType := readInt32(t, r)
	if scType != int32(model.ShortcutTypeItem) {
		t.Errorf("type = %d; want %d", scType, model.ShortcutTypeItem)
	}

	absSlot := readInt32(t, r)
	if absSlot != 0 {
		t.Errorf("absSlot = %d; want 0", absSlot)
	}

	objectID := readInt32(t, r)
	if objectID != 700 {
		t.Errorf("objectID = %d; want 700", objectID)
	}

	// trailing int = 1
	trailing := readInt32(t, r)
	if trailing != 1 {
		t.Errorf("trailing = %d; want 1", trailing)
	}
}

func TestShortCutInitMultipleShortcuts(t *testing.T) {
	shortcuts := []*model.Shortcut{
		{Slot: 0, Page: 0, Type: model.ShortcutTypeAction, ID: 1},
		{Slot: 1, Page: 0, Type: model.ShortcutTypeSkill, ID: 100, Level: 3},
		{Slot: 2, Page: 1, Type: model.ShortcutTypeItem, ID: 200},
	}

	pkt := NewShortCutInit(shortcuts)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(data)

	if _, err := r.ReadByte(); err != nil {
		t.Fatalf("ReadByte: %v", err)
	}

	count := readInt32(t, r)
	if count != 3 {
		t.Errorf("count = %d; want 3", count)
	}
}

func TestShortCutRegisterRecipe(t *testing.T) {
	sc := &model.Shortcut{
		Slot: 10,
		Page: 5,
		Type: model.ShortcutTypeRecipe,
		ID:   300,
	}

	pkt := NewShortCutRegister(sc)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(data)

	opcode, err := r.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte: %v", err)
	}
	if opcode != OpcodeShortCutRegister {
		t.Errorf("opcode = 0x%02X; want 0x%02X", opcode, OpcodeShortCutRegister)
	}

	scType := readInt32(t, r)
	if scType != int32(model.ShortcutTypeRecipe) {
		t.Errorf("type = %d; want %d", scType, model.ShortcutTypeRecipe)
	}

	// absSlot = 10 + 5*12 = 70
	absSlot := readInt32(t, r)
	if absSlot != 70 {
		t.Errorf("absSlot = %d; want 70", absSlot)
	}

	recipeID := readInt32(t, r)
	if recipeID != 300 {
		t.Errorf("recipeID = %d; want 300", recipeID)
	}

	// trailing int = 1
	trailing := readInt32(t, r)
	if trailing != 1 {
		t.Errorf("trailing = %d; want 1", trailing)
	}
}
