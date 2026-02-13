package serverpackets

import (
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/game/manor"
)

func TestExShowSeedInfo_Write_Empty(t *testing.T) {
	t.Parallel()

	p := &ExShowSeedInfo{
		ManorID:     1,
		HideButtons: false,
		Seeds:       nil,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", data[0])
	}

	subOp := int16(data[1]) | int16(data[2])<<8
	if subOp != SubOpcodeExShowSeedInfo {
		t.Errorf("subOpcode = 0x%04X; want 0x%04X", subOp, SubOpcodeExShowSeedInfo)
	}

	// hideButtons at offset 3
	if data[3] != 0 {
		t.Errorf("hideButtons = %d; want 0", data[3])
	}

	// manorID at offset 4
	manorID := int32(binary.LittleEndian.Uint32(data[4:8]))
	if manorID != 1 {
		t.Errorf("manorID = %d; want 1", manorID)
	}

	// count at offset 12 (4 + 4 unknown + 4 count)
	count := int32(binary.LittleEndian.Uint32(data[12:16]))
	if count != 0 {
		t.Errorf("count = %d; want 0", count)
	}
}

func TestExShowSeedInfo_Write_WithSeeds(t *testing.T) {
	t.Parallel()

	sp1 := manor.NewSeedProduction(5016, 100, 3000, 200)
	sp2 := manor.NewSeedProduction(5017, 50, 5000, 100)

	p := &ExShowSeedInfo{
		ManorID:     1,
		HideButtons: true,
		Seeds:       []*manor.SeedProduction{sp1, sp2},
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[3] != 1 {
		t.Errorf("hideButtons = %d; want 1", data[3])
	}

	count := int32(binary.LittleEndian.Uint32(data[12:16]))
	if count != 2 {
		t.Errorf("count = %d; want 2", count)
	}

	// First seed entry starts at offset 16
	seedID := int32(binary.LittleEndian.Uint32(data[16:20]))
	if seedID != 5016 {
		t.Errorf("seedID[0] = %d; want 5016", seedID)
	}

	amount := int32(binary.LittleEndian.Uint32(data[20:24]))
	if amount != 100 {
		t.Errorf("amount[0] = %d; want 100", amount)
	}

	startAmount := int32(binary.LittleEndian.Uint32(data[24:28]))
	if startAmount != 200 {
		t.Errorf("startAmount[0] = %d; want 200", startAmount)
	}

	price := int32(binary.LittleEndian.Uint32(data[28:32]))
	if price != 3000 {
		t.Errorf("price[0] = %d; want 3000", price)
	}
}

func TestExShowCropInfo_Write_Empty(t *testing.T) {
	t.Parallel()

	p := &ExShowCropInfo{
		ManorID:     2,
		HideButtons: false,
		Crops:       nil,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", data[0])
	}

	subOp := int16(data[1]) | int16(data[2])<<8
	if subOp != SubOpcodeExShowCropInfo {
		t.Errorf("subOpcode = 0x%04X; want 0x%04X", subOp, SubOpcodeExShowCropInfo)
	}

	count := int32(binary.LittleEndian.Uint32(data[12:16]))
	if count != 0 {
		t.Errorf("count = %d; want 0", count)
	}
}

func TestExShowCropInfo_Write_WithCrops(t *testing.T) {
	t.Parallel()

	cp := manor.NewCropProcure(5078, 50, 1, 100, 2000)

	p := &ExShowCropInfo{
		ManorID:     1,
		HideButtons: true,
		Crops:       []*manor.CropProcure{cp},
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[3] != 1 {
		t.Errorf("hideButtons = %d; want 1", data[3])
	}

	count := int32(binary.LittleEndian.Uint32(data[12:16]))
	if count != 1 {
		t.Errorf("count = %d; want 1", count)
	}

	// First crop entry starts at offset 16
	cropID := int32(binary.LittleEndian.Uint32(data[16:20]))
	if cropID != 5078 {
		t.Errorf("cropID = %d; want 5078", cropID)
	}

	amount := int32(binary.LittleEndian.Uint32(data[20:24]))
	if amount != 50 {
		t.Errorf("amount = %d; want 50", amount)
	}

	startAmount := int32(binary.LittleEndian.Uint32(data[24:28]))
	if startAmount != 100 {
		t.Errorf("startAmount = %d; want 100", startAmount)
	}

	price := int32(binary.LittleEndian.Uint32(data[28:32]))
	if price != 2000 {
		t.Errorf("price = %d; want 2000", price)
	}

	// rewardType at offset 32
	if data[32] != 1 {
		t.Errorf("rewardType = %d; want 1", data[32])
	}
}

func TestExShowSeedSetting_Write_Empty(t *testing.T) {
	t.Parallel()

	p := &ExShowSeedSetting{
		ManorID: 1,
		Seeds:   nil,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", data[0])
	}

	subOp := int16(data[1]) | int16(data[2])<<8
	if subOp != SubOpcodeExShowSeedSetting {
		t.Errorf("subOpcode = 0x%04X; want 0x%04X", subOp, SubOpcodeExShowSeedSetting)
	}

	manorID := int32(binary.LittleEndian.Uint32(data[3:7]))
	if manorID != 1 {
		t.Errorf("manorID = %d; want 1", manorID)
	}

	count := int32(binary.LittleEndian.Uint32(data[7:11]))
	if count != 0 {
		t.Errorf("count = %d; want 0", count)
	}
}

func TestExShowCropSetting_Write_Empty(t *testing.T) {
	t.Parallel()

	p := &ExShowCropSetting{
		ManorID: 3,
		Seeds:   nil,
	}

	data, err := p.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != 0xFE {
		t.Errorf("opcode = 0x%02X; want 0xFE", data[0])
	}

	subOp := int16(data[1]) | int16(data[2])<<8
	if subOp != SubOpcodeExShowCropSetting {
		t.Errorf("subOpcode = 0x%04X; want 0x%04X", subOp, SubOpcodeExShowCropSetting)
	}

	manorID := int32(binary.LittleEndian.Uint32(data[3:7]))
	if manorID != 3 {
		t.Errorf("manorID = %d; want 3", manorID)
	}

	count := int32(binary.LittleEndian.Uint32(data[7:11]))
	if count != 0 {
		t.Errorf("count = %d; want 0", count)
	}
}

func TestExShowSeedInfo_HideButtonsFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		hide bool
		want byte
	}{
		{"visible", false, 0},
		{"hidden", true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ExShowSeedInfo{ManorID: 1, HideButtons: tt.hide}
			data, err := p.Write()
			if err != nil {
				t.Fatalf("Write() error: %v", err)
			}
			if data[3] != tt.want {
				t.Errorf("hideButtons = %d; want %d", data[3], tt.want)
			}
		})
	}
}
