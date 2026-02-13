package serverpackets

import (
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// --- ExSendManorList ---

func TestExSendManorList_Write(t *testing.T) {
	t.Parallel()

	pkt := &ExSendManorList{
		CastleNames: []string{"Gludio", "Dion", "Giran"},
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", data[0])
	}

	subOp := int16(binary.LittleEndian.Uint16(data[1:3]))
	if subOp != 0x1B {
		t.Errorf("sub-opcode = 0x%04X; want 0x001B", subOp)
	}

	count := int32(binary.LittleEndian.Uint32(data[3:7]))
	if count != 3 {
		t.Errorf("count = %d; want 3", count)
	}

	// Проверяем первый castleID (1-based) через Reader для простоты чтения строк
	r := packet.NewReader(data[7:])

	castleID, err := r.ReadInt()
	if err != nil {
		t.Fatalf("reading castleID[0]: %v", err)
	}
	if castleID != 1 {
		t.Errorf("castleID[0] = %d; want 1", castleID)
	}

	name, err := r.ReadString()
	if err != nil {
		t.Fatalf("reading name[0]: %v", err)
	}
	if name != "Gludio" {
		t.Errorf("name[0] = %q; want %q", name, "Gludio")
	}

	castleID, err = r.ReadInt()
	if err != nil {
		t.Fatalf("reading castleID[1]: %v", err)
	}
	if castleID != 2 {
		t.Errorf("castleID[1] = %d; want 2", castleID)
	}

	name, err = r.ReadString()
	if err != nil {
		t.Fatalf("reading name[1]: %v", err)
	}
	if name != "Dion" {
		t.Errorf("name[1] = %q; want %q", name, "Dion")
	}

	castleID, err = r.ReadInt()
	if err != nil {
		t.Fatalf("reading castleID[2]: %v", err)
	}
	if castleID != 3 {
		t.Errorf("castleID[2] = %d; want 3", castleID)
	}

	name, err = r.ReadString()
	if err != nil {
		t.Fatalf("reading name[2]: %v", err)
	}
	if name != "Giran" {
		t.Errorf("name[2] = %q; want %q", name, "Giran")
	}
}

func TestExSendManorList_WriteEmpty(t *testing.T) {
	t.Parallel()

	pkt := &ExSendManorList{}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", data[0])
	}

	count := int32(binary.LittleEndian.Uint32(data[3:7]))
	if count != 0 {
		t.Errorf("count = %d; want 0", count)
	}

	// 1 (opcode) + 2 (sub-opcode) + 4 (count) = 7 bytes
	if len(data) != 7 {
		t.Errorf("len(data) = %d; want 7", len(data))
	}
}

// --- ExEnchantSkillInfo ---

func TestExEnchantSkillInfo_Write_WithBook(t *testing.T) {
	t.Parallel()

	pkt := &ExEnchantSkillInfo{
		SkillID:    56,
		SkillLevel: 101,
		SpCost:     2000,
		ExpCost:    50000,
		Rate:       76,
		HasBookReq: true,
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
	if subOp != 0x18 {
		t.Errorf("sub-opcode = 0x%04X; want 0x0018", subOp)
	}

	skillID, err := r.ReadInt()
	if err != nil {
		t.Fatalf("reading SkillID: %v", err)
	}
	if skillID != 56 {
		t.Errorf("SkillID = %d; want 56", skillID)
	}

	skillLevel, err := r.ReadInt()
	if err != nil {
		t.Fatalf("reading SkillLevel: %v", err)
	}
	if skillLevel != 101 {
		t.Errorf("SkillLevel = %d; want 101", skillLevel)
	}

	spCost, err := r.ReadInt()
	if err != nil {
		t.Fatalf("reading SpCost: %v", err)
	}
	if spCost != 2000 {
		t.Errorf("SpCost = %d; want 2000", spCost)
	}

	expCost, err := r.ReadLong()
	if err != nil {
		t.Fatalf("reading ExpCost: %v", err)
	}
	if expCost != 50000 {
		t.Errorf("ExpCost = %d; want 50000", expCost)
	}

	rate, err := r.ReadInt()
	if err != nil {
		t.Fatalf("reading Rate: %v", err)
	}
	if rate != 76 {
		t.Errorf("Rate = %d; want 76", rate)
	}

	reqCount, err := r.ReadInt()
	if err != nil {
		t.Fatalf("reading reqCount: %v", err)
	}
	if reqCount != 1 {
		t.Errorf("reqCount = %d; want 1", reqCount)
	}

	// Проверяем требование: книга заточки
	reqType, err := r.ReadInt()
	if err != nil {
		t.Fatalf("reading req.Type: %v", err)
	}
	if reqType != 99 {
		t.Errorf("req.Type = %d; want 99", reqType)
	}

	itemID, err := r.ReadInt()
	if err != nil {
		t.Fatalf("reading req.ItemID: %v", err)
	}
	if itemID != 6622 {
		t.Errorf("req.ItemID = %d; want 6622", itemID)
	}

	itemCount, err := r.ReadLong()
	if err != nil {
		t.Fatalf("reading req.Count: %v", err)
	}
	if itemCount != 1 {
		t.Errorf("req.Count = %d; want 1", itemCount)
	}
}

func TestExEnchantSkillInfo_Write_NoBook(t *testing.T) {
	t.Parallel()

	pkt := &ExEnchantSkillInfo{
		SkillID:    100,
		SkillLevel: 102,
		SpCost:     500,
		ExpCost:    10000,
		Rate:       90,
		HasBookReq: false,
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
	if subOp != 0x18 {
		t.Errorf("sub-opcode = 0x%04X; want 0x0018", subOp)
	}

	r.ReadInt() // skillID
	r.ReadInt() // skillLevel
	r.ReadInt() // spCost
	r.ReadLong() // expCost
	r.ReadInt() // rate

	reqCount, err := r.ReadInt()
	if err != nil {
		t.Fatalf("reading reqCount: %v", err)
	}
	if reqCount != 0 {
		t.Errorf("reqCount = %d; want 0", reqCount)
	}

	// Должно быть 0 оставшихся байт (книга не записана)
	if r.Remaining() != 0 {
		t.Errorf("remaining bytes = %d; want 0", r.Remaining())
	}
}

// --- ExEnchantSkillResult ---

func TestExEnchantSkillResult_Write_Success(t *testing.T) {
	t.Parallel()

	pkt := &ExEnchantSkillResult{Result: 1}

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
	if subOp != 0x19 {
		t.Errorf("sub-opcode = 0x%04X; want 0x0019", subOp)
	}

	result, err := r.ReadInt()
	if err != nil {
		t.Fatalf("reading Result: %v", err)
	}
	if result != 1 {
		t.Errorf("Result = %d; want 1", result)
	}
}

func TestExEnchantSkillResult_Write_Failure(t *testing.T) {
	t.Parallel()

	pkt := &ExEnchantSkillResult{Result: 0}

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
	if subOp != 0x19 {
		t.Errorf("sub-opcode = 0x%04X; want 0x0019", subOp)
	}

	result, err := r.ReadInt()
	if err != nil {
		t.Fatalf("reading Result: %v", err)
	}
	if result != 0 {
		t.Errorf("Result = %d; want 0", result)
	}

	// 1 (opcode) + 2 (sub-opcode) + 4 (result) = 7 bytes
	if len(data) != 7 {
		t.Errorf("len(data) = %d; want 7", len(data))
	}
}
