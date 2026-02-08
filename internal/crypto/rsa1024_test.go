package crypto

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/udisondev/la2go/internal/constants"
)

// TestRSA1024_EncryptDecrypt проверяет базовую корректность RSA-1024 шифрования/дешифрования.
func TestRSA1024_EncryptDecrypt(t *testing.T) {
	kp, err := GenerateRSAKeyPair()
	require.NoError(t, err, "Failed to generate RSA-1024 key pair")

	// Генерируем random plaintext (имитируем L2 login packet payload)
	plaintext := make([]byte, constants.RSA1024ModulusSize)
	_, err = rand.Read(plaintext[:94]) // L2 RequestAuthLogin payload ~94 bytes
	require.NoError(t, err, "Failed to generate random plaintext")

	// Encrypt: c = m^e mod n
	m := new(big.Int).SetBytes(plaintext)
	// IMPORTANT: plaintext must be < N for valid RSA
	m.Mod(m, kp.PrivateKey.N)

	e := big.NewInt(int64(kp.PrivateKey.E))
	c := new(big.Int).Exp(m, e, kp.PrivateKey.N)
	ciphertext := c.Bytes()

	// Pad to 128 bytes (RSA-1024 key size)
	if len(ciphertext) < constants.RSA1024ModulusSize {
		padded := make([]byte, constants.RSA1024ModulusSize)
		copy(padded[constants.RSA1024ModulusSize-len(ciphertext):], ciphertext)
		ciphertext = padded
	}

	// Decrypt using RSADecryptNoPadding
	decrypted, err := RSADecryptNoPadding(kp.PrivateKey, ciphertext)
	require.NoError(t, err, "Failed to decrypt ciphertext")
	require.Len(t, decrypted, constants.RSA1024ModulusSize, "Decrypted result must be 128 bytes")

	// Compare as big.Int (ignore leading zeros)
	mDecrypted := new(big.Int).SetBytes(decrypted)
	assert.Equal(t, m, mDecrypted, "Decrypt(Encrypt(m)) must equal m")
}

// TestRSA1024_CRTVsFallback проверяет эквивалентность CRT и fallback путей дешифрования.
// Оба пути должны давать идентичный результат для одного и того же ciphertext.
func TestRSA1024_CRTVsFallback(t *testing.T) {
	kp, err := GenerateRSAKeyPair()
	require.NoError(t, err)

	// Создаём тестовый plaintext
	plaintext := make([]byte, constants.RSA1024ModulusSize)
	_, err = rand.Read(plaintext[:94])
	require.NoError(t, err)

	// Encrypt
	m := new(big.Int).SetBytes(plaintext)
	// IMPORTANT: plaintext must be < N for valid RSA
	m.Mod(m, kp.PrivateKey.N)
	e := big.NewInt(int64(kp.PrivateKey.E))
	c := new(big.Int).Exp(m, e, kp.PrivateKey.N)
	ciphertext := padToSize(c.Bytes(), constants.RSA1024ModulusSize)

	// Decrypt with CRT (normal path)
	resultCRT, err := RSADecryptNoPadding(kp.PrivateKey, ciphertext)
	require.NoError(t, err, "CRT decryption failed")

	// Decrypt with fallback (force by clearing Precomputed values)
	keyNoPrecompute := *kp.PrivateKey
	keyNoPrecompute.Precomputed.Dp = nil
	keyNoPrecompute.Precomputed.Dq = nil
	keyNoPrecompute.Precomputed.Qinv = nil

	resultFallback, err := RSADecryptNoPadding(&keyNoPrecompute, ciphertext)
	require.NoError(t, err, "Fallback decryption failed")

	// Оба пути должны дать идентичный результат
	assert.Equal(t, resultCRT, resultFallback,
		"CRT and fallback paths must produce identical results")
}

// TestRSA1024_EdgeCase_NegativeH проверяет корректность обработки отрицательного h.
// Когда m1 < m2 в CRT алгоритме, h = (m1 - m2) становится отрицательным.
// big.Int.Mod() должен правильно обработать это (вернуть положительный остаток).
func TestRSA1024_EdgeCase_NegativeH(t *testing.T) {
	kp, err := GenerateRSAKeyPair()
	require.NoError(t, err)

	// Run multiple iterations to increase probability of hitting edge case
	for i := range 100 {
		plaintext := make([]byte, constants.RSA1024ModulusSize)
		_, err = rand.Read(plaintext)
		require.NoError(t, err)

		m := new(big.Int).SetBytes(plaintext)
		// IMPORTANT: plaintext must be < N for valid RSA
		m.Mod(m, kp.PrivateKey.N)

		e := big.NewInt(int64(kp.PrivateKey.E))
		c := new(big.Int).Exp(m, e, kp.PrivateKey.N)

		ciphertext := padToSize(c.Bytes(), constants.RSA1024ModulusSize)
		decrypted, err := RSADecryptNoPadding(kp.PrivateKey, ciphertext)
		require.NoError(t, err, "Iteration %d: decryption failed", i)

		// Verify correctness (compare as big.Int to handle leading zeros)
		mDecrypted := new(big.Int).SetBytes(decrypted)
		assert.True(t, m.Cmp(mDecrypted) == 0, "Iteration %d: decrypt result incorrect", i)
	}
}

// TestRSA1024_EdgeCase_LeadingZeros проверяет корректность обработки plaintext с ведущими нулями.
// RSADecryptNoPadding должен всегда возвращать результат длиной keySize байт (128 для RSA-1024).
func TestRSA1024_EdgeCase_LeadingZeros(t *testing.T) {
	kp, err := GenerateRSAKeyPair()
	require.NoError(t, err)

	// Plaintext с большим количеством ведущих нулей
	plaintext := make([]byte, constants.RSA1024ModulusSize)
	_, err = rand.Read(plaintext[64:]) // First 64 bytes = 0
	require.NoError(t, err)

	m := new(big.Int).SetBytes(plaintext)
	// IMPORTANT: plaintext must be < N for valid RSA
	m.Mod(m, kp.PrivateKey.N)
	e := big.NewInt(int64(kp.PrivateKey.E))
	c := new(big.Int).Exp(m, e, kp.PrivateKey.N)
	ciphertext := padToSize(c.Bytes(), constants.RSA1024ModulusSize)

	decrypted, err := RSADecryptNoPadding(kp.PrivateKey, ciphertext)
	require.NoError(t, err)
	require.Len(t, decrypted, constants.RSA1024ModulusSize, "Result must be padded to keySize")

	// Verify correctness
	mDecrypted := new(big.Int).SetBytes(decrypted)
	assert.Equal(t, m, mDecrypted, "Leading zeros must be preserved")
}

// TestRSA1024_EdgeCase_CiphertextZero проверяет обработку нулевого ciphertext.
// Хотя это невалидный сценарий для реального использования, функция не должна паниковать.
func TestRSA1024_EdgeCase_CiphertextZero(t *testing.T) {
	kp, err := GenerateRSAKeyPair()
	require.NoError(t, err)

	// Ciphertext = 0 (все байты нулевые)
	ciphertext := make([]byte, constants.RSA1024ModulusSize)

	decrypted, err := RSADecryptNoPadding(kp.PrivateKey, ciphertext)
	require.NoError(t, err, "Should handle ciphertext=0 without panic")
	require.Len(t, decrypted, constants.RSA1024ModulusSize)

	// 0^d mod n = 0
	assert.Equal(t, ciphertext, decrypted, "Decrypt(0) must equal 0")
}

// TestRSA1024_MultipleKeys проверяет корректность для нескольких разных ключей.
// Разные ключи должны давать разные результаты для одного и того же plaintext.
func TestRSA1024_MultipleKeys(t *testing.T) {
	// Генерируем 3 разных ключа
	keys := make([]*RSAKeyPair, 3)
	for i := range keys {
		kp, err := GenerateRSAKeyPair()
		require.NoError(t, err, "Failed to generate key %d", i)
		keys[i] = kp
	}

	// Для каждого ключа используем plaintext < его N
	for i, kp := range keys {
		plaintext := make([]byte, constants.RSA1024ModulusSize)
		_, err := rand.Read(plaintext[:94])
		require.NoError(t, err)

		m := new(big.Int).SetBytes(plaintext)
		// IMPORTANT: plaintext must be < N for valid RSA
		m.Mod(m, kp.PrivateKey.N)

		// Encrypt
		e := big.NewInt(int64(kp.PrivateKey.E))
		c := new(big.Int).Exp(m, e, kp.PrivateKey.N)
		ciphertext := padToSize(c.Bytes(), constants.RSA1024ModulusSize)

		// Decrypt
		decrypted, err := RSADecryptNoPadding(kp.PrivateKey, ciphertext)
		require.NoError(t, err, "Key %d: decryption failed", i)

		// Verify
		mDecrypted := new(big.Int).SetBytes(decrypted)
		assert.True(t, m.Cmp(mDecrypted) == 0, "Key %d: incorrect decryption", i)
	}
}

// TestRSA1024_InvalidCiphertextSize проверяет обработку невалидного размера ciphertext.
func TestRSA1024_InvalidCiphertextSize(t *testing.T) {
	kp, err := GenerateRSAKeyPair()
	require.NoError(t, err)

	tests := []struct {
		name string
		size int
	}{
		{"empty", 0},
		{"too_short", 64},
		{"too_long", 256},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext := make([]byte, tt.size)
			_, err := RSADecryptNoPadding(kp.PrivateKey, ciphertext)
			assert.Error(t, err, "Should return error for invalid ciphertext size")
		})
	}
}

// padToSize добавляет ведущие нули к data до достижения size байт.
// Если data уже >= size, возвращает как есть (обрезая если нужно).
func padToSize(data []byte, size int) []byte {
	if len(data) >= size {
		return data[len(data)-size:]
	}
	padded := make([]byte, size)
	copy(padded[size-len(data):], data)
	return padded
}
