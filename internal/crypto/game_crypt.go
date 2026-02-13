package crypto

import (
	"encoding/binary"
	"sync/atomic"
)

// GameCrypt implements the XOR rolling cipher used by the L2 GameServer protocol.
// After the Init handshake, ALL GameServer packets are encrypted/decrypted with this cipher
// (NOT Blowfish — Blowfish is only used for LoginServer).
//
// Algorithm (from L2J Mobius Encryption.java):
//   - Encrypt: encrypted[i] = raw[i] ^ outKey[i & 0x0F] ^ encrypted[i-1]
//   - Decrypt: decrypted[i] = encrypted[i] ^ inKey[i & 0x0F] ^ encrypted[i-1]
//   - After each call, key bytes [8:12] (LE uint32) are incremented by packet size.
//   - The FIRST encrypt call is skipped (Init packet sent unencrypted).
type GameCrypt struct {
	inKey     [16]byte
	outKey    [16]byte
	isEnabled atomic.Bool
}

// NewGameCrypt creates a new GameCrypt instance (disabled until SetKey + first Encrypt).
func NewGameCrypt() *GameCrypt {
	return &GameCrypt{}
}

// SetKey initializes both inKey and outKey from the same 16-byte key.
// Must be called before Encrypt/Decrypt.
func (gc *GameCrypt) SetKey(key []byte) {
	copy(gc.inKey[:], key[:16])
	copy(gc.outKey[:], key[:16])
}

// Encrypt encrypts data in-place using the XOR rolling cipher.
// The FIRST call is skipped (sets isEnabled=true and returns) because
// the Init packet is sent unencrypted. All subsequent calls encrypt.
func (gc *GameCrypt) Encrypt(data []byte) {
	if !gc.isEnabled.Swap(true) {
		return
	}

	var prev byte
	for i := range len(data) {
		prev = data[i] ^ gc.outKey[i&0x0F] ^ prev
		data[i] = prev
	}

	shiftKey(gc.outKey[:], len(data))
}

// Decrypt decrypts data in-place using the XOR rolling cipher.
// If encryption is not yet enabled (before first Encrypt call), this is a no-op.
func (gc *GameCrypt) Decrypt(data []byte) {
	if !gc.isEnabled.Load() {
		return
	}

	var xor byte
	for i := range len(data) {
		encrypted := data[i]
		data[i] = encrypted ^ gc.inKey[i&0x0F] ^ xor
		xor = encrypted
	}

	shiftKey(gc.inKey[:], len(data))
}

// IsEnabled returns true if the cipher has been activated (first Encrypt was called).
func (gc *GameCrypt) IsEnabled() bool {
	return gc.isEnabled.Load()
}

// shiftKey increments key bytes [8:12] (interpreted as LE uint32) by size.
// This causes the key to evolve after each packet, preventing replay attacks.
func shiftKey(key []byte, size int) {
	old := binary.LittleEndian.Uint32(key[8:12])
	old += uint32(size)
	binary.LittleEndian.PutUint32(key[8:12], old)
}
