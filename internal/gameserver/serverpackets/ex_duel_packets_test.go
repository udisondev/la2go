package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestExDuelAskStart_Write(t *testing.T) {
	pkt := ExDuelAskStart{
		RequestorName: "Alice",
		PartyDuel:     false,
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	r := packet.NewReader(data)

	opcode, err := r.ReadByte()
	if err != nil {
		t.Fatalf("reading opcode: %v", err)
	}
	if opcode != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", opcode)
	}

	subOp, err := r.ReadShort()
	if err != nil {
		t.Fatalf("reading sub-opcode: %v", err)
	}
	if subOp != SubOpcodeExDuelAskStart {
		t.Errorf("sub-opcode = 0x%04X; want 0x%04X", subOp, SubOpcodeExDuelAskStart)
	}

	name, err := r.ReadString()
	if err != nil {
		t.Fatalf("reading name: %v", err)
	}
	if name != "Alice" {
		t.Errorf("name = %q; want %q", name, "Alice")
	}

	partyDuel, err := r.ReadInt()
	if err != nil {
		t.Fatalf("reading partyDuel: %v", err)
	}
	if partyDuel != 0 {
		t.Errorf("partyDuel = %d; want 0", partyDuel)
	}
}

func TestExDuelAskStart_Write_Party(t *testing.T) {
	pkt := ExDuelAskStart{
		RequestorName: "Bob",
		PartyDuel:     true,
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	r := packet.NewReader(data)
	r.ReadByte()  // opcode
	r.ReadShort() // sub-opcode
	r.ReadString() // name

	partyDuel, err := r.ReadInt()
	if err != nil {
		t.Fatalf("reading partyDuel: %v", err)
	}
	if partyDuel != 1 {
		t.Errorf("partyDuel = %d; want 1", partyDuel)
	}
}

func TestExDuelReady_Write(t *testing.T) {
	for _, partyDuel := range []bool{false, true} {
		pkt := ExDuelReady{PartyDuel: partyDuel}
		data, err := pkt.Write()
		if err != nil {
			t.Fatalf("Write(party=%v): %v", partyDuel, err)
		}

		r := packet.NewReader(data)
		opcode, _ := r.ReadByte()
		if opcode != 0xFE {
			t.Errorf("opcode = 0x%02X; want 0xFE", opcode)
		}
		subOp, _ := r.ReadShort()
		if subOp != SubOpcodeExDuelReady {
			t.Errorf("sub-opcode = 0x%04X; want 0x%04X", subOp, SubOpcodeExDuelReady)
		}
		val, _ := r.ReadInt()
		expected := int32(0)
		if partyDuel {
			expected = 1
		}
		if val != expected {
			t.Errorf("partyDuel = %d; want %d", val, expected)
		}
	}
}

func TestExDuelStart_Write(t *testing.T) {
	pkt := ExDuelStart{PartyDuel: true}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	r := packet.NewReader(data)
	opcode, _ := r.ReadByte()
	if opcode != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", opcode)
	}
	subOp, _ := r.ReadShort()
	if subOp != SubOpcodeExDuelStart {
		t.Errorf("sub-opcode = 0x%04X; want 0x%04X", subOp, SubOpcodeExDuelStart)
	}
	val, _ := r.ReadInt()
	if val != 1 {
		t.Errorf("partyDuel = %d; want 1", val)
	}
}

func TestExDuelEnd_Write(t *testing.T) {
	pkt := ExDuelEnd{PartyDuel: false}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	r := packet.NewReader(data)
	opcode, _ := r.ReadByte()
	if opcode != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", opcode)
	}
	subOp, _ := r.ReadShort()
	if subOp != SubOpcodeExDuelEnd {
		t.Errorf("sub-opcode = 0x%04X; want 0x%04X", subOp, SubOpcodeExDuelEnd)
	}
	val, _ := r.ReadInt()
	if val != 0 {
		t.Errorf("partyDuel = %d; want 0", val)
	}
}

func TestExDuelUpdateUserInfo_Write(t *testing.T) {
	pkt := ExDuelUpdateUserInfo{
		ObjectID:  12345,
		Name:      "Alice",
		CurrentHP: 800,
		MaxHP:     1000,
		CurrentMP: 400,
		MaxMP:     500,
		CurrentCP: 600,
		MaxCP:     700,
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	r := packet.NewReader(data)
	opcode, _ := r.ReadByte()
	if opcode != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", opcode)
	}
	subOp, _ := r.ReadShort()
	if subOp != SubOpcodeExDuelUpdateUserInfo {
		t.Errorf("sub-opcode = 0x%04X; want 0x%04X", subOp, SubOpcodeExDuelUpdateUserInfo)
	}
	objID, _ := r.ReadInt()
	if uint32(objID) != 12345 {
		t.Errorf("objectID = %d; want 12345", objID)
	}
	name, _ := r.ReadString()
	if name != "Alice" {
		t.Errorf("name = %q; want %q", name, "Alice")
	}
	hp, _ := r.ReadInt()
	if hp != 800 {
		t.Errorf("currentHP = %d; want 800", hp)
	}
	maxHP, _ := r.ReadInt()
	if maxHP != 1000 {
		t.Errorf("maxHP = %d; want 1000", maxHP)
	}
	mp, _ := r.ReadInt()
	if mp != 400 {
		t.Errorf("currentMP = %d; want 400", mp)
	}
	maxMP, _ := r.ReadInt()
	if maxMP != 500 {
		t.Errorf("maxMP = %d; want 500", maxMP)
	}
	cp, _ := r.ReadInt()
	if cp != 600 {
		t.Errorf("currentCP = %d; want 600", cp)
	}
	maxCP, _ := r.ReadInt()
	if maxCP != 700 {
		t.Errorf("maxCP = %d; want 700", maxCP)
	}
}
