package testutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"unicode/utf16"
)

// AssertPacketOpcode проверяет, что первый байт пакета соответствует ожидаемому opcode.
func AssertPacketOpcode(t testing.TB, expected byte, packet []byte) {
	t.Helper()

	if len(packet) == 0 {
		t.Fatalf("packet is empty, expected opcode 0x%02X", expected)
	}

	actual := packet[0]
	if actual != expected {
		t.Fatalf("packet opcode mismatch: expected 0x%02X, got 0x%02X", expected, actual)
	}
}

// AssertInt32LE проверяет, что int32 значение в пакете (little-endian) соответствует ожидаемому.
func AssertInt32LE(t testing.TB, expected int32, packet []byte, offset int) {
	t.Helper()

	if len(packet) < offset+4 {
		t.Fatalf("packet too short: need %d bytes for int32 at offset %d, got %d",
			offset+4, offset, len(packet))
	}

	actual := int32(binary.LittleEndian.Uint32(packet[offset:]))
	if actual != expected {
		t.Fatalf("int32 mismatch at offset %d: expected %d, got %d", offset, expected, actual)
	}
}

// AssertInt64LE проверяет, что int64 значение в пакете (little-endian) соответствует ожидаемому.
func AssertInt64LE(t testing.TB, expected int64, packet []byte, offset int) {
	t.Helper()

	if len(packet) < offset+8 {
		t.Fatalf("packet too short: need %d bytes for int64 at offset %d, got %d",
			offset+8, offset, len(packet))
	}

	actual := int64(binary.LittleEndian.Uint64(packet[offset:]))
	if actual != expected {
		t.Fatalf("int64 mismatch at offset %d: expected %d, got %d", offset, expected, actual)
	}
}

// AssertByteAtOffset проверяет, что байт в пакете соответствует ожидаемому.
func AssertByteAtOffset(t testing.TB, expected byte, packet []byte, offset int) {
	t.Helper()

	if len(packet) <= offset {
		t.Fatalf("packet too short: need %d bytes, got %d", offset+1, len(packet))
	}

	actual := packet[offset]
	if actual != expected {
		t.Fatalf("byte mismatch at offset %d: expected 0x%02X, got 0x%02X", offset, expected, actual)
	}
}

// AssertUTF16String проверяет, что UTF-16LE строка в пакете соответствует ожидаемой.
// Строка должна заканчиваться нулевым терминатором (0x00 0x00).
func AssertUTF16String(t testing.TB, expected string, packet []byte, offset int) {
	t.Helper()

	// Ищем null terminator
	nullIdx := -1
	for i := offset; i < len(packet)-1; i += 2 {
		if packet[i] == 0 && packet[i+1] == 0 {
			nullIdx = i
			break
		}
	}

	if nullIdx == -1 {
		t.Fatalf("UTF-16 string at offset %d has no null terminator", offset)
	}

	// Декодируем UTF-16LE
	utf16Data := packet[offset:nullIdx]
	if len(utf16Data)%2 != 0 {
		t.Fatalf("UTF-16 string at offset %d has odd length: %d", offset, len(utf16Data))
	}

	runes := make([]uint16, len(utf16Data)/2)
	for i := range runes {
		runes[i] = binary.LittleEndian.Uint16(utf16Data[i*2:])
	}

	actual := string(utf16.Decode(runes))
	if actual != expected {
		t.Fatalf("UTF-16 string mismatch at offset %d: expected %q, got %q", offset, expected, actual)
	}
}

// AssertBytesEqual проверяет, что два байтовых слайса равны.
func AssertBytesEqual(t testing.TB, expected, actual []byte, msg string) {
	t.Helper()

	if !bytes.Equal(expected, actual) {
		t.Fatalf("%s: bytes mismatch\nexpected: %v\nactual:   %v", msg, expected, actual)
	}
}

// AssertPacketLength проверяет, что длина пакета соответствует ожидаемой.
func AssertPacketLength(t testing.TB, expected int, packet []byte) {
	t.Helper()

	actual := len(packet)
	if actual != expected {
		t.Fatalf("packet length mismatch: expected %d bytes, got %d bytes", expected, actual)
	}
}

// AssertPacketMinLength проверяет, что пакет не короче минимальной длины.
func AssertPacketMinLength(t testing.TB, minLength int, packet []byte) {
	t.Helper()

	actual := len(packet)
	if actual < minLength {
		t.Fatalf("packet too short: expected at least %d bytes, got %d bytes", minLength, actual)
	}
}

// DumpPacket возвращает hex dump пакета для отладки.
func DumpPacket(packet []byte) string {
	var buf bytes.Buffer
	for i := 0; i < len(packet); i += 16 {
		end := i + 16
		if end > len(packet) {
			end = len(packet)
		}
		chunk := packet[i:end]

		// Offset
		fmt.Fprintf(&buf, "%04x  ", i)

		// Hex
		for j, b := range chunk {
			if j == 8 {
				buf.WriteString(" ")
			}
			fmt.Fprintf(&buf, "%02x ", b)
		}

		// Padding
		for j := len(chunk); j < 16; j++ {
			if j == 8 {
				buf.WriteString(" ")
			}
			buf.WriteString("   ")
		}

		// ASCII
		buf.WriteString(" |")
		for _, b := range chunk {
			if b >= 32 && b <= 126 {
				buf.WriteByte(b)
			} else {
				buf.WriteByte('.')
			}
		}
		buf.WriteString("|\n")
	}
	return buf.String()
}
