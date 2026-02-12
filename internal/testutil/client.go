package testutil

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"math/rand/v2"
	"net"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/login/serverpackets"
	"github.com/udisondev/la2go/internal/protocol"
)

// LoginClient упрощает написание integration тестов для LoginServer.
// Автоматически управляет подключением, шифрованием и чтением/записью пакетов.
type LoginClient struct {
	t       testing.TB
	conn    net.Conn
	enc     *crypto.LoginEncryption
	readBuf []byte
	writeBuf []byte

	// Данные из Init пакета
	sessionID       int32
	rsaModulus      []byte // unscrambled modulus (128 bytes)
	blowfishKey     []byte
	protocolVersion uint32

	// Timeout для операций
	timeout time.Duration
}

// NewLoginClient создаёт LoginClient и подключается к LoginServer по указанному адресу.
// Автоматически читает Init пакет и сохраняет sessionID, RSA modulus, Blowfish key.
// Использует t.Cleanup() для автоматического закрытия соединения.
func NewLoginClient(t testing.TB, addr string) (*LoginClient, error) {
	t.Helper()

	// Retry dial с экспоненциальным бэкофф + jitter: macOS TCP стек может не успевать
	// освободить порты при массовых подключениях
	var conn net.Conn
	var err error
	for attempt := range 10 {
		conn, err = net.DialTimeout("tcp", addr, 5*time.Second)
		if err == nil {
			break
		}
		if attempt < 9 {
			base := time.Duration(20<<min(attempt, 6)) * time.Millisecond // 20, 40, 80, ..., 1280ms
			jitter := time.Duration(rand.IntN(int(base/2)+1)) * time.Millisecond
			time.Sleep(base + jitter)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("dial login server: %w", err)
	}

	// SO_LINGER=0: немедленный RST вместо TIME_WAIT, предотвращает исчерпание эфемерных портов в тестах
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetLinger(0); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("set linger: %w", err)
		}
	}

	client := &LoginClient{
		t:        t,
		conn:     conn,
		readBuf:  make([]byte, 4096),
		writeBuf: make([]byte, 4096),
		timeout:  5 * time.Second,
	}

	t.Cleanup(func() {
		_ = client.Close()
	})

	// Автоматически читаем Init пакет
	if err := client.ReadInitPacket(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("read init packet: %w", err)
	}

	return client, nil
}

// ReadInitPacket читает Init пакет (opcode 0x00) от сервера.
// Извлекает sessionID, scrambled RSA modulus, Blowfish key.
// Инициализирует LoginEncryption с dynamic Blowfish key.
func (c *LoginClient) ReadInitPacket() error {
	c.t.Helper()

	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}

	// Init пакет ЗАШИФРОВАН с static Blowfish key + encXORPass
	var header [2]byte
	if _, err := io.ReadFull(c.conn, header[:]); err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	totalLen := int(binary.LittleEndian.Uint16(header[:]))
	if totalLen < 2 {
		return fmt.Errorf("invalid init packet length: %d", totalLen)
	}

	payloadLen := totalLen - 2
	if payloadLen > len(c.readBuf) {
		return fmt.Errorf("init packet too large: %d", payloadLen)
	}

	payload := c.readBuf[:payloadLen]
	if _, err := io.ReadFull(c.conn, payload); err != nil {
		return fmt.Errorf("read init payload: %w", err)
	}

	// Расшифровываем Init пакет (только Blowfish, БЕЗ проверки checksum т.к. используется encXORPass)
	staticCipher, err := crypto.NewBlowfishCipher(crypto.StaticBlowfishKey)
	if err != nil {
		return fmt.Errorf("create static cipher: %w", err)
	}
	if err := staticCipher.Decrypt(payload, 0, payloadLen); err != nil {
		return fmt.Errorf("blowfish decrypt init: %w", err)
	}

	// Применяем decXORPass для завершения расшифровки
	DecXORPass(payload, 0, payloadLen)

	// Validate opcode
	if payload[0] != serverpackets.InitOpcode {
		return fmt.Errorf("expected Init opcode 0x00, got 0x%02X", payload[0])
	}

	// Parse Init packet (minimum 170 bytes plaintext)
	if len(payload) < 170 {
		return fmt.Errorf("init packet too short: %d", len(payload))
	}

	c.sessionID = int32(binary.LittleEndian.Uint32(payload[1:5]))
	c.protocolVersion = binary.LittleEndian.Uint32(payload[5:9])

	// Scrambled RSA modulus (128 bytes at offset 9)
	scrambledModulus := payload[9 : 9+128]

	// Unscramble RSA modulus to get the original public key modulus
	c.rsaModulus = crypto.UnscrambleModulus(scrambledModulus)

	// Blowfish key (16 bytes at offset 153)
	// Init packet structure: opcode(1) + sessionID(4) + protocolRev(4) + rsaKey(128) + ggData(16) = 153
	c.blowfishKey = make([]byte, 16)
	copy(c.blowfishKey, payload[153:153+16])

	// Инициализируем LoginEncryption с dynamic key
	enc, err := crypto.NewLoginEncryption(c.blowfishKey)
	if err != nil {
		return fmt.Errorf("create login encryption: %w", err)
	}
	c.enc = enc

	return nil
}

// SendAuthGameGuard отправляет пакет AuthGameGuard (opcode 0x07) серверу.
// Включает sessionID для валидации.
func (c *LoginClient) SendAuthGameGuard() error {
	c.t.Helper()

	// AuthGameGuard packet: opcode (1) + sessionID (4) + data1 (4) + data2 (4) + data3 (4) + data4 (4)
	payload := c.writeBuf[2 : 2+21]
	payload[0] = 0x07 // AuthGameGuard opcode
	binary.LittleEndian.PutUint32(payload[1:], uint32(c.sessionID))
	// data1-data4 можно заполнить нулями для теста
	clear(payload[5:21])

	// Encrypt using client method (appendChecksum + dynamicCipher, no encXORPass)
	encSize, err := c.enc.EncryptPacketClient(c.writeBuf, 2, 21)
	if err != nil {
		return fmt.Errorf("encrypt AuthGameGuard: %w", err)
	}

	// Write length header + encrypted packet
	totalLen := 2 + encSize
	binary.LittleEndian.PutUint16(c.writeBuf[:2], uint16(totalLen))

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}

	if _, err := c.conn.Write(c.writeBuf[:totalLen]); err != nil {
		return fmt.Errorf("write AuthGameGuard: %w", err)
	}

	return nil
}

// ReadGGAuth читает ответ GGAuth (opcode 0x0B) от сервера.
func (c *LoginClient) ReadGGAuth() error {
	c.t.Helper()

	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}

	payload, err := protocol.ReadPacket(c.conn, c.enc, c.readBuf)
	if err != nil {
		return fmt.Errorf("read GGAuth: %w", err)
	}

	if len(payload) < 1 {
		return fmt.Errorf("GGAuth packet too short")
	}

	if payload[0] != 0x0B {
		return fmt.Errorf("expected GGAuth opcode 0x0B, got 0x%02X", payload[0])
	}

	return nil
}

// SendRequestAuthLogin отправляет пакет RequestAuthLogin (opcode 0x00).
// Шифрует login/password через RSA и отправляет серверу.
func (c *LoginClient) SendRequestAuthLogin(login, password string) error {
	c.t.Helper()

	// Prepare plaintext block (128 bytes for RSA-1024)
	plaintext := make([]byte, 128)

	// Login at offset 0x5E (14 bytes)
	copy(plaintext[0x5E:0x5E+14], login)

	// Password at offset 0x6C (16 bytes)
	copy(plaintext[0x6C:0x6C+16], password)

	// Encrypt with RSA
	encrypted, err := RSAEncryptWithModulus(c.rsaModulus, plaintext)
	if err != nil {
		return fmt.Errorf("RSA encrypt: %w", err)
	}

	// RequestAuthLogin packet: opcode (1) + encrypted_block (128)
	payload := c.writeBuf[2 : 2+1+128]
	payload[0] = 0x00 // RequestAuthLogin opcode
	copy(payload[1:], encrypted)

	// Encrypt using client method (appendChecksum + dynamicCipher)
	encSize, err := c.enc.EncryptPacketClient(c.writeBuf, 2, 1+128)
	if err != nil {
		return fmt.Errorf("encrypt RequestAuthLogin: %w", err)
	}

	// Write length header + encrypted packet
	totalLen := 2 + encSize
	binary.LittleEndian.PutUint16(c.writeBuf[:2], uint16(totalLen))

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}

	if _, err := c.conn.Write(c.writeBuf[:totalLen]); err != nil {
		return fmt.Errorf("write RequestAuthLogin: %w", err)
	}

	return nil
}

// ReadLoginOk читает пакет LoginOk (opcode 0x03) и возвращает loginOkID1, loginOkID2.
// Если сервер отправил другой пакет (например LoginFail), возвращает ошибку.
func (c *LoginClient) ReadLoginOk() (loginOkID1, loginOkID2 int32, err error) {
	c.t.Helper()

	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return 0, 0, fmt.Errorf("set read deadline: %w", err)
	}

	payload, err := protocol.ReadPacket(c.conn, c.enc, c.readBuf)
	if err != nil {
		return 0, 0, fmt.Errorf("read LoginOk: %w", err)
	}

	if len(payload) < 1 {
		return 0, 0, fmt.Errorf("LoginOk packet too short")
	}

	opcode := payload[0]

	// Проверяем что это LoginOk или ServerList (в случае ShowLicence=false)
	if opcode != serverpackets.LoginOkOpcode && opcode != 0x04 { // 0x04 = ServerList
		// Возможно LoginFail или AccountKicked
		if opcode == 0x01 {
			reason := payload[1]
			return 0, 0, fmt.Errorf("received LoginFail, reason: 0x%02X", reason)
		}
		if opcode == 0x02 {
			reason := payload[1]
			return 0, 0, fmt.Errorf("received AccountKicked, reason: 0x%02X", reason)
		}
		return 0, 0, fmt.Errorf("expected LoginOk or ServerList, got opcode 0x%02X", opcode)
	}

	if opcode == serverpackets.LoginOkOpcode {
		// LoginOk: opcode (1) + loginOkID1 (4) + loginOkID2 (4) + ...
		if len(payload) < 9 {
			return 0, 0, fmt.Errorf("LoginOk payload too short: %d", len(payload))
		}

		loginOkID1 = int32(binary.LittleEndian.Uint32(payload[1:5]))
		loginOkID2 = int32(binary.LittleEndian.Uint32(payload[5:9]))
		return loginOkID1, loginOkID2, nil
	}

	// ServerList (ShowLicence=false) — loginOkID1/loginOkID2 не передаются напрямую
	// Но для тестов нам нужны эти значения. В случае ServerList они НЕ передаются.
	// Для simplicity, вернём ошибку если нужны loginOkID.
	return 0, 0, fmt.Errorf("received ServerList instead of LoginOk (ShowLicence=false mode)")
}

// ReadServerList читает пакет ServerList (opcode 0x04) и возвращает сырой payload.
// Для детального парсинга caller должен разобрать payload самостоятельно.
func (c *LoginClient) ReadServerList() ([]byte, error) {
	c.t.Helper()

	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return nil, fmt.Errorf("set read deadline: %w", err)
	}

	payload, err := protocol.ReadPacket(c.conn, c.enc, c.readBuf)
	if err != nil {
		return nil, fmt.Errorf("read ServerList: %w", err)
	}

	if len(payload) < 1 {
		return nil, fmt.Errorf("ServerList packet too short")
	}

	if payload[0] != 0x04 {
		return nil, fmt.Errorf("expected ServerList opcode 0x04, got 0x%02X", payload[0])
	}

	return payload, nil
}

// SendRequestServerList отправляет пакет RequestServerList (opcode 0x05).
// Включает loginOkID1, loginOkID2 для валидации сессии.
func (c *LoginClient) SendRequestServerList(loginOkID1, loginOkID2 int32) error {
	c.t.Helper()

	// RequestServerList packet: opcode (1) + loginOkID1 (4) + loginOkID2 (4)
	payload := c.writeBuf[2 : 2+9]
	payload[0] = 0x05 // RequestServerList opcode
	binary.LittleEndian.PutUint32(payload[1:], uint32(loginOkID1))
	binary.LittleEndian.PutUint32(payload[5:], uint32(loginOkID2))

	// Encrypt using client method
	encSize, err := c.enc.EncryptPacketClient(c.writeBuf, 2, 9)
	if err != nil {
		return fmt.Errorf("encrypt RequestServerList: %w", err)
	}

	// Write length header + encrypted packet
	totalLen := 2 + encSize
	binary.LittleEndian.PutUint16(c.writeBuf[:2], uint16(totalLen))

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}

	if _, err := c.conn.Write(c.writeBuf[:totalLen]); err != nil {
		return fmt.Errorf("write RequestServerList: %w", err)
	}

	return nil
}

// SendRequestServerLogin отправляет пакет RequestServerLogin (opcode 0x02).
// Включает loginOkID1, loginOkID2, serverID для выбора GameServer.
func (c *LoginClient) SendRequestServerLogin(loginOkID1, loginOkID2 int32, serverID byte) error {
	c.t.Helper()

	// RequestServerLogin packet: opcode (1) + loginOkID1 (4) + loginOkID2 (4) + serverID (1)
	payload := c.writeBuf[2 : 2+10]
	payload[0] = 0x02 // RequestServerLogin opcode
	binary.LittleEndian.PutUint32(payload[1:], uint32(loginOkID1))
	binary.LittleEndian.PutUint32(payload[5:], uint32(loginOkID2))
	payload[9] = serverID

	// Encrypt using client method
	encSize, err := c.enc.EncryptPacketClient(c.writeBuf, 2, 10)
	if err != nil {
		return fmt.Errorf("encrypt RequestServerLogin: %w", err)
	}

	// Write length header + encrypted packet
	totalLen := 2 + encSize
	binary.LittleEndian.PutUint16(c.writeBuf[:2], uint16(totalLen))

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}

	if _, err := c.conn.Write(c.writeBuf[:totalLen]); err != nil {
		return fmt.Errorf("write RequestServerLogin: %w", err)
	}

	return nil
}

// ReadPlayOk читает пакет PlayOk (opcode 0x07) и возвращает playOkID1, playOkID2.
func (c *LoginClient) ReadPlayOk() (playOkID1, playOkID2 int32, err error) {
	c.t.Helper()

	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return 0, 0, fmt.Errorf("set read deadline: %w", err)
	}

	payload, err := protocol.ReadPacket(c.conn, c.enc, c.readBuf)
	if err != nil {
		return 0, 0, fmt.Errorf("read PlayOk: %w", err)
	}

	if len(payload) < 1 {
		return 0, 0, fmt.Errorf("PlayOk packet too short")
	}

	if payload[0] != serverpackets.PlayOkOpcode {
		// Возможно PlayFail
		if payload[0] == 0x06 {
			reason := payload[1]
			return 0, 0, fmt.Errorf("received PlayFail, reason: 0x%02X", reason)
		}
		return 0, 0, fmt.Errorf("expected PlayOk opcode 0x07, got 0x%02X", payload[0])
	}

	if len(payload) < 9 {
		return 0, 0, fmt.Errorf("PlayOk payload too short: %d", len(payload))
	}

	playOkID1 = int32(binary.LittleEndian.Uint32(payload[1:5]))
	playOkID2 = int32(binary.LittleEndian.Uint32(payload[5:9]))

	return playOkID1, playOkID2, nil
}

// ReadPacket читает любой пакет от сервера и возвращает payload.
// Используется для generic чтения когда opcode неизвестен.
func (c *LoginClient) ReadPacket() ([]byte, error) {
	c.t.Helper()

	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return nil, fmt.Errorf("set read deadline: %w", err)
	}

	payload, err := protocol.ReadPacket(c.conn, c.enc, c.readBuf)
	if err != nil {
		return nil, fmt.Errorf("read packet: %w", err)
	}

	return payload, nil
}

// ReadLoginFail читает пакет LoginFail (opcode 0x01) и возвращает reason code.
// Если получен другой пакет, возвращает ошибку.
func (c *LoginClient) ReadLoginFail() (byte, error) {
	c.t.Helper()

	payload, err := c.ReadPacket()
	if err != nil {
		return 0, err
	}

	if len(payload) < 2 {
		return 0, fmt.Errorf("LoginFail packet too short: %d", len(payload))
	}

	if payload[0] != serverpackets.LoginFailOpcode {
		return 0, fmt.Errorf("expected LoginFail opcode 0x01, got 0x%02X", payload[0])
	}

	return payload[1], nil
}

// ReadPlayFail читает пакет PlayFail (opcode 0x06) и возвращает reason code.
// Если получен другой пакет, возвращает ошибку.
func (c *LoginClient) ReadPlayFail() (byte, error) {
	c.t.Helper()

	payload, err := c.ReadPacket()
	if err != nil {
		return 0, err
	}

	if len(payload) < 2 {
		return 0, fmt.Errorf("PlayFail packet too short: %d", len(payload))
	}

	if payload[0] != 0x06 { // PlayFail opcode
		return 0, fmt.Errorf("expected PlayFail opcode 0x06, got 0x%02X", payload[0])
	}

	return payload[1], nil
}

// Close закрывает соединение с сервером.
func (c *LoginClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SessionID возвращает sessionID из Init пакета.
func (c *LoginClient) SessionID() int32 {
	return c.sessionID
}

// RSAModulus возвращает unscrambled RSA modulus (128 bytes) из Init пакета.
func (c *LoginClient) RSAModulus() []byte {
	return c.rsaModulus
}

// BlowfishKey возвращает Blowfish key (16 bytes) из Init пакета.
func (c *LoginClient) BlowfishKey() []byte {
	return c.blowfishKey
}

// RSAEncryptWithModulus шифрует plaintext используя RSA modulus (128 bytes для RSA-1024).
// Используется для шифрования login/password в RequestAuthLogin.
func RSAEncryptWithModulus(modulus, plaintext []byte) ([]byte, error) {
	if len(modulus) != 128 {
		return nil, fmt.Errorf("RSA modulus must be 128 bytes, got %d", len(modulus))
	}

	if len(plaintext) != 128 {
		return nil, fmt.Errorf("RSA plaintext must be 128 bytes, got %d", len(plaintext))
	}

	// Конвертируем modulus в big.Int
	n := new(big.Int).SetBytes(modulus)

	// Public exponent (hardcoded в L2 = 65537)
	e := big.NewInt(65537)

	// RSA encrypt: ciphertext = plaintext^e mod n
	m := new(big.Int).SetBytes(plaintext)
	c := new(big.Int).Exp(m, e, n)

	ciphertext := c.Bytes()

	// Pad to 128 bytes if needed
	if len(ciphertext) < 128 {
		padded := make([]byte, 128)
		copy(padded[128-len(ciphertext):], ciphertext)
		ciphertext = padded
	}

	return ciphertext, nil
}
