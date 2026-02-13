package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestParseRequestChangePartyLeader(t *testing.T) {
	w := packet.NewWriter(64)
	w.WriteString("NewLeader")
	data := w.Bytes()

	pkt, err := ParseRequestChangePartyLeader(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.Name != "NewLeader" {
		t.Errorf("Name = %q, want %q", pkt.Name, "NewLeader")
	}
}

func TestParseRequestChangePartyLeader_Empty(t *testing.T) {
	w := packet.NewWriter(8)
	w.WriteString("")
	data := w.Bytes()

	pkt, err := ParseRequestChangePartyLeader(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.Name != "" {
		t.Errorf("Name = %q, want empty", pkt.Name)
	}
}
