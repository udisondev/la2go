package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeCharDeleteFail is the opcode for CharDeleteFail (S2C 0x24).
// Java reference: ServerPackets.CHAR_DELETE_FAIL(0x24).
const OpcodeCharDeleteFail = 0x24

// CharDeleteFail reason codes.
const (
	CharDeleteReasonFailed     int32 = 1 // Generic deletion failure
	CharDeleteReasonClanMember int32 = 2 // "You may not delete a clan member."
	CharDeleteReasonClanLeader int32 = 3 // "Clan leaders may not be deleted."
)

// CharDeleteFail reports character deletion failure with reason.
//
// Packet structure:
//   - opcode (byte) — 0x24
//   - reason (int32) — error code
type CharDeleteFail struct {
	Reason int32
}

// NewCharDeleteFail creates a CharDeleteFail packet.
func NewCharDeleteFail(reason int32) *CharDeleteFail {
	return &CharDeleteFail{Reason: reason}
}

// Write serializes CharDeleteFail packet.
func (p *CharDeleteFail) Write() ([]byte, error) {
	w := packet.NewWriter(8)
	w.WriteByte(OpcodeCharDeleteFail)
	w.WriteInt(p.Reason)
	return w.Bytes(), nil
}
