package gameserver

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/protocol"
)

// Server is the GameServer that accepts game client connections on port 7777.
type Server struct {
	cfg            config.GameServer
	sessionManager *login.SessionManager
	charRepo       CharacterRepository // Phase 4.6: character database access
	persister      PlayerPersister     // Phase 6.0: DB persistence

	sendPool  *BytePool
	readPool  *BytePool
	writePool *BytePool // Phase 7.0: shared pool for encrypted outgoing packets
	handler   *Handler

	clientManager *ClientManager // Phase 4.5 PR4: broadcast infrastructure

	listener net.Listener
	mu       sync.Mutex
}

// NewServer creates a new GameServer.
func NewServer(cfg config.GameServer, sessionManager *login.SessionManager, charRepo CharacterRepository, persister PlayerPersister) (*Server, error) {
	clientMgr := NewClientManager()

	writePool := NewBytePool(constants.GameServerWriteBufSize)
	clientMgr.SetWritePool(writePool)

	s := &Server{
		cfg:            cfg,
		sessionManager: sessionManager,
		charRepo:       charRepo,
		persister:      persister,
		sendPool:       NewBytePool(constants.DefaultSendBufSize),
		readPool:       NewBytePool(constants.DefaultReadBufSize),
		writePool:      writePool,
		handler:        NewHandler(sessionManager, clientMgr, charRepo, persister),
		clientManager:  clientMgr,
	}

	return s, nil
}

// generateBlowfishKey creates a fresh 16-byte random Blowfish key.
func generateBlowfishKey() ([]byte, error) {
	key := make([]byte, constants.BlowfishKeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generating blowfish key: %w", err)
	}
	// Ensure no zero bytes (L2 client requirement: bytes 1-255)
	for i, b := range key {
		if b == 0 {
			key[i] = 1
		}
	}
	return key, nil
}

// Addr returns the address the server is listening on.
// Returns nil if the server hasn't started yet.
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// ClientManager returns the client manager for this server.
// Used for broadcast operations and client tracking.
func (s *Server) ClientManager() *ClientManager {
	return s.clientManager
}

// Close closes the listener and saves all online players.
func (s *Server) Close() error {
	// Phase 6.0: Save all online players before shutdown
	s.saveAllPlayers()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// saveAllPlayers saves all currently online players to DB.
// Called during graceful shutdown.
func (s *Server) saveAllPlayers() {
	if s.persister == nil {
		return
	}

	saved := 0
	s.clientManager.ForEachClient(func(client *GameClient) bool {
		player := client.ActivePlayer()
		if player == nil {
			return true
		}
		saveCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.persister.SavePlayer(saveCtx, player); err != nil {
			slog.Error("save player on shutdown",
				"player", player.Name(),
				"error", err)
		} else {
			saved++
		}
		return true
	})

	if saved > 0 {
		slog.Info("saved players on shutdown", "count", saved)
	}
}

// Run begins listening for game client connections.
// Creates a listener on cfg.BindAddress:cfg.Port and starts the accept loop.
func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.BindAddress, s.cfg.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", addr, err)
	}

	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()

	return s.Serve(ctx, ln)
}

// Serve accepts connections from the given listener and starts the accept loop.
// Used for testing with custom listeners.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	go func() {
		<-ctx.Done()
		// Phase 6.0: Save all online players before closing listener
		s.saveAllPlayers()
		ln.Close()
	}()

	var wg sync.WaitGroup
	go func() {
		slog.Info("game server started", "address", ln.Addr())
		acceptLoop(ctx, &wg, s, ln)
	}()

	wg.Wait()

	return nil
}

func acceptLoop(
	ctx context.Context,
	wg *sync.WaitGroup,
	srv *Server,
	ln net.Listener,
) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				slog.Error("failed to accept new connection", "error", err)
				continue
			}

			// Enable TCP keepalive (detect dead connections)
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				if err := tcpConn.SetKeepAlive(true); err != nil {
					slog.Warn("set keepalive failed", "error", err)
				}
				if err := tcpConn.SetKeepAlivePeriod(30 * time.Second); err != nil {
					slog.Warn("set keepalive period failed", "error", err)
				}
			}

			wg.Go(func() {
				handleConnection(ctx, srv, conn)
			})
		}
	}
}

func handleConnection(ctx context.Context, srv *Server, conn net.Conn) {
	defer conn.Close()

	// Cleanup function to unregister client on disconnect
	var accountName string
	var client *GameClient
	defer func() {
		if accountName != "" {
			srv.clientManager.Unregister(accountName)
			slog.Debug("client unregistered", "account", accountName)
		}
		// Call OnDisconnection for graceful cleanup (Phase 4.17.7)
		// Handles delayed removal if player in combat (15-second delay to prevent combat logging)
		if client != nil {
			OnDisconnection(ctx, client, srv.persister)
		}
	}()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		slog.Error("failed to split host port", "connection", conn.RemoteAddr(), "error", err)
		return
	}

	slog.Info("new game client connection", "remote", host)

	// Generate fresh Blowfish key for this connection
	blowfishKey, err := generateBlowfishKey()
	if err != nil {
		slog.Error("failed to generate blowfish key", "error", err)
		return
	}

	// Create GameClient state with write pool and config
	client, err = NewGameClient(conn, blowfishKey, srv.writePool, srv.cfg.SendQueueSize, srv.cfg.WriteTimeout)
	if err != nil {
		slog.Error("failed to create game client", "error", err)
		return
	}

	// Send KeyPacket (first packet: opcode 0x2E + protocol version + blowfish key)
	keyPkt := serverpackets.NewKeyPacket(blowfishKey)
	keyData, err := keyPkt.Write()
	if err != nil {
		slog.Error("failed to write KeyPacket", "error", err)
		return
	}

	// KeyPacket is NOT encrypted (sent in plaintext)
	if _, err := conn.Write(keyData); err != nil {
		slog.Error("failed to send KeyPacket", "error", err)
		return
	}

	slog.Debug("sent KeyPacket", "client", client.IP())

	// Start writePump AFTER KeyPacket (plaintext) is sent directly
	go client.writePump()
	defer client.Close() // CloseAsync + conn.Close → stops writePump

	// Resolve read timeout from config
	readTimeout := srv.cfg.ReadTimeout
	if readTimeout <= 0 {
		readTimeout = defaultReadTimeout
	}

	// Enter packet handling loop (read → decrypt → handle → encrypt → queue)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := handlePacket(ctx, srv, client, readTimeout); err != nil {
				// Store account name for cleanup
				accountName = client.AccountName()

				if err == io.EOF {
					slog.Info("client disconnected", "account", accountName, "client", client.IP())
				} else {
					slog.Error("packet handling error", "error", err, "client", client.IP())
				}
				return
			}

			// Update account name if client authenticated during this iteration
			if client.State() >= ClientStateAuthenticated && accountName == "" {
				accountName = client.AccountName()
			}
		}
	}
}

func handlePacket(ctx context.Context, srv *Server, client *GameClient, readTimeout time.Duration) error {
	readBuf := srv.readPool.Get(constants.DefaultReadBufSize)
	defer srv.readPool.Put(readBuf)

	// Read timeout: idle client disconnects (Cloudflare recommendation)
	if err := client.Conn().SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		return fmt.Errorf("setting read deadline: %w", err)
	}

	// Read and decrypt packet
	payload, err := protocol.ReadPacket(client.Conn(), client.Encryption(), readBuf)
	if err != nil {
		return fmt.Errorf("reading packet: %w", err)
	}

	// Dispatch to handler
	sendBuf := srv.sendPool.Get(constants.DefaultSendBufSize)
	defer srv.sendPool.Put(sendBuf)

	n, keepOpen, err := srv.handler.HandlePacket(ctx, client, payload, sendBuf[constants.PacketHeaderSize:])
	if err != nil {
		return fmt.Errorf("handling packet: %w", err)
	}

	// Send response via write queue (zero steady-state alloc with pool)
	if n > 0 {
		encPkt, encErr := srv.writePool.EncryptToPooled(client.Encryption(), sendBuf[constants.PacketHeaderSize:], n)
		if encErr != nil {
			return fmt.Errorf("encrypting response: %w", encErr)
		}
		// SendSync takes ownership — will return to pool
		writeTimeout := srv.cfg.WriteTimeout
		if writeTimeout <= 0 {
			writeTimeout = defaultWriteTimeout
		}
		if err := client.SendSync(encPkt, writeTimeout); err != nil {
			return fmt.Errorf("queueing response: %w", err)
		}
	}

	if !keepOpen {
		return fmt.Errorf("handler requested connection close")
	}

	return nil
}
