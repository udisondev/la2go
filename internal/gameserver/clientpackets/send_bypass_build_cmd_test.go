package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestParseSendBypassBuildCmd(t *testing.T) {
	w := packet.NewWriter(64)
	w.WriteString("spawn 20001 1")
	data := w.Bytes()

	pkt, err := ParseSendBypassBuildCmd(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.Command != "spawn 20001 1" {
		t.Errorf("Command = %q, want %q", pkt.Command, "spawn 20001 1")
	}
}

func TestParseSendBypassBuildCmd_Empty(t *testing.T) {
	w := packet.NewWriter(8)
	w.WriteString("")
	data := w.Bytes()

	pkt, err := ParseSendBypassBuildCmd(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.Command != "" {
		t.Errorf("Command = %q, want empty", pkt.Command)
	}
}
