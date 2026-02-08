package gslistener

import (
	"context"
	"fmt"
	"log/slog"
	mathrand "math/rand/v2"
	"net"
	"sync"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/gslistener/serverpackets"
	"github.com/udisondev/la2go/internal/login"
)

const (
	rsaKeyPairCount = 10
)

// Server представляет GameServer↔LoginServer TCP listener
type Server struct {
	cfg     config.LoginServer
	db      *db.DB
	gsTable *gameserver.GameServerTable

	rsaKeyPairs [rsaKeyPairCount]*crypto.RSAKeyPair
	sendPool    *BytePool
	readPool    *BytePool
	handler     *Handler

	listener net.Listener
	mu       sync.Mutex
}

// NewServer создаёт новый GS listener server с предварительно сгенерированными RSA-512 ключами
func NewServer(
	cfg config.LoginServer,
	database *db.DB,
	gsTable *gameserver.GameServerTable,
	sessionManager *login.SessionManager,
) (*Server, error) {
	s := &Server{
		cfg:      cfg,
		db:       database,
		gsTable:  gsTable,
		sendPool: NewBytePool(constants.GSListenerSendBufSize),
		readPool: NewBytePool(constants.GSListenerReadBufSize),
		handler:  NewHandler(database, gsTable, sessionManager),
	}

	// Pre-generate RSA-512 key pairs
	slog.Info("generating RSA-512 key pairs for GS listener", "count", rsaKeyPairCount)
	for i := range rsaKeyPairCount {
		kp, err := crypto.GenerateRSAKeyPair512()
		if err != nil {
			return nil, fmt.Errorf("generating RSA-512 key pair %d: %w", i, err)
		}
		s.rsaKeyPairs[i] = kp
	}

	return s, nil
}

// Addr возвращает адрес, на котором слушает GS listener.
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

// Run начинает прослушивание подключений от GameServer.
// Создаёт listener на cfg.GSListenHost:cfg.GSListenPort и запускает accept loop.
func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.GSListenHost, s.cfg.GSListenPort)
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
	// Graceful shutdown
	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	var wg sync.WaitGroup
	go func() {
		slog.Info("GS listener started", "address", ln.Addr())
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
				slog.Error("failed to accept GS connection", "error", err)
				continue
			}
			wg.Go(func() {
				handleConnection(ctx, srv, conn)
			})
		}
	}
}

func handleConnection(ctx context.Context, srv *Server, conn net.Conn) {
	defer conn.Close()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		slog.Error("failed to split host port", "connection", conn.RemoteAddr(), "error", err)
		return
	}

	slog.Info("GS connected", "remote", host)

	// Select random RSA-512 key
	rsaKeyPair := srv.rsaKeyPairs[mathrand.IntN(rsaKeyPairCount)]

	// Create GSConnection
	gsConn, err := NewGSConnection(conn, rsaKeyPair)
	if err != nil {
		slog.Error("failed to create GS connection", "err", err, "remote", host)
		return
	}

	// Send InitLS packet — write payload into sendBuf[constants.PacketHeaderSize:], then WritePacket encrypts in-place
	sendBuf := srv.sendPool.Get(constants.GSListenerSendBufSize)
	n := serverpackets.InitLS(sendBuf[constants.PacketHeaderSize:], constants.ProtocolRevisionInterlude, rsaKeyPair.ScrambledModulus)
	if err := WritePacket(conn, gsConn.BlowfishCipher(), sendBuf, n); err != nil {
		srv.sendPool.Put(sendBuf)
		slog.Error("failed to send InitLS packet", "err", err, "remote", host)
		return
	}
	srv.sendPool.Put(sendBuf)
	slog.Debug("InitLS packet sent", "remote", host)

	// Packet loop
	for {
		select {
		case <-ctx.Done():
			cleanup(gsConn, host)
			return
		default:
			if ok, err := handlePacket(ctx, gsConn, srv); !ok {
				cleanup(gsConn, host)
				return
			} else if err != nil {
				slog.Error("failed to handle packet", "remote", host, "error", err)
			}
		}
	}
}

func cleanup(conn *GSConnection, host string) {
	info := conn.GameServerInfo()
	if info != nil {
		info.SetDown()
		slog.Info("GS disconnected", "id", info.ID(), "remote", host)
	} else {
		slog.Info("GS disconnected", "remote", host)
	}
}

func handlePacket(
	ctx context.Context,
	conn *GSConnection,
	srv *Server,
) (bool, error) {
	sendBuf := srv.sendPool.Get(constants.GSListenerSendBufSize)
	defer srv.sendPool.Put(sendBuf)
	readBuf := srv.readPool.Get(constants.GSListenerReadBufSize)
	defer srv.readPool.Put(readBuf)

	data, err := ReadPacket(conn.conn, conn.BlowfishCipher(), readBuf)
	if err != nil {
		return false, fmt.Errorf("read packet: %w", err)
	}

	// Handler writes response payload into sendBuf[constants.PacketHeaderSize:]
	n, ok, err := srv.handler.HandlePacket(ctx, conn, data, sendBuf[constants.PacketHeaderSize:])
	if err != nil {
		return false, fmt.Errorf("handle packet: %w", err)
	}

	if n > 0 {
		if err := WritePacket(conn.conn, conn.BlowfishCipher(), sendBuf, n); err != nil {
			return false, fmt.Errorf("write packet: %w", err)
		}
	}

	return ok, nil
}
