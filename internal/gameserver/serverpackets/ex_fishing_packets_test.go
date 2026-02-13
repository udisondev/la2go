package serverpackets

import "testing"

func TestExFishingStart_Write(t *testing.T) {
	t.Parallel()

	p := ExFishingStart{
		ObjectID:  100,
		FishType:  1,
		X:         10000,
		Y:         20000,
		Z:         -3000,
		NightLure: false,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", data[0])
	}

	// Check subopcode (little-endian short)
	subOp := int16(data[1]) | int16(data[2])<<8
	if subOp != SubOpcodeExFishingStart {
		t.Errorf("subOpcode = 0x%04X; want 0x%04X", subOp, SubOpcodeExFishingStart)
	}
}

func TestExFishingStart_NightLure(t *testing.T) {
	t.Parallel()

	p := ExFishingStart{
		ObjectID:  200,
		NightLure: true,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// NightLure byte at offset 23 (1 + 2 + 5*4 = 23)
	if data[23] != 1 {
		t.Errorf("NightLure byte = %d; want 1", data[23])
	}
}

func TestExFishingEnd_Win(t *testing.T) {
	t.Parallel()

	p := ExFishingEnd{ObjectID: 100, Win: true}
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", data[0])
	}

	// Win byte at offset 7 (1 + 2 + 4 = 7)
	if data[7] != 1 {
		t.Errorf("win byte = %d; want 1", data[7])
	}
}

func TestExFishingEnd_Lose(t *testing.T) {
	t.Parallel()

	p := ExFishingEnd{ObjectID: 100, Win: false}
	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[7] != 0 {
		t.Errorf("win byte = %d; want 0", data[7])
	}
}

func TestExFishingStartCombat_Write(t *testing.T) {
	t.Parallel()

	p := ExFishingStartCombat{
		ObjectID:      100,
		Time:          30,
		HP:            500,
		Mode:          1,
		LureType:      1,
		DeceptiveMode: 0,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", data[0])
	}

	subOp := int16(data[1]) | int16(data[2])<<8
	if subOp != SubOpcodeExFishingStartCombat {
		t.Errorf("subOpcode = 0x%04X; want 0x%04X", subOp, SubOpcodeExFishingStartCombat)
	}

	// Mode byte at offset 15 (1 + 2 + 3*4 = 15)
	if data[15] != 1 {
		t.Errorf("mode byte = %d; want 1", data[15])
	}
}

func TestExFishingHpRegen_Write(t *testing.T) {
	t.Parallel()

	p := ExFishingHpRegen{
		ObjectID:   100,
		Time:       25,
		FishHP:     400,
		HPMode:     1,
		GoodUse:    1,
		Anim:       2,
		Penalty:    5,
		HPBarColor: 0,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", data[0])
	}

	subOp := int16(data[1]) | int16(data[2])<<8
	if subOp != SubOpcodeExFishingHpRegen {
		t.Errorf("subOpcode = 0x%04X; want 0x%04X", subOp, SubOpcodeExFishingHpRegen)
	}

	// HPMode at offset 15, GoodUse at 16, Anim at 17
	if data[15] != 1 {
		t.Errorf("hpMode byte = %d; want 1", data[15])
	}
	if data[16] != 1 {
		t.Errorf("goodUse byte = %d; want 1", data[16])
	}
	if data[17] != 2 {
		t.Errorf("anim byte = %d; want 2", data[17])
	}
}
