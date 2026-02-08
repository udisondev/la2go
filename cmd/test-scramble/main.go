package main

import (
	"encoding/hex"
	"fmt"
)

func main() {
	// Тестовый modulus (128 bytes)
	original := make([]byte, 128)
	for i := range original {
		original[i] = byte(i)
	}

	fmt.Printf("Original modulus (first 16 bytes): %s\n", hex.EncodeToString(original[:16]))
	fmt.Printf("Original modulus (last 16 bytes): %s\n", hex.EncodeToString(original[112:]))

	// Применяем scramble
	scrambled := make([]byte, 128)
	copy(scrambled, original)
	scramble(scrambled)

	fmt.Printf("\nScrambled modulus (first 16 bytes): %s\n", hex.EncodeToString(scrambled[:16]))
	fmt.Printf("Scrambled modulus (last 16 bytes): %s\n", hex.EncodeToString(scrambled[112:]))

	// Применяем unscramble
	unscrambled := make([]byte, 128)
	copy(unscrambled, scrambled)
	unscramble(unscrambled)

	fmt.Printf("\nUnscrambled modulus (first 16 bytes): %s\n", hex.EncodeToString(unscrambled[:16]))
	fmt.Printf("Unscrambled modulus (last 16 bytes): %s\n", hex.EncodeToString(unscrambled[112:]))

	// Проверяем что unscramble восстановил оригинал
	match := true
	for i := range original {
		if original[i] != unscrambled[i] {
			fmt.Printf("\n❌ FAIL at byte %d: original=0x%02X, unscrambled=0x%02X\n", i, original[i], unscrambled[i])
			match = false
			break
		}
	}

	if match {
		fmt.Println("\n✅ SUCCESS: Unscramble restored original modulus!")
	}
}

// scramble как в Java ScrambledKeyPair.scrambleModulus()
func scramble(data []byte) {
	// Step 1: swap bytes [0x00-0x03] <-> [0x4d-0x50]
	for i := 0; i < 4; i++ {
		data[0x00+i], data[0x4d+i] = data[0x4d+i], data[0x00+i]
	}

	// Step 2: XOR first 0x40 bytes with last 0x40 bytes
	for i := 0; i < 0x40; i++ {
		data[i] ^= data[0x40+i]
	}

	// Step 3: XOR bytes [0x0d-0x10] with bytes [0x34-0x38]
	for i := 0; i < 4; i++ {
		data[0x0d+i] ^= data[0x34+i]
	}

	// Step 4: XOR last 0x40 bytes with first 0x40 bytes
	for i := 0; i < 0x40; i++ {
		data[0x40+i] ^= data[i]
	}
}

// unscramble — обратная операция
func unscramble(data []byte) {
	// Применяем операции В ОБРАТНОМ ПОРЯДКЕ

	// Step 4 reverse: XOR last 0x40 bytes with first 0x40 bytes
	// B = B XOR A, чтобы восстановить B: B_new = B_old XOR A
	for i := 0; i < 0x40; i++ {
		data[0x40+i] ^= data[i]
	}

	// Step 3 reverse: XOR bytes [0x0d-0x10] with bytes [0x34-0x38]
	for i := 0; i < 4; i++ {
		data[0x0d+i] ^= data[0x34+i]
	}

	// Step 2 reverse: XOR first 0x40 bytes with last 0x40 bytes
	for i := 0; i < 0x40; i++ {
		data[i] ^= data[0x40+i]
	}

	// Step 1 reverse: swap bytes [0x00-0x03] <-> [0x4d-0x50]
	for i := 0; i < 4; i++ {
		data[0x00+i], data[0x4d+i] = data[0x4d+i], data[0x00+i]
	}
}
