package login

import (
	"fmt"
	"math/rand/v2"
	"net"
	"sync"

	"github.com/udisondev/la2go/internal/crypto"
)

// Client represents a single client connection to the login server.
type Client struct {
	conn       net.Conn
	ip         string
	sessionID  int32
	rsaKeyPair *crypto.RSAKeyPair

	state      ConnectionState
	sessionKey SessionKey
	account    string

	mu sync.Mutex
}

// NewClient creates a new login client state for the given connection.
func NewClient(conn net.Conn, rsaKeyPair *crypto.RSAKeyPair) (*Client, error) {
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return nil, fmt.Errorf("splitting host port: %w", err)
	}

	return &Client{
		conn:       conn,
		ip:         host,
		sessionID:  rand.Int32(),
		rsaKeyPair: rsaKeyPair,
		state:      StateConnected,
	}, nil
}

// IP returns the client's remote IP address.
func (c *Client) IP() string {
	return c.ip
}

// SessionID returns the session ID assigned to this client.
func (c *Client) SessionID() int32 {
	return c.sessionID
}

// State returns the current connection state.
func (c *Client) State() ConnectionState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// SetState sets the connection state.
func (c *Client) SetState(s ConnectionState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = s
}

// Account returns the logged-in account name.
func (c *Client) Account() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.account
}

// SetAccount sets the account name.
func (c *Client) SetAccount(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.account = name
}

// SessionKey returns the session key.
func (c *Client) SessionKey() SessionKey {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sessionKey
}

// SetSessionKey sets the session key.
func (c *Client) SetSessionKey(sk SessionKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessionKey = sk
}

// RSAKeyPair returns the RSA key pair assigned to this client.
func (c *Client) RSAKeyPair() *crypto.RSAKeyPair {
	return c.rsaKeyPair
}
