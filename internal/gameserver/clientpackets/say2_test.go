package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestSay2_Parse_General(t *testing.T) {
	// Build packet: text + chatType(0) — no target for GENERAL
	w := packet.NewWriter(64)
	w.WriteString("Hello World")
	w.WriteInt(0) // GENERAL

	pkt, err := ParseSay2(w.Bytes())
	if err != nil {
		t.Fatalf("ParseSay2() error: %v", err)
	}

	if pkt.Text != "Hello World" {
		t.Errorf("Text = %q, want %q", pkt.Text, "Hello World")
	}
	if pkt.ChatType != 0 {
		t.Errorf("ChatType = %d, want 0", pkt.ChatType)
	}
	if pkt.Target != "" {
		t.Errorf("Target = %q, want empty", pkt.Target)
	}
}

func TestSay2_Parse_Whisper(t *testing.T) {
	// Build packet: text + chatType(2) + target
	w := packet.NewWriter(128)
	w.WriteString("secret message")
	w.WriteInt(2) // WHISPER
	w.WriteString("TargetPlayer")

	pkt, err := ParseSay2(w.Bytes())
	if err != nil {
		t.Fatalf("ParseSay2() error: %v", err)
	}

	if pkt.Text != "secret message" {
		t.Errorf("Text = %q, want %q", pkt.Text, "secret message")
	}
	if pkt.ChatType != 2 {
		t.Errorf("ChatType = %d, want 2", pkt.ChatType)
	}
	if pkt.Target != "TargetPlayer" {
		t.Errorf("Target = %q, want %q", pkt.Target, "TargetPlayer")
	}
}

func TestSay2_Parse_Shout(t *testing.T) {
	w := packet.NewWriter(64)
	w.WriteString("WTS +10 SoE 10kk")
	w.WriteInt(1) // SHOUT

	pkt, err := ParseSay2(w.Bytes())
	if err != nil {
		t.Fatalf("ParseSay2() error: %v", err)
	}

	if pkt.Text != "WTS +10 SoE 10kk" {
		t.Errorf("Text = %q, want %q", pkt.Text, "WTS +10 SoE 10kk")
	}
	if pkt.ChatType != 1 {
		t.Errorf("ChatType = %d, want 1", pkt.ChatType)
	}
	if pkt.Target != "" {
		t.Errorf("Target = %q, want empty", pkt.Target)
	}
}

func TestSay2_Parse_EmptyText(t *testing.T) {
	w := packet.NewWriter(16)
	w.WriteString("")
	w.WriteInt(0) // GENERAL

	pkt, err := ParseSay2(w.Bytes())
	if err != nil {
		t.Fatalf("ParseSay2() error: %v", err)
	}

	if pkt.Text != "" {
		t.Errorf("Text = %q, want empty", pkt.Text)
	}
}

func TestSay2_Parse_Trade(t *testing.T) {
	w := packet.NewWriter(64)
	w.WriteString("WTB Top D weapon")
	w.WriteInt(8) // TRADE

	pkt, err := ParseSay2(w.Bytes())
	if err != nil {
		t.Fatalf("ParseSay2() error: %v", err)
	}

	if pkt.Text != "WTB Top D weapon" {
		t.Errorf("Text = %q, want %q", pkt.Text, "WTB Top D weapon")
	}
	if pkt.ChatType != 8 {
		t.Errorf("ChatType = %d, want 8", pkt.ChatType)
	}
}
