package crypto

import (
	"bytes"
	"testing"
)

func TestDefaultGSBlowfishKey_Length(t *testing.T) {
	// DefaultGSBlowfishKey должен быть 22 байта
	expectedLen := 22
	if len(DefaultGSBlowfishKey) != expectedLen {
		t.Errorf("Expected DefaultGSBlowfishKey length %d, got %d", expectedLen, len(DefaultGSBlowfishKey))
	}
}

func TestDefaultGSBlowfishKey_Value(t *testing.T) {
	// Проверяем, что DefaultGSBlowfishKey соответствует строке "_;v.]05-31!|+-%xT!^[$\x00"
	expected := []byte("_;v.]05-31!|+-%xT!^[$\x00")

	if !bytes.Equal(DefaultGSBlowfishKey, expected) {
		t.Errorf("DefaultGSBlowfishKey mismatch.\nExpected: %v\nGot:      %v", expected, DefaultGSBlowfishKey)
	}
}

func TestDefaultGSBlowfishKey_CreateCipher(t *testing.T) {
	// Проверяем, что можно создать BlowfishCipher с DefaultGSBlowfishKey
	cipher, err := NewBlowfishCipher(DefaultGSBlowfishKey)
	if err != nil {
		t.Fatalf("Failed to create BlowfishCipher with DefaultGSBlowfishKey: %v", err)
	}

	// Простой тест encrypt/decrypt
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	original := make([]byte, len(data))
	copy(original, data)

	// Encrypt
	if err := cipher.Encrypt(data, 0, len(data)); err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Проверяем, что данные изменились
	if bytes.Equal(data, original) {
		t.Error("Data should be encrypted (different from original)")
	}

	// Decrypt
	if err := cipher.Decrypt(data, 0, len(data)); err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	// Проверяем, что расшифрованное совпадает с оригиналом
	if !bytes.Equal(data, original) {
		t.Error("Decrypted data does not match original")
	}
}

func TestDefaultGSBlowfishKey_EncryptDecryptMultipleBlocks(t *testing.T) {
	cipher, err := NewBlowfishCipher(DefaultGSBlowfishKey)
	if err != nil {
		t.Fatalf("Failed to create cipher: %v", err)
	}

	// Тестируем с несколькими 8-байтными блоками
	data := make([]byte, 32)
	for i := range data {
		data[i] = byte(i)
	}
	original := make([]byte, len(data))
	copy(original, data)

	// Encrypt
	if err := cipher.Encrypt(data, 0, len(data)); err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Decrypt
	if err := cipher.Decrypt(data, 0, len(data)); err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	// Проверяем
	if !bytes.Equal(data, original) {
		t.Error("Decrypted multi-block data does not match original")
	}
}
