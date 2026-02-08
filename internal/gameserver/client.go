package gameserver

import (
	"fmt"
	"math/rand/v2"
	"net"
	"sync"
	"sync/atomic"

	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/login"
)

// GameClient represents a single game client connection to the game server.
type GameClient struct {
	conn       net.Conn
	ip         string
	sessionID  int32
	encryption *crypto.LoginEncryption

	// state использует atomic.Int32 для lock-free reads в hot path
	state atomic.Int32

	// mu защищает только accountName и sessionKey (редкие операции)
	mu          sync.Mutex
	accountName string
	sessionKey  *login.SessionKey
	// activeChar будет добавлен позже в Phase 4.2 (model.Character)
}

// NewGameClient creates a new game client state for the given connection.
func NewGameClient(conn net.Conn, blowfishKey []byte) (*GameClient, error) {
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return nil, fmt.Errorf("splitting host port: %w", err)
	}

	// Initialize encryption with Blowfish key
	enc, err := crypto.NewLoginEncryption(blowfishKey)
	if err != nil {
		return nil, fmt.Errorf("creating login encryption: %w", err)
	}

	client := &GameClient{
		conn:       conn,
		ip:         host,
		sessionID:  rand.Int32(),
		encryption: enc,
	}
	client.state.Store(int32(ClientStateConnected))
	return client, nil
}

// Conn returns the underlying network connection.
func (c *GameClient) Conn() net.Conn {
	return c.conn
}

// IP returns the client's remote IP address.
func (c *GameClient) IP() string {
	return c.ip
}

// SessionID returns the session ID assigned to this client.
func (c *GameClient) SessionID() int32 {
	return c.sessionID
}

// Encryption returns the encryption context for this client.
func (c *GameClient) Encryption() *crypto.LoginEncryption {
	return c.encryption
}

// State returns the current connection state.
// Использует atomic для lock-free reads (hot path).
func (c *GameClient) State() ClientConnectionState {
	return ClientConnectionState(c.state.Load())
}

// SetState sets the connection state.
// Использует atomic для lock-free writes.
func (c *GameClient) SetState(s ClientConnectionState) {
	c.state.Store(int32(s))
}

// AccountName returns the logged-in account name.
func (c *GameClient) AccountName() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.accountName
}

// SetAccountName sets the account name.
func (c *GameClient) SetAccountName(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accountName = name
}

// SessionKey returns the session key.
func (c *GameClient) SessionKey() *login.SessionKey {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sessionKey
}

// SetSessionKey sets the session key.
func (c *GameClient) SetSessionKey(sk *login.SessionKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessionKey = sk
}

// Close closes the connection.
func (c *GameClient) Close() error {
	// Проверяем state без lock (atomic read)
	if ClientConnectionState(c.state.Load()) == ClientStateDisconnected {
		return nil
	}

	// Устанавливаем disconnected state (atomic write)
	c.state.Store(int32(ClientStateDisconnected))
	return c.conn.Close()
}
