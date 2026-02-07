package serverpackets

import "encoding/binary"

const InitOpcode = 0x00

// Init writes the Init packet (opcode 0x00) into buf.
// Contains: sessionId, protocol version, scrambled RSA modulus, GG constants, blowfish key.
// Returns the number of bytes written.
func Init(buf []byte, sessionID int32, scrambledModulus, blowfishKey []byte) int {
	buf[0] = InitOpcode
	binary.LittleEndian.PutUint32(buf[1:], uint32(sessionID))
	binary.LittleEndian.PutUint32(buf[5:], 0x0000C621) // protocol revision

	copy(buf[9:], scrambledModulus) // 128 bytes
	clear(buf[137:153])            // 16 bytes padding after modulus

	// GG constants
	binary.LittleEndian.PutUint32(buf[153:], 0x29DD954E)
	binary.LittleEndian.PutUint32(buf[157:], 0x77C39CFC)
	binary.LittleEndian.PutUint32(buf[161:], 0x97ADB620) // -0x685249E0 as uint32
	binary.LittleEndian.PutUint32(buf[165:], 0x07BDE0F7)

	copy(buf[169:], blowfishKey) // 16 bytes
	buf[185] = 0x00              // null terminator

	return 186
}
