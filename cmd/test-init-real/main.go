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
	// Подключаемся к РЕАЛЬНОМУ LoginServer
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
	fmt.Printf("Init packet total length: %d bytes\n", totalLen)

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

	fmt.Printf("\n=== ENCRYPTED payload ===\n")
	fmt.Printf("Length: %d bytes\n", payloadLen)
	fmt.Printf("First 64 bytes:\n%s\n", hex.Dump(payload[:64]))
	fmt.Printf("Last 64 bytes:\n%s\n", hex.Dump(payload[payloadLen-64:]))

	// Расшифровываем Init пакет (статический Blowfish ТОЛЬКО, БЕЗ decXORPass пока)
	staticCipher, err := crypto.NewBlowfishCipher(crypto.StaticBlowfishKey)
	if err != nil {
		slog.Error("failed to create static cipher", "err", err)
		return
	}

	if err := staticCipher.Decrypt(payload, 0, payloadLen); err != nil {
		slog.Error("failed to decrypt Init", "err", err)
		return
	}

	fmt.Printf("\n=== After Blowfish decrypt (BEFORE decXORPass) ===\n")
	fmt.Printf("First 64 bytes:\n%s\n", hex.Dump(payload[:64]))
	fmt.Printf("Last 64 bytes:\n%s\n", hex.Dump(payload[payloadLen-64:]))

	// Смотрим на offset 169 (blowfish key) ПЕРЕД decXORPass
	fmt.Printf("\n=== Blowfish key area BEFORE decXORPass ===\n")
	fmt.Printf("Offset 169-184 (16 bytes):\n%s\n", hex.Dump(payload[169:185]))

	// Проверяем финальный XOR ключ (последние 4 байта перед padding)
	// Для encryptedSize=192: финальный ключ на offset 184-188
	xorKeyOffset := payloadLen - 8
	fmt.Printf("\n=== XOR key location ===\n")
	fmt.Printf("PayloadLen: %d\n", payloadLen)
	fmt.Printf("XOR key offset (payloadLen-8): %d\n", xorKeyOffset)
	fmt.Printf("XOR key (4 bytes): %s\n", hex.EncodeToString(payload[xorKeyOffset:xorKeyOffset+4]))

	// Применяем decXORPass
	crypto.DecXORPass(payload, 0, payloadLen)

	fmt.Printf("\n=== After decXORPass ===\n")
	fmt.Printf("First 64 bytes:\n%s\n", hex.Dump(payload[:64]))
	fmt.Printf("Last 64 bytes:\n%s\n", hex.Dump(payload[payloadLen-64:]))

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
	fmt.Printf("\n=== Init packet contents ===\n")
	fmt.Printf("SessionID: 0x%08X\n", sessionID)

	// Blowfish key (16 bytes at offset 153 or 169?)
	// Давайте проверим обе позиции
	fmt.Printf("\nBlowfish key at offset 153 (opcode=1, sessionID=4, protocolRev=4, rsaKey=128, ggData=16 = 153):\n")
	fmt.Printf("%s\n", hex.EncodeToString(payload[153:169]))

	fmt.Printf("\nBlowfish key at offset 169 (153 + 16 = 169):\n")
	fmt.Printf("%s\n", hex.EncodeToString(payload[169:185]))

	// Правильный offset для blowfish key:
	// opcode(1) + sessionID(4) + protocolRev(4) + rsaKey(128) + ggData(16) = 153
	blowfishOffset := 153
	blowfishKey := payload[blowfishOffset : blowfishOffset+16]
	fmt.Printf("\n=== FINAL Blowfish key (offset %d) ===\n", blowfishOffset)
	fmt.Printf("%s\n", hex.EncodeToString(blowfishKey))

	fmt.Println("\n=== SUCCESS ===")
}
