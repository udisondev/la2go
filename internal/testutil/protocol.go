package testutil

import (
	"encoding/binary"
	"unicode/utf16"

	"github.com/udisondev/la2go/internal/constants"
)

// EncodeUTF16LE кодирует строку в UTF-16LE с null terminator.
func EncodeUTF16LE(s string) []byte {
	runes := []rune(s)
	encoded := utf16.Encode(runes)

	// 2 байта на символ + 2 байта null terminator
	result := make([]byte, (len(encoded)+1)*2)

	for i, r := range encoded {
		binary.LittleEndian.PutUint16(result[i*2:], r)
	}

	// Null terminator (0x00 0x00)
	binary.LittleEndian.PutUint16(result[len(encoded)*2:], 0)

	return result
}

// MakeBlowFishKeyPacket создаёт пакет BlowFishKey (GS→LS).
func MakeBlowFishKeyPacket(key []byte) []byte {
	if len(key) != constants.BlowfishKeySize {
		panic("blowfish key must be 16 bytes")
	}

	packet := make([]byte, 1+constants.BlowfishKeySize)
	packet[0] = 0x00 // BlowFishKey opcode
	copy(packet[1:], key)

	return packet
}

// MakeGameServerAuthPacket создаёт пакет GameServerAuth (GS→LS).
func MakeGameServerAuthPacket(serverID byte, hexID string) []byte {
	// Opcode (1) + ServerID (1) + AcceptAlternate (1) + Reserved (1) + MaxPlayers (2) + Port (2) + GameHosts (variable)
	// Minimal: opcode + serverID + acceptAlternate + reserved + maxPlayers + port + 0-terminated host list

	packet := []byte{
		0x01,                         // GameServerAuth opcode
		serverID,                     // ServerID
		0x00,                         // AcceptAlternate
		0x00,                         // Reserved
		constants.TestMaxPlayersLE1,  // MaxPlayers low byte (1000 = 0x03E8 LE)
		constants.TestMaxPlayersLE2,  // MaxPlayers high byte
		constants.TestServerPortLE1,  // Port low byte (8180 = 0x1FF4 LE)
		constants.TestServerPortLE2,  // Port high byte
		// GameHosts (пустой список, null-terminated)
		0x00,
	}

	// HexID (64 символа = 32 байта)
	hexIDBytes := make([]byte, 32)
	copy(hexIDBytes, hexID)
	packet = append(packet, hexIDBytes...)

	return packet
}

// MakePlayerAuthRequestPacket создаёт пакет PlayerAuthRequest (GS→LS).
func MakePlayerAuthRequestPacket(account string, playOkID1, playOkID2, loginOkID1, loginOkID2 int32) []byte {
	// Opcode (1) + Account (UTF-16LE) + PlayOkID1 (4) + PlayOkID2 (4) + LoginOkID1 (4) + LoginOkID2 (4)
	accountBytes := EncodeUTF16LE(account)

	packet := make([]byte, 1+len(accountBytes)+16)
	packet[0] = 0x05 // PlayerAuthRequest opcode

	offset := 1
	copy(packet[offset:], accountBytes)
	offset += len(accountBytes)

	binary.LittleEndian.PutUint32(packet[offset:], uint32(playOkID1))
	offset += 4
	binary.LittleEndian.PutUint32(packet[offset:], uint32(playOkID2))
	offset += 4
	binary.LittleEndian.PutUint32(packet[offset:], uint32(loginOkID1))
	offset += 4
	binary.LittleEndian.PutUint32(packet[offset:], uint32(loginOkID2))

	return packet
}

// MakePlayerInGamePacket создаёт пакет PlayerInGame (GS→LS).
func MakePlayerInGamePacket(account string) []byte {
	accountBytes := EncodeUTF16LE(account)
	packet := make([]byte, 1+len(accountBytes))
	packet[0] = 0x03 // PlayerInGame opcode
	copy(packet[1:], accountBytes)
	return packet
}

// MakePlayerLogoutPacket создаёт пакет PlayerLogout (GS→LS).
func MakePlayerLogoutPacket(account string) []byte {
	accountBytes := EncodeUTF16LE(account)
	packet := make([]byte, 1+len(accountBytes))
	packet[0] = 0x04 // PlayerLogout opcode
	copy(packet[1:], accountBytes)
	return packet
}

// MakeServerStatusPacket создаёт пакет ServerStatus (GS→LS).
func MakeServerStatusPacket(serverID byte, statusCode int32) []byte {
	packet := make([]byte, 1+1+4)
	packet[0] = 0x06 // ServerStatus opcode
	packet[1] = serverID
	binary.LittleEndian.PutUint32(packet[2:], uint32(statusCode))
	return packet
}
