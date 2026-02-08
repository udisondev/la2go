package gslistener

import (
	"fmt"
	"net"
	"sync"

	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/gameserver"
)

// GSConnection представляет подключение одного GameServer к LoginServer
type GSConnection struct {
	conn       net.Conn
	ip         string
	rsaKeyPair *crypto.RSAKeyPair

	mu             sync.Mutex
	state          gameserver.GSConnectionState
	blowfishCipher *crypto.BlowfishCipher
	info           *gameserver.GameServerInfo // attached after auth
	accounts       map[string]struct{}        // online accounts
}

// NewGSConnection создаёт новое подключение GameServer
func NewGSConnection(conn net.Conn, rsaKeyPair *crypto.RSAKeyPair) (*GSConnection, error) {
	// Extract IP
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		host = conn.RemoteAddr().String()
	}

	// Initial Blowfish cipher (22 bytes default key)
	cipher, err := crypto.NewBlowfishCipher(crypto.DefaultGSBlowfishKey)
	if err != nil {
		return nil, fmt.Errorf("creating initial Blowfish cipher: %w", err)
	}

	return &GSConnection{
		conn:           conn,
		ip:             host,
		rsaKeyPair:     rsaKeyPair,
		state:          gameserver.GSStateConnected,
		blowfishCipher: cipher,
		accounts:       make(map[string]struct{}),
	}, nil
}

// IP returns the remote IP address
func (c *GSConnection) IP() string {
	return c.ip
}

// State возвращает текущее состояние соединения
func (c *GSConnection) State() gameserver.GSConnectionState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// SetState устанавливает новое состояние соединения
func (c *GSConnection) SetState(s gameserver.GSConnectionState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = s
}

// BlowfishCipher возвращает текущий Blowfish cipher
func (c *GSConnection) BlowfishCipher() *crypto.BlowfishCipher {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.blowfishCipher
}

// SetBlowfishCipher устанавливает новый Blowfish cipher
func (c *GSConnection) SetBlowfishCipher(cipher *crypto.BlowfishCipher) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.blowfishCipher = cipher
}

// RSAKeyPair возвращает RSA ключ для этого соединения
func (c *GSConnection) RSAKeyPair() *crypto.RSAKeyPair {
	return c.rsaKeyPair
}

// AttachGameServerInfo привязывает информацию о GameServer после аутентификации
func (c *GSConnection) AttachGameServerInfo(info *gameserver.GameServerInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.info = info
}

// GameServerInfo возвращает информацию о GameServer (nil если не аутентифицирован)
func (c *GSConnection) GameServerInfo() *gameserver.GameServerInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.info
}

// AddAccount добавляет аккаунт в список онлайн игроков
func (c *GSConnection) AddAccount(account string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accounts[account] = struct{}{}
}

// RemoveAccount удаляет аккаунт из списка онлайн игроков
func (c *GSConnection) RemoveAccount(account string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.accounts, account)
}

// HasAccount проверяет, находится ли аккаунт в списке онлайн игроков
func (c *GSConnection) HasAccount(account string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.accounts[account]
	return ok
}
