package serverpackets

import "encoding/binary"

const GGAuthOpcode = 0x0B

// GGAuth writes the GGAuth response packet (opcode 0x0B) into buf.
// Returns the number of bytes written.
func GGAuth(buf []byte, sessionID int32) int {
	buf[0] = GGAuthOpcode
	binary.LittleEndian.PutUint32(buf[1:], uint32(sessionID))
	binary.LittleEndian.PutUint32(buf[5:], 0)
	binary.LittleEndian.PutUint32(buf[9:], 0)
	binary.LittleEndian.PutUint32(buf[13:], 0)
	binary.LittleEndian.PutUint32(buf[17:], 0)
	return 21
}
