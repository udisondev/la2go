package gameserver

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/udisondev/la2go/internal/game/augment"
	"github.com/udisondev/la2go/internal/game/bbs"
	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/game/cursed"
	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/game/duel"
	"github.com/udisondev/la2go/internal/game/offlinetrade"
	"github.com/udisondev/la2go/internal/game/geo"
	"github.com/udisondev/la2go/internal/game/hall"
	"github.com/udisondev/la2go/internal/game/instance"
	"github.com/udisondev/la2go/internal/game/olympiad"
	"github.com/udisondev/la2go/internal/game/party"
	"github.com/udisondev/la2go/internal/game/quest"
	"github.com/udisondev/la2go/internal/game/sevensigns"
	"github.com/udisondev/la2go/internal/game/siege"
	"github.com/udisondev/la2go/internal/game/zone"
	"github.com/udisondev/la2go/internal/gameserver/admin"
	"github.com/udisondev/la2go/internal/gameserver/admin/commands"
	"github.com/udisondev/la2go/internal/gameserver/clan"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/html"
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
	partyManager  *party.Manager // Phase 7.3: party system
	zoneManager   *zone.Manager  // Phase 7.2: zone logic
	geoEngine     *geo.Engine    // Phase 7.1: GeoEngine pathfinding & LOS

	listener     net.Listener
	mu           sync.Mutex
	saveOnce     sync.Once
}

// NewServer creates a new GameServer.
func NewServer(cfg config.GameServer, sessionManager *login.SessionManager, charRepo CharacterRepository, persister PlayerPersister, dialogMgr *html.DialogManager, questMgr *quest.Manager) (*Server, error) {
	clientMgr := NewClientManager()

	writePool := NewBytePool(constants.GameServerWriteBufSize)
	clientMgr.SetWritePool(writePool)

	partyMgr := party.NewManager()

	zoneMgr := zone.NewManager()
	if err := zoneMgr.Init(); err != nil {
		slog.Warn("zone manager init failed (zones disabled)", "error", err)
	}

	geoEng := geo.NewEngine()
	if cfg.GeodataDir != "" {
		if err := geoEng.LoadGeodata(cfg.GeodataDir); err != nil {
			slog.Warn("geodata loading failed (pathfinding disabled)", "error", err)
		}
	}

	// Phase 17: Admin/User command handler
	adminHandler := admin.NewHandler()
	adminAdapter := NewAdminClientAdapter(clientMgr)
	commands.RegisterAll(adminHandler, adminAdapter)

	// Phase 18: Clan system
	clanTbl := clan.NewTable()

	// Phase 21: Siege system
	siegeMgr := siege.NewManager(siege.DefaultManagerConfig())

	// Phase 20: Duel system
	duelMgr := duel.NewManager()

	// Phase 22: Clan Halls
	hallTbl := hall.NewTable()

	// Phase 25: Seven Signs
	ssMgr := sevensigns.NewManager()

	// Phase 26: Instance Zones
	instanceMgr := instance.NewManager()

	// Phase 24: Olympiad + Hero System
	olympiadMgr := olympiad.NewOlympiad()

	// Phase 28: Augmentation System
	augmentSvc := augment.NewService()

	// Phase 30: Community Board
	bbsHandler := bbs.NewHandler()

	// Phase 32: Cursed Weapons
	cursedMgr := cursed.NewManager()

	// Phase 31: Offline Trade
	offlineCfg := offlinetrade.Config{
		Enabled:            cfg.OfflineTradeEnabled,
		MaxDays:            cfg.OfflineMaxDays,
		DisconnectFinished: cfg.OfflineDisconnectFinished,
		SetNameColor:       cfg.OfflineSetNameColor,
		NameColor:          cfg.OfflineNameColor,
		RealtimeSave:       cfg.OfflineTradeRealtimeSave,
		RestoreOnStartup:   cfg.OfflineRestoreOnStartup,
	}
	offlineSvc := offlinetrade.NewService(offlineCfg, nil) // repo set after DB init

	s := &Server{
		cfg:            cfg,
		sessionManager: sessionManager,
		charRepo:       charRepo,
		persister:      persister,
		sendPool:       NewBytePool(constants.DefaultSendBufSize),
		readPool:       NewBytePool(constants.DefaultReadBufSize),
		writePool:      writePool,
		handler:        NewHandler(sessionManager, clientMgr, charRepo, persister, partyMgr, zoneMgr, geoEng, dialogMgr, adminHandler, nil, questMgr, clanTbl, siegeMgr, duelMgr, hallTbl, ssMgr, instanceMgr, olympiadMgr, augmentSvc, bbsHandler, offlineSvc, cursedMgr, nil, nil),
		clientManager:  clientMgr,
		partyManager:   partyMgr,
		zoneManager:    zoneMgr,
		geoEngine:      geoEng,
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
// Called during graceful shutdown. Uses sync.Once to prevent double save.
func (s *Server) saveAllPlayers() {
	s.saveOnce.Do(func() {
		s.doSaveAllPlayers()
	})
}

func (s *Server) doSaveAllPlayers() {
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
	wg.Go(func() {
		slog.Info("game server started", "address", ln.Addr())
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
					return // listener closed, stop accept loop
				}
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
	done := make(chan struct{})
	defer close(done)
	defer conn.Close()

	// Cleanup function to unregister client on disconnect
	var accountName string
	var client *GameClient
	defer func() {
		if accountName != "" {
			srv.clientManager.Unregister(accountName)
			slog.Debug("client unregistered", "account", accountName)
		}
		// Phase 26: Remove player from instance on disconnect.
		if client != nil {
			if player := client.ActivePlayer(); player != nil {
				if srv.handler.instanceManager != nil && srv.handler.instanceManager.IsInInstance(player.ObjectID()) {
					if _, err := srv.handler.instanceManager.ExitInstance(player.ObjectID(), player.CharacterID()); err != nil {
						slog.Debug("instance cleanup on disconnect", "error", err)
					}
				}
			}
		}
		// Call OnDisconnection for graceful cleanup (Phase 4.17.7)
		// Handles delayed removal if player in combat (15-second delay to prevent combat logging)
		// Phase 31: Pass offline trade service for offline trade mode detection.
		if client != nil {
			OnDisconnection(ctx, client, srv.persister, srv.handler.offlineSvc)
		}
	}()

	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
		}
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
