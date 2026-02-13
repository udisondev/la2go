package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestLinkHtml is the C2S opcode 0x20 — click on NPC HTML link.
const OpcodeRequestLinkHtml byte = 0x20

// RequestLinkHtml represents a request to load a linked HTML page.
// Sent when the player clicks an <a action="link ..."> link in an NPC dialog.
type RequestLinkHtml struct {
	Link string // relative path, e.g. "merchant/30001-01.htm"
}

// ParseRequestLinkHtml parses the packet from raw bytes.
func ParseRequestLinkHtml(data []byte) (*RequestLinkHtml, error) {
	r := packet.NewReader(data)

	link, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading Link: %w", err)
	}

	return &RequestLinkHtml{Link: link}, nil
}
