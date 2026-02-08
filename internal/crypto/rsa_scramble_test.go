package crypto

import (
	"bytes"
	"testing"
)

func TestScrambleUnscrambleModulus(t *testing.T) {
	// Создаём тестовый modulus (128 bytes)
	original := make([]byte, 128)
	for i := range original {
		original[i] = byte(i)
	}

	// Scramble
	scrambled := ScrambleModulus(original)

	// Проверяем что scrambled отличается от original
	if bytes.Equal(original, scrambled) {
		t.Error("ScrambleModulus returned unchanged data")
	}

	// Unscramble
	unscrambled := UnscrambleModulus(scrambled)

	// Проверяем что unscrambled совпадает с original
	if !bytes.Equal(original, unscrambled) {
		t.Error("UnscrambleModulus did not restore original modulus")
		t.Logf("Original (first 16 bytes):    %x", original[:16])
		t.Logf("Unscrambled (first 16 bytes): %x", unscrambled[:16])

		// Найти первое несовпадение
		for i := range original {
			if original[i] != unscrambled[i] {
				t.Errorf("First mismatch at byte %d: original=0x%02X, unscrambled=0x%02X", i, original[i], unscrambled[i])
				break
			}
		}
	}
}

func TestScrambleUnscrambleRealRSAKey(t *testing.T) {
	// Генерируем реальный RSA key pair
	keyPair, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	// Получаем original modulus
	originalModulus := keyPair.PrivateKey.PublicKey.N.Bytes()

	// Pad to 128 bytes if needed
	if len(originalModulus) < 128 {
		padded := make([]byte, 128)
		copy(padded[128-len(originalModulus):], originalModulus)
		originalModulus = padded
	} else if len(originalModulus) == 129 && originalModulus[0] == 0 {
		originalModulus = originalModulus[1:]
	}

	// Scramble (уже сделано в GenerateRSAKeyPair, но проверим)
	scrambled := ScrambleModulus(originalModulus)

	// Unscramble
	unscrambled := UnscrambleModulus(scrambled)

	// Проверяем что unscrambled совпадает с original
	if !bytes.Equal(originalModulus, unscrambled) {
		t.Error("UnscrambleModulus did not restore original RSA modulus")
		t.Logf("Original (first 16 bytes):    %x", originalModulus[:16])
		t.Logf("Unscrambled (first 16 bytes): %x", unscrambled[:16])
	}

	// Проверяем что ScrambledModulus из keyPair совпадает с scrambled
	if !bytes.Equal(keyPair.ScrambledModulus, scrambled) {
		t.Error("keyPair.ScrambledModulus does not match ScrambleModulus(originalModulus)")
	}
}

func TestScrambleModulusPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ScrambleModulus did not panic on wrong size")
		}
	}()

	// Should panic
	ScrambleModulus(make([]byte, 64))
}

func TestUnscrambleModulusPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("UnscrambleModulus did not panic on wrong size")
		}
	}()

	// Should panic
	UnscrambleModulus(make([]byte, 64))
}
