package serverpackets

import (
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/crypto"
)

// TestInitEncryptionDecryption проверяет полный flow шифрования/дешифрования Init пакета
func TestInitEncryptionDecryption(t *testing.T) {
	sessionID := int32(0x12345678)
	scrambledModulus := make([]byte, 128)
	for i := range scrambledModulus {
		scrambledModulus[i] = byte(i)
	}
	blowfishKey := []byte{
		0x04, 0xa1, 0xc3, 0x42, 0xad, 0xaa, 0xf2, 0x34,
		0x30, 0x78, 0x9f, 0x61, 0xb8, 0x92, 0x53, 0x32,
	}

	// 1. Создаём plaintext Init пакет
	buf := make([]byte, 256)
	plaintextSize := Init(buf[2:], sessionID, scrambledModulus, blowfishKey)
	t.Logf("Plaintext size: %d bytes", plaintextSize)

	if plaintextSize != 170 {
		t.Fatalf("expected plaintext size 170, got %d", plaintextSize)
	}

	// Сохраняем оригинальный blowfish key для сравнения
	originalKey := make([]byte, 16)
	copy(originalKey, buf[2+153:2+153+16])
	t.Logf("Original blowfish key: %x", originalKey)

	// 2. Шифруем как на сервере (encXORPass + static Blowfish)
	enc, err := crypto.NewLoginEncryption(blowfishKey)
	if err != nil {
		t.Fatalf("failed to create login encryption: %v", err)
	}

	encSize, err := enc.EncryptPacket(buf, 2, plaintextSize)
	if err != nil {
		t.Fatalf("failed to encrypt packet: %v", err)
	}
	t.Logf("Encrypted size: %d bytes", encSize)

	// encryptedSize должен быть 192 по формуле:
	// 170 + 8 (static) = 178
	// 178 + padding to 8 = 184 (padding=6)
	// 184 + 8 (final) = 192
	if encSize != 192 {
		t.Fatalf("expected encrypted size 192, got %d", encSize)
	}

	// 3. Симулируем получение на клиенте и расшифровку
	encrypted := make([]byte, encSize)
	copy(encrypted, buf[2:2+encSize])

	// Blowfish decrypt
	staticCipher, err := crypto.NewBlowfishCipher(crypto.StaticBlowfishKey)
	if err != nil {
		t.Fatalf("failed to create static cipher: %v", err)
	}

	if err := staticCipher.Decrypt(encrypted, 0, encSize); err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	// decXORPass
	crypto.DecXORPass(encrypted, 0, encSize)

	// 4. Проверяем что данные восстановились корректно
	if encrypted[0] != InitOpcode {
		t.Errorf("opcode = 0x%02X, expected 0x00", encrypted[0])
	}

	gotSessionID := int32(binary.LittleEndian.Uint32(encrypted[1:5]))
	if gotSessionID != sessionID {
		t.Errorf("sessionID = 0x%08X, expected 0x%08X", gotSessionID, sessionID)
	}

	// Проверяем blowfish key на offset 153
	recoveredKey := encrypted[153 : 153+16]
	t.Logf("Recovered blowfish key: %x", recoveredKey)

	for i := 0; i < 16; i++ {
		if recoveredKey[i] != originalKey[i] {
			t.Errorf("blowfishKey[%d] = 0x%02X, expected 0x%02X", i, recoveredKey[i], originalKey[i])
		}
	}

	t.Log("✅ Init packet encryption/decryption flow validated successfully")
}
