package serverpackets

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/udisondev/la2go/internal/constants"
)

func TestInitLS(t *testing.T) {
	buf := make([]byte, 256)
	modulus := make([]byte, 64)
	for i := range modulus {
		modulus[i] = byte(i)
	}

	n := InitLS(buf, 0x0106, modulus)

	// Verify size: opcode(1) + revision(4) + keySize(4) + modulus(64) = 73
	require.Equal(t, 73, n)

	// Verify opcode
	assert.Equal(t, byte(0x00), buf[0])

	// Verify revision (LE)
	revision := binary.LittleEndian.Uint32(buf[1:5])
	assert.Equal(t, uint32(0x0106), revision)

	// Verify key size (LE)
	keySize := binary.LittleEndian.Uint32(buf[5:9])
	assert.Equal(t, uint32(constants.RSA512ModulusSize), keySize)

	// Verify modulus
	assert.Equal(t, modulus, buf[9:9+constants.RSA512ModulusSize])
}

func TestInitLSModulusSize(t *testing.T) {
	buf := make([]byte, 256)

	// Test different modulus sizes
	tests := []struct {
		name      string
		modSize   int
		shouldPad bool
	}{
		{"exact 64 bytes", 64, false},
		{"less than 64 bytes", 60, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modulus := make([]byte, tt.modSize)
			n := InitLS(buf, 0x0106, modulus)

			// Always expect 73 bytes
			assert.Equal(t, 73, n)

			// Key size should always be constants.RSA512ModulusSize
			keySize := binary.LittleEndian.Uint32(buf[5:9])
			assert.Equal(t, uint32(constants.RSA512ModulusSize), keySize)
		})
	}
}
