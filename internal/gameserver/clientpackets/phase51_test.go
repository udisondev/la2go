package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// --- ParseRequestJoinAlly ---

func TestParseRequestJoinAlly(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(8)
	w.WriteInt(12345)

	pkt, err := ParseRequestJoinAlly(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestJoinAlly: %v", err)
	}
	if pkt.ObjectID != 12345 {
		t.Errorf("ObjectID = %d; want 12345", pkt.ObjectID)
	}
}

func TestParseRequestJoinAlly_TooShort(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestJoinAlly([]byte{0x01})
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestParseRequestJoinAlly_NilData(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestJoinAlly(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
}

// --- ParseRequestAnswerJoinAlly ---

func TestParseRequestAnswerJoinAlly_Accept(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(8)
	w.WriteInt(1)

	pkt, err := ParseRequestAnswerJoinAlly(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestAnswerJoinAlly: %v", err)
	}
	if pkt.Response != 1 {
		t.Errorf("Response = %d; want 1", pkt.Response)
	}
}

func TestParseRequestAnswerJoinAlly_Decline(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(8)
	w.WriteInt(0)

	pkt, err := ParseRequestAnswerJoinAlly(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestAnswerJoinAlly: %v", err)
	}
	if pkt.Response != 0 {
		t.Errorf("Response = %d; want 0", pkt.Response)
	}
}

func TestParseRequestAnswerJoinAlly_TooShort(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestAnswerJoinAlly([]byte{})
	if err == nil {
		t.Error("expected error for empty data")
	}
}

// --- ParseAllyDismiss ---

func TestParseAllyDismiss(t *testing.T) {
	t.Parallel()

	// ReadString читает UTF-16LE null-terminated строку.
	// Используем packet.Writer для корректного формирования данных.
	w := packet.NewWriter(64)
	w.WriteString("TestClan")

	pkt, err := ParseAllyDismiss(w.Bytes())
	if err != nil {
		t.Fatalf("ParseAllyDismiss: %v", err)
	}
	if pkt.ClanName != "TestClan" {
		t.Errorf("ClanName = %q; want %q", pkt.ClanName, "TestClan")
	}
}

func TestParseAllyDismiss_EmptyString(t *testing.T) {
	t.Parallel()

	// Пустая строка — только null-terminator (0x00 0x00)
	w := packet.NewWriter(8)
	w.WriteString("")

	pkt, err := ParseAllyDismiss(w.Bytes())
	if err != nil {
		t.Fatalf("ParseAllyDismiss: %v", err)
	}
	if pkt.ClanName != "" {
		t.Errorf("ClanName = %q; want empty", pkt.ClanName)
	}
}

func TestParseAllyDismiss_TooShort(t *testing.T) {
	t.Parallel()

	// Один байт — ReadString требует минимум 2 байта для null-terminator
	_, err := ParseAllyDismiss([]byte{0x41})
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestParseAllyDismiss_NoNullTerminator(t *testing.T) {
	t.Parallel()

	// UTF-16LE 'T' без null-terminator — обрыв данных
	_, err := ParseAllyDismiss([]byte{'T', 0x00})
	if err == nil {
		t.Error("expected error for data without null terminator")
	}
}

// --- ParseRequestSetAllyCrest ---

func TestParseRequestSetAllyCrest(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(16)
	w.WriteInt(3)                            // length = 3
	w.WriteBytes([]byte{0xAA, 0xBB, 0xCC}) // 3 bytes of crest data

	pkt, err := ParseRequestSetAllyCrest(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestSetAllyCrest: %v", err)
	}
	if pkt.Length != 3 {
		t.Errorf("Length = %d; want 3", pkt.Length)
	}
	if len(pkt.Data) != 3 {
		t.Fatalf("len(Data) = %d; want 3", len(pkt.Data))
	}
	if pkt.Data[0] != 0xAA || pkt.Data[1] != 0xBB || pkt.Data[2] != 0xCC {
		t.Errorf("Data = %x; want [AA BB CC]", pkt.Data)
	}
}

func TestParseRequestSetAllyCrest_Empty(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(8)
	w.WriteInt(0) // length = 0, no data

	pkt, err := ParseRequestSetAllyCrest(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestSetAllyCrest: %v", err)
	}
	if pkt.Length != 0 {
		t.Errorf("Length = %d; want 0", pkt.Length)
	}
	if len(pkt.Data) != 0 {
		t.Errorf("len(Data) = %d; want 0", len(pkt.Data))
	}
}

func TestParseRequestSetAllyCrest_TooShort(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestSetAllyCrest([]byte{0x01})
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestParseRequestSetAllyCrest_DataTruncated(t *testing.T) {
	t.Parallel()

	// length = 10, но данных после int32 только 2 байта
	w := packet.NewWriter(8)
	w.WriteInt(10)
	w.WriteBytes([]byte{0xAA, 0xBB})

	_, err := ParseRequestSetAllyCrest(w.Bytes())
	if err == nil {
		t.Error("expected error for truncated crest data")
	}
}

// --- Opcode Constants ---

func TestAllianceOpcodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		got    byte
		want   byte
	}{
		{"OpcodeRequestJoinAlly", OpcodeRequestJoinAlly, 0x82},
		{"OpcodeRequestAnswerJoinAlly", OpcodeRequestAnswerJoinAlly, 0x83},
		{"OpcodeAllyLeave", OpcodeAllyLeave, 0x84},
		{"OpcodeAllyDismiss", OpcodeAllyDismiss, 0x85},
		{"OpcodeRequestDismissAlly", OpcodeRequestDismissAlly, 0x86},
		{"OpcodeRequestSetAllyCrest", OpcodeRequestSetAllyCrest, 0x87},
		{"OpcodeRequestAllyCrest", OpcodeRequestAllyCrest, 0x88},
		{"OpcodeRequestAllyInfo", OpcodeRequestAllyInfo, 0x8E},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = 0x%02X; want 0x%02X", tt.name, tt.got, tt.want)
			}
		})
	}
}
