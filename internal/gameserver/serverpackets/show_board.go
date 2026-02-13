package serverpackets

import (
	"github.com/udisondev/la2go/internal/game/bbs"
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeShowBoard is the opcode for ShowBoard packet (S2C 0x6E).
// Sends community board HTML content to the client.
//
// Java reference: ShowBoard.java
const OpcodeShowBoard = 0x6E

// ShowBoard sends community board content to the client.
//
// Packet structure:
//   - opcode (byte) — 0x6E
//   - showBoard (byte) — 1=show, 0=hide
//   - 8 navigation bypass strings (string × 8) — fixed top panel buttons
//   - content (string) — HTML content with ID prefix separated by \u0008
//
// Java reference: ShowBoard.java
type ShowBoard struct {
	Show    bool   // true = show board, false = hide
	Content string // "ID\u0008HTML" format
}

// NewShowBoard creates a ShowBoard with HTML content.
// id: "101", "102", "103" for multi-part content, or "1001" for single.
// html: HTML content string.
func NewShowBoard(id, html string) ShowBoard {
	return ShowBoard{
		Show:    true,
		Content: bbs.FormatContent(id, html),
	}
}

// NewShowBoardHide creates a ShowBoard that hides the community board.
func NewShowBoardHide() ShowBoard {
	return ShowBoard{Show: false}
}

// Write serializes ShowBoard packet to bytes.
func (p ShowBoard) Write() ([]byte, error) {
	// Estimate: 1 opcode + 1 show + 8 strings (~40 bytes each) + content
	w := packet.NewWriter(1 + 1 + 8*80 + len(p.Content)*2 + 2)

	w.WriteByte(OpcodeShowBoard)

	if p.Show {
		w.WriteByte(0x01)
	} else {
		w.WriteByte(0x00)
	}

	// 8 фиксированных bypass-кнопок верхней панели
	for _, btn := range bbs.NavigationButtons {
		w.WriteString(btn)
	}

	w.WriteString(p.Content)

	return w.Bytes(), nil
}
