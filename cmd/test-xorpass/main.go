package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

func main() {
	// Симулируем encXORPass и decXORPass
	blowfishKey := []byte{
		0x04, 0xa1, 0xc3, 0x42, 0xad, 0xaa, 0xf2, 0x34,
		0x30, 0x78, 0x9f, 0x61, 0xb8, 0x92, 0x53, 0x32,
	}

	// Создаём Init пакет (plaintext)
	encryptedSize := 192 // после padding (plaintext 170 bytes → 192 encrypted)

	plaintext := make([]byte, encryptedSize)

	// Заполняем тестовыми данными (упрощённо)
	plaintext[0] = 0x00 // opcode
	binary.LittleEndian.PutUint32(plaintext[1:], 0x12345678) // sessionID
	// ... остальные данные ...

	// Blowfish key на offset 153
	blowfishOffset := 153
	copy(plaintext[blowfishOffset:], blowfishKey)

	fmt.Printf("Original plaintext blowfish key: %s\n", hex.EncodeToString(plaintext[blowfishOffset:blowfishOffset+16]))

	// Применяем encXORPass как на сервере
	xorKey := int32(0x0BCDEF01)
	encXORPass(plaintext, 0, encryptedSize, xorKey)

	fmt.Printf("After encXORPass blowfish key: %s\n", hex.EncodeToString(plaintext[blowfishOffset:blowfishOffset+16]))

	// Теперь применяем decXORPass как на клиенте
	decXORPass(plaintext, 0, encryptedSize)

	fmt.Printf("After decXORPass blowfish key: %s\n", hex.EncodeToString(plaintext[blowfishOffset:blowfishOffset+16]))

	// Сравниваем
	if hex.EncodeToString(plaintext[blowfishOffset:blowfishOffset+16]) == hex.EncodeToString(blowfishKey) {
		fmt.Println("\n✅ SUCCESS: Blowfish key recovered correctly!")
	} else {
		fmt.Println("\n❌ FAIL: Blowfish key NOT recovered!")
		fmt.Printf("Expected: %s\n", hex.EncodeToString(blowfishKey))
		fmt.Printf("Got:      %s\n", hex.EncodeToString(plaintext[blowfishOffset:blowfishOffset+16]))
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

// decXORPass как в нашем Go коде
func decXORPass(data []byte, offset, size int) {
	stop := offset + size - 8
	pos := offset + 4 // Skip first 4 bytes

	// Читаем финальный накопленный ключ из последних 4 байт
	ecx := binary.LittleEndian.Uint32(data[stop:])

	// Идём с конца в обратном направлении
	for i := stop - 4; i >= pos; i -= 4 {
		edx := binary.LittleEndian.Uint32(data[i:])
		edx ^= ecx
		binary.LittleEndian.PutUint32(data[i:], edx)
		ecx -= edx
	}
}
