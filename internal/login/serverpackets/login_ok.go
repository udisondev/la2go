package serverpackets

import "encoding/binary"

const LoginOkOpcode = 0x03

// LoginOk writes the LoginOk packet (opcode 0x03) into buf.
// Returns the number of bytes written.
func LoginOk(buf []byte, loginOkID1, loginOkID2 int32) int {
	buf[0] = LoginOkOpcode
	binary.LittleEndian.PutUint32(buf[1:], uint32(loginOkID1))
	binary.LittleEndian.PutUint32(buf[5:], uint32(loginOkID2))
	binary.LittleEndian.PutUint32(buf[9:], 0)
	binary.LittleEndian.PutUint32(buf[13:], 0)
	binary.LittleEndian.PutUint32(buf[17:], 0x000003EA)
	binary.LittleEndian.PutUint32(buf[21:], 0)
	binary.LittleEndian.PutUint32(buf[25:], 0)
	binary.LittleEndian.PutUint32(buf[29:], 0)
	clear(buf[33:49]) // 16 zero bytes padding
	return 49
}
