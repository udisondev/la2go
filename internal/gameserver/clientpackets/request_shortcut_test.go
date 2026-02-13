package clientpackets

import (
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func writeInt32(buf []byte, offset int, val int32) {
	binary.LittleEndian.PutUint32(buf[offset:], uint32(val))
}

func TestParseRequestShortCutReg(t *testing.T) {
	tests := []struct {
		name      string
		typeID    int32
		slotPage  int32
		id        int32
		level     int32
		wantType  model.ShortcutType
		wantSlot  int8
		wantPage  int8
		wantID    int32
		wantLevel int32
	}{
		{
			name:      "skill slot 3 page 0",
			typeID:    2, // ShortcutTypeSkill
			slotPage:  3,
			id:        1001,
			level:     5,
			wantType:  model.ShortcutTypeSkill,
			wantSlot:  3,
			wantPage:  0,
			wantID:    1001,
			wantLevel: 5,
		},
		{
			name:      "item slot 0 page 2",
			typeID:    1, // ShortcutTypeItem
			slotPage:  24, // 0 + 2*12
			id:        500,
			level:     0,
			wantType:  model.ShortcutTypeItem,
			wantSlot:  0,
			wantPage:  2,
			wantID:    500,
			wantLevel: 0,
		},
		{
			name:      "action slot 11 page 9",
			typeID:    3, // ShortcutTypeAction
			slotPage:  119, // 11 + 9*12
			id:        42,
			level:     0,
			wantType:  model.ShortcutTypeAction,
			wantSlot:  11,
			wantPage:  9,
			wantID:    42,
			wantLevel: 0,
		},
		{
			name:      "invalid type defaults to None",
			typeID:    99,
			slotPage:  0,
			id:        1,
			level:     0,
			wantType:  model.ShortcutTypeNone,
			wantSlot:  0,
			wantPage:  0,
			wantID:    1,
			wantLevel: 0,
		},
		{
			name:      "recipe type",
			typeID:    5, // ShortcutTypeRecipe
			slotPage:  5,
			id:        300,
			level:     0,
			wantType:  model.ShortcutTypeRecipe,
			wantSlot:  5,
			wantPage:  0,
			wantID:    300,
			wantLevel: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, 16) // 4 int32 values
			writeInt32(data, 0, tt.typeID)
			writeInt32(data, 4, tt.slotPage)
			writeInt32(data, 8, tt.id)
			writeInt32(data, 12, tt.level)

			pkt, err := ParseRequestShortCutReg(data)
			if err != nil {
				t.Fatalf("ParseRequestShortCutReg() error: %v", err)
			}
			if pkt.Type != tt.wantType {
				t.Errorf("Type = %d; want %d", pkt.Type, tt.wantType)
			}
			if pkt.Slot != tt.wantSlot {
				t.Errorf("Slot = %d; want %d", pkt.Slot, tt.wantSlot)
			}
			if pkt.Page != tt.wantPage {
				t.Errorf("Page = %d; want %d", pkt.Page, tt.wantPage)
			}
			if pkt.ID != tt.wantID {
				t.Errorf("ID = %d; want %d", pkt.ID, tt.wantID)
			}
			if pkt.Level != tt.wantLevel {
				t.Errorf("Level = %d; want %d", pkt.Level, tt.wantLevel)
			}
		})
	}
}

func TestParseRequestShortCutRegTruncated(t *testing.T) {
	tests := []struct {
		name    string
		dataLen int
	}{
		{"empty", 0},
		{"only type", 4},
		{"type+slot", 8},
		{"type+slot+id", 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.dataLen)
			_, err := ParseRequestShortCutReg(data)
			if err == nil {
				t.Error("expected error for truncated data, got nil")
			}
		})
	}
}

func TestParseRequestShortCutDel(t *testing.T) {
	tests := []struct {
		name     string
		slotPage int32
		wantSlot int8
		wantPage int8
	}{
		{"slot 0 page 0", 0, 0, 0},
		{"slot 5 page 0", 5, 5, 0},
		{"slot 0 page 1", 12, 0, 1},
		{"slot 11 page 9", 119, 11, 9},
		{"slot 7 page 4", 55, 7, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, 4)
			writeInt32(data, 0, tt.slotPage)

			pkt, err := ParseRequestShortCutDel(data)
			if err != nil {
				t.Fatalf("ParseRequestShortCutDel() error: %v", err)
			}
			if pkt.Slot != tt.wantSlot {
				t.Errorf("Slot = %d; want %d", pkt.Slot, tt.wantSlot)
			}
			if pkt.Page != tt.wantPage {
				t.Errorf("Page = %d; want %d", pkt.Page, tt.wantPage)
			}
		})
	}
}

func TestParseRequestShortCutDelEmpty(t *testing.T) {
	_, err := ParseRequestShortCutDel(nil)
	if err == nil {
		t.Error("expected error for empty data, got nil")
	}
}
