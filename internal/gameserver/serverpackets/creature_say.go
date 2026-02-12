package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeCreatureSay is the server packet opcode for chat messages (S2C 0x4A).
// Java reference: CreatureSay.java (opcode 0x4A).
const OpcodeCreatureSay = 0x4A

// CreatureSay represents a chat message packet sent to the client.
//
// Packet structure (S2C 0x4A):
//   - opcode     byte    0x4A
//   - objectID   int32   sender objectID (0 if system)
//   - chatType   int32   chat channel type
//   - senderName string  sender name (UTF-16LE null-terminated)
//   - text       string  message text (UTF-16LE null-terminated)
//
// Phase 5.11: Chat System.
// Java reference: CreatureSay.java:86-114.
type CreatureSay struct {
	ObjectID   int32
	ChatType   int32
	SenderName string
	Text       string
}

// NewCreatureSay creates a new CreatureSay packet.
func NewCreatureSay(objectID int32, chatType int32, senderName, text string) CreatureSay {
	return CreatureSay{
		ObjectID:   objectID,
		ChatType:   chatType,
		SenderName: senderName,
		Text:       text,
	}
}

// Write serializes the CreatureSay packet to bytes.
func (p *CreatureSay) Write() ([]byte, error) {
	// opcode(1) + objectID(4) + chatType(4) + name(~32) + text(~256)
	w := packet.NewWriter(128)

	w.WriteByte(OpcodeCreatureSay)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.ChatType)
	w.WriteString(p.SenderName)
	w.WriteString(p.Text)

	return w.Bytes(), nil
}
