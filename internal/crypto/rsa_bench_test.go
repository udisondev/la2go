package crypto

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/udisondev/la2go/internal/constants"
)

// BenchmarkRSADecrypt_1024 — P0: дешифрование при логине клиента
// После добавления Precompute(): улучшение -20-30% (было ~298µs)
func BenchmarkRSADecrypt_1024(b *testing.B) {
	b.ReportAllocs()

	// Setup: генерируем ключи и тестовые данные
	keyPair, err := GenerateRSAKeyPair()
	if err != nil {
		b.Fatalf("failed to generate key pair: %v", err)
	}

	// Создаем тестовый ciphertext (шифруем случайные данные)
	plaintext := make([]byte, constants.RSA1024ModulusSize)
	if _, err := rand.Read(plaintext); err != nil {
		b.Fatal(err)
	}

	// RSA encrypt: plaintext^e mod n
	m := new(big.Int).SetBytes(plaintext)
	e := big.NewInt(int64(constants.RSAPublicExponent))
	c := new(big.Int).Exp(m, e, keyPair.PrivateKey.N)
	ciphertext := c.Bytes()

	// Pad to 128 bytes
	if len(ciphertext) < constants.RSA1024ModulusSize {
		padded := make([]byte, constants.RSA1024ModulusSize)
		copy(padded[constants.RSA1024ModulusSize-len(ciphertext):], ciphertext)
		ciphertext = padded
	}

	b.ResetTimer()
	for range b.N {
		_, err := RSADecryptNoPadding(keyPair.PrivateKey, ciphertext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRSADecrypt_512 — дешифрование при регистрации GameServer
func BenchmarkRSADecrypt_512(b *testing.B) {
	b.ReportAllocs()

	keyPair, err := GenerateRSAKeyPair512()
	if err != nil {
		b.Fatalf("failed to generate key pair: %v", err)
	}

	plaintext := make([]byte, constants.RSA512ModulusSize)
	if _, err := rand.Read(plaintext); err != nil {
		b.Fatal(err)
	}

	m := new(big.Int).SetBytes(plaintext)
	e := big.NewInt(int64(constants.RSAPublicExponent))
	c := new(big.Int).Exp(m, e, keyPair.PrivateKey.N)
	ciphertext := c.Bytes()

	if len(ciphertext) < constants.RSA512ModulusSize {
		padded := make([]byte, constants.RSA512ModulusSize)
		copy(padded[constants.RSA512ModulusSize-len(ciphertext):], ciphertext)
		ciphertext = padded
	}

	b.ResetTimer()
	for range b.N {
		_, err := RSADecryptNoPadding(keyPair.PrivateKey, ciphertext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateRSAKeyPair — генерация ключей при старте LoginServer
func BenchmarkGenerateRSAKeyPair(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for range b.N {
		_, err := GenerateRSAKeyPair()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateRSAKeyPair512 — генерация ключей при старте GS↔LS listener
func BenchmarkGenerateRSAKeyPair512(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for range b.N {
		_, err := GenerateRSAKeyPair512()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkScrambleModulus — скремблирование модуля для Init пакета
func BenchmarkScrambleModulus(b *testing.B) {
	b.ReportAllocs()

	modulus := make([]byte, constants.RSA1024ModulusSize)
	if _, err := rand.Read(modulus); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for range b.N {
		_ = ScrambleModulus(modulus)
	}
}

// BenchmarkUnscrambleModulus — дескремблирование модуля на стороне клиента
func BenchmarkUnscrambleModulus(b *testing.B) {
	b.ReportAllocs()

	modulus := make([]byte, constants.RSA1024ModulusSize)
	if _, err := rand.Read(modulus); err != nil {
		b.Fatal(err)
	}
	scrambled := ScrambleModulus(modulus)

	b.ResetTimer()
	for range b.N {
		_ = UnscrambleModulus(scrambled)
	}
}

// BenchmarkRSAKeyPairGeneration_Complete — полный цикл генерации ключей + скремблирование
func BenchmarkRSAKeyPairGeneration_Complete(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for range b.N {
		keyPair, err := GenerateRSAKeyPair()
		if err != nil {
			b.Fatal(err)
		}
		// Скремблированный модуль уже есть в keyPair.ScrambledModulus
		_ = keyPair.ScrambledModulus
	}
}

// BenchmarkRSA_FullLoginCycle — полный цикл: генерация ключей + шифрование + дешифрование
func BenchmarkRSA_FullLoginCycle(b *testing.B) {
	b.ReportAllocs()

	// Setup: генерируем ключи заранее (как в реальном сервере)
	keyPair, err := GenerateRSAKeyPair()
	if err != nil {
		b.Fatalf("failed to generate key pair: %v", err)
	}

	plaintext := make([]byte, constants.RSA1024ModulusSize)

	b.ResetTimer()
	for range b.N {
		// 1. Клиент генерирует случайные данные
		if _, err := rand.Read(plaintext); err != nil {
			b.Fatal(err)
		}

		// 2. Клиент шифрует RSA (публичный ключ)
		m := new(big.Int).SetBytes(plaintext)
		e := big.NewInt(int64(constants.RSAPublicExponent))
		c := new(big.Int).Exp(m, e, keyPair.PrivateKey.N)
		ciphertext := c.Bytes()

		if len(ciphertext) < constants.RSA1024ModulusSize {
			padded := make([]byte, constants.RSA1024ModulusSize)
			copy(padded[constants.RSA1024ModulusSize-len(ciphertext):], ciphertext)
			ciphertext = padded
		}

		// 3. Сервер дешифрует RSA (приватный ключ)
		_, err := RSADecryptNoPadding(keyPair.PrivateKey, ciphertext)
		if err != nil {
			b.Fatal(err)
		}
	}
}
