package serverpackets

import (
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func TestExAutoSoulShot_Write(t *testing.T) {
	t.Parallel()

	pkt := &ExAutoSoulShot{ItemID: 3947, Type: 1}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", data[0])
	}
	subOp := int16(binary.LittleEndian.Uint16(data[1:3]))
	if subOp != SubOpcodeExAutoSoulShot {
		t.Errorf("sub-opcode = 0x%04X; want 0x%04X", subOp, SubOpcodeExAutoSoulShot)
	}
	itemID := int32(binary.LittleEndian.Uint32(data[3:7]))
	if itemID != 3947 {
		t.Errorf("ItemID = %d; want 3947", itemID)
	}
	typ := int32(binary.LittleEndian.Uint32(data[7:11]))
	if typ != 1 {
		t.Errorf("Type = %d; want 1", typ)
	}
}

func TestExAutoSoulShot_WriteDisable(t *testing.T) {
	t.Parallel()

	pkt := &ExAutoSoulShot{ItemID: 3948, Type: 0}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	typ := int32(binary.LittleEndian.Uint32(data[7:11]))
	if typ != 0 {
		t.Errorf("Type = %d; want 0 (disable)", typ)
	}
}

func TestSendMacroList_WriteWithMacro(t *testing.T) {
	t.Parallel()

	macro := &model.Macro{
		ID:      5,
		Name:    "Heal",
		Desc:    "Quick heal",
		Acronym: "H",
		Icon:    2,
		Commands: []model.MacroCmd{
			{Entry: 0, Type: model.MacroCmdSkill, D1: 1001, D2: 0, Command: "/use 1001"},
		},
	}

	pkt := &SendMacroList{
		Revision: 3,
		Count:    2,
		Macro:    macro,
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if data[0] != OpcodeSendMacroList {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeSendMacroList)
	}

	revision := int32(binary.LittleEndian.Uint32(data[1:5]))
	if revision != 3 {
		t.Errorf("Revision = %d; want 3", revision)
	}

	// byte 5 = unknown (0), byte 6 = count, byte 7 = hasMacro
	if data[6] != 2 {
		t.Errorf("Count = %d; want 2", data[6])
	}
	if data[7] != 1 {
		t.Errorf("hasMacro = %d; want 1", data[7])
	}

	// Macro ID starts at offset 8
	macroID := int32(binary.LittleEndian.Uint32(data[8:12]))
	if macroID != 5 {
		t.Errorf("MacroID = %d; want 5", macroID)
	}

	// Minimum size check: should have at least opcode + revision + unknown + count + hasMacro + macroID
	if len(data) < 12 {
		t.Errorf("data length = %d; want at least 12", len(data))
	}
}

func TestSendMacroList_WriteEmpty(t *testing.T) {
	t.Parallel()

	pkt := &SendMacroList{
		Revision: 1,
		Count:    0,
		Macro:    nil,
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if data[0] != OpcodeSendMacroList {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeSendMacroList)
	}

	// hasMacro byte should be 0
	if data[7] != 0 {
		t.Errorf("hasMacro = %d; want 0", data[7])
	}

	// Total length: opcode(1) + revision(4) + unknown(1) + count(1) + hasMacro(1) = 8
	if len(data) != 8 {
		t.Errorf("data length = %d; want 8 for nil macro", len(data))
	}
}

func TestL2FriendSay_Write(t *testing.T) {
	t.Parallel()

	pkt := &L2FriendSay{
		Sender:   "Alice",
		Receiver: "Bob",
		Message:  "Hello!",
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if data[0] != OpcodeL2FriendSay {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeL2FriendSay)
	}

	// After opcode: int32(0) = 4 bytes at offset 1-4
	unk := int32(binary.LittleEndian.Uint32(data[1:5]))
	if unk != 0 {
		t.Errorf("unknown = %d; want 0", unk)
	}

	// Minimum size: opcode(1) + int32(4) + 3 strings (each at least 2 bytes null-term)
	if len(data) < 11 {
		t.Errorf("data length = %d; too short", len(data))
	}
}
