package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestParseRequestBypassToServer(t *testing.T) {
	// Build test packet with bypass string
	bypass := "npc_12345_Shop"
	w := packet.NewWriter(64)
	w.WriteString(bypass)

	pkt, err := ParseRequestBypassToServer(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestBypassToServer() error: %v", err)
	}

	if pkt.Bypass != bypass {
		t.Errorf("Bypass = %q, want %q", pkt.Bypass, bypass)
	}
}

func TestParseRequestBypassToServer_BBS(t *testing.T) {
	w := packet.NewWriter(32)
	w.WriteString("_bbshome")

	pkt, err := ParseRequestBypassToServer(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestBypassToServer() error: %v", err)
	}

	if pkt.Bypass != "_bbshome" {
		t.Errorf("Bypass = %q, want %q", pkt.Bypass, "_bbshome")
	}
}

func TestParseRequestBypassToServer_EmptyData(t *testing.T) {
	_, err := ParseRequestBypassToServer(nil)
	if err == nil {
		t.Error("expected error for empty data, got nil")
	}
}
