package clientpackets

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

// writeUTF16LE writes a Go string as L2-style UTF-16LE with null-terminator.
func writeUTF16LE(s string) []byte {
	runes := utf16.Encode([]rune(s))
	b := make([]byte, (len(runes)+1)*2)
	for i, r := range runes {
		binary.LittleEndian.PutUint16(b[i*2:], r)
	}
	// null terminator
	binary.LittleEndian.PutUint16(b[len(runes)*2:], 0)
	return b
}

func TestParseRequestAutoSoulShot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		data       []byte
		wantItemID int32
		wantType   int32
		wantErr    bool
	}{
		{
			name:       "enable soulshot",
			data:       makeInt32Pair(3947, 1),
			wantItemID: 3947,
			wantType:   1,
		},
		{
			name:       "disable soulshot",
			data:       makeInt32Pair(3948, 0),
			wantItemID: 3948,
			wantType:   0,
		},
		{
			name:    "too short",
			data:    []byte{1, 2},
			wantErr: true,
		},
		{
			name:    "empty",
			data:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pkt, err := ParseRequestAutoSoulShot(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pkt.ItemID != tt.wantItemID {
				t.Errorf("ItemID = %d; want %d", pkt.ItemID, tt.wantItemID)
			}
			if pkt.Type != tt.wantType {
				t.Errorf("Type = %d; want %d", pkt.Type, tt.wantType)
			}
		})
	}
}

func TestParseRequestDeleteMacro(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		data      []byte
		wantID    int32
		wantErr   bool
	}{
		{"valid", makeInt32Bytes(42), 42, false},
		{"zero", makeInt32Bytes(0), 0, false},
		{"too short", []byte{1}, 0, true},
		{"empty", nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pkt, err := ParseRequestDeleteMacro(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pkt.MacroID != tt.wantID {
				t.Errorf("MacroID = %d; want %d", pkt.MacroID, tt.wantID)
			}
		})
	}
}

func TestParseRequestCrystallizeItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     []byte
		wantObj  int32
		wantCnt  int64
		wantErr  bool
	}{
		{
			name:    "valid",
			data:    makeInt32Pair(100, 1),
			wantObj: 100,
			wantCnt: 1,
		},
		{
			name:    "too short",
			data:    []byte{1, 2, 3},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pkt, err := ParseRequestCrystallizeItem(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pkt.ObjectID != tt.wantObj {
				t.Errorf("ObjectID = %d; want %d", pkt.ObjectID, tt.wantObj)
			}
			if pkt.Count != tt.wantCnt {
				t.Errorf("Count = %d; want %d", pkt.Count, tt.wantCnt)
			}
		})
	}
}

func TestParseRequestSendFriendMsg(t *testing.T) {
	t.Parallel()

	// Build packet data: message string + receiver string (both UTF-16LE)
	msg := writeUTF16LE("Hello friend!")
	recv := writeUTF16LE("TargetPlayer")
	data := append(msg, recv...)

	pkt, err := ParseRequestSendFriendMsg(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.Message != "Hello friend!" {
		t.Errorf("Message = %q; want %q", pkt.Message, "Hello friend!")
	}
	if pkt.Receiver != "TargetPlayer" {
		t.Errorf("Receiver = %q; want %q", pkt.Receiver, "TargetPlayer")
	}
}

func TestParseRequestSendFriendMsg_Empty(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestSendFriendMsg(nil)
	if err == nil {
		t.Error("expected error for empty data")
	}
}

func TestParseRequestMakeMacro(t *testing.T) {
	t.Parallel()

	// Build macro packet:
	// int32 id, string name, string desc, string acronym, byte icon, byte count
	// then count × (byte entry, byte type, int32 d1, byte d2, string command)

	var buf []byte
	// ID
	b4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(b4, uint32(7))
	buf = append(buf, b4...)
	// Name
	buf = append(buf, writeUTF16LE("TestMacro")...)
	// Desc
	buf = append(buf, writeUTF16LE("A test")...)
	// Acronym
	buf = append(buf, writeUTF16LE("TM")...)
	// Icon (1 byte)
	buf = append(buf, 3)
	// Count (1 byte)
	buf = append(buf, 1)
	// Command 1: entry=0, type=1(skill), d1=101, d2=0, command="/use 101"
	buf = append(buf, 0) // entry
	buf = append(buf, 1) // type
	binary.LittleEndian.PutUint32(b4, uint32(101))
	buf = append(buf, b4...) // d1
	buf = append(buf, 0)     // d2
	buf = append(buf, writeUTF16LE("/use 101")...)

	pkt, err := ParseRequestMakeMacro(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.Macro == nil {
		t.Fatal("Macro is nil")
	}
	if pkt.Macro.ID != 7 {
		t.Errorf("Macro.ID = %d; want 7", pkt.Macro.ID)
	}
	if pkt.Macro.Name != "TestMacro" {
		t.Errorf("Macro.Name = %q; want %q", pkt.Macro.Name, "TestMacro")
	}
	if pkt.Macro.Icon != 3 {
		t.Errorf("Macro.Icon = %d; want 3", pkt.Macro.Icon)
	}
	if len(pkt.Macro.Commands) != 1 {
		t.Fatalf("len(Commands) = %d; want 1", len(pkt.Macro.Commands))
	}
	cmd := pkt.Macro.Commands[0]
	if cmd.Type != 1 {
		t.Errorf("Command.Type = %d; want 1", cmd.Type)
	}
	if cmd.D1 != 101 {
		t.Errorf("Command.D1 = %d; want 101", cmd.D1)
	}
	if cmd.Command != "/use 101" {
		t.Errorf("Command.Command = %q; want %q", cmd.Command, "/use 101")
	}
}

// makeInt32Pair creates an 8-byte LE buffer with two int32 values.
func makeInt32Pair(a, b int32) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data[:4], uint32(a))
	binary.LittleEndian.PutUint32(data[4:8], uint32(b))
	return data
}
