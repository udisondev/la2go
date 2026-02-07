package login

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	mathrand "math/rand/v2"
	"net"
	"sync"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/login/serverpackets"
	"github.com/udisondev/la2go/internal/protocol"
)

const (
	rsaKeyPairCount    = 10
	blowfishKeySize    = 16
	defaultSendBufSize = 512
	defaultReadBufSize = 512
)

// Server is the LoginServer that accepts client connections on port 2106.
type Server struct {
	cfg config.LoginServer
	db  *db.DB

	rsaKeyPairs [rsaKeyPairCount]*crypto.RSAKeyPair
	sendPool    *BytePool
	readPool    *BytePool
	handler     *Handler
}

// NewServer creates a new LoginServer with pre-generated RSA key pairs.
// Blowfish keys are generated fresh per connection.
func NewServer(cfg config.LoginServer, database *db.DB) (*Server, error) {
	s := &Server{
		cfg:      cfg,
		db:       database,
		sendPool: NewBytePool(defaultSendBufSize),
		readPool: NewBytePool(defaultReadBufSize),
		handler:  NewHandler(database, cfg),
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

// generateBlowfishKey creates a fresh 16-byte random Blowfish key.
func generateBlowfishKey() ([]byte, error) {
	key := make([]byte, blowfishKeySize)
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

// Run begins listening for client connections.
func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.BindAddress, s.cfg.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", addr, err)
	}
	defer ln.Close()

	var wg sync.WaitGroup
	go func() {
		slog.Info("login server started", "address", addr)
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
	defer conn.Close()
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
		conn.Close()
		return
	}

	sendBuf := srv.sendPool.Get(defaultSendBufSize)
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
	sendBuf := srv.sendPool.Get(defaultSendBufSize)
	defer srv.sendPool.Put(sendBuf)
	readBuf := srv.readPool.Get(defaultReadBufSize)
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
