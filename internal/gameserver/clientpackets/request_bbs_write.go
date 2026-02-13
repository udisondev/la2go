package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestBBSwrite is the client packet opcode for RequestBBSwrite (0x22).
// Client sends this when submitting a form on the Community Board.
//
// Java reference: RequestBBSwrite.java
const OpcodeRequestBBSwrite = 0x22

// RequestBBSwrite represents the client's BBS write request.
//
// Packet structure:
//   - url (string) — write URL ("Topic", "Post", "Mail", "Region", "Notice")
//   - arg1..arg5 (string × 5) — form arguments
//
// Phase 30: Community Board.
type RequestBBSwrite struct {
	URL  string
	Args [5]string
}

// ParseRequestBBSwrite parses a RequestBBSwrite packet from raw bytes.
func ParseRequestBBSwrite(data []byte) (*RequestBBSwrite, error) {
	r := packet.NewReader(data)

	url, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading url: %w", err)
	}

	var args [5]string
	for i := range 5 {
		s, err := r.ReadString()
		if err != nil {
			return nil, fmt.Errorf("reading arg%d: %w", i+1, err)
		}
		args[i] = s
	}

	return &RequestBBSwrite{
		URL:  url,
		Args: args,
	}, nil
}
