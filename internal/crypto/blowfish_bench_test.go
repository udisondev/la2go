package crypto

import (
	"testing"

	"github.com/udisondev/la2go/internal/constants"
)

// BenchmarkBlowfishEncrypt — P0 hotpath: шифрование каждого исходящего пакета
func BenchmarkBlowfishEncrypt(b *testing.B) {
	b.ReportAllocs()

	cipher, err := NewBlowfishCipher(DefaultGSBlowfishKey)
	if err != nil {
		b.Fatalf("failed to create cipher: %v", err)
	}

	data := make([]byte, 256) // Типичный размер пакета

	b.ResetTimer()
	for range b.N {
		if err := cipher.Encrypt(data, 0, len(data)); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBlowfishEncrypt_Sizes — тест производительности для разных размеров пакетов
func BenchmarkBlowfishEncrypt_Sizes(b *testing.B) {
	sizes := []int{64, 128, 256, 512, 1024, 2048}

	for _, size := range sizes {
		b.Run(formatSize(size), func(b *testing.B) {
			b.ReportAllocs()

			cipher, err := NewBlowfishCipher(DefaultGSBlowfishKey)
			if err != nil {
				b.Fatalf("failed to create cipher: %v", err)
			}

			data := make([]byte, size)
			b.SetBytes(int64(size)) // Для расчета throughput (MB/s)

			b.ResetTimer()
			for range b.N {
				if err := cipher.Encrypt(data, 0, size); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkBlowfishDecrypt — P0 hotpath: дешифрование каждого входящего пакета
func BenchmarkBlowfishDecrypt(b *testing.B) {
	b.ReportAllocs()

	cipher, err := NewBlowfishCipher(DefaultGSBlowfishKey)
	if err != nil {
		b.Fatalf("failed to create cipher: %v", err)
	}

	data := make([]byte, 256)
	// Сначала шифруем данные
	if err := cipher.Encrypt(data, 0, len(data)); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for range b.N {
		if err := cipher.Decrypt(data, 0, len(data)); err != nil {
			b.Fatal(err)
		}
		// Re-encrypt для следующей итерации
		if err := cipher.Encrypt(data, 0, len(data)); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBlowfishDecrypt_Sizes — тест производительности дешифрования для разных размеров
func BenchmarkBlowfishDecrypt_Sizes(b *testing.B) {
	sizes := []int{64, 128, 256, 512, 1024, 2048}

	for _, size := range sizes {
		b.Run(formatSize(size), func(b *testing.B) {
			b.ReportAllocs()

			cipher, err := NewBlowfishCipher(DefaultGSBlowfishKey)
			if err != nil {
				b.Fatalf("failed to create cipher: %v", err)
			}

			data := make([]byte, size)
			if err := cipher.Encrypt(data, 0, size); err != nil {
				b.Fatal(err)
			}

			b.SetBytes(int64(size))

			b.ResetTimer()
			for range b.N {
				if err := cipher.Decrypt(data, 0, size); err != nil {
					b.Fatal(err)
				}
				if err := cipher.Encrypt(data, 0, size); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkAppendChecksum — вызывается для каждого исходящего пакета
func BenchmarkAppendChecksum(b *testing.B) {
	b.ReportAllocs()

	// Данные + 4 байта для checksum
	data := make([]byte, 256+constants.PacketChecksumSize)

	b.ResetTimer()
	for range b.N {
		AppendChecksum(data, 0, len(data))
	}
}

// BenchmarkAppendChecksum_Sizes — разные размеры пакетов
func BenchmarkAppendChecksum_Sizes(b *testing.B) {
	sizes := []int{64, 128, 256, 512, 1024, 2048}

	for _, size := range sizes {
		b.Run(formatSize(size), func(b *testing.B) {
			b.ReportAllocs()

			data := make([]byte, size+constants.PacketChecksumSize)
			b.SetBytes(int64(size))

			b.ResetTimer()
			for range b.N {
				AppendChecksum(data, 0, len(data))
			}
		})
	}
}

// BenchmarkVerifyChecksum — вызывается для каждого входящего пакета
func BenchmarkVerifyChecksum(b *testing.B) {
	b.ReportAllocs()

	data := make([]byte, 256+constants.PacketChecksumSize)
	AppendChecksum(data, 0, len(data))

	b.ResetTimer()
	for range b.N {
		if !VerifyChecksum(data, 0, len(data)) {
			b.Fatal("checksum verification failed")
		}
	}
}

// BenchmarkVerifyChecksum_Sizes — разные размеры пакетов
func BenchmarkVerifyChecksum_Sizes(b *testing.B) {
	sizes := []int{64, 128, 256, 512, 1024, 2048}

	for _, size := range sizes {
		b.Run(formatSize(size), func(b *testing.B) {
			b.ReportAllocs()

			data := make([]byte, size+constants.PacketChecksumSize)
			AppendChecksum(data, 0, len(data))
			b.SetBytes(int64(size))

			b.ResetTimer()
			for range b.N {
				if !VerifyChecksum(data, 0, len(data)) {
					b.Fatal("checksum verification failed")
				}
			}
		})
	}
}

// BenchmarkEncXORPass — XOR шифрование для Init пакета (вызывается при каждом новом подключении)
func BenchmarkEncXORPass(b *testing.B) {
	b.ReportAllocs()

	data := make([]byte, 256)
	key := int32(0x12345678)

	b.ResetTimer()
	for range b.N {
		EncXORPass(data, 0, len(data), key)
	}
}

// BenchmarkDecXORPass — XOR дешифрование для Init пакета
func BenchmarkDecXORPass(b *testing.B) {
	b.ReportAllocs()

	data := make([]byte, 256)
	key := int32(0x12345678)
	EncXORPass(data, 0, len(data), key)

	b.ResetTimer()
	for range b.N {
		DecXORPass(data, 0, len(data))
		// Re-encrypt для следующей итерации
		EncXORPass(data, 0, len(data), key)
	}
}

// BenchmarkBlowfishCipherCreation — проверка overhead создания cipher
func BenchmarkBlowfishCipherCreation(b *testing.B) {
	b.ReportAllocs()

	key := DefaultGSBlowfishKey

	b.ResetTimer()
	for range b.N {
		_, err := NewBlowfishCipher(key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// formatSize форматирует размер в байтах для имени бенчмарка
func formatSize(size int) string {
	if size >= 1024 {
		return string(rune('0'+size/1024)) + "KB"
	}
	return string(rune('0'+size/64)) + "x64B"
}
