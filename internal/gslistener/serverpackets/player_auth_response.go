package serverpackets

import "unicode/utf16"

const (
	opcodePlayerAuthResponse = 0x03
)

// PlayerAuthResponse [0x03] — LS → GS результат валидации сессии игрока
//
// Format:
//   [opcodePlayerAuthResponse]           // opcode
//   [account UTF-16LE null-terminated]
//   [result byte]                        // 1 = success, 0 = failure
//
// Returns: number of bytes written to buf
func PlayerAuthResponse(buf []byte, account string, success bool) int {
	pos := 0

	// Opcode
	buf[pos] = opcodePlayerAuthResponse
	pos++

	// Account (UTF-16LE null-terminated)
	accountRunes := utf16.Encode([]rune(account))
	for _, r := range accountRunes {
		buf[pos] = byte(r)
		buf[pos+1] = byte(r >> 8)
		pos += 2
	}

	// Null terminator
	buf[pos] = 0
	buf[pos+1] = 0
	pos += 2

	// Result
	if success {
		buf[pos] = 1
	} else {
		buf[pos] = 0
	}
	pos++

	return pos
}
