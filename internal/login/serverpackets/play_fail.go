package serverpackets

const PlayFailOpcode = 0x06

// PlayFail writes the PlayFail packet (opcode 0x06) into buf.
// Returns the number of bytes written.
func PlayFail(buf []byte, reason byte) int {
	buf[0] = PlayFailOpcode
	buf[1] = reason
	return 2
}
