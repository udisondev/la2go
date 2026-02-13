package clientpackets

import (
	"encoding/binary"
	"testing"
)

func TestParseRequestSocialAction(t *testing.T) {
	tests := []struct {
		name     string
		actionID int32
	}{
		{"greeting", 2},
		{"victory", 3},
		{"charm", 15},
		{"shyness", 16},
		{"negative", -1},
		{"zero", 0},
		{"high_value", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, 4)
			binary.LittleEndian.PutUint32(data, uint32(tt.actionID))

			pkt, err := ParseRequestSocialAction(data)
			if err != nil {
				t.Fatalf("ParseRequestSocialAction() error = %v", err)
			}

			if pkt.ActionID != tt.actionID {
				t.Errorf("ActionID = %d; want %d", pkt.ActionID, tt.actionID)
			}
		})
	}
}

func TestParseRequestSocialAction_TooShort(t *testing.T) {
	data := make([]byte, 2) // need at least 4 bytes for int32
	_, err := ParseRequestSocialAction(data)
	if err == nil {
		t.Fatal("ParseRequestSocialAction() expected error for short data, got nil")
	}
}

func TestParseRequestSocialAction_EmptyData(t *testing.T) {
	_, err := ParseRequestSocialAction(nil)
	if err == nil {
		t.Fatal("ParseRequestSocialAction() expected error for nil data, got nil")
	}
}

func TestOpcodeRequestSocialAction(t *testing.T) {
	if OpcodeRequestSocialAction != 0x1B {
		t.Errorf("OpcodeRequestSocialAction = 0x%02X; want 0x1B", OpcodeRequestSocialAction)
	}
}
