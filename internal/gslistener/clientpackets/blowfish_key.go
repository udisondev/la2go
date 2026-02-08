package clientpackets

import (
	"fmt"
)

// BlowFishKey [0x00] — GS → LS зашифрованный Blowfish ключ
//
// Format:
//   [opcode 0x00]
//   [encryptedKey byte[64]] // RSA-512 зашифрованный Blowfish ключ (фиксированный размер 64 байта)
type BlowFishKey struct {
	EncryptedKey []byte
}

// Parse парсит пакет BlowFishKey из body (без opcode).
func (p *BlowFishKey) Parse(body []byte) error {
	// BlowFishKey packet format (после удаления opcode):
	// encrypted_key (64 bytes для RSA-512)
	const expectedSize = 64

	if len(body) < expectedSize {
		return fmt.Errorf("BlowFishKey packet too short: got %d, want %d", len(body), expectedSize)
	}

	p.EncryptedKey = make([]byte, expectedSize)
	copy(p.EncryptedKey, body[:expectedSize])
	return nil
}
