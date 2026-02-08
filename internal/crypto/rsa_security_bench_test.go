package crypto

import (
	"crypto/rand"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/constants"
)

// BenchmarkRSADecrypt_TimingVariance измеряет variance времени выполнения для анализа timing attacks.
// Высокая вариативность (CV > 5%) может указывать на timing leak.
func BenchmarkRSADecrypt_TimingVariance(b *testing.B) {
	kp, err := GenerateRSAKeyPair()
	if err != nil {
		b.Fatalf("Failed to generate key: %v", err)
	}

	// Генерируем 100 разных ciphertext'ов для измерения variance
	const numSamples = 100
	ciphertexts := make([][]byte, numSamples)
	for i := range ciphertexts {
		plaintext := make([]byte, constants.RSA1024ModulusSize)
		if _, err := rand.Read(plaintext); err != nil {
			b.Fatalf("Failed to generate plaintext: %v", err)
		}

		m := new(big.Int).SetBytes(plaintext)
		e := big.NewInt(int64(kp.PrivateKey.E))
		c := new(big.Int).Exp(m, e, kp.PrivateKey.N)
		ciphertexts[i] = padToSize(c.Bytes(), constants.RSA1024ModulusSize)
	}

	b.ResetTimer()

	// Измеряем timing для каждого ciphertext
	times := make([]time.Duration, 0, b.N)
	for i := 0; i < b.N; i++ {
		ct := ciphertexts[i%numSamples]
		start := time.Now()
		if _, err := RSADecryptNoPadding(kp.PrivateKey, ct); err != nil {
			b.Fatalf("Decryption failed: %v", err)
		}
		times = append(times, time.Since(start))
	}

	b.StopTimer()

	// Compute statistics (среднее и стандартное отклонение)
	if len(times) > 0 {
		mean, stddev := computeTimingStats(times)
		cvFloat := float64(stddev) / float64(mean) // Coefficient of variation

		b.ReportMetric(mean.Seconds()*1e6, "mean_µs")
		b.ReportMetric(stddev.Seconds()*1e6, "stddev_µs")
		b.ReportMetric(cvFloat*100, "cv_%")

		// Warning if variance > 5%
		if cvFloat > 0.05 {
			b.Logf("⚠️  WARNING: High timing variance detected: %.2f%% (mean=%.2fµs, stddev=%.2fµs)",
				cvFloat*100, mean.Seconds()*1e6, stddev.Seconds()*1e6)
		}
	}
}

// BenchmarkRSA_CRTVsFallbackTiming измеряет разницу во времени между CRT и fallback путями.
// Существенная разница (>2x) создаёт timing attack vector.
func BenchmarkRSA_CRTVsFallbackTiming(b *testing.B) {
	kp, err := GenerateRSAKeyPair()
	if err != nil {
		b.Fatalf("Failed to generate key: %v", err)
	}

	// Генерируем тестовый ciphertext
	plaintext := make([]byte, constants.RSA1024ModulusSize)
	if _, err := rand.Read(plaintext); err != nil {
		b.Fatalf("Failed to generate plaintext: %v", err)
	}

	m := new(big.Int).SetBytes(plaintext)
	e := big.NewInt(int64(kp.PrivateKey.E))
	c := new(big.Int).Exp(m, e, kp.PrivateKey.N)
	ciphertext := padToSize(c.Bytes(), constants.RSA1024ModulusSize)

	b.Run("CRT", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := RSADecryptNoPadding(kp.PrivateKey, ciphertext); err != nil {
				b.Fatalf("CRT decryption failed: %v", err)
			}
		}
	})

	b.Run("Fallback", func(b *testing.B) {
		// Создаём ключ без Precomputed values для форсирования fallback пути
		keyNoPrecompute := *kp.PrivateKey
		keyNoPrecompute.Precomputed.Dp = nil
		keyNoPrecompute.Precomputed.Dq = nil
		keyNoPrecompute.Precomputed.Qinv = nil

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := RSADecryptNoPadding(&keyNoPrecompute, ciphertext); err != nil {
				b.Fatalf("Fallback decryption failed: %v", err)
			}
		}
	})
}

// BenchmarkRSA_InputVariety измеряет timing для разных типов входных данных.
// Проверяет, есть ли корреляция между содержимым ciphertext и временем выполнения.
func BenchmarkRSA_InputVariety(b *testing.B) {
	kp, err := GenerateRSAKeyPair()
	if err != nil {
		b.Fatalf("Failed to generate key: %v", err)
	}

	// Создаём разные типы plaintext для проверки timing bias
	plaintexts := map[string][]byte{
		"all_zeros":      make([]byte, constants.RSA1024ModulusSize),
		"all_ones":       makeFilledBytes(constants.RSA1024ModulusSize, 0xFF),
		"random":         makeRandomBytes(constants.RSA1024ModulusSize),
		"leading_zeros":  makeLeadingZeros(constants.RSA1024ModulusSize, 64),
		"trailing_zeros": makeTrailingZeros(constants.RSA1024ModulusSize, 64),
	}

	e := big.NewInt(int64(kp.PrivateKey.E))

	for name, plaintext := range plaintexts {
		b.Run(name, func(b *testing.B) {
			m := new(big.Int).SetBytes(plaintext)
			c := new(big.Int).Exp(m, e, kp.PrivateKey.N)
			ciphertext := padToSize(c.Bytes(), constants.RSA1024ModulusSize)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := RSADecryptNoPadding(kp.PrivateKey, ciphertext); err != nil {
					b.Fatalf("Decryption failed: %v", err)
				}
			}
		})
	}
}

// computeTimingStats вычисляет среднее и стандартное отклонение для среза time.Duration.
func computeTimingStats(times []time.Duration) (mean, stddev time.Duration) {
	if len(times) == 0 {
		return 0, 0
	}

	// Compute mean
	var sum time.Duration
	for _, t := range times {
		sum += t
	}
	mean = sum / time.Duration(len(times))

	// Compute standard deviation
	var variance float64
	for _, t := range times {
		diff := float64(t - mean)
		variance += diff * diff
	}
	variance /= float64(len(times))
	stddev = time.Duration(math.Sqrt(variance))

	return mean, stddev
}

// makeFilledBytes создаёт слайс из count байт, заполненных значением value.
func makeFilledBytes(count int, value byte) []byte {
	b := make([]byte, count)
	for i := range b {
		b[i] = value
	}
	return b
}

// makeRandomBytes создаёт слайс из count случайных байт.
func makeRandomBytes(count int) []byte {
	b := make([]byte, count)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}

// makeLeadingZeros создаёт слайс длиной total с zeroCount ведущими нулями.
func makeLeadingZeros(total, zeroCount int) []byte {
	b := make([]byte, total)
	if _, err := rand.Read(b[zeroCount:]); err != nil {
		panic(err)
	}
	return b
}

// makeTrailingZeros создаёт слайс длиной total с zeroCount замыкающими нулями.
func makeTrailingZeros(total, zeroCount int) []byte {
	b := make([]byte, total)
	if _, err := rand.Read(b[:total-zeroCount]); err != nil {
		panic(err)
	}
	return b
}
