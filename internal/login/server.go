package login

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	mathrand "math/rand/v2"
	"net"
	"sync"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/login/serverpackets"
	"github.com/udisondev/la2go/internal/protocol"
)

const (
	rsaKeyPairCount = 10
)

// ServerOption is a functional option for Server configuration.
type ServerOption func(**SessionManager)

// WithSessionManager sets a custom SessionManager (useful for testing with shared SessionManager).
func WithSessionManager(sm *SessionManager) ServerOption {
	return func(current **SessionManager) {
		*current = sm
	}
}

// Server is the LoginServer that accepts client connections on port 2106.
type Server struct {
	cfg            config.LoginServer
	db             *db.DB
	sessionManager *SessionManager

	rsaKeyPairs [rsaKeyPairCount]*crypto.RSAKeyPair
	sendPool    *BytePool
	readPool    *BytePool
	handler     *Handler

	listener net.Listener
	mu       sync.Mutex
}

// NewServer creates a new LoginServer with pre-generated RSA key pairs.
// Blowfish keys are generated fresh per connection.
func NewServer(cfg config.LoginServer, database *db.DB, opts ...ServerOption) (*Server, error) {
	sessionManager := NewSessionManager()

	// Применяем опции
	for _, opt := range opts {
		if opt != nil {
			opt(&sessionManager)
		}
	}

	// Создаём AccountRepository для Handler
	accountRepo := db.NewPostgresAccountRepository(database.Pool())

	s := &Server{
		cfg:            cfg,
		db:             database,
		sessionManager: sessionManager,
		sendPool:       NewBytePool(constants.DefaultSendBufSize),
		readPool:       NewBytePool(constants.DefaultReadBufSize),
		handler:        NewHandler(accountRepo, cfg, sessionManager),
	}

	// Pre-generate RSA key pairs (expensive operation — ~10-50ms each)
	slog.Info("generating RSA key pairs", "count", rsaKeyPairCount)
	for i := range rsaKeyPairCount {
		kp, err := crypto.GenerateRSAKeyPair()
		if err != nil {
			return nil, fmt.Errorf("generating RSA key pair %d: %w", i, err)
		}
		s.rsaKeyPairs[i] = kp
	}

	return s, nil
}

// SessionManager возвращает менеджер сессий (для интеграции с gslistener).
func (s *Server) SessionManager() *SessionManager {
	return s.sessionManager
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

// Addr возвращает адрес, на котором слушает сервер.
// Возвращает nil если сервер ещё не запущен.
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// Close закрывает listener и останавливает сервер.
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Run begins listening for client connections.
// Создаёт listener на cfg.BindAddress:cfg.Port и запускает accept loop.
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

// Serve принимает готовый listener и запускает accept loop.
// Используется для тестирования с произвольным listener.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	var wg sync.WaitGroup
	wg.Go(func() {
		slog.Info("login server started", "address", ln.Addr())
		acceptLoop(ctx, &wg, s, ln)
	})

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
				if errors.Is(err, net.ErrClosed) {
					return
				}
				slog.Error("Failed to accept new connection", "error", err)
				continue
			}
			wg.Go(func() {
				handleConnection(ctx, srv, conn)
			})
		}
	}
}

func handleConnection(ctx context.Context, srv *Server, conn net.Conn) {
	done := make(chan struct{})
	defer close(done)
	defer conn.Close()

	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
		}
	}()

	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		slog.Error("Failed to split host port", "connection", conn.RemoteAddr(), "error", err)
		return
	}

	slog.Info("new connection", "remote", host)

	rsaKeyPair := srv.rsaKeyPairs[mathrand.IntN(rsaKeyPairCount)]
	bfKey, err := generateBlowfishKey()
	if err != nil {
		slog.Error("failed to generate blowfish key", "err", err, "remote", host)
		return
	}

	enc, err := crypto.NewLoginEncryption(bfKey)
	if err != nil {
		slog.Error("failed to create login encryption", "err", err, "remote", host)
		return
	}

	client, err := NewClient(conn, rsaKeyPair)
	if err != nil {
		slog.Error("failed to create client", "err", err, "remote", host)
		return
	}

	sendBuf := srv.sendPool.Get(constants.DefaultSendBufSize)
	// Send Init packet — write payload into sendBuf[2:], then WritePacket encrypts in-place
	n := serverpackets.Init(sendBuf[2:], client.SessionID(), rsaKeyPair.ScrambledModulus, bfKey)
	if err := protocol.WritePacket(conn, enc, sendBuf, n); err != nil {
		srv.sendPool.Put(sendBuf)
		slog.Error("failed to send Init packet", "err", err, "remote", host)
		return
	}
	srv.sendPool.Put(sendBuf)
	slog.Debug("Init packet sent", "remote", host, "sessionId", client.SessionID())

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if ok, err := handlePacket(ctx, client, enc, srv); !ok {
				return
			} else if err != nil {
				slog.Error("Failed to handle packet", "remote", conn.RemoteAddr(), "error", err)
			}
		}
	}
}

func handlePacket(
	ctx context.Context,
	cli *Client,
	enc *crypto.LoginEncryption,
	srv *Server,
) (bool, error) {
	sendBuf := srv.sendPool.Get(constants.DefaultSendBufSize)
	defer srv.sendPool.Put(sendBuf)
	readBuf := srv.readPool.Get(constants.DefaultReadBufSize)
	defer srv.readPool.Put(readBuf)
	data, err := protocol.ReadPacket(cli.conn, enc, readBuf)
	if err != nil {
		return false, fmt.Errorf("read packet: %w", err)
	}

	// Handler writes response payload into sendBuf[2:]
	n, ok, err := srv.handler.HandlePacket(ctx, cli, data, sendBuf[2:])
	if err != nil {
		return false, fmt.Errorf("handle packet: %w", err)
	}

	if n > 0 {
		if err := protocol.WritePacket(cli.conn, enc, sendBuf, n); err != nil {
			return false, fmt.Errorf("write packet: %w", err)
		}
	}

	return ok, nil
}
