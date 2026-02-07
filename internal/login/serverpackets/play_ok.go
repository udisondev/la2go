package serverpackets

import "encoding/binary"

const PlayOkOpcode = 0x07

// PlayOk writes the PlayOk packet (opcode 0x07) into buf.
// Returns the number of bytes written.
func PlayOk(buf []byte, playOkID1, playOkID2 int32) int {
	buf[0] = PlayOkOpcode
	binary.LittleEndian.PutUint32(buf[1:], uint32(playOkID1))
	binary.LittleEndian.PutUint32(buf[5:], uint32(playOkID2))
	return 9
}
