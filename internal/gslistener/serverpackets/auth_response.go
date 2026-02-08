package serverpackets

import "unicode/utf16"

const (
	opcodeAuthResponse = 0x02
)

// AuthResponse [0x02] — LS → GS регистрация подтверждена
//
// Format:
//   [opcodeAuthResponse]                 // opcode
//   [serverID byte]                      // 1..127
//   [serverName UTF-16LE null-terminated]
//
// Returns: number of bytes written to buf
func AuthResponse(buf []byte, serverID byte, serverName string) int {
	pos := 0

	// Opcode
	buf[pos] = opcodeAuthResponse
	pos++

	// Server ID
	buf[pos] = serverID
	pos++

	// Server name (UTF-16LE null-terminated)
	nameRunes := utf16.Encode([]rune(serverName))
	for _, r := range nameRunes {
		buf[pos] = byte(r)
		buf[pos+1] = byte(r >> 8)
		pos += 2
	}

	// Null terminator
	buf[pos] = 0
	buf[pos+1] = 0
	pos += 2

	return pos
}
