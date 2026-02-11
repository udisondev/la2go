package serverpackets

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

func TestSystemMessage_NumberParam(t *testing.T) {
	msg := NewSystemMessage(SysMsgYouEarnedS1Exp).AddNumber(1500)
	data, err := msg.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// opcode
	if data[0] != OpcodeSystemMessage {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSystemMessage)
	}

	// messageID
	messageID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if messageID != SysMsgYouEarnedS1Exp {
		t.Errorf("messageID = %d, want %d", messageID, SysMsgYouEarnedS1Exp)
	}

	// paramCount
	paramCount := int32(binary.LittleEndian.Uint32(data[5:9]))
	if paramCount != 1 {
		t.Errorf("paramCount = %d, want 1", paramCount)
	}

	// param type
	paramType := int32(binary.LittleEndian.Uint32(data[9:13]))
	if paramType != ParamTypeNumber {
		t.Errorf("paramType = %d, want %d", paramType, ParamTypeNumber)
	}

	// param value
	paramValue := int32(binary.LittleEndian.Uint32(data[13:17]))
	if paramValue != 1500 {
		t.Errorf("paramValue = %d, want 1500", paramValue)
	}
}

func TestSystemMessage_StringParam(t *testing.T) {
	msg := NewSystemMessage(SysMsgTargetIsNotFound).AddString("TestPlayer")
	data, err := msg.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// paramCount
	paramCount := int32(binary.LittleEndian.Uint32(data[5:9]))
	if paramCount != 1 {
		t.Errorf("paramCount = %d, want 1", paramCount)
	}

	// param type
	paramType := int32(binary.LittleEndian.Uint32(data[9:13]))
	if paramType != ParamTypeText {
		t.Errorf("paramType = %d, want %d (Text)", paramType, ParamTypeText)
	}

	// param value = UTF-16LE string
	str := readSysMsgUTF16(t, data, 13)
	if str != "TestPlayer" {
		t.Errorf("string param = %q, want %q", str, "TestPlayer")
	}
}

func TestSystemMessage_MultipleParams(t *testing.T) {
	msg := NewSystemMessage(SysMsgYouEarnedS1ExpAndS2SP).
		AddNumber(5000).
		AddNumber(200)

	data, err := msg.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	paramCount := int32(binary.LittleEndian.Uint32(data[5:9]))
	if paramCount != 2 {
		t.Errorf("paramCount = %d, want 2", paramCount)
	}

	// First param: type=1 (Number), value=5000
	p1Type := int32(binary.LittleEndian.Uint32(data[9:13]))
	p1Value := int32(binary.LittleEndian.Uint32(data[13:17]))
	if p1Type != ParamTypeNumber || p1Value != 5000 {
		t.Errorf("param1 type=%d value=%d, want type=%d value=5000", p1Type, p1Value, ParamTypeNumber)
	}

	// Second param: type=1 (Number), value=200
	p2Type := int32(binary.LittleEndian.Uint32(data[17:21]))
	p2Value := int32(binary.LittleEndian.Uint32(data[21:25]))
	if p2Type != ParamTypeNumber || p2Value != 200 {
		t.Errorf("param2 type=%d value=%d, want type=%d value=200", p2Type, p2Value, ParamTypeNumber)
	}
}

func TestSystemMessage_LongParam(t *testing.T) {
	msg := NewSystemMessage(SysMsgYouEarnedS1Exp).AddLong(999999999999)
	data, err := msg.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// param type
	paramType := int32(binary.LittleEndian.Uint32(data[9:13]))
	if paramType != ParamTypeLong {
		t.Errorf("paramType = %d, want %d (Long)", paramType, ParamTypeLong)
	}

	// param value (8 bytes, int64)
	longValue := int64(binary.LittleEndian.Uint64(data[13:21]))
	if longValue != 999999999999 {
		t.Errorf("longValue = %d, want 999999999999", longValue)
	}
}

func TestSystemMessage_SkillParam(t *testing.T) {
	msg := NewSystemMessage(0).AddSkillName(1064, 3)
	data, err := msg.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// param type
	paramType := int32(binary.LittleEndian.Uint32(data[9:13]))
	if paramType != ParamTypeSkill {
		t.Errorf("paramType = %d, want %d (Skill)", paramType, ParamTypeSkill)
	}

	// skill ID
	skillID := int32(binary.LittleEndian.Uint32(data[13:17]))
	if skillID != 1064 {
		t.Errorf("skillID = %d, want 1064", skillID)
	}

	// skill level
	level := int32(binary.LittleEndian.Uint32(data[17:21]))
	if level != 3 {
		t.Errorf("skillLevel = %d, want 3", level)
	}
}

func TestSystemMessage_ItemNameParam(t *testing.T) {
	msg := NewSystemMessage(0).AddItemName(57) // Adena
	data, err := msg.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	paramType := int32(binary.LittleEndian.Uint32(data[9:13]))
	if paramType != ParamTypeItemName {
		t.Errorf("paramType = %d, want %d (ItemName)", paramType, ParamTypeItemName)
	}

	itemID := int32(binary.LittleEndian.Uint32(data[13:17]))
	if itemID != 57 {
		t.Errorf("itemID = %d, want 57", itemID)
	}
}

func TestSystemMessage_NpcNameParam(t *testing.T) {
	msg := NewSystemMessage(0).AddNpcName(20001)
	data, err := msg.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	paramType := int32(binary.LittleEndian.Uint32(data[9:13]))
	if paramType != ParamTypeNpcName {
		t.Errorf("paramType = %d, want %d (NpcName)", paramType, ParamTypeNpcName)
	}

	npcID := int32(binary.LittleEndian.Uint32(data[13:17]))
	if npcID != 20001 {
		t.Errorf("npcID = %d, want 20001", npcID)
	}
}

func TestSystemMessage_PlayerNameParam(t *testing.T) {
	msg := NewSystemMessage(0).AddPlayerName("Warrior")
	data, err := msg.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	paramType := int32(binary.LittleEndian.Uint32(data[9:13]))
	if paramType != ParamTypePlayerName {
		t.Errorf("paramType = %d, want %d (PlayerName)", paramType, ParamTypePlayerName)
	}

	str := readSysMsgUTF16(t, data, 13)
	if str != "Warrior" {
		t.Errorf("playerName = %q, want %q", str, "Warrior")
	}
}

func TestSystemMessage_NoParams(t *testing.T) {
	msg := NewSystemMessage(SysMsgYourLevelHasIncreased)
	data, err := msg.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// Should be opcode(1) + messageID(4) + paramCount(4) = 9 bytes
	if len(data) != 9 {
		t.Errorf("len = %d, want 9", len(data))
	}

	paramCount := int32(binary.LittleEndian.Uint32(data[5:9]))
	if paramCount != 0 {
		t.Errorf("paramCount = %d, want 0", paramCount)
	}
}

// readSysMsgUTF16 reads a UTF-16LE null-terminated string from data at given offset.
func readSysMsgUTF16(t *testing.T, data []byte, offset int) string {
	t.Helper()

	var runes []uint16
	pos := offset
	for {
		if pos+2 > len(data) {
			t.Fatalf("unexpected end of data reading string at offset %d", offset)
		}
		r := binary.LittleEndian.Uint16(data[pos:])
		pos += 2
		if r == 0 {
			break
		}
		runes = append(runes, r)
	}
	return string(utf16.Decode(runes))
}
