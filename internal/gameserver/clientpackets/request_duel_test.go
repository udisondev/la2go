package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestParseRequestDuelStart(t *testing.T) {
	tests := []struct {
		name      string
		playerName string
		partyDuel int32
		wantParty bool
	}{
		{"1v1 duel", "Alice", 0, false},
		{"party duel", "Bob", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := packet.NewWriter(64)
			w.WriteString(tt.playerName)
			w.WriteInt(tt.partyDuel)

			pkt, err := ParseRequestDuelStart(w.Bytes())
			if err != nil {
				t.Fatalf("ParseRequestDuelStart: %v", err)
			}
			if pkt.Name != tt.playerName {
				t.Errorf("Name = %q; want %q", pkt.Name, tt.playerName)
			}
			if pkt.PartyDuel != tt.wantParty {
				t.Errorf("PartyDuel = %v; want %v", pkt.PartyDuel, tt.wantParty)
			}
		})
	}
}

func TestParseRequestDuelStart_ShortData(t *testing.T) {
	_, err := ParseRequestDuelStart(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
}

func TestParseRequestDuelAnswerStart(t *testing.T) {
	tests := []struct {
		name      string
		partyDuel int32
		response  int32
		wantParty bool
		wantAccepted bool
	}{
		{"accept 1v1", 0, 1, false, true},
		{"decline 1v1", 0, 0, false, false},
		{"accept party", 1, 1, true, true},
		{"decline party", 1, 0, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := packet.NewWriter(8)
			w.WriteInt(tt.partyDuel)
			w.WriteInt(tt.response)

			pkt, err := ParseRequestDuelAnswerStart(w.Bytes())
			if err != nil {
				t.Fatalf("ParseRequestDuelAnswerStart: %v", err)
			}
			if pkt.PartyDuel != tt.wantParty {
				t.Errorf("PartyDuel = %v; want %v", pkt.PartyDuel, tt.wantParty)
			}
			if pkt.Accepted != tt.wantAccepted {
				t.Errorf("Accepted = %v; want %v", pkt.Accepted, tt.wantAccepted)
			}
		})
	}
}

func TestParseRequestDuelAnswerStart_ShortData(t *testing.T) {
	_, err := ParseRequestDuelAnswerStart([]byte{0x01})
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestParseRequestDuelSurrender(t *testing.T) {
	pkt, err := ParseRequestDuelSurrender(nil)
	if err != nil {
		t.Fatalf("ParseRequestDuelSurrender: %v", err)
	}
	if pkt == nil {
		t.Error("expected non-nil packet")
	}
}
