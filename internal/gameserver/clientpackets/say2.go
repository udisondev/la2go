package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeSay2 is the client packet opcode for chat messages (C2S 0x38).
//
// Packet structure (C2S 0x38):
//   - text   string  message text (UTF-16LE null-terminated)
//   - type   int32   chat channel type
//   - target string  recipient name (ONLY for WHISPER, type == 2)
//
// Phase 5.11: Chat System.
// Java reference: Say2.java:98-103.
const OpcodeSay2 = 0x38

// Say2 represents a client chat message packet.
type Say2 struct {
	Text     string
	ChatType int32
	Target   string // non-empty only for WHISPER (ChatType == 2)
}

// ParseSay2 parses Say2 packet from raw bytes.
// Opcode already stripped by HandlePacket.
func ParseSay2(data []byte) (*Say2, error) {
	r := packet.NewReader(data)

	text, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading text: %w", err)
	}

	chatType, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading chatType: %w", err)
	}

	var target string
	// WHISPER (type 2) includes target name
	if chatType == 2 {
		target, err = r.ReadString()
		if err != nil {
			return nil, fmt.Errorf("reading whisper target: %w", err)
		}
	}

	return &Say2{
		Text:     text,
		ChatType: chatType,
		Target:   target,
	}, nil
}
