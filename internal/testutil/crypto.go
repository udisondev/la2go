package testutil

import (
	"fmt"

	"github.com/udisondev/la2go/internal/crypto"
)

// BlowfishEncrypt шифрует data in-place используя указанный Blowfish cipher.
// Возвращает зашифрованные данные для удобства использования в тестах.
// data должен быть кратен 8 байтам (Blowfish block size).
func BlowfishEncrypt(cipher *crypto.BlowfishCipher, data []byte) ([]byte, error) {
	if len(data)%8 != 0 {
		return nil, fmt.Errorf("data length must be multiple of 8, got %d", len(data))
	}

	if err := cipher.Encrypt(data, 0, len(data)); err != nil {
		return nil, fmt.Errorf("blowfish encrypt: %w", err)
	}

	return data, nil
}

// BlowfishDecrypt расшифровывает data in-place используя указанный Blowfish cipher.
// Возвращает расшифрованные данные для удобства использования в тестах.
// data должен быть кратен 8 байтам (Blowfish block size).
func BlowfishDecrypt(cipher *crypto.BlowfishCipher, data []byte) ([]byte, error) {
	if len(data)%8 != 0 {
		return nil, fmt.Errorf("data length must be multiple of 8, got %d", len(data))
	}

	if err := cipher.Decrypt(data, 0, len(data)); err != nil {
		return nil, fmt.Errorf("blowfish decrypt: %w", err)
	}

	return data, nil
}

// PadToBlowfishBlock дополняет data до ближайшего кратного 8 байт нулями.
// Полезно для подготовки данных перед шифрованием Blowfish.
func PadToBlowfishBlock(data []byte) []byte {
	remainder := len(data) % 8
	if remainder == 0 {
		return data
	}

	padding := 8 - remainder
	padded := make([]byte, len(data)+padding)
	copy(padded, data)

	return padded
}

// DecXORPass применяет обратную операцию к encXORPass для расшифровки Init пакета.
// Алгоритм обратный к EncXORPass из crypto package.
// DEPRECATED: Используйте crypto.DecXORPass напрямую.
func DecXORPass(data []byte, offset, size int) {
	crypto.DecXORPass(data, offset, size)
}
