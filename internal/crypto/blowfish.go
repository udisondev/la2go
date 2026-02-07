package crypto

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/crypto/blowfish"
)

// BlowfishCipher wraps Blowfish ECB encryption/decryption for L2 protocol.
type BlowfishCipher struct {
	cipher *blowfish.Cipher
}

// NewBlowfishCipher creates a new Blowfish ECB cipher from the given key.
func NewBlowfishCipher(key []byte) (*BlowfishCipher, error) {
	c, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating blowfish cipher: %w", err)
	}
	return &BlowfishCipher{cipher: c}, nil
}

// Encrypt encrypts data in-place using Blowfish ECB mode.
// Data length must be a multiple of 8.
func (b *BlowfishCipher) Encrypt(data []byte, offset, size int) error {
	if size%8 != 0 {
		return fmt.Errorf("blowfish encrypt: size %d is not a multiple of 8", size)
	}
	if offset+size > len(data) {
		return fmt.Errorf("blowfish encrypt: offset %d + size %d exceeds data length %d", offset, size, len(data))
	}
	for i := offset; i < offset+size; i += 8 {
		b.cipher.Encrypt(data[i:i+8], data[i:i+8])
	}
	return nil
}

// Decrypt decrypts data in-place using Blowfish ECB mode.
// Data length must be a multiple of 8.
func (b *BlowfishCipher) Decrypt(data []byte, offset, size int) error {
	if size%8 != 0 {
		return fmt.Errorf("blowfish decrypt: size %d is not a multiple of 8", size)
	}
	if offset+size > len(data) {
		return fmt.Errorf("blowfish decrypt: offset %d + size %d exceeds data length %d", offset, size, len(data))
	}
	for i := offset; i < offset+size; i += 8 {
		b.cipher.Decrypt(data[i:i+8], data[i:i+8])
	}
	return nil
}

// AppendChecksum calculates and appends a 32-bit XOR checksum to the data.
// The data must have at least 4 extra bytes at the end for the checksum.
// Size must be a multiple of 4.
func AppendChecksum(data []byte, offset, size int) {
	var checksum uint32
	for i := offset; i < offset+size-4; i += 4 {
		checksum ^= binary.LittleEndian.Uint32(data[i:])
	}
	binary.LittleEndian.PutUint32(data[offset+size-4:], checksum)
}

// VerifyChecksum verifies that XOR of all 32-bit words in the range equals zero.
func VerifyChecksum(data []byte, offset, size int) bool {
	if size%4 != 0 || size <= 4 {
		return false
	}
	var checksum uint32
	for i := offset; i < offset+size; i += 4 {
		checksum ^= binary.LittleEndian.Uint32(data[i:])
	}
	return checksum == 0
}

// EncXORPass applies the pre-encryption XOR pass used for the Init packet.
// The key is a 32-bit integer used as the initial accumulator.
func EncXORPass(data []byte, offset, size int, key int32) {
	ecx := uint32(key)
	for i := offset; i < offset+size; i += 4 {
		edx := binary.LittleEndian.Uint32(data[i:])
		ecx += edx
		edx ^= ecx
		binary.LittleEndian.PutUint32(data[i:], edx)
	}
}
