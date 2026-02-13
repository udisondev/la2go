package serverpackets

import (
	"testing"
)

func TestExShowVariationMakeWindow_Write(t *testing.T) {
	t.Parallel()

	pkt := ExShowVariationMakeWindow{}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if len(data) < 3 {
		t.Fatalf("Write() len = %d, want >= 3", len(data))
	}
	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X, want 0xFE", data[0])
	}
	subOp := int16(data[1]) | int16(data[2])<<8
	if subOp != SubOpcodeExShowVariationMakeWindow {
		t.Errorf("subOpcode = 0x%04X, want 0x%04X", subOp, SubOpcodeExShowVariationMakeWindow)
	}
}

func TestExShowVariationCancelWindow_Write(t *testing.T) {
	t.Parallel()

	pkt := ExShowVariationCancelWindow{}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X, want 0xFE", data[0])
	}
	subOp := int16(data[1]) | int16(data[2])<<8
	if subOp != SubOpcodeExShowVariationCancelWindow {
		t.Errorf("subOpcode = 0x%04X, want 0x%04X", subOp, SubOpcodeExShowVariationCancelWindow)
	}
}

func TestExVariationResult_Write(t *testing.T) {
	t.Parallel()

	pkt := ExVariationResult{
		Stat12: 12345,
		Stat34: 100,
		Result: 1,
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// 1 opcode + 2 subop + 3*4 data = 15 bytes
	if len(data) < 15 {
		t.Fatalf("Write() len = %d, want >= 15", len(data))
	}
	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X, want 0xFE", data[0])
	}
	subOp := int16(data[1]) | int16(data[2])<<8
	if subOp != SubOpcodeExVariationResult {
		t.Errorf("subOpcode = 0x%04X, want 0x%04X", subOp, SubOpcodeExVariationResult)
	}
}

func TestExVariationCancelResult_Write(t *testing.T) {
	t.Parallel()

	pkt := ExVariationCancelResult{Result: 1}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// 1 opcode + 2 subop + 4 closeWindow + 4 result = 11 bytes
	if len(data) < 11 {
		t.Fatalf("Write() len = %d, want >= 11", len(data))
	}
	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X, want 0xFE", data[0])
	}
}

func TestExPutItemResultForVariation_Write(t *testing.T) {
	t.Parallel()

	pkt := ExPutItemResultForVariation{
		ItemObjectID: 42,
		ItemID:       100,
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// 1 + 2 + 4 + 4 + 4 = 15 bytes (added hardcoded success flag)
	if len(data) < 15 {
		t.Fatalf("Write() len = %d, want >= 15", len(data))
	}
	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X, want 0xFE", data[0])
	}
}

func TestExPutCommissionResult_Write(t *testing.T) {
	t.Parallel()

	pkt := ExPutCommissionResult{
		GemstoneObjectID: 1458,
		GemstoneCount:    20,
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	// 1 + 2 + 5*4 = 23 bytes (gemObjId, 0, gemCount, 0, 1)
	if len(data) < 23 {
		t.Fatalf("Write() len = %d, want >= 23", len(data))
	}
}

func TestExPutIntensiveResult_Write(t *testing.T) {
	t.Parallel()

	pkt := ExPutIntensiveResult{
		RefinerObjectID: 50,
		LifeStoneItemID: 8723,
		GemstoneItemID:  2131,
		GemstoneCount:   20,
	}
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	// 1 + 2 + 5*4 = 23 bytes (refinerObjId, lifeStoneId, gemId, gemCount, 1)
	if len(data) < 23 {
		t.Fatalf("Write() len = %d, want >= 23", len(data))
	}
}

func TestServerSubOpcodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  int16
		want int16
	}{
		{"ShowVariationMakeWindow", SubOpcodeExShowVariationMakeWindow, 0x50},
		{"ShowVariationCancelWindow", SubOpcodeExShowVariationCancelWindow, 0x51},
		{"PutItemResult", SubOpcodeExPutItemResultForVariation, 0x52},
		{"PutIntensiveResult", SubOpcodeExPutIntensiveResult, 0x53},
		{"PutCommissionResult", SubOpcodeExPutCommissionResult, 0x54},
		{"VariationResult", SubOpcodeExVariationResult, 0x55},
		{"PutItemForVariationCancel", SubOpcodeExPutItemForVariationCancel, 0x56},
		{"VariationCancelResult", SubOpcodeExVariationCancelResult, 0x57},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = 0x%02X, want 0x%02X", tt.name, tt.got, tt.want)
			}
		})
	}
}
