package crypto

import (
	"fmt"
	"math/rand/v2"
)

// StaticBlowfishKey is the key hardcoded in the L2 client for the first Init packet.
var StaticBlowfishKey = []byte{
	0x6b, 0x60, 0xcb, 0x5b,
	0x82, 0xce, 0x90, 0xb1,
	0xcc, 0x2b, 0x6c, 0x55,
	0x6c, 0x6c, 0x6c, 0x6c,
}

// LoginEncryption handles Blowfish encryption/decryption for login protocol.
// The first outgoing packet (Init) uses the static key + encXORPass.
// All subsequent packets use the dynamic key + checksum.
type LoginEncryption struct {
	staticCipher  *BlowfishCipher
	dynamicCipher *BlowfishCipher
	firstPacket   bool
}

// NewLoginEncryption creates a LoginEncryption with the given dynamic Blowfish key.
func NewLoginEncryption(dynamicKey []byte) (*LoginEncryption, error) {
	sc, err := NewBlowfishCipher(StaticBlowfishKey)
	if err != nil {
		return nil, fmt.Errorf("creating static blowfish cipher: %w", err)
	}
	dc, err := NewBlowfishCipher(dynamicKey)
	if err != nil {
		return nil, fmt.Errorf("creating dynamic blowfish cipher: %w", err)
	}
	return &LoginEncryption{
		staticCipher:  sc,
		dynamicCipher: dc,
		firstPacket:   true,
	}, nil
}

// EncryptPacket encrypts an outgoing packet in-place.
// For the first packet (Init): encXORPass + static Blowfish.
// For subsequent packets: appendChecksum + dynamic Blowfish.
// Returns the total size to send (may include padding).
func (le *LoginEncryption) EncryptPacket(data []byte, offset, size int) (int, error) {
	// Pad to multiple of 8
	padded := size
	if padded%8 != 0 {
		padded += 8 - (padded % 8)
	}
	// Ensure we have enough space for checksum (4 bytes) and padding
	needed := padded + 8 // extra 8 bytes for checksum/padding room
	if offset+needed > len(data) {
		return 0, fmt.Errorf("encrypt packet: buffer too small (need %d, have %d)", offset+needed, len(data))
	}

	if le.firstPacket {
		le.firstPacket = false
		// encXORPass with a random key, then encrypt with static Blowfish
		// Formula from Java LoginEncryption.encryptedSize():
		// dataSize += 8 (for _static=true)
		// dataSize += 8 - (dataSize % 8) (padding to multiple of 8)
		// dataSize += 8 (final 8 bytes)
		xorKey := rand.Int32()
		encSize := size + 8 // +8 for static
		if encSize%8 != 0 {
			encSize += 8 - (encSize % 8)
		}
		encSize += 8 // final 8 bytes
		EncXORPass(data, offset, encSize, xorKey)
		if err := le.staticCipher.Encrypt(data, offset, encSize); err != nil {
			return 0, fmt.Errorf("encrypting init packet: %w", err)
		}
		return encSize, nil
	}

	// Subsequent packets: checksum + dynamic Blowfish
	checksumSize := size + 4 // 4 bytes for checksum
	if checksumSize%8 != 0 {
		checksumSize += 8 - (checksumSize % 8)
	}
	// Zero out padding bytes
	for i := offset + size; i < offset+checksumSize; i++ {
		data[i] = 0
	}
	AppendChecksum(data, offset, checksumSize)
	if err := le.dynamicCipher.Encrypt(data, offset, checksumSize); err != nil {
		return 0, fmt.Errorf("encrypting packet: %w", err)
	}
	return checksumSize, nil
}

// DecryptPacket decrypts an incoming packet in-place using the dynamic Blowfish key.
// Returns true if the checksum is valid.
func (le *LoginEncryption) DecryptPacket(data []byte, offset, size int) (bool, error) {
	// Incoming packets are always encrypted with the dynamic key
	if size%8 != 0 {
		return false, fmt.Errorf("decrypt packet: size %d is not multiple of 8", size)
	}
	if err := le.dynamicCipher.Decrypt(data, offset, size); err != nil {
		return false, fmt.Errorf("decrypting packet: %w", err)
	}
	return VerifyChecksum(data, offset, size), nil
}

// EncryptPacketClient encrypts an outgoing packet from client to server.
// For clients, ALL packets use: appendChecksum + dynamic Blowfish (no encXORPass, no firstPacket logic).
// Returns the total size to send (includes padding to multiple of 8).
func (le *LoginEncryption) EncryptPacketClient(data []byte, offset, size int) (int, error) {
	// Add 4 bytes for checksum, then pad to multiple of 8
	checksumSize := size + 4
	if checksumSize%8 != 0 {
		checksumSize += 8 - (checksumSize % 8)
	}

	// Ensure we have enough space
	if offset+checksumSize > len(data) {
		return 0, fmt.Errorf("encrypt packet client: buffer too small (need %d, have %d)", offset+checksumSize, len(data))
	}

	// Zero out padding bytes
	for i := offset + size; i < offset+checksumSize; i++ {
		data[i] = 0
	}

	// Append checksum (XOR of all 32-bit words)
	AppendChecksum(data, offset, checksumSize)

	// Encrypt with dynamic Blowfish
	if err := le.dynamicCipher.Encrypt(data, offset, checksumSize); err != nil {
		return 0, fmt.Errorf("encrypting client packet: %w", err)
	}

	return checksumSize, nil
}
