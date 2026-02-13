package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeCharCreateFail is the opcode for CharCreateFail (S2C 0x1A).
// Java reference: ServerPackets.CHAR_CREATE_FAIL(0x1A).
const OpcodeCharCreateFail = 0x1A

// CharCreateFail reason codes.
const (
	CharCreateReasonFailed          int32 = 0x00 // "Your character creation has failed."
	CharCreateReasonTooMany         int32 = 0x01 // "You cannot create another character."
	CharCreateReasonNameExists      int32 = 0x02 // "This name already exists."
	CharCreateReasonNameTooLong     int32 = 0x03 // "Your title cannot exceed 16 characters."
	CharCreateReasonIncorrectName   int32 = 0x04 // "Incorrect name. Please try again."
	CharCreateReasonNotAllowed      int32 = 0x05 // "Characters cannot be created from this server."
	CharCreateReasonChooseAnotherSv int32 = 0x06 // "Unable to create character."
)

// CharCreateFail reports character creation failure with reason.
//
// Packet structure:
//   - opcode (byte) — 0x1A
//   - reason (int32) — error code
type CharCreateFail struct {
	Reason int32
}

// NewCharCreateFail creates a CharCreateFail packet.
func NewCharCreateFail(reason int32) *CharCreateFail {
	return &CharCreateFail{Reason: reason}
}

// Write serializes CharCreateFail packet.
func (p *CharCreateFail) Write() ([]byte, error) {
	w := packet.NewWriter(8)
	w.WriteByte(OpcodeCharCreateFail)
	w.WriteInt(p.Reason)
	return w.Bytes(), nil
}
