package serverpackets

import (
	"encoding/binary"
	"testing"
)

// funcPairOffset возвращает байтовое смещение для N-й пары (activeFlag, level).
// Структура пакета: opcode(1) + hallID(4) + 9 пар по 8 байт.
func funcPairOffset(index int) int {
	return 1 + 4 + index*8
}

func TestAgitDecoInfo_Write_AllFunctions(t *testing.T) {
	t.Parallel()

	p := &AgitDecoInfo{
		HallID:       21,
		HPLevel:      3,
		MPLevel:      2,
		ExpLevel:     5,
		SPLevel:      1,
		TeleLevel:    4,
		CurtainLevel: 2,
		FrontLevel:   3,
		ItemLevel:    1,
		SupportLevel: 8,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Opcode 0xF7 (regular packet, not extended).
	if data[0] != OpcodeAgitDecoInfo {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeAgitDecoInfo)
	}
	if data[0] != 0xF7 {
		t.Errorf("opcode = 0x%02X, want 0xF7", data[0])
	}

	// HallID.
	hallID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if hallID != 21 {
		t.Errorf("hallID = %d, want 21", hallID)
	}

	// Все 9 функций должны иметь activeFlag=1 и соответствующий level.
	wantLevels := []struct {
		name  string
		level int32
	}{
		{"HP", 3},
		{"MP", 2},
		{"Exp", 5},
		{"SP", 1},
		{"Tele", 4},
		{"Curtain", 2},
		{"Front", 3},
		{"Item", 1},
		{"Support", 8},
	}

	for i, tt := range wantLevels {
		off := funcPairOffset(i)

		active := int32(binary.LittleEndian.Uint32(data[off : off+4]))
		if active != 1 {
			t.Errorf("%s active = %d, want 1", tt.name, active)
		}

		level := int32(binary.LittleEndian.Uint32(data[off+4 : off+8]))
		if level != tt.level {
			t.Errorf("%s level = %d, want %d", tt.name, level, tt.level)
		}
	}

	// Общая длина: 1 + 4 + 9*8 = 77 байт.
	wantLen := 1 + 4 + 9*8
	if len(data) != wantLen {
		t.Errorf("len(data) = %d, want %d", len(data), wantLen)
	}
}

func TestAgitDecoInfo_Write_NoFunctions(t *testing.T) {
	t.Parallel()

	p := &AgitDecoInfo{
		HallID: 5,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if data[0] != 0xF7 {
		t.Errorf("opcode = 0x%02X, want 0xF7", data[0])
	}

	hallID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if hallID != 5 {
		t.Errorf("hallID = %d, want 5", hallID)
	}

	// Все 9 функций: activeFlag=0, level=0.
	funcNames := []string{"HP", "MP", "Exp", "SP", "Tele", "Curtain", "Front", "Item", "Support"}
	for i, name := range funcNames {
		off := funcPairOffset(i)

		active := int32(binary.LittleEndian.Uint32(data[off : off+4]))
		if active != 0 {
			t.Errorf("%s active = %d, want 0", name, active)
		}

		level := int32(binary.LittleEndian.Uint32(data[off+4 : off+8]))
		if level != 0 {
			t.Errorf("%s level = %d, want 0", name, level)
		}
	}
}

func TestAgitDecoInfo_Write_MixedFunctions(t *testing.T) {
	t.Parallel()

	p := &AgitDecoInfo{
		HallID:       10,
		HPLevel:      2,
		MPLevel:      0,
		ExpLevel:     3,
		SPLevel:      0,
		TeleLevel:    1,
		CurtainLevel: 0,
		FrontLevel:   0,
		ItemLevel:    4,
		SupportLevel: 0,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if data[0] != 0xF7 {
		t.Errorf("opcode = 0x%02X, want 0xF7", data[0])
	}

	hallID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if hallID != 10 {
		t.Errorf("hallID = %d, want 10", hallID)
	}

	wantPairs := []struct {
		name       string
		wantActive int32
		wantLevel  int32
	}{
		{"HP", 1, 2},
		{"MP", 0, 0},
		{"Exp", 1, 3},
		{"SP", 0, 0},
		{"Tele", 1, 1},
		{"Curtain", 0, 0},
		{"Front", 0, 0},
		{"Item", 1, 4},
		{"Support", 0, 0},
	}

	for i, tt := range wantPairs {
		off := funcPairOffset(i)

		active := int32(binary.LittleEndian.Uint32(data[off : off+4]))
		if active != tt.wantActive {
			t.Errorf("%s active = %d, want %d", tt.name, active, tt.wantActive)
		}

		level := int32(binary.LittleEndian.Uint32(data[off+4 : off+8]))
		if level != tt.wantLevel {
			t.Errorf("%s level = %d, want %d", tt.name, level, tt.wantLevel)
		}
	}
}
