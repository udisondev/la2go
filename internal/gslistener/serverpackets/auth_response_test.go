package serverpackets

import (
	"testing"
	"unicode/utf16"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthResponse(t *testing.T) {
	buf := make([]byte, 256)

	tests := []struct {
		name       string
		serverID   byte
		serverName string
	}{
		{"server 1", 1, "Bartz"},
		{"server 2", 2, "Sieghardt"},
		{"empty name", 3, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := AuthResponse(buf, tt.serverID, tt.serverName)

			// Verify opcode
			assert.Equal(t, byte(0x02), buf[0])

			// Verify serverID
			assert.Equal(t, tt.serverID, buf[1])

			// Verify server name (UTF-16LE null-terminated)
			nameRunes := utf16.Encode([]rune(tt.serverName))
			nameRunes = append(nameRunes, 0) // null terminator

			expectedSize := 2 + len(nameRunes)*2
			require.Equal(t, expectedSize, n)

			// Decode and verify
			decodedRunes := make([]uint16, len(nameRunes))
			for i := range nameRunes {
				decodedRunes[i] = uint16(buf[2+i*2]) | uint16(buf[2+i*2+1])<<8
			}
			decodedStr := string(utf16.Decode(decodedRunes[:len(decodedRunes)-1]))
			assert.Equal(t, tt.serverName, decodedStr)
		})
	}
}
