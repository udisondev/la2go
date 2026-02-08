package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/udisondev/la2go/internal/crypto"
)

func main() {
	// Пример: создаём dynamic Blowfish key (16 bytes)
	dynamicKey := []byte{
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C,
		0x0D, 0x0E, 0x0F, 0x10,
	}

	fmt.Printf("Dynamic key: %s\n", hex.EncodeToString(dynamicKey))

	// Создаём LoginEncryption
	enc, err := crypto.NewLoginEncryption(dynamicKey)
	if err != nil {
		log.Fatalf("create login encryption: %v", err)
	}

	// Создаём тестовый пакет AuthGameGuard (opcode 0x07 + sessionID + 4×data)
	sessionID := int32(0x12345678)
	plaintext := make([]byte, 2+21+16) // 2-byte header + 21-byte packet + 16-byte padding room
	plaintext[2] = 0x07                 // opcode
	binary.LittleEndian.PutUint32(plaintext[3:], uint32(sessionID))
	// data1-data4 = 0

	fmt.Printf("Plaintext packet (opcode + sessionID + data): %s\n", hex.EncodeToString(plaintext[2:2+21]))

	// Шифруем как клиент (appendChecksum + dynamicCipher)
	encSize, err := enc.EncryptPacketClient(plaintext, 2, 21)
	if err != nil {
		log.Fatalf("encrypt packet client: %v", err)
	}

	fmt.Printf("Encrypted size: %d bytes\n", encSize)
	fmt.Printf("Encrypted packet: %s\n", hex.EncodeToString(plaintext[2:2+encSize]))

	// Теперь попробуем расшифровать на сервере
	encryptedPacket := make([]byte, encSize)
	copy(encryptedPacket, plaintext[2:2+encSize])

	// Создаём НОВЫЙ LoginEncryption с ТЕМ ЖЕ ключом (как на сервере)
	serverEnc, err := crypto.NewLoginEncryption(dynamicKey)
	if err != nil {
		log.Fatalf("create server login encryption: %v", err)
	}

	// Дешифруем
	ok, err := serverEnc.DecryptPacket(encryptedPacket, 0, encSize)
	if err != nil {
		log.Fatalf("decrypt packet: %v", err)
	}

	if !ok {
		slog.Warn("checksum verification FAILED")
	} else {
		slog.Info("checksum verification OK")
	}

	fmt.Printf("Decrypted packet: %s\n", hex.EncodeToString(encryptedPacket[:21]))

	// Проверяем opcode и sessionID
	opcode := encryptedPacket[0]
	receivedSessionID := int32(binary.LittleEndian.Uint32(encryptedPacket[1:5]))

	fmt.Printf("Opcode: 0x%02X (expected 0x07)\n", opcode)
	fmt.Printf("SessionID: 0x%08X (expected 0x%08X)\n", receivedSessionID, sessionID)

	if opcode != 0x07 {
		fmt.Printf("ERROR: opcode mismatch!\n")
		os.Exit(1)
	}

	if receivedSessionID != sessionID {
		fmt.Printf("ERROR: sessionID mismatch!\n")
		os.Exit(1)
	}

	fmt.Println("SUCCESS: packet encrypted/decrypted correctly")
}
