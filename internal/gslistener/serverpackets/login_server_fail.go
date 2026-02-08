package serverpackets

const (
	opcodeLoginServerFail = 0x01
)

// LoginServerFail [0x01] — LS → GS регистрация отклонена
//
// Format:
//   [opcodeLoginServerFail]  // opcode
//   [reason byte]            // gameserver.ReasonXXX constants
//
// Returns: number of bytes written to buf
func LoginServerFail(buf []byte, reason byte) int {
	buf[0] = opcodeLoginServerFail
	buf[1] = reason
	return 2
}
