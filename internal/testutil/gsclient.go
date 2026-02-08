package testutil

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"net"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/login"
)

// GSClient упрощает написание integration тестов для gslistener.
// Автоматически управляет подключением, шифрованием и чтением/записью пакетов.
type GSClient struct {
	t        testing.TB
	conn     net.Conn
	cipher   *crypto.BlowfishCipher
	readBuf  []byte
	writeBuf []byte

	// Данные из InitLS пакета
	rsaModulus      []byte // RSA-512 modulus (64 bytes)
	protocolVersion int32

	// Server info
	serverID byte

	// Timeout для операций
	timeout time.Duration
}

// NewGSClient создаёт GSClient и подключается к gslistener по указанному адресу.
// Автоматически читает InitLS пакет и сохраняет RSA modulus.
// Использует t.Cleanup() для автоматического закрытия соединения.
func NewGSClient(t testing.TB, addr string) (*GSClient, error) {
	t.Helper()

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("dial gslistener: %w", err)
	}

	client := &GSClient{
		t:        t,
		conn:     conn,
		readBuf:  make([]byte, 4096),
		writeBuf: make([]byte, 4096),
		timeout:  5 * time.Second,
	}

	t.Cleanup(func() {
		_ = client.Close()
	})

	// Автоматически читаем InitLS пакет
	if err := client.ReadInitLS(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("read InitLS: %w", err)
	}

	return client, nil
}

// ReadInitLS читает пакет InitLS (opcode 0x00) от LoginServer.
// Извлекает protocol version и RSA-512 modulus.
func (c *GSClient) ReadInitLS() error {
	c.t.Helper()

	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}

	// InitLS пакет ЗАШИФРОВАН с DefaultGSBlowfishKey
	// Создаём cipher для расшифровки InitLS
	defaultCipher, err := crypto.NewBlowfishCipher(crypto.DefaultGSBlowfishKey)
	if err != nil {
		return fmt.Errorf("create default GS cipher: %w", err)
	}

	var header [2]byte
	if _, err := io.ReadFull(c.conn, header[:]); err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	totalLen := int(binary.LittleEndian.Uint16(header[:]))
	if totalLen < 2 {
		return fmt.Errorf("invalid InitLS packet length: %d", totalLen)
	}

	payloadLen := totalLen - 2
	if payloadLen > len(c.readBuf) {
		return fmt.Errorf("InitLS packet too large: %d", payloadLen)
	}

	payload := c.readBuf[:payloadLen]
	if _, err := io.ReadFull(c.conn, payload); err != nil {
		return fmt.Errorf("read InitLS payload: %w", err)
	}

	// Расшифровываем InitLS пакет
	if err := defaultCipher.Decrypt(payload, 0, payloadLen); err != nil {
		return fmt.Errorf("decrypt InitLS: %w", err)
	}

	// Validate opcode
	if payload[0] != 0x00 {
		return fmt.Errorf("expected InitLS opcode 0x00, got 0x%02X", payload[0])
	}

	// Parse InitLS packet
	// Format: opcode (1) + revision (4) + keySize (4) + rsaModulus (64)
	if len(payload) < 1+4+4+64 {
		return fmt.Errorf("InitLS packet too short: %d", len(payload))
	}

	c.protocolVersion = int32(binary.LittleEndian.Uint32(payload[1:5]))
	keySize := int32(binary.LittleEndian.Uint32(payload[5:9]))

	if keySize != 64 {
		return fmt.Errorf("unexpected RSA key size: %d (expected 64)", keySize)
	}

	// RSA modulus (64 bytes at offset 9)
	c.rsaModulus = make([]byte, 64)
	copy(c.rsaModulus, payload[9:9+64])

	return nil
}

// SendBlowFishKey отправляет пакет BlowFishKey (opcode 0x00) с зашифрованным Blowfish ключом.
// Ключ шифруется RSA-512 и отправляется серверу.
// После успешной отправки переключает cipher на новый Blowfish.
func (c *GSClient) SendBlowFishKey(key []byte) error {
	c.t.Helper()

	if len(key) != 40 {
		return fmt.Errorf("blowfish key must be 40 bytes, got %d", len(key))
	}

	// RSA encrypt key with RSA-512 modulus
	encrypted, err := RSAEncryptWithModulus512(c.rsaModulus, key)
	if err != nil {
		return fmt.Errorf("RSA encrypt blowfish key: %w", err)
	}

	// BlowFishKey packet: opcode (1) + encrypted_key (64)
	payloadOffset := 2
	payload := c.writeBuf[payloadOffset:]
	payload[0] = 0x00 // BlowFishKey opcode
	copy(payload[1:], encrypted)
	payloadLen := 1 + 64

	// Первый пакет от GS должен быть зашифрован с DefaultGSBlowfishKey
	defaultCipher, err := crypto.NewBlowfishCipher(crypto.DefaultGSBlowfishKey)
	if err != nil {
		return fmt.Errorf("create default cipher: %w", err)
	}

	// Append checksum и pad до 8 байт
	checksumSize := payloadLen + 4 // +4 для checksum
	if checksumSize%8 != 0 {
		checksumSize += 8 - (checksumSize % 8)
	}
	for i := payloadLen; i < checksumSize; i++ {
		payload[i] = 0
	}
	crypto.AppendChecksum(payload, 0, checksumSize)

	// Шифруем с DefaultGSBlowfishKey
	if err := defaultCipher.Encrypt(payload, 0, checksumSize); err != nil {
		return fmt.Errorf("encrypt BlowFishKey: %w", err)
	}

	// Write packet
	totalLen := 2 + checksumSize
	binary.LittleEndian.PutUint16(c.writeBuf[:2], uint16(totalLen))

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}

	if _, err := c.conn.Write(c.writeBuf[:totalLen]); err != nil {
		return fmt.Errorf("write BlowFishKey: %w", err)
	}

	// Создаём Blowfish cipher для дальнейшего шифрования
	cipher, err := crypto.NewBlowfishCipher(key)
	if err != nil {
		return fmt.Errorf("create blowfish cipher: %w", err)
	}
	c.cipher = cipher

	return nil
}

// SendGameServerAuth отправляет пакет GameServerAuth (opcode 0x01).
// Регистрирует GameServer с указанным serverID и hexID.
// Параметр acceptAlternate указывает, может ли сервер принять альтернативный ID.
func (c *GSClient) SendGameServerAuth(serverID byte, hexID string, acceptAlternate bool) error {
	c.t.Helper()

	c.serverID = serverID

	// GameServerAuth packet:
	// opcode (1) + serverID (1) + acceptAlternate (1) + reserved (1) +
	// maxPlayers (2) + port (2) + gameHosts (variable, null-terminated) + hexID (32)

	payload := c.writeBuf[2:]
	pos := 0

	payload[pos] = 0x01 // GameServerAuth opcode
	pos++

	payload[pos] = serverID
	pos++

	if acceptAlternate {
		payload[pos] = 0x01
	} else {
		payload[pos] = 0x00
	}
	pos++

	payload[pos] = 0x00 // reserved
	pos++

	// maxPlayers (1000 = 0x03E8 LE)
	binary.LittleEndian.PutUint16(payload[pos:], 1000)
	pos += 2

	// port (7777 = 0x1E61 LE)
	binary.LittleEndian.PutUint16(payload[pos:], 7777)
	pos += 2

	// gameHosts (empty list, null-terminated)
	payload[pos] = 0x00
	pos++

	// hexID (32 bytes)
	hexIDBytes := make([]byte, 32)
	copy(hexIDBytes, hexID)
	copy(payload[pos:], hexIDBytes)
	pos += 32

	// Шифруем пакет с checksum
	if c.cipher == nil {
		return fmt.Errorf("cipher not initialized (call SendBlowFishKey first)")
	}

	payloadLen := pos

	// Append checksum и pad до 8 байт
	checksumSize := payloadLen + 4 // +4 для checksum
	if checksumSize%8 != 0 {
		checksumSize += 8 - (checksumSize % 8)
	}
	for i := payloadLen; i < checksumSize; i++ {
		payload[i] = 0
	}
	crypto.AppendChecksum(payload, 0, checksumSize)

	// Шифруем
	if err := c.cipher.Encrypt(payload, 0, checksumSize); err != nil {
		return fmt.Errorf("encrypt GameServerAuth: %w", err)
	}

	totalLen := 2 + checksumSize
	binary.LittleEndian.PutUint16(c.writeBuf[:2], uint16(totalLen))

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}

	if _, err := c.conn.Write(c.writeBuf[:totalLen]); err != nil {
		return fmt.Errorf("write GameServerAuth: %w", err)
	}

	return nil
}

// ReadAuthResponse читает пакет AuthResponse (opcode 0x02) и возвращает serverID.
// Если сервер отправил LoginServerFail, возвращает ошибку.
func (c *GSClient) ReadAuthResponse() (byte, error) {
	c.t.Helper()

	payload, err := c.readEncryptedPacket()
	if err != nil {
		return 0, fmt.Errorf("read AuthResponse: %w", err)
	}

	if len(payload) < 1 {
		return 0, fmt.Errorf("AuthResponse packet too short")
	}

	opcode := payload[0]

	if opcode == 0x01 { // LoginServerFail
		if len(payload) < 2 {
			return 0, fmt.Errorf("received LoginServerFail (truncated)")
		}
		reason := payload[1]
		return 0, fmt.Errorf("received LoginServerFail, reason: 0x%02X", reason)
	}

	if opcode != 0x02 {
		return 0, fmt.Errorf("expected AuthResponse opcode 0x02, got 0x%02X", opcode)
	}

	if len(payload) < 2 {
		return 0, fmt.Errorf("AuthResponse payload too short")
	}

	serverID := payload[1]
	return serverID, nil
}

// ReadLoginServerFail читает пакет LoginServerFail (opcode 0x01) и возвращает reason code.
// Если получен другой пакет, возвращает ошибку.
func (c *GSClient) ReadLoginServerFail() (byte, error) {
	c.t.Helper()

	payload, err := c.readEncryptedPacket()
	if err != nil {
		return 0, fmt.Errorf("read LoginServerFail: %w", err)
	}

	if len(payload) < 2 {
		return 0, fmt.Errorf("LoginServerFail packet too short: %d", len(payload))
	}

	if payload[0] != 0x01 {
		return 0, fmt.Errorf("expected LoginServerFail opcode 0x01, got 0x%02X", payload[0])
	}

	return payload[1], nil
}

// SendPlayerAuthRequest отправляет пакет PlayerAuthRequest (opcode 0x05).
// Запрашивает валидацию session key для указанного account.
func (c *GSClient) SendPlayerAuthRequest(account string, sk login.SessionKey) error {
	c.t.Helper()

	// PlayerAuthRequest packet:
	// opcode (1) + account (UTF-16LE null-terminated) +
	// playOkID1 (4) + playOkID2 (4) + loginOkID1 (4) + loginOkID2 (4)

	payload := c.writeBuf[2:]
	pos := 0

	payload[pos] = 0x05 // PlayerAuthRequest opcode
	pos++

	// Account (UTF-16LE null-terminated)
	accountRunes := utf16.Encode([]rune(account))
	for _, r := range accountRunes {
		payload[pos] = byte(r)
		payload[pos+1] = byte(r >> 8)
		pos += 2
	}
	// Null terminator
	payload[pos] = 0
	payload[pos+1] = 0
	pos += 2

	// Session key
	binary.LittleEndian.PutUint32(payload[pos:], uint32(sk.PlayOkID1))
	pos += 4
	binary.LittleEndian.PutUint32(payload[pos:], uint32(sk.PlayOkID2))
	pos += 4
	binary.LittleEndian.PutUint32(payload[pos:], uint32(sk.LoginOkID1))
	pos += 4
	binary.LittleEndian.PutUint32(payload[pos:], uint32(sk.LoginOkID2))
	pos += 4

	return c.sendEncryptedPacket(payload[:pos])
}

// ReadPlayerAuthResponse читает пакет PlayerAuthResponse (opcode 0x03).
// Возвращает account и result (true = success, false = failure).
func (c *GSClient) ReadPlayerAuthResponse() (account string, result bool, err error) {
	c.t.Helper()

	payload, err := c.readEncryptedPacket()
	if err != nil {
		return "", false, fmt.Errorf("read PlayerAuthResponse: %w", err)
	}

	if len(payload) < 1 {
		return "", false, fmt.Errorf("PlayerAuthResponse packet too short")
	}

	if payload[0] != 0x03 {
		return "", false, fmt.Errorf("expected PlayerAuthResponse opcode 0x03, got 0x%02X", payload[0])
	}

	// Parse account (UTF-16LE null-terminated)
	pos := 1
	accountRunes := []uint16{}
	for pos+1 < len(payload) {
		r := binary.LittleEndian.Uint16(payload[pos : pos+2])
		pos += 2
		if r == 0 {
			break
		}
		accountRunes = append(accountRunes, r)
	}

	account = string(utf16.Decode(accountRunes))

	// Result byte
	if pos >= len(payload) {
		return "", false, fmt.Errorf("PlayerAuthResponse missing result byte")
	}

	result = payload[pos] == 1

	return account, result, nil
}

// SendPlayerInGame отправляет пакет PlayerInGame (opcode 0x03).
func (c *GSClient) SendPlayerInGame(account string) error {
	c.t.Helper()

	payload := c.writeBuf[2:]
	pos := 0

	payload[pos] = 0x03 // PlayerInGame opcode
	pos++

	// Account (UTF-16LE null-terminated)
	accountRunes := utf16.Encode([]rune(account))
	for _, r := range accountRunes {
		payload[pos] = byte(r)
		payload[pos+1] = byte(r >> 8)
		pos += 2
	}
	// Null terminator
	payload[pos] = 0
	payload[pos+1] = 0
	pos += 2

	return c.sendEncryptedPacket(payload[:pos])
}

// SendPlayerLogout отправляет пакет PlayerLogout (opcode 0x04).
func (c *GSClient) SendPlayerLogout(account string) error {
	c.t.Helper()

	payload := c.writeBuf[2:]
	pos := 0

	payload[pos] = 0x04 // PlayerLogout opcode
	pos++

	// Account (UTF-16LE null-terminated)
	accountRunes := utf16.Encode([]rune(account))
	for _, r := range accountRunes {
		payload[pos] = byte(r)
		payload[pos+1] = byte(r >> 8)
		pos += 2
	}
	// Null terminator
	payload[pos] = 0
	payload[pos+1] = 0
	pos += 2

	return c.sendEncryptedPacket(payload[:pos])
}

// SendServerStatus отправляет пакет ServerStatus (opcode 0x06).
// attributes — map[attributeID]value (например, map[0x01]maxPlayers).
func (c *GSClient) SendServerStatus(serverID byte, attributes map[int]int32) error {
	c.t.Helper()

	payload := c.writeBuf[2:]
	pos := 0

	payload[pos] = 0x06 // ServerStatus opcode
	pos++

	// Server ID
	payload[pos] = serverID
	pos++

	// Attributes count
	payload[pos] = byte(len(attributes))
	pos++

	// Attributes
	for attrID, value := range attributes {
		payload[pos] = byte(attrID)
		pos++
		binary.LittleEndian.PutUint32(payload[pos:], uint32(value))
		pos += 4
	}

	return c.sendEncryptedPacket(payload[:pos])
}

// CompleteRegistration — helper для быстрой регистрации GameServer.
// Выполняет полный цикл: ReadInitLS → SendBlowFishKey → SendGameServerAuth → ReadAuthResponse.
func (c *GSClient) CompleteRegistration(serverID byte, hexID string) error {
	c.t.Helper()

	// Генерируем случайный Blowfish key (40 bytes)
	bfKey := make([]byte, 40)
	if _, err := rand.Read(bfKey); err != nil {
		return fmt.Errorf("generate blowfish key: %w", err)
	}

	// Отправляем BlowFishKey
	if err := c.SendBlowFishKey(bfKey); err != nil {
		return fmt.Errorf("send BlowFishKey: %w", err)
	}

	// Отправляем GameServerAuth (acceptAlternate=false)
	if err := c.SendGameServerAuth(serverID, hexID, false); err != nil {
		return fmt.Errorf("send GameServerAuth: %w", err)
	}

	// Читаем AuthResponse
	receivedID, err := c.ReadAuthResponse()
	if err != nil {
		return fmt.Errorf("read AuthResponse: %w", err)
	}

	if receivedID != serverID {
		return fmt.Errorf("server ID mismatch: expected %d, got %d", serverID, receivedID)
	}

	return nil
}

// Close закрывает соединение с сервером.
func (c *GSClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// sendEncryptedPacket шифрует и отправляет payload.
func (c *GSClient) sendEncryptedPacket(payload []byte) error {
	c.t.Helper()

	if c.cipher == nil {
		return fmt.Errorf("cipher not initialized")
	}

	payloadLen := len(payload)

	// Копируем payload в writeBuf (начиная с offset 2 для length header)
	buf := c.writeBuf[2:]
	copy(buf, payload)

	// Append checksum и pad до 8 байт
	checksumSize := payloadLen + 4 // +4 для checksum
	if checksumSize%8 != 0 {
		checksumSize += 8 - (checksumSize % 8)
	}
	for i := payloadLen; i < checksumSize; i++ {
		buf[i] = 0
	}
	crypto.AppendChecksum(buf, 0, checksumSize)

	if err := c.cipher.Encrypt(buf, 0, checksumSize); err != nil {
		return fmt.Errorf("encrypt packet: %w", err)
	}

	totalLen := 2 + checksumSize
	binary.LittleEndian.PutUint16(c.writeBuf[:2], uint16(totalLen))

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}

	if _, err := c.conn.Write(c.writeBuf[:totalLen]); err != nil {
		return fmt.Errorf("write packet: %w", err)
	}

	return nil
}

// readEncryptedPacket читает и расшифровывает пакет.
func (c *GSClient) readEncryptedPacket() ([]byte, error) {
	c.t.Helper()

	if c.cipher == nil {
		return nil, fmt.Errorf("cipher not initialized")
	}

	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return nil, fmt.Errorf("set read deadline: %w", err)
	}

	var header [2]byte
	if _, err := io.ReadFull(c.conn, header[:]); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	totalLen := int(binary.LittleEndian.Uint16(header[:]))
	if totalLen < 2 {
		return nil, fmt.Errorf("invalid packet length: %d", totalLen)
	}

	payloadLen := totalLen - 2
	if payloadLen > len(c.readBuf) {
		return nil, fmt.Errorf("packet too large: %d", payloadLen)
	}

	payload := c.readBuf[:payloadLen]
	if _, err := io.ReadFull(c.conn, payload); err != nil {
		return nil, fmt.Errorf("read payload: %w", err)
	}

	if err := c.cipher.Decrypt(payload, 0, payloadLen); err != nil {
		return nil, fmt.Errorf("decrypt packet: %w", err)
	}

	return payload, nil
}

// RSAEncryptWithModulus512 шифрует plaintext используя RSA-512 modulus (64 bytes).
// Plaintext должен быть <= 40 bytes (для Blowfish key).
// Добавляет padding для RSA-512.
func RSAEncryptWithModulus512(modulus, plaintext []byte) ([]byte, error) {
	if len(modulus) != 64 {
		return nil, fmt.Errorf("RSA-512 modulus must be 64 bytes, got %d", len(modulus))
	}

	if len(plaintext) > 40 {
		return nil, fmt.Errorf("plaintext too large for RSA-512, got %d bytes", len(plaintext))
	}

	// Pad plaintext to 64 bytes (left-pad with zeros)
	padded := make([]byte, 64)
	copy(padded[64-len(plaintext):], plaintext)

	// Конвертируем modulus в big.Int
	n := new(big.Int).SetBytes(modulus)

	// Public exponent (hardcoded в L2 = 65537)
	e := big.NewInt(65537)

	// RSA encrypt: ciphertext = plaintext^e mod n
	m := new(big.Int).SetBytes(padded)
	c := new(big.Int).Exp(m, e, n)

	ciphertext := c.Bytes()

	// Pad to 64 bytes if needed
	if len(ciphertext) < 64 {
		paddedCipher := make([]byte, 64)
		copy(paddedCipher[64-len(ciphertext):], ciphertext)
		ciphertext = paddedCipher
	}

	return ciphertext, nil
}
