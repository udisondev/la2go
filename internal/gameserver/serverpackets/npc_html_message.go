package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeNpcHtmlMessage is the opcode for NpcHtmlMessage packet (S2C 0x0F).
// Sends HTML dialog to client from NPC interaction.
//
// Phase 8.2: NPC Dialogues.
// Java reference: NpcHtmlMessage.java
const OpcodeNpcHtmlMessage = 0x0F

// NpcHtmlMessage sends an HTML dialog window to the client.
//
// Packet structure:
//   - opcode (byte) — 0x0F
//   - npcObjectID (int32) — NPC object ID (0 if not from NPC)
//   - html (string) — HTML content (UTF-16LE null-terminated)
//   - itemID (int32) — item object ID for item dialogs (0 if not item dialog)
//
// Phase 8.2: NPC Dialogues.
type NpcHtmlMessage struct {
	NpcObjectID int32
	Html        string
	ItemID      int32
}

// NewNpcHtmlMessage creates NpcHtmlMessage from NPC object ID and HTML content.
func NewNpcHtmlMessage(npcObjectID int32, html string) NpcHtmlMessage {
	return NpcHtmlMessage{
		NpcObjectID: npcObjectID,
		Html:        html,
	}
}

// Write serializes NpcHtmlMessage packet to bytes.
//
// Phase 8.2: NPC Dialogues.
func (p NpcHtmlMessage) Write() ([]byte, error) {
	// 1 opcode + 4 npcObjID + html (variable) + 4 itemID
	w := packet.NewWriter(9 + len(p.Html)*2 + 2) // *2 for UTF-16LE + null terminator

	w.WriteByte(OpcodeNpcHtmlMessage)
	w.WriteInt(p.NpcObjectID)
	w.WriteString(p.Html)
	w.WriteInt(p.ItemID)

	return w.Bytes(), nil
}
