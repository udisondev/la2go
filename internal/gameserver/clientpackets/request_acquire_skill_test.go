package clientpackets

import (
	"encoding/binary"
	"testing"
)

func TestParseRequestAcquireSkillInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		skillID   int32
		level     int32
		skillType int32
	}{
		{"class_skill", 1, 3, 0},
		{"fishing_skill", 1315, 1, 1},
		{"pledge_skill", 300, 5, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := make([]byte, 12)
			binary.LittleEndian.PutUint32(data[0:], uint32(tt.skillID))
			binary.LittleEndian.PutUint32(data[4:], uint32(tt.level))
			binary.LittleEndian.PutUint32(data[8:], uint32(tt.skillType))

			pkt, err := ParseRequestAcquireSkillInfo(data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pkt.SkillID != tt.skillID {
				t.Errorf("SkillID = %d; want %d", pkt.SkillID, tt.skillID)
			}
			if pkt.Level != tt.level {
				t.Errorf("Level = %d; want %d", pkt.Level, tt.level)
			}
			if pkt.SkillType != AcquireSkillType(tt.skillType) {
				t.Errorf("SkillType = %d; want %d", pkt.SkillType, tt.skillType)
			}
		})
	}
}

func TestParseRequestAcquireSkillInfo_TooShort(t *testing.T) {
	t.Parallel()
	_, err := ParseRequestAcquireSkillInfo([]byte{0x01, 0x00})
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestParseRequestAcquireSkill(t *testing.T) {
	t.Parallel()

	data := make([]byte, 12)
	binary.LittleEndian.PutUint32(data[0:], 56)  // skillID
	binary.LittleEndian.PutUint32(data[4:], 2)   // level
	binary.LittleEndian.PutUint32(data[8:], 0)   // CLASS type

	pkt, err := ParseRequestAcquireSkill(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.SkillID != 56 {
		t.Errorf("SkillID = %d; want 56", pkt.SkillID)
	}
	if pkt.Level != 2 {
		t.Errorf("Level = %d; want 2", pkt.Level)
	}
	if pkt.SkillType != AcquireSkillTypeClass {
		t.Errorf("SkillType = %d; want %d", pkt.SkillType, AcquireSkillTypeClass)
	}
}

func TestParseRequestAcquireSkill_TooShort(t *testing.T) {
	t.Parallel()
	_, err := ParseRequestAcquireSkill([]byte{0x01})
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestParseDlgAnswer(t *testing.T) {
	t.Parallel()

	data := make([]byte, 12)
	binary.LittleEndian.PutUint32(data[0:], 1024) // messageID
	binary.LittleEndian.PutUint32(data[4:], 1)    // answer = Yes
	binary.LittleEndian.PutUint32(data[8:], 500)  // requesterID

	pkt, err := ParseDlgAnswer(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.MessageID != 1024 {
		t.Errorf("MessageID = %d; want 1024", pkt.MessageID)
	}
	if pkt.Answer != 1 {
		t.Errorf("Answer = %d; want 1", pkt.Answer)
	}
	if pkt.RequesterID != 500 {
		t.Errorf("RequesterID = %d; want 500", pkt.RequesterID)
	}
}

func TestParseDlgAnswer_TooShort(t *testing.T) {
	t.Parallel()
	_, err := ParseDlgAnswer([]byte{0x01, 0x00, 0x00, 0x00})
	if err == nil {
		t.Error("expected error for short data")
	}
}
