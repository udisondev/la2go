package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// Симулируем создание Init пакета как в Java
func main() {
	// Данные Init пакета (БЕЗ шифрования):
	sessionID := int32(0x12345678)
	protocolRev := uint32(0x0000c621)
	rsaPublicKey := make([]byte, 128) // 128 байт scrambled RSA modulus
	for i := range rsaPublicKey {
		rsaPublicKey[i] = byte(i) // Тестовые данные
	}
	ggData := []uint32{0x29DD954E, 0x77C39CFC, 0x97ADB620, 0x07BDE0F7} // 4×int32 = 16 bytes
	blowfishKey := []byte{
		0x04, 0xa1, 0xc3, 0x42, 0xad, 0xaa, 0xf2, 0x34,
		0x30, 0x78, 0x9f, 0x61, 0xb8, 0x92, 0x53, 0x32,
	} // 16 bytes

	// Расчёт размера plaintext Init пакета:
	// opcode(1) + sessionID(4) + protocolRev(4) + rsaPublicKey(128) + ggData(16) + blowfishKey(16) + null(1) = 170 bytes
	plaintextSize := 1 + 4 + 4 + 128 + 16 + 16 + 1
	fmt.Printf("Plaintext Init packet size: %d bytes\n", plaintextSize)

	// Создаём plaintext пакет
	plaintext := make([]byte, plaintextSize)
	pos := 0

	// opcode
	plaintext[pos] = 0x00
	pos++

	// sessionID
	binary.LittleEndian.PutUint32(plaintext[pos:], uint32(sessionID))
	pos += 4

	// protocolRev
	binary.LittleEndian.PutUint32(plaintext[pos:], protocolRev)
	pos += 4

	// rsaPublicKey
	copy(plaintext[pos:], rsaPublicKey)
	pos += 128

	// ggData
	for _, v := range ggData {
		binary.LittleEndian.PutUint32(plaintext[pos:], v)
		pos += 4
	}

	// blowfishKey
	copy(plaintext[pos:], blowfishKey)
	pos += 16

	// null termination
	plaintext[pos] = 0x00
	pos++

	fmt.Printf("Final position: %d (should be %d)\n", pos, plaintextSize)
	fmt.Printf("\nPlaintext Init packet (first 50 bytes):\n%s\n", hex.Dump(plaintext[:50]))
	fmt.Printf("\nPlaintext Init packet (last 50 bytes):\n%s\n", hex.Dump(plaintext[plaintextSize-50:]))

	// Теперь применяем encXORPass как в Java
	// LoginEncryption.encryptedSize(170):
	// dataSize = 170
	// dataSize += 8 (для _static=true)  → 178
	// dataSize += 8 - (178 % 8) = 8 - 2 = 6 → 184
	// dataSize += 8 → 192
	// НО: это не совсем правильно, давайте посчитаем точно по формуле Java

	dataSize := plaintextSize
	fmt.Printf("\n=== encryptedSize calculation ===\n")
	fmt.Printf("Initial dataSize: %d\n", dataSize)
	dataSize += 8 // _static = true
	fmt.Printf("After +8 (static): %d\n", dataSize)
	dataSize += 8 - (dataSize % 8)
	fmt.Printf("After padding to 8: %d\n", dataSize)
	dataSize += 8
	fmt.Printf("After +8 (final): %d\n", dataSize)

	// Создаём буфер для шифрования (с padding)
	encBuf := make([]byte, dataSize)
	copy(encBuf, plaintext)

	fmt.Printf("\nBefore encXORPass (blowfish key at offset 153-168):\n")
	fmt.Printf("Offset 153-168: %s\n", hex.EncodeToString(encBuf[153:169]))

	// Применяем encXORPass
	// encXORPass(data, offset=0, size=dataSize, key=randomInt)
	xorKey := int32(0x0BCDEF01) // Тестовый ключ (положительное значение)
	encXORPass(encBuf, 0, dataSize, xorKey)

	fmt.Printf("\nAfter encXORPass (blowfish key at offset 153-168):\n")
	fmt.Printf("Offset 153-168: %s\n", hex.EncodeToString(encBuf[153:169]))

	fmt.Printf("\nencXORPass wrote final key at offset %d--%d:\n", dataSize-8, dataSize-4)
	fmt.Printf("Final XOR key: %s\n", hex.EncodeToString(encBuf[dataSize-8:dataSize-4]))

	// Проверяем что blowfish key НЕ повреждён
	// Blowfish key начинается с offset:
	// opcode(1) + sessionID(4) + protocolRev(4) + rsaPublicKey(128) + ggData(16) = 153
	blowfishOffset := 153
	fmt.Printf("\nBlowfish key offset: %d\n", blowfishOffset)
	fmt.Printf("Blowfish key (expected): %s\n", hex.EncodeToString(blowfishKey))
	fmt.Printf("Blowfish key (in buffer after encXORPass): %s\n", hex.EncodeToString(encBuf[blowfishOffset:blowfishOffset+16]))

	// Анализ encXORPass диапазона:
	// encXORPass(data, offset, size, key):
	// - stop = size - 8
	// - pos = 4 + offset (skips first 4 bytes)
	// - while pos < stop: процесс XOR
	// - writeInt(pos, ecx) — записывает финальный ключ

	stop := dataSize - 8
	startPos := 4 + 0
	fmt.Printf("\nencXORPass range:\n")
	fmt.Printf("  start (pos=4+offset): %d\n", startPos)
	fmt.Printf("  stop (size-8): %d\n", stop)
	fmt.Printf("  final key written at: %d--%d\n", stop, stop+4)
	fmt.Printf("  blowfish key range: %d--%d\n", blowfishOffset, blowfishOffset+16)

	if blowfishOffset+16 > stop && blowfishOffset < stop+4 {
		fmt.Printf("\n❌ WARNING: encXORPass OVERWRITES part of blowfish key!\n")
		fmt.Printf("   Blowfish key ends at %d, encXORPass writes at %d--%d\n", blowfishOffset+16, stop, stop+4)
	} else {
		fmt.Printf("\n✅ OK: encXORPass does NOT overwrite blowfish key\n")
	}
}

// encXORPass как в Java
func encXORPass(data []byte, offset, size int, key int32) {
	stop := size - 8
	pos := 4 + offset
	ecx := key

	for pos < stop {
		edx := int32(binary.LittleEndian.Uint32(data[pos:]))
		ecx += edx
		edx ^= ecx
		binary.LittleEndian.PutUint32(data[pos:], uint32(edx))
		pos += 4
	}

	binary.LittleEndian.PutUint32(data[pos:], uint32(ecx))
}
