package login

import (
	"context"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/login/serverpackets"
)

// Client packet opcodes
const (
	OpcodeRequestAuthLogin   = 0x00
	OpcodeRequestServerLogin = 0x02
	OpcodeRequestServerList  = 0x05
	OpcodeAuthGameGuard      = 0x07
)

// Handler processes login packets. Singleton — один на сервер.
type Handler struct {
	accounts       AccountRepository
	cfg            config.LoginServer
	sessionManager *SessionManager
}

// NewHandler creates a packet handler.
func NewHandler(accounts AccountRepository, cfg config.LoginServer, sessionManager *SessionManager) *Handler {
	return &Handler{
		accounts:       accounts,
		cfg:            cfg,
		sessionManager: sessionManager,
	}
}

// HandlePacket dispatches a decrypted packet to the appropriate handler.
// Writes response into buf. Returns: n — bytes written to buf (0 = nothing to send),
// ok — true if connection stays open (false = close after sending).
func (h *Handler) HandlePacket(
	ctx context.Context,
	client *Client,
	data, buf []byte,
) (int, bool, error) {
	if len(data) == 0 {
		return 0, false, fmt.Errorf("empty packet data")
	}

	opcode := data[0]
	body := data[1:]

	switch opcode {
	case OpcodeAuthGameGuard:
		return handleAuthGameGuard(client, body, buf)
	case OpcodeRequestAuthLogin:
		return h.handleRequestAuthLogin(ctx, client, body, buf)
	case OpcodeRequestServerList:
		return handleRequestServerList(client, body, buf, h.cfg.GameServers)
	case OpcodeRequestServerLogin:
		return handleRequestServerLogin(
			client,
			body,
			buf,
			h.cfg.ShowLicence,
			h.cfg.GameServers,
		)
	default:
		slog.Warn("unknown login packet opcode", "opcode", fmt.Sprintf("0x%02X", opcode), "client", client.IP())
		return 0, true, nil
	}
}

func closeFail(buf []byte, reason byte) (int, bool) {
	return serverpackets.LoginFail(buf, reason), false
}

// handleAuthGameGuard processes opcode 0x07 in state CONNECTED.
func handleAuthGameGuard(client *Client, data, buf []byte) (int, bool, error) {
	if client.State() != StateConnected {
		slog.Warn("AuthGameGuard in wrong state", "state", client.State(), "client", client.IP())
		return 0, true, nil
	}

	if len(data) < 4 {
		return 0, false, fmt.Errorf("AuthGameGuard packet too short: %d", len(data))
	}

	sessionID := int32(binary.LittleEndian.Uint32(data[:4]))

	if sessionID != client.SessionID() {
		slog.Warn("session ID mismatch in AuthGameGuard",
			"expected", client.SessionID(),
			"got", sessionID,
			"client", client.IP())
		n, ok := closeFail(buf, serverpackets.ReasonAccessFailed)
		return n, ok, nil
	}

	client.SetState(StateAuthedGG)
	slog.Debug("GameGuard auth OK", "client", client.IP())
	return serverpackets.GGAuth(buf, client.SessionID()), true, nil
}

// handleRequestAuthLogin processes opcode 0x00 in state AUTHED_GG.
func (h *Handler) handleRequestAuthLogin(
	ctx context.Context,
	client *Client,
	data, buf []byte,
) (int, bool, error) {
	if client.State() != StateAuthedGG {
		slog.Warn("RequestAuthLogin in wrong state", "state", client.State(), "client", client.IP())
		return 0, true, nil
	}

	remaining := len(data)
	var raw1 []byte

	if remaining >= 256 {
		raw1 = data[:128]
	} else if remaining >= 128 {
		raw1 = data[:128]
	} else {
		slog.Warn("RequestAuthLogin packet too short", "size", remaining, "client", client.IP())
		n, ok := closeFail(buf, serverpackets.ReasonAccessFailed)
		return n, ok, nil
	}

	decrypted, err := crypto.RSADecryptNoPadding(client.RSAKeyPair().PrivateKey, raw1)
	if err != nil {
		slog.Warn("RSA decryption failed", "err", err, "client", client.IP())
		n, ok := closeFail(buf, serverpackets.ReasonAccessFailed)
		return n, ok, nil
	}

	login := strings.TrimRight(string(decrypted[constants.AuthLoginUsernameOffset:constants.AuthLoginUsernameOffset+constants.AuthLoginUsernameMaxLength]), "\x00 ")
	password := strings.TrimRight(string(decrypted[constants.AuthLoginPasswordOffset:constants.AuthLoginPasswordOffset+constants.AuthLoginPasswordMaxLength]), "\x00 ")

	login = strings.ToLower(strings.TrimSpace(login))
	password = strings.TrimSpace(password)

	if login == "" || password == "" {
		slog.Warn("empty login or password", "client", client.IP())
		n, ok := closeFail(buf, serverpackets.ReasonUserOrPassWrong)
		return n, ok, nil
	}

	slog.Info("auth attempt", "login", login, "client", client.IP())

	passHash := db.HashPassword(password)

	acc, err := h.accounts.GetAccount(ctx, login)
	if err != nil {
		slog.Error("database error during auth", "err", err, "client", client.IP())
		n, ok := closeFail(buf, serverpackets.ReasonSystemError)
		return n, ok, nil
	}

	if acc == nil {
		if h.cfg.AutoCreateAccounts {
			// Атомарная операция: получить существующий или создать новый
			// Thread-safe: использует INSERT ... ON CONFLICT для защиты от race conditions
			acc, err = h.accounts.GetOrCreateAccount(ctx, login, passHash, client.IP())
			if err != nil {
				slog.Error("failed to get or create account", "err", err, "client", client.IP())
				n, ok := closeFail(buf, serverpackets.ReasonSystemError)
				return n, ok, nil
			}
		} else {
			n, ok := closeFail(buf, serverpackets.ReasonUserOrPassWrong)
			return n, ok, nil
		}
	}

	if subtle.ConstantTimeCompare([]byte(acc.PasswordHash), []byte(passHash)) != 1 {
		slog.Warn("wrong password", "login", login, "client", client.IP())
		n, ok := closeFail(buf, serverpackets.ReasonUserOrPassWrong)
		return n, ok, nil
	}

	if acc.AccessLevel < 0 {
		slog.Warn("account banned", "login", login, "client", client.IP())
		n := serverpackets.AccountKicked(buf, serverpackets.ReasonPermanentlyBanned)
		return n, false, nil
	}

	client.SetAccount(login)
	client.SetState(StateAuthedLogin)
	sk := NewSessionKey()
	client.SetSessionKey(sk)

	// Сохраняем сессию для последующей валидации через GameServer
	h.sessionManager.Store(login, sk, client)

	if err := h.accounts.UpdateLastLogin(ctx, login, client.IP()); err != nil {
		slog.Error("failed to update last login", "err", err)
	}

	slog.Info("auth success", "login", login, "client", client.IP())

	if h.cfg.ShowLicence {
		return serverpackets.LoginOk(
			buf,
			sk.LoginOkID1,
			sk.LoginOkID2,
		), true, nil
	}
	return buildServerList(buf, h.cfg.GameServers), true, nil
}

// handleRequestServerList processes opcode 0x05 in state AUTHED_LOGIN.
func handleRequestServerList(
	client *Client,
	data, buf []byte,
	gameServers []config.GameServerEntry,
) (int, bool, error) {
	if client.State() != StateAuthedLogin {
		slog.Warn("RequestServerList in wrong state", "state", client.State(), "client", client.IP())
		return 0, true, nil
	}

	if len(data) < 8 {
		return 0, false, fmt.Errorf("RequestServerList packet too short: %d", len(data))
	}

	skey1 := int32(binary.LittleEndian.Uint32(data[:4]))
	skey2 := int32(binary.LittleEndian.Uint32(data[4:8]))

	sk := client.SessionKey()
	if !sk.CheckLoginPair(skey1, skey2) {
		slog.Warn("login pair mismatch in RequestServerList", "client", client.IP())
		n, ok := closeFail(buf, serverpackets.ReasonAccessFailed)
		return n, ok, nil
	}

	return buildServerList(buf, gameServers), true, nil
}

// handleRequestServerLogin processes opcode 0x02 in state AUTHED_LOGIN.
func handleRequestServerLogin(
	client *Client,
	data, buf []byte,
	showLicence bool,
	gameServers []config.GameServerEntry,
) (int, bool, error) {
	if client.State() != StateAuthedLogin {
		slog.Warn("RequestServerLogin in wrong state", "state", client.State(), "client", client.IP())
		return 0, true, nil
	}

	if len(data) < 9 {
		return 0, false, fmt.Errorf("RequestServerLogin packet too short: %d", len(data))
	}

	skey1 := int32(binary.LittleEndian.Uint32(data[:4]))
	skey2 := int32(binary.LittleEndian.Uint32(data[4:8]))
	serverIDByte := data[8]

	sk := client.SessionKey()
	if !sk.CheckLoginPair(skey1, skey2) {
		slog.Warn("login pair mismatch in RequestServerLogin", "client", client.IP())
		n, ok := closeFail(buf, serverpackets.ReasonAccessFailed)
		return n, ok, nil
	}

	var found bool
	for _, gs := range gameServers {
		if gs.ID == int(serverIDByte) {
			found = true
			break
		}
	}
	if !found {
		slog.Warn("unknown server requested", "serverId", serverIDByte, "client", client.IP())
		return serverpackets.PlayFail(buf, serverpackets.ReasonServerOverloaded), true, nil
	}

	slog.Info("server login OK", "login", client.Account(), "serverId", serverIDByte, "client", client.IP())
	return serverpackets.PlayOk(buf, sk.PlayOkID1, sk.PlayOkID2), true, nil
}

// buildServerList writes the server list packet into buf from config.
func buildServerList(buf []byte, gameServers []config.GameServerEntry) int {
	servers := make([]serverpackets.ServerInfo, 0, len(gameServers))
	for _, gs := range gameServers {
		servers = append(servers, serverpackets.ServerInfo{
			ID:             byte(gs.ID),
			IP:             net.ParseIP(gs.Host),
			Port:           int32(gs.Port),
			AgeLimit:       0,
			PvP:            false,
			CurrentPlayers: 0,
			MaxPlayers:     int16(constants.DefaultMaxPlayers),
			Status:         byte(constants.DefaultServerStatus),
			ServerType:     int32(constants.DefaultServerType),
			Brackets:       false,
			CharCount:      0,
		})
	}
	var lastServer byte
	if len(servers) > 0 {
		lastServer = servers[0].ID
	}
	return serverpackets.ServerList(buf, servers, lastServer)
}
