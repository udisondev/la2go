package serverpackets

import (
	"encoding/binary"
	"testing"
)

func TestPlaySound_Write_Normal(t *testing.T) {
	pkt := NewPlaySound(SoundQuestAccept)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != OpcodePlaySound {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePlaySound)
	}

	soundType := int32(binary.LittleEndian.Uint32(data[1:5]))
	if soundType != SoundTypeNormal {
		t.Errorf("soundType = %d, want %d", soundType, SoundTypeNormal)
	}

	// Проверяем что creatureID = 0 (после строки)
	// Строка в UTF-16LE + null terminator, потом creatureID
}

func TestPlaySound_Write_3D(t *testing.T) {
	pkt := NewPlaySound3D("test_sound", 42, 100, 200, 300)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != OpcodePlaySound {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePlaySound)
	}

	soundType := int32(binary.LittleEndian.Uint32(data[1:5]))
	if soundType != SoundType3D {
		t.Errorf("soundType = %d, want %d", soundType, SoundType3D)
	}
}

func TestPlaySound_SoundConstants(t *testing.T) {
	// Проверяем что константы не пустые
	sounds := []string{
		SoundQuestAccept,
		SoundQuestMiddle,
		SoundQuestFinish,
		SoundQuestGiveUp,
		SoundQuestItemGet,
		SoundQuestTutorial,
		SoundQuestFanfareM,
		SoundQuestFanfare1,
		SoundQuestFanfare2,
	}

	for _, s := range sounds {
		if s == "" {
			t.Errorf("sound constant is empty")
		}
	}
}
