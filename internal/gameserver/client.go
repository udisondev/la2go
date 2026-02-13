package gameserver

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/model"
)

// Default write queue / timeout constants.
// Overridden by config values when available.
const (
	defaultSendQueueSize = 256
	defaultWriteTimeout  = 5 * time.Second
	defaultReadTimeout   = 120 * time.Second
)

// GameClient represents a single game client connection to the game server.
type GameClient struct {
	conn       net.Conn
	ip         string
	sessionID  int32
	encryption *crypto.LoginEncryption
	gameCrypt  *crypto.GameCrypt

	// state использует atomic.Int32 для lock-free reads в hot path
	state atomic.Int32

	// markedForDisconnection indicates client should be disconnected after sending current packet
	// Phase 4.17.5: Used by Logout/RequestRestart to gracefully close connection
	markedForDisconnection atomic.Bool

	// detached indicates TCP connection is closed but Player stays in world (offline trade).
	// Phase 31: Offline Trade — mirrors Java GameClient.isDetached.
	detached atomic.Bool

	// mu защищает только accountName, sessionKey, selectedCharacter, activePlayer (редкие операции)
	mu                sync.Mutex
	accountName       string
	sessionKey        *login.SessionKey
	selectedCharacter int32         // Character slot index (0-7), -1 = not selected
	activePlayer      *model.Player // Active player (set after EnterWorld)

	// Session-scoped character cache (Phase 4.18 Optimization 3)
	// Eliminates 3× redundant LoadByAccountName() calls per login
	// Protected by cacheMu (separate mutex for cache operations)
	cacheMu          sync.RWMutex
	cachedCharacters []*model.Player
	cacheAccountName string

	// Per-client write queue (Phase 7.0: Async Write Architecture)
	// Pattern: Leaf, Zinx, Gorilla Chat, L2J MMOCore
	sendCh    chan []byte  // buffered channel with encrypted packets (pool-backed)
	closeCh   chan struct{}
	closeOnce sync.Once

	writePool    *BytePool     // shared pool for returning buffers after write
	writeTimeout time.Duration // per-write deadline
}

// NewGameClient creates a new game client state for the given connection.
func NewGameClient(conn net.Conn, blowfishKey []byte, writePool *BytePool, sendQueueSize int, writeTimeout time.Duration) (*GameClient, error) {
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return nil, fmt.Errorf("splitting host port: %w", err)
	}

	// Initialize encryption with Blowfish key
	enc, err := crypto.NewLoginEncryption(blowfishKey)
	if err != nil {
		return nil, fmt.Errorf("creating login encryption: %w", err)
	}

	if sendQueueSize <= 0 {
		sendQueueSize = defaultSendQueueSize
	}
	if writeTimeout <= 0 {
		writeTimeout = defaultWriteTimeout
	}

	client := &GameClient{
		conn:              conn,
		ip:                host,
		sessionID:         rand.Int32(),
		encryption:        enc,
		gameCrypt:         crypto.NewGameCrypt(),
		selectedCharacter: -1, // Not selected yet
		sendCh:            make(chan []byte, sendQueueSize),
		closeCh:           make(chan struct{}),
		writePool:         writePool,
		writeTimeout:      writeTimeout,
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

// Encryption returns the Blowfish encryption context for this client (LoginServer protocol).
func (c *GameClient) Encryption() *crypto.LoginEncryption {
	return c.encryption
}

// GameCrypt returns the XOR rolling cipher for this client (GameServer protocol).
// After the Init handshake, call GameCrypt().SetKey(key) to initialize, then
// all subsequent packets are encrypted/decrypted via this cipher.
func (c *GameClient) GameCrypt() *crypto.GameCrypt {
	return c.gameCrypt
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

// SelectedCharacter returns the selected character slot index (-1 if not selected).
func (c *GameClient) SelectedCharacter() int32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.selectedCharacter
}

// SetSelectedCharacter sets the selected character slot index.
func (c *GameClient) SetSelectedCharacter(slot int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.selectedCharacter = slot
}

// ActivePlayer returns the active player (nil if not in game).
func (c *GameClient) ActivePlayer() *model.Player {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.activePlayer
}

// SetActivePlayer sets the active player (called after EnterWorld).
func (c *GameClient) SetActivePlayer(player *model.Player) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.activePlayer = player
}

// writePump is a dedicated writer goroutine for this client.
// Reads encrypted packets from sendCh and writes them to conn.
// Uses net.Buffers (writev syscall) for batching and pool.Put for buffer return.
//
// Pattern: Gorilla WebSocket Chat + net.Buffers + drain batching.
func (c *GameClient) writePump() {
	// Pre-allocate scratch slices (one-time, reused across iterations)
	bufs := make(net.Buffers, 0, 64)
	poolBufs := make([][]byte, 0, 64)

	defer func() {
		// Drain remaining packets and return to pool
		for {
			select {
			case pkt := <-c.sendCh:
				if c.writePool != nil {
					c.writePool.Put(pkt)
				}
			default:
				return
			}
		}
	}()

	for {
		select {
		case pkt, ok := <-c.sendCh:
			if !ok {
				return // channel closed = graceful shutdown
			}

			if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
				slog.Warn("set write deadline failed", "client", c.ip, "error", err)
				if c.writePool != nil {
					c.writePool.Put(pkt)
				}
				return
			}

			// Batching: drain all queued packets (Gorilla Chat pattern)
			queued := len(c.sendCh)
			if queued == 0 {
				// Single packet — direct write (hot path, zero-alloc)
				_, err := c.conn.Write(pkt)
				if c.writePool != nil {
					c.writePool.Put(pkt)
				}
				if err != nil {
					slog.Warn("write failed", "client", c.ip, "error", err)
					return
				}
				continue
			}

			// Multiple packets — net.Buffers (writev syscall, zero-copy)
			bufs = bufs[:0]
			poolBufs = poolBufs[:0]

			bufs = append(bufs, pkt)
			poolBufs = append(poolBufs, pkt)
			for range queued {
				p := <-c.sendCh
				bufs = append(bufs, p)
				poolBufs = append(poolBufs, p)
			}

			_, err := bufs.WriteTo(c.conn)

			// ALWAYS return buffers to pool (even on error)
			if c.writePool != nil {
				for _, b := range poolBufs {
					c.writePool.Put(b)
				}
			}

			if err != nil {
				slog.Warn("batch write failed", "client", c.ip, "error", err)
				return
			}

		case <-c.closeCh:
			return
		}
	}
}

// Send queues an encrypted packet for async delivery.
// Non-blocking: returns error if queue is full (slow client → disconnect).
// OWNERSHIP: takes ownership of encryptedPkt (pool buffer). writePump will return it to pool.
func (c *GameClient) Send(encryptedPkt []byte) error {
	select {
	case c.sendCh <- encryptedPkt:
		return nil
	default:
		if c.writePool != nil {
			c.writePool.Put(encryptedPkt)
		}
		slog.Warn("send queue full, disconnecting slow client", "client", c.ip)
		c.CloseAsync()
		return fmt.Errorf("send queue full")
	}
}

// SendSync queues an encrypted packet and blocks until accepted or timeout.
// Used for handler responses that MUST be delivered.
// OWNERSHIP: takes ownership of encryptedPkt.
func (c *GameClient) SendSync(encryptedPkt []byte, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case c.sendCh <- encryptedPkt:
		return nil
	case <-timer.C:
		if c.writePool != nil {
			c.writePool.Put(encryptedPkt)
		}
		return fmt.Errorf("send timeout after %v", timeout)
	case <-c.closeCh:
		if c.writePool != nil {
			c.writePool.Put(encryptedPkt)
		}
		return fmt.Errorf("client closed")
	}
}

// CloseAsync signals the writePump to stop without blocking.
// Safe to call multiple times.
func (c *GameClient) CloseAsync() {
	c.closeOnce.Do(func() {
		c.state.Store(int32(ClientStateDisconnected))
		close(c.closeCh)
	})
}

// Close closes the connection and stops the writePump.
func (c *GameClient) Close() error {
	c.CloseAsync()
	return c.conn.Close()
}

// MarkForDisconnection marks client for graceful disconnection after sending current packet.
// Used by Logout and RequestRestart handlers (Phase 4.17.5).
// Server will close TCP connection after sending response packet.
func (c *GameClient) MarkForDisconnection() {
	c.markedForDisconnection.Store(true)
}

// IsMarkedForDisconnection returns true if client is marked for disconnection.
// Server checks this flag after sending each packet and closes connection if true.
func (c *GameClient) IsMarkedForDisconnection() bool {
	return c.markedForDisconnection.Load()
}

// Detach marks the client as detached (TCP closed, player stays in world).
// Phase 31: Offline Trade — player's WorldObject remains visible while trading.
func (c *GameClient) Detach() {
	c.detached.Store(true)
}

// IsDetached returns true if client is detached (offline trade mode).
func (c *GameClient) IsDetached() bool {
	return c.detached.Load()
}

// GetCharacters returns cached characters for the account or loads from repository.
// Phase 4.18 Optimization 3: Eliminates 3× redundant DB queries per login.
// Cache is session-scoped and cleared on logout/disconnect.
//
// Performance impact:
// - Before: 3 × LoadByAccountName() = 3 × 500µs = 1.5ms per login
// - After: 1 × LoadByAccountName() + 2 × cache hits = 500µs
// - Improvement: -66.7% (3× faster), -200K DB queries/sec @ 100K logins/sec
func (c *GameClient) GetCharacters(accountName string, loader func(string) ([]*model.Player, error)) ([]*model.Player, error) {
	// Fast path: check cache with read lock
	c.cacheMu.RLock()
	if c.cacheAccountName == accountName && c.cachedCharacters != nil {
		chars := c.cachedCharacters
		c.cacheMu.RUnlock()
		return chars, nil
	}
	c.cacheMu.RUnlock()

	// Cache miss — load from repository
	chars, err := loader(accountName)
	if err != nil {
		return nil, err
	}

	// Update cache with write lock
	c.cacheMu.Lock()
	c.cachedCharacters = chars
	c.cacheAccountName = accountName
	c.cacheMu.Unlock()

	return chars, nil
}

// ClearCharacterCache clears the session character cache.
// Called on logout/disconnect to free memory.
func (c *GameClient) ClearCharacterCache() {
	c.cacheMu.Lock()
	c.cachedCharacters = nil
	c.cacheAccountName = ""
	c.cacheMu.Unlock()
}

// InvalidateCharacterCache forces re-loading characters from DB on next GetCharacters call.
// Called after creating or deleting a character.
func (c *GameClient) InvalidateCharacterCache() {
	c.ClearCharacterCache()
}
