package crypto

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/crypto/blowfish"

	"github.com/udisondev/la2go/internal/constants"
)

// DefaultGSBlowfishKey — статический ключ Blowfish для GameServer↔LoginServer
// до обмена динамическим ключом. Соответствует ключу из Java L2J:
// "_;v.]05-31!|+-%xT!^[$\x00"
var DefaultGSBlowfishKey = []byte{
	0x5F, 0x3B, 0x76, 0x2E, 0x5D, 0x30, 0x35, 0x2D,
	0x33, 0x31, 0x21, 0x7C, 0x2B, 0x2D, 0x25, 0x78,
	0x54, 0x21, 0x5E, 0x5B, 0x24, 0x00,
}

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
	if size%constants.BlowfishBlockSize != 0 {
		return fmt.Errorf("blowfish encrypt: size %d is not a multiple of %d", size, constants.BlowfishBlockSize)
	}
	if offset+size > len(data) {
		return fmt.Errorf("blowfish encrypt: offset %d + size %d exceeds data length %d", offset, size, len(data))
	}
	for i := offset; i < offset+size; i += constants.BlowfishBlockSize {
		b.cipher.Encrypt(data[i:i+constants.BlowfishBlockSize], data[i:i+constants.BlowfishBlockSize])
	}
	return nil
}

// Decrypt decrypts data in-place using Blowfish ECB mode.
// Data length must be a multiple of 8.
func (b *BlowfishCipher) Decrypt(data []byte, offset, size int) error {
	if size%constants.BlowfishBlockSize != 0 {
		return fmt.Errorf("blowfish decrypt: size %d is not a multiple of %d", size, constants.BlowfishBlockSize)
	}
	if offset+size > len(data) {
		return fmt.Errorf("blowfish decrypt: offset %d + size %d exceeds data length %d", offset, size, len(data))
	}
	for i := offset; i < offset+size; i += constants.BlowfishBlockSize {
		b.cipher.Decrypt(data[i:i+constants.BlowfishBlockSize], data[i:i+constants.BlowfishBlockSize])
	}
	return nil
}

// AppendChecksum calculates and appends a 32-bit XOR checksum to the data.
// The data must have at least 4 extra bytes at the end for the checksum.
// Size must be a multiple of 4.
func AppendChecksum(data []byte, offset, size int) {
	var checksum uint32
	for i := offset; i < offset+size-constants.PacketChecksumSize; i += constants.PacketChecksumSize {
		checksum ^= binary.LittleEndian.Uint32(data[i:])
	}
	binary.LittleEndian.PutUint32(data[offset+size-constants.PacketChecksumSize:], checksum)
}

// VerifyChecksum verifies that XOR of all 32-bit words in the range equals zero.
func VerifyChecksum(data []byte, offset, size int) bool {
	if size%constants.PacketChecksumSize != 0 || size <= constants.PacketChecksumSize {
		return false
	}
	var checksum uint32
	for i := offset; i < offset+size; i += constants.PacketChecksumSize {
		checksum ^= binary.LittleEndian.Uint32(data[i:])
	}
	return checksum == 0
}

// EncXORPass applies the pre-encryption XOR pass used for the Init packet.
// The key is a 32-bit integer used as the initial accumulator.
// Algorithm matches L2J NewCrypt.encXORPass: skips first 4 bytes, stops 8 bytes before end,
// and writes the final accumulated key in the last 4 bytes.
func EncXORPass(data []byte, offset, size int, key int32) {
	ecx := uint32(key)
	stop := offset + size - constants.XOREncryptStopOffset
	pos := offset + constants.XOREncryptSkipBytes // Skip first 4 bytes (sessionId is not encrypted)

	for pos < stop {
		edx := binary.LittleEndian.Uint32(data[pos:])
		ecx += edx
		edx ^= ecx
		binary.LittleEndian.PutUint32(data[pos:], edx)
		pos += constants.PacketChecksumSize
	}

	// Write final accumulated key in the last 4 bytes
	binary.LittleEndian.PutUint32(data[stop:], ecx)
}

// DecXORPass applies the reverse operation of EncXORPass to decrypt the Init packet.
// Reads the final accumulated key from the last 4 bytes and reverses the XOR process.
// Algorithm is the reverse of EncXORPass: processes from end to start.
func DecXORPass(data []byte, offset, size int) {
	stop := offset + size - constants.XOREncryptStopOffset
	pos := offset + constants.XOREncryptSkipBytes // Skip first 4 bytes

	// Read the final accumulated key from the last 4 bytes
	ecx := binary.LittleEndian.Uint32(data[stop:])

	// Process from end to start (reverse order)
	for i := stop - constants.PacketChecksumSize; i >= pos; i -= constants.PacketChecksumSize {
		edx := binary.LittleEndian.Uint32(data[i:])
		edx ^= ecx
		binary.LittleEndian.PutUint32(data[i:], edx)
		ecx -= edx
	}
}
