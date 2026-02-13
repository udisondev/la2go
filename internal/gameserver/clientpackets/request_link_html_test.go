package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestParseRequestLinkHtml(t *testing.T) {
	t.Parallel()

	// Build packet with a string
	w := packet.NewWriter(64)
	w.WriteString("merchant/30001-01.htm")
	data := w.Bytes()

	pkt, err := ParseRequestLinkHtml(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.Link != "merchant/30001-01.htm" {
		t.Errorf("Link = %q; want %q", pkt.Link, "merchant/30001-01.htm")
	}
}

func TestParseRequestLinkHtml_Empty(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(4)
	w.WriteString("")
	data := w.Bytes()

	pkt, err := ParseRequestLinkHtml(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.Link != "" {
		t.Errorf("Link = %q; want empty", pkt.Link)
	}
}

func TestParseRequestLinkHtml_TooShort(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestLinkHtml([]byte{})
	if err == nil {
		t.Error("expected error for empty data")
	}
}
