package serverpackets

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

func TestPlayerAuthResponse(t *testing.T) {
	tests := []struct {
		name    string
		account string
		success bool
	}{
		{
			name:    "success response",
			account: "testuser",
			success: true,
		},
		{
			name:    "failure response",
			account: "baduser",
			success: false,
		},
		{
			name:    "empty account",
			account: "",
			success: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 512)
			n := PlayerAuthResponse(buf, tt.account, tt.success)

			// Проверяем opcode
			if buf[0] != opcodePlayerAuthResponse {
				t.Errorf("opcode = 0x%02x, want 0x%02x", buf[0], opcodePlayerAuthResponse)
			}

			// Проверяем account
			pos := 1
			var decodedRunes []uint16
			for {
				if pos+2 > n {
					t.Fatal("unexpected end of data while reading account")
				}
				rune := binary.LittleEndian.Uint16(buf[pos:])
				pos += 2

				if rune == 0 {
					break
				}
				decodedRunes = append(decodedRunes, rune)
			}

			decoded := string(utf16.Decode(decodedRunes))
			if decoded != tt.account {
				t.Errorf("account = %q, want %q", decoded, tt.account)
			}

			// Проверяем result
			expectedResult := byte(0)
			if tt.success {
				expectedResult = 1
			}

			if buf[pos] != expectedResult {
				t.Errorf("result = %d, want %d", buf[pos], expectedResult)
			}

			// Проверяем общую длину
			expectedPos := pos + 1
			if n != expectedPos {
				t.Errorf("returned length = %d, want %d", n, expectedPos)
			}
		})
	}
}
