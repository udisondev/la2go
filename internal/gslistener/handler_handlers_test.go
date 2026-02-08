package gslistener

import (
	"context"
	"encoding/binary"
	"math/big"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/login"
)

// Вспомогательные функции для создания тестовых пакетов

func encodeUTF16LE(s string) []byte {
	runes := []rune(s)
	encoded := make([]byte, 0, (len(runes)+1)*2)

	for _, r := range runes {
		encoded = append(encoded, byte(r), byte(r>>8))
	}

	// Null terminator
	encoded = append(encoded, 0x00, 0x00)
	return encoded
}

func makeBlowFishKeyPacket(rsaKeyPair *crypto.RSAKeyPair) []byte {
	// Генерируем 40-байтовый ключ
	blowfishKey := make([]byte, 40)
	for i := range blowfishKey {
		blowfishKey[i] = byte(i + 1)
	}

	// Для RSA-512 нужно зашифровать в 64 байта
	// Паддим 40 байт до 64, добавляя 24 нуля в начало
	pubKey := rsaKeyPair.PrivateKey.PublicKey
	keySize := 64 // RSA-512

	// Создаём блок данных: [24 нуля][40 байт ключа]
	// При расшифровке handler должен взять последние 40 байт
	plaintext := make([]byte, keySize)
	copy(plaintext[keySize-40:], blowfishKey)

	// m^e mod n
	m := new(big.Int).SetBytes(plaintext)
	e := big.NewInt(int64(pubKey.E))
	c := new(big.Int).Exp(m, e, pubKey.N)

	encrypted := c.Bytes()
	if len(encrypted) < keySize {
		padded := make([]byte, keySize)
		copy(padded[keySize-len(encrypted):], encrypted)
		encrypted = padded
	}

	// Формируем пакет: size(int32) + encryptedKey
	buf := make([]byte, 4+len(encrypted))
	binary.LittleEndian.PutUint32(buf[0:4], uint32(len(encrypted)))
	copy(buf[4:], encrypted)

	return buf
}

func makeGameServerAuthPacket(id byte, acceptAlt bool, port int16, maxPlayers int32, hexID []byte) []byte {
	buf := make([]byte, 0, 256)

	// id
	buf = append(buf, id)

	// acceptAlternate
	if acceptAlt {
		buf = append(buf, 0x01)
	} else {
		buf = append(buf, 0x00)
	}

	// reserved
	buf = append(buf, 0x00)

	// maxPlayers (uint16, 2 bytes) — СНАЧАЛА maxPlayers
	maxPlayersBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(maxPlayersBytes, uint16(maxPlayers))
	buf = append(buf, maxPlayersBytes...)

	// port (uint16, 2 bytes) — ПОТОМ port
	portBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(portBytes, uint16(port))
	buf = append(buf, portBytes...)

	// gameHosts (empty list, null-terminated)
	buf = append(buf, 0x00)

	// hexID (фиксированный размер 32 байта)
	hexIDBytes := make([]byte, 32)
	copy(hexIDBytes, hexID)
	buf = append(buf, hexIDBytes...)

	return buf
}

func makePlayerAuthRequestPacket(account string, sessionKey login.SessionKey) []byte {
	buf := make([]byte, 0, 256)

	// account (UTF-16LE null-terminated)
	buf = append(buf, encodeUTF16LE(account)...)

	// playOkID1
	playOk1Bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(playOk1Bytes, uint32(sessionKey.PlayOkID1))
	buf = append(buf, playOk1Bytes...)

	// playOkID2
	playOk2Bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(playOk2Bytes, uint32(sessionKey.PlayOkID2))
	buf = append(buf, playOk2Bytes...)

	// loginOkID1
	loginOk1Bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(loginOk1Bytes, uint32(sessionKey.LoginOkID1))
	buf = append(buf, loginOk1Bytes...)

	// loginOkID2
	loginOk2Bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(loginOk2Bytes, uint32(sessionKey.LoginOkID2))
	buf = append(buf, loginOk2Bytes...)

	return buf
}

func makePlayerInGamePacket(accounts []string) []byte {
	buf := make([]byte, 0, 256)

	// count
	countBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(countBytes, uint16(len(accounts)))
	buf = append(buf, countBytes...)

	// accounts
	for _, account := range accounts {
		buf = append(buf, encodeUTF16LE(account)...)
	}

	return buf
}

func makePlayerLogoutPacket(account string) []byte {
	return encodeUTF16LE(account)
}

func makeServerStatusPacket(attrs map[int32]int32) []byte {
	buf := make([]byte, 0, 256)

	// count
	countBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(countBytes, uint32(len(attrs)))
	buf = append(buf, countBytes...)

	// attributes
	for id, value := range attrs {
		idBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(idBytes, uint32(id))
		buf = append(buf, idBytes...)

		valueBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(valueBytes, uint32(value))
		buf = append(buf, valueBytes...)
	}

	return buf
}


// Тесты

func TestHandleBlowFishKey(t *testing.T) {
	var database *db.DB
	gsTable := gameserver.NewGameServerTable(database)
	sessionManager := login.NewSessionManager()
	handler := NewHandler(database, gsTable, sessionManager)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	ctx := context.Background()
	responseBuf := make([]byte, 1024)

	// Устанавливаем состояние CONNECTED
	conn.SetState(gameserver.GSStateConnected)

	// Создаём пакет BlowFishKey
	body := makeBlowFishKeyPacket(rsaKey)

	// Вызываем handler
	n, ok, err := handleBlowFishKey(ctx, handler, conn, body, responseBuf)
	require.NoError(t, err)
	assert.True(t, ok, "connection should stay open")
	assert.Equal(t, 0, n, "no response should be sent")

	// Проверяем, что состояние изменилось
	assert.Equal(t, gameserver.GSStateBFConnected, conn.State())

	// Проверяем, что cipher изменился (не равен дефолтному)
	// Это косвенная проверка через факт успешного выполнения
}

func TestHandleGameServerAuth_NewServer(t *testing.T) {
	var database *db.DB
	gsTable := gameserver.NewGameServerTable(database)
	sessionManager := login.NewSessionManager()
	handler := NewHandler(database, gsTable, sessionManager)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	ctx := context.Background()
	responseBuf := make([]byte, 1024)

	// Устанавливаем состояние BF_CONNECTED
	conn.SetState(gameserver.GSStateBFConnected)

	// Создаём пакет GameServerAuth для нового сервера
	hexID := make([]byte, 32)
	copy(hexID, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})
	body := makeGameServerAuthPacket(1, false, 7777, 100, hexID)

	// Вызываем handler
	n, ok, err := handleGameServerAuth(ctx, handler, conn, body, responseBuf)
	require.NoError(t, err)
	assert.True(t, ok, "connection should stay open")
	assert.Greater(t, n, 0, "AuthResponse should be sent")

	// Проверяем, что состояние изменилось на AUTHED
	assert.Equal(t, gameserver.GSStateAuthed, conn.State())

	// Проверяем, что сервер зарегистрирован
	info, exists := gsTable.GetByID(1)
	assert.True(t, exists)
	assert.NotNil(t, info)
	assert.True(t, info.IsAuthed())
	assert.Equal(t, 7777, info.Port())
	assert.Equal(t, 100, info.MaxPlayers())

	// Проверяем opcode AuthResponse (0x02)
	assert.Equal(t, byte(0x02), responseBuf[0])
}

func TestHandleGameServerAuth_ExistingServer_ValidHexID(t *testing.T) {
	var database *db.DB
	gsTable := gameserver.NewGameServerTable(database)
	sessionManager := login.NewSessionManager()
	handler := NewHandler(database, gsTable, sessionManager)

	// Предварительно регистрируем сервер
	hexID := make([]byte, 32)
	copy(hexID, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})
	existingInfo := gameserver.NewGameServerInfo(1, hexID)
	gsTable.Register(1, existingInfo)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	ctx := context.Background()
	responseBuf := make([]byte, 1024)

	conn.SetState(gameserver.GSStateBFConnected)

	// Пытаемся зарегистрироваться с тем же hexID
	body := makeGameServerAuthPacket(1, false, 7777, 100, hexID)

	n, ok, err := handleGameServerAuth(ctx, handler, conn, body, responseBuf)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Greater(t, n, 0)

	// Должно быть успешно
	assert.Equal(t, gameserver.GSStateAuthed, conn.State())
	assert.True(t, existingInfo.IsAuthed())
}

func TestHandleGameServerAuth_WrongHexID_NoAlternative(t *testing.T) {
	var database *db.DB
	gsTable := gameserver.NewGameServerTable(database)
	sessionManager := login.NewSessionManager()
	handler := NewHandler(database, gsTable, sessionManager)

	// Регистрируем сервер с одним hexID
	correctHexID := make([]byte, 32)
	copy(correctHexID, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})
	existingInfo := gameserver.NewGameServerInfo(1, correctHexID)
	gsTable.Register(1, existingInfo)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	ctx := context.Background()
	responseBuf := make([]byte, 1024)

	conn.SetState(gameserver.GSStateBFConnected)

	// Пытаемся зарегистрироваться с неправильным hexID
	wrongHexID := make([]byte, 32)
	for i := range wrongHexID {
		wrongHexID[i] = 0xFF
	}
	body := makeGameServerAuthPacket(1, false, 7777, 100, wrongHexID)

	n, ok, err := handleGameServerAuth(ctx, handler, conn, body, responseBuf)
	require.NoError(t, err)
	assert.False(t, ok, "connection should close")
	assert.Greater(t, n, 0)

	// Проверяем opcode LoginServerFail (0x01)
	assert.Equal(t, byte(0x01), responseBuf[0])
	// Проверяем reason: ReasonWrongHexID (3)
	assert.Equal(t, byte(gameserver.ReasonWrongHexID), responseBuf[1])
}

func TestHandleGameServerAuth_AlreadyAuthenticated(t *testing.T) {
	var database *db.DB
	gsTable := gameserver.NewGameServerTable(database)
	sessionManager := login.NewSessionManager()
	handler := NewHandler(database, gsTable, sessionManager)

	hexID := make([]byte, 32)
	copy(hexID, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})
	existingInfo := gameserver.NewGameServerInfo(1, hexID)
	existingInfo.SetAuthed(true) // Уже подключен
	gsTable.Register(1, existingInfo)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	ctx := context.Background()
	responseBuf := make([]byte, 1024)

	conn.SetState(gameserver.GSStateBFConnected)

	body := makeGameServerAuthPacket(1, false, 7777, 100, hexID)

	n, ok, err := handleGameServerAuth(ctx, handler, conn, body, responseBuf)
	require.NoError(t, err)
	assert.False(t, ok, "connection should close")
	assert.Greater(t, n, 0)

	// Проверяем opcode LoginServerFail (0x01)
	assert.Equal(t, byte(0x01), responseBuf[0])
	// Проверяем reason: ReasonAlreadyLoggedIn (7)
	assert.Equal(t, byte(gameserver.ReasonAlreadyLoggedIn), responseBuf[1])
}

func TestHandlePlayerAuthRequest_ValidSession(t *testing.T) {
	var database *db.DB
	gsTable := gameserver.NewGameServerTable(database)
	sessionManager := login.NewSessionManager()
	handler := NewHandler(database, gsTable, sessionManager)

	// Сохраняем сессию
	account := "testuser"
	sessionKey := login.SessionKey{
		LoginOkID1: 12345,
		LoginOkID2: 67890,
		PlayOkID1:  11111,
		PlayOkID2:  22222,
	}
	sessionManager.Store(account, sessionKey, nil)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	ctx := context.Background()
	responseBuf := make([]byte, 1024)

	conn.SetState(gameserver.GSStateAuthed)

	// Создаём пакет PlayerAuthRequest с правильными ключами
	body := makePlayerAuthRequestPacket(account, sessionKey)

	n, ok, err := handlePlayerAuthRequest(ctx, handler, conn, body, responseBuf)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Greater(t, n, 0)

	// Проверяем opcode PlayerAuthResponse (0x03)
	assert.Equal(t, byte(0x03), responseBuf[0])

	// Проверяем, что account в ответе (UTF-16LE)
	expectedAccount := encodeUTF16LE(account)
	assert.Equal(t, expectedAccount, responseBuf[1:1+len(expectedAccount)])

	// Проверяем result (должно быть 1 = success)
	resultPos := 1 + len(expectedAccount)
	assert.Equal(t, byte(1), responseBuf[resultPos])

	// Проверяем, что сессия удалена
	assert.Equal(t, 0, sessionManager.Count())
}

func TestHandlePlayerAuthRequest_InvalidSession(t *testing.T) {
	var database *db.DB
	gsTable := gameserver.NewGameServerTable(database)
	sessionManager := login.NewSessionManager()
	handler := NewHandler(database, gsTable, sessionManager)

	account := "testuser"
	wrongSessionKey := login.SessionKey{
		LoginOkID1: 99999,
		LoginOkID2: 88888,
		PlayOkID1:  77777,
		PlayOkID2:  66666,
	}

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	ctx := context.Background()
	responseBuf := make([]byte, 1024)

	conn.SetState(gameserver.GSStateAuthed)

	body := makePlayerAuthRequestPacket(account, wrongSessionKey)

	n, ok, err := handlePlayerAuthRequest(ctx, handler, conn, body, responseBuf)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Greater(t, n, 0)

	// Проверяем opcode PlayerAuthResponse (0x03)
	assert.Equal(t, byte(0x03), responseBuf[0])

	// Проверяем result (должно быть 0 = failure)
	expectedAccount := encodeUTF16LE(account)
	resultPos := 1 + len(expectedAccount)
	assert.Equal(t, byte(0), responseBuf[resultPos])
}

func TestHandlePlayerInGame(t *testing.T) {
	var database *db.DB
	gsTable := gameserver.NewGameServerTable(database)
	sessionManager := login.NewSessionManager()
	handler := NewHandler(database, gsTable, sessionManager)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	ctx := context.Background()
	responseBuf := make([]byte, 1024)

	conn.SetState(gameserver.GSStateAuthed)

	// Создаём пакет с 3 игроками
	accounts := []string{"user1", "user2", "user3"}
	body := makePlayerInGamePacket(accounts)

	n, ok, err := handlePlayerInGame(ctx, handler, conn, body, responseBuf)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 0, n, "no response should be sent")

	// Проверяем, что все игроки добавлены
	for _, account := range accounts {
		assert.True(t, conn.HasAccount(account), "account %s should be in online list", account)
	}
}

func TestHandlePlayerLogout(t *testing.T) {
	var database *db.DB
	gsTable := gameserver.NewGameServerTable(database)
	sessionManager := login.NewSessionManager()
	handler := NewHandler(database, gsTable, sessionManager)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	ctx := context.Background()
	responseBuf := make([]byte, 1024)

	conn.SetState(gameserver.GSStateAuthed)

	// Добавляем игрока
	account := "testuser"
	conn.AddAccount(account)
	require.True(t, conn.HasAccount(account))

	// Создаём пакет PlayerLogout
	body := makePlayerLogoutPacket(account)

	n, ok, err := handlePlayerLogout(ctx, handler, conn, body, responseBuf)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 0, n)

	// Проверяем, что игрок удалён
	assert.False(t, conn.HasAccount(account))
}

func TestHandleServerStatus(t *testing.T) {
	var database *db.DB
	gsTable := gameserver.NewGameServerTable(database)
	sessionManager := login.NewSessionManager()
	handler := NewHandler(database, gsTable, sessionManager)

	// Создаём и регистрируем GameServer
	hexID := make([]byte, 32)
	copy(hexID, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})
	gsInfo := gameserver.NewGameServerInfo(1, hexID)
	gsInfo.SetAuthed(true)
	gsTable.Register(1, gsInfo)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	// Привязываем GameServerInfo к соединению
	conn.AttachGameServerInfo(gsInfo)

	ctx := context.Background()
	responseBuf := make([]byte, 1024)

	conn.SetState(gameserver.GSStateAuthed)

	// Создаём пакет ServerStatus со всеми атрибутами
	attrs := map[int32]int32{
		0: 1,    // showingBrackets = true
		1: 0x01, // serverType = NORMAL
		2: 0x01, // status = GOOD
		3: 0x12, // ageLimit = 18
		4: 500,  // maxPlayers = 500
	}
	body := makeServerStatusPacket(attrs)

	n, ok, err := handleServerStatus(ctx, handler, conn, body, responseBuf)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 0, n)

	// Проверяем, что все атрибуты обновлены
	assert.True(t, gsInfo.ShowingBrackets())
	assert.Equal(t, 0x01, gsInfo.ServerType())
	assert.Equal(t, 0x01, gsInfo.Status())
	assert.Equal(t, 0x12, gsInfo.AgeLimit())
	assert.Equal(t, 500, gsInfo.MaxPlayers())
}
