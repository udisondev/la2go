package serverpackets

import "encoding/binary"

const AccountKickedOpcode = 0x02

// AccountKickedReason codes.
const (
	ReasonDataStealer       int32 = 0x01
	ReasonGenericViolation  int32 = 0x08
	Reason7DaysSuspended    int32 = 0x10
	ReasonPermanentlyBanned int32 = 0x20
)

// AccountKicked writes the AccountKicked packet (opcode 0x02) into buf.
// Returns the number of bytes written.
func AccountKicked(buf []byte, reason int32) int {
	buf[0] = AccountKickedOpcode
	binary.LittleEndian.PutUint32(buf[1:], uint32(reason))
	return 5
}
