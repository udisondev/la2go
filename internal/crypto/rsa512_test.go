package crypto

import (
	"crypto/rand"
	"math/big"
	"testing"
)

func TestGenerateRSAKeyPair512(t *testing.T) {
	kp, err := GenerateRSAKeyPair512()
	if err != nil {
		t.Fatalf("Failed to generate RSA-512 key pair: %v", err)
	}

	// Проверяем размер ключа (512 бит = 64 байта)
	if kp.PrivateKey.N.BitLen() != 512 {
		t.Errorf("Expected 512-bit key, got %d bits", kp.PrivateKey.N.BitLen())
	}

	// Проверяем размер modulus (должно быть 64 байта)
	if len(kp.ScrambledModulus) != 64 {
		t.Errorf("Expected modulus length 64 bytes, got %d", len(kp.ScrambledModulus))
	}

	// Проверяем, что exponent = 65537 (F4)
	if kp.PrivateKey.E != 65537 {
		t.Errorf("Expected exponent 65537, got %d", kp.PrivateKey.E)
	}
}

func TestRSA512_EncryptDecrypt(t *testing.T) {
	kp, err := GenerateRSAKeyPair512()
	if err != nil {
		t.Fatalf("Failed to generate RSA-512 key pair: %v", err)
	}

	// Генерируем случайные данные < N (для raw RSA plaintext должен быть < modulus)
	// Используем 40 байт (размер Blowfish ключа для GS↔LS)
	plaintext := make([]byte, 40)
	if _, err := rand.Read(plaintext); err != nil {
		t.Fatalf("Failed to generate random data: %v", err)
	}

	// Pad до 64 байт для шифрования
	padded := make([]byte, 64)
	copy(padded[64-len(plaintext):], plaintext)

	// Encrypt (raw RSA operation: m^e mod n)
	m := new(big.Int).SetBytes(padded)
	// Убедимся что m < N
	if m.Cmp(kp.PrivateKey.N) >= 0 {
		t.Fatal("Plaintext >= N, invalid for RSA")
	}
	c := new(big.Int).Exp(m, big.NewInt(int64(kp.PrivateKey.E)), kp.PrivateKey.N)
	ciphertext := c.Bytes()

	// Decrypt (raw RSA operation: c^d mod n)
	decrypted, err := RSADecryptNoPadding(kp.PrivateKey, padTo64(ciphertext))
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	// Проверяем, что расшифрованное совпадает с оригиналом
	// (старшие байты могут быть нулями, поэтому сравниваем big.Int)
	originalInt := new(big.Int).SetBytes(padded)
	decryptedInt := new(big.Int).SetBytes(decrypted)

	if originalInt.Cmp(decryptedInt) != 0 {
		t.Error("Decrypted data does not match original")
	}
}

func TestRSA512_ModulusNotScrambled(t *testing.T) {
	kp, err := GenerateRSAKeyPair512()
	if err != nil {
		t.Fatalf("Failed to generate RSA-512 key pair: %v", err)
	}

	// Проверяем, что ScrambledModulus == raw modulus (без scrambling)
	rawModulus := kp.PrivateKey.N.Bytes()
	if len(rawModulus) < 64 {
		padded := make([]byte, 64)
		copy(padded[64-len(rawModulus):], rawModulus)
		rawModulus = padded
	}

	if len(rawModulus) > 64 && rawModulus[0] == 0 {
		rawModulus = rawModulus[1:]
	}

	// ScrambledModulus должен быть равен raw modulus (без scrambling)
	for i := range kp.ScrambledModulus {
		if kp.ScrambledModulus[i] != rawModulus[i] {
			t.Error("ScrambledModulus is not equal to raw modulus (expected no scrambling for GS)")
			break
		}
	}
}

func TestRSADecryptNoPadding_512(t *testing.T) {
	kp, err := GenerateRSAKeyPair512()
	if err != nil {
		t.Fatalf("Failed to generate RSA-512 key pair: %v", err)
	}

	// Создаём тестовое сообщение (40 байт — размер Blowfish ключа)
	message := make([]byte, 40)
	for i := range message {
		message[i] = byte(i + 1)
	}

	// Pad до 64 байт
	padded := make([]byte, 64)
	copy(padded[64-len(message):], message)

	// Encrypt: m^e mod n
	m := new(big.Int).SetBytes(padded)
	c := new(big.Int).Exp(m, big.NewInt(int64(kp.PrivateKey.E)), kp.PrivateKey.N)
	ciphertext := c.Bytes()

	// Pad ciphertext to 64 bytes if needed
	if len(ciphertext) < 64 {
		paddedCipher := make([]byte, 64)
		copy(paddedCipher[64-len(ciphertext):], ciphertext)
		ciphertext = paddedCipher
	}

	// Decrypt
	decrypted, err := RSADecryptNoPadding(kp.PrivateKey, ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	// Проверяем, что расшифрованное совпадает с padded
	if len(decrypted) != 64 {
		t.Errorf("Expected decrypted length 64, got %d", len(decrypted))
	}

	// Сравниваем как big.Int (игнорируем leading zeros)
	originalInt := new(big.Int).SetBytes(padded)
	decryptedInt := new(big.Int).SetBytes(decrypted)

	if originalInt.Cmp(decryptedInt) != 0 {
		t.Error("Decrypted data does not match original padded message")
	}
}

// Helper: pad bytes to 64 bytes
func padTo64(data []byte) []byte {
	if len(data) >= 64 {
		if len(data) == 65 && data[0] == 0 {
			return data[1:]
		}
		return data[:64]
	}
	padded := make([]byte, 64)
	copy(padded[64-len(data):], data)
	return padded
}
