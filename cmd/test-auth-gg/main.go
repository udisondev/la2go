package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/login/serverpackets"
)

func main() {
	// Подключаемся к РЕАЛЬНОМУ LoginServer (должен быть запущен)
	// Используйте: `./loginserver` в другом терминале
	conn, err := net.Dial("tcp", "127.0.0.1:2106")
	if err != nil {
		slog.Error("failed to connect to LoginServer", "err", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to LoginServer")

	// Читаем Init пакет
	buf := make([]byte, 4096)

	// Читаем 2-byte header
	var header [2]byte
	if _, err := io.ReadFull(conn, header[:]); err != nil {
		slog.Error("failed to read header", "err", err)
		return
	}

	totalLen := int(binary.LittleEndian.Uint16(header[:]))
	fmt.Printf("Init packet length: %d bytes\n", totalLen)

	if totalLen < 2 {
		slog.Error("invalid packet length", "len", totalLen)
		return
	}

	payloadLen := totalLen - 2
	payload := buf[:payloadLen]
	if _, err := io.ReadFull(conn, payload); err != nil {
		slog.Error("failed to read payload", "err", err)
		return
	}

	fmt.Printf("Encrypted Init payload (first 32 bytes): %s\n", hex.EncodeToString(payload[:min(32, len(payload))]))

	// Расшифровываем Init пакет (статический Blowfish + decXORPass)
	staticCipher, err := crypto.NewBlowfishCipher(crypto.StaticBlowfishKey)
	if err != nil {
		slog.Error("failed to create static cipher", "err", err)
		return
	}

	if err := staticCipher.Decrypt(payload, 0, payloadLen); err != nil {
		slog.Error("failed to decrypt Init", "err", err)
		return
	}

	// Применяем decXORPass
	crypto.DecXORPass(payload, 0, payloadLen)

	fmt.Printf("Decrypted Init payload (first 32 bytes): %s\n", hex.EncodeToString(payload[:min(32, len(payload))]))

	// Validate opcode
	if payload[0] != serverpackets.InitOpcode {
		slog.Error("invalid opcode", "opcode", fmt.Sprintf("0x%02X", payload[0]))
		return
	}

	if len(payload) < 186 {
		slog.Error("Init packet too short", "len", len(payload))
		return
	}

	sessionID := int32(binary.LittleEndian.Uint32(payload[1:5]))
	fmt.Printf("SessionID: 0x%08X\n", sessionID)

	// Scrambled RSA modulus (128 bytes at offset 9)
	scrambledModulus := payload[9 : 9+128]
	_ = scrambledModulus // not needed for this test

	// Blowfish key (16 bytes at offset 169)
	blowfishKey := payload[169 : 169+16]
	fmt.Printf("Blowfish key: %s\n", hex.EncodeToString(blowfishKey))

	// Инициализируем LoginEncryption с dynamic key
	enc, err := crypto.NewLoginEncryption(blowfishKey)
	if err != nil {
		slog.Error("failed to create login encryption", "err", err)
		return
	}

	fmt.Println("\n=== Sending AuthGameGuard ===")

	// Создаём AuthGameGuard пакет: opcode (1) + sessionID (4) + data1-4 (16) = 21 bytes
	writeBuf := make([]byte, 2+64) // 2-byte header + payload + padding
	authGGPayload := writeBuf[2 : 2+21]
	authGGPayload[0] = 0x07 // opcode
	binary.LittleEndian.PutUint32(authGGPayload[1:], uint32(sessionID))
	// data1-4 = 0 (already zeroed)

	fmt.Printf("Plaintext AuthGameGuard: %s\n", hex.EncodeToString(authGGPayload))

	// Шифруем как клиент (appendChecksum + dynamicCipher)
	encSize, err := enc.EncryptPacketClient(writeBuf, 2, 21)
	if err != nil {
		slog.Error("failed to encrypt AuthGameGuard", "err", err)
		return
	}

	fmt.Printf("Encrypted AuthGameGuard size: %d bytes\n", encSize)
	fmt.Printf("Encrypted AuthGameGuard: %s\n", hex.EncodeToString(writeBuf[2:2+encSize]))

	// Записываем length header + encrypted packet
	totalLen = 2 + encSize
	binary.LittleEndian.PutUint16(writeBuf[:2], uint16(totalLen))

	if _, err := conn.Write(writeBuf[:totalLen]); err != nil {
		slog.Error("failed to write AuthGameGuard", "err", err)
		return
	}

	fmt.Println("AuthGameGuard sent")

	// Читаем GGAuth ответ
	fmt.Println("\n=== Reading GGAuth response ===")

	// Читаем header
	if _, err := io.ReadFull(conn, header[:]); err != nil {
		slog.Error("failed to read GGAuth header", "err", err)
		return
	}

	totalLen = int(binary.LittleEndian.Uint16(header[:]))
	fmt.Printf("GGAuth packet length: %d bytes\n", totalLen)

	if totalLen < 2 {
		slog.Error("invalid GGAuth length", "len", totalLen)
		return
	}

	payloadLen = totalLen - 2
	payload = buf[:payloadLen]
	if _, err := io.ReadFull(conn, payload); err != nil {
		slog.Error("failed to read GGAuth payload", "err", err)
		return
	}

	fmt.Printf("Encrypted GGAuth: %s\n", hex.EncodeToString(payload[:min(32, len(payload))]))

	// Расшифровываем GGAuth (dynamicCipher + verifyChecksum)
	ok, err := enc.DecryptPacket(payload, 0, payloadLen)
	if err != nil {
		slog.Error("failed to decrypt GGAuth", "err", err)
		return
	}

	if !ok {
		slog.Warn("GGAuth checksum verification FAILED")
	} else {
		slog.Info("GGAuth checksum verification OK")
	}

	fmt.Printf("Decrypted GGAuth: %s\n", hex.EncodeToString(payload[:min(32, len(payload))]))

	if len(payload) < 1 {
		slog.Error("GGAuth packet too short")
		return
	}

	opcode := payload[0]
	fmt.Printf("Opcode: 0x%02X (expected 0x0B)\n", opcode)

	if opcode != 0x0B {
		slog.Error("unexpected opcode", "opcode", fmt.Sprintf("0x%02X", opcode))
		return
	}

	fmt.Println("\n=== SUCCESS ===")
	fmt.Println("GGAuth handshake completed successfully!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
