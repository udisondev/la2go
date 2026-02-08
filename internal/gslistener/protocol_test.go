package gslistener

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/crypto"
)

func TestWritePacket(t *testing.T) {
	cipher, err := crypto.NewBlowfishCipher(crypto.DefaultGSBlowfishKey)
	require.NoError(t, err)

	// Prepare test payload
	buf := make([]byte, 1024)
	payload := []byte{0x01, 0x02, 0x03, 0x04, 0x05} // 5 bytes
	copy(buf[constants.PacketHeaderSize:], payload)

	var w bytes.Buffer
	err = WritePacket(&w, cipher, buf, len(payload))
	require.NoError(t, err)

	written := w.Bytes()

	// Verify length header (2 bytes LE)
	require.GreaterOrEqual(t, len(written), 2)
	totalLen := binary.LittleEndian.Uint16(written[0:2])
	assert.Equal(t, len(written), int(totalLen), "total length should match header")

	// Verify size is 2 (header) + encrypted data (aligned to 8)
	// payload=5, checksum=4, padding=7 → 16 bytes encrypted → 18 total
	assert.Equal(t, 18, len(written))
}

func TestReadPacket(t *testing.T) {
	cipher, err := crypto.NewBlowfishCipher(crypto.DefaultGSBlowfishKey)
	require.NoError(t, err)

	// First write a packet
	buf := make([]byte, 1024)
	payload := []byte{0xAA, 0xBB, 0xCC}
	copy(buf[constants.PacketHeaderSize:], payload)

	var w bytes.Buffer
	err = WritePacket(&w, cipher, buf, len(payload))
	require.NoError(t, err)

	// Now read it back
	readBuf := make([]byte, 1024)
	decrypted, err := ReadPacket(&w, cipher, readBuf)
	require.NoError(t, err)

	// Verify payload matches
	require.Equal(t, payload, decrypted[:len(payload)])
}

func TestPacketChecksum(t *testing.T) {
	buf := make([]byte, 64)
	payload := []byte{0x01, 0x02, 0x03, 0x04}
	copy(buf[constants.PacketHeaderSize:], payload)

	// Append checksum
	dataLen := len(payload) + constants.PacketChecksumSize // payload + checksum space
	crypto.AppendChecksum(buf, constants.PacketHeaderSize, dataLen)

	// Verify checksum
	valid := crypto.VerifyChecksum(buf, constants.PacketHeaderSize, dataLen)
	assert.True(t, valid, "checksum should be valid")

	// Corrupt data
	buf[2] ^= 0xFF

	// Verify checksum fails
	valid = crypto.VerifyChecksum(buf, constants.PacketHeaderSize, dataLen)
	assert.False(t, valid, "checksum should be invalid after corruption")
}

func TestPacketPadding(t *testing.T) {
	cipher, err := crypto.NewBlowfishCipher(crypto.DefaultGSBlowfishKey)
	require.NoError(t, err)

	tests := []struct {
		name        string
		payloadSize int
		expectPad   int
	}{
		{"no padding needed", 4, 0},      // 4 + 4 = 8 (% 8 == 0)
		{"1 byte payload", 1, 3},         // 1 + 4 = 5 → pad 3 → 8
		{"5 byte payload", 5, 7},         // 5 + 4 = 9 → pad 7 → 16
		{"12 byte payload", 12, 0},       // 12 + 4 = 16 (% 8 == 0)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 1024)
			payload := make([]byte, tt.payloadSize)
			copy(buf[constants.PacketHeaderSize:], payload)

			var w bytes.Buffer
			err := WritePacket(&w, cipher, buf, tt.payloadSize)
			require.NoError(t, err)

			written := w.Bytes()
			// length = constants.PacketHeaderSize + payloadSize + constants.PacketChecksumSize + padding
			expectedLen := constants.PacketHeaderSize + tt.payloadSize + constants.PacketChecksumSize + tt.expectPad
			assert.Equal(t, expectedLen, len(written), "packet size should include correct padding")
		})
	}
}

func TestPacketLengthHeader(t *testing.T) {
	cipher, err := crypto.NewBlowfishCipher(crypto.DefaultGSBlowfishKey)
	require.NoError(t, err)

	buf := make([]byte, 1024)
	payload := []byte{0x01, 0x02, 0x03}
	copy(buf[constants.PacketHeaderSize:], payload)

	var w bytes.Buffer
	err = WritePacket(&w, cipher, buf, len(payload))
	require.NoError(t, err)

	written := w.Bytes()

	// Read length header (LE)
	totalLen := binary.LittleEndian.Uint16(written[0:2])

	// Verify LE encoding: should be total packet size
	assert.Equal(t, uint16(len(written)), totalLen)
}

func TestReadWritePacketRoundtrip(t *testing.T) {
	cipher, err := crypto.NewBlowfishCipher(crypto.DefaultGSBlowfishKey)
	require.NoError(t, err)

	testCases := [][]byte{
		{0x00},                           // 1 byte
		{0x01, 0x02},                     // 2 bytes
		{0xAA, 0xBB, 0xCC, 0xDD, 0xEE},   // 5 bytes
		make([]byte, 100),                // 100 bytes
	}

	for i, payload := range testCases {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			// Write
			buf := make([]byte, 1024)
			copy(buf[constants.PacketHeaderSize:], payload)

			var w bytes.Buffer
			err := WritePacket(&w, cipher, buf, len(payload))
			require.NoError(t, err)

			// Read
			readBuf := make([]byte, 1024)
			decrypted, err := ReadPacket(&w, cipher, readBuf)
			require.NoError(t, err)

			// Verify
			require.Equal(t, payload, decrypted[:len(payload)])
		})
	}
}
