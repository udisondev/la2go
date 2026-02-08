package gslistener

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/gameserver"
	"github.com/udisondev/la2go/internal/gslistener/clientpackets"
	"github.com/udisondev/la2go/internal/gslistener/serverpackets"
	"github.com/udisondev/la2go/internal/login"
)

// Handler обрабатывает входящие пакеты от GameServer
type Handler struct {
	db             *db.DB
	gsTable        *gameserver.GameServerTable
	sessionManager *login.SessionManager
}

// NewHandler создаёт новый handler для GS↔LS пакетов
func NewHandler(database *db.DB, gsTable *gameserver.GameServerTable, sessionManager *login.SessionManager) *Handler {
	return &Handler{
		db:             database,
		gsTable:        gsTable,
		sessionManager: sessionManager,
	}
}

// HandlePacket диспетчеризирует пакет по (state, opcode) → handler function.
// Writes response into buf. Returns: n — bytes written to buf (0 = nothing to send),
// ok — true if connection stays open (false = close after sending).
func (h *Handler) HandlePacket(
	ctx context.Context,
	conn *GSConnection,
	data, buf []byte,
) (int, bool, error) {
	if len(data) == 0 {
		return 0, false, fmt.Errorf("empty packet")
	}

	opcode := data[0]
	body := data[1:]
	state := conn.State()

	switch state {
	case gameserver.GSStateConnected:
		switch opcode {
		case OpcodeGSBlowFishKey:
			return handleBlowFishKey(ctx, h, conn, body, buf)
		default:
			return 0, true, fmt.Errorf("invalid opcode 0x%02x for state CONNECTED", opcode)
		}

	case gameserver.GSStateBFConnected:
		switch opcode {
		case OpcodeGSGameServerAuth:
			return handleGameServerAuth(ctx, h, conn, body, buf)
		default:
			return 0, true, fmt.Errorf("invalid opcode 0x%02x for state BF_CONNECTED", opcode)
		}

	case gameserver.GSStateAuthed:
		switch opcode {
		case OpcodeGSPlayerInGame:
			return handlePlayerInGame(ctx, h, conn, body, buf)
		case OpcodeGSPlayerLogout:
			return handlePlayerLogout(ctx, h, conn, body, buf)
		case OpcodeGSPlayerAuthRequest:
			return handlePlayerAuthRequest(ctx, h, conn, body, buf)
		case OpcodeGSServerStatus:
			return handleServerStatus(ctx, h, conn, body, buf)
		default:
			return 0, false, fmt.Errorf("unknown opcode 0x%02x", opcode)
		}

	default:
		return 0, true, fmt.Errorf("invalid connection state: %v", state)
	}
}

// Placeholder handlers (P3.5 scope — stub implementations)
// Будут реализованы в P3.7-P3.9

func handleBlowFishKey(_ context.Context, _ *Handler, conn *GSConnection, body []byte, _ []byte) (int, bool, error) {
	// Парсим пакет
	var pkt clientpackets.BlowFishKey
	if err := pkt.Parse(body); err != nil {
		return 0, false, fmt.Errorf("parsing BlowFishKey packet: %w", err)
	}

	// RSA расшифровка ключа
	rsaKeyPair := conn.RSAKeyPair()
	decryptedBlock, err := crypto.RSADecryptNoPadding(rsaKeyPair.PrivateKey, pkt.EncryptedKey)
	if err != nil {
		return 0, false, fmt.Errorf("RSA decrypt failed: %w", err)
	}

	// RSA-512 расшифровывает в 64 байта, берём последние 40 байт (как в Java)
	const blowfishKeySize = 40
	if len(decryptedBlock) < blowfishKeySize {
		return 0, false, fmt.Errorf("decrypted block too short: got %d, want at least %d", len(decryptedBlock), blowfishKeySize)
	}

	// Берём последние 40 байт
	decryptedKey := decryptedBlock[len(decryptedBlock)-blowfishKeySize:]

	// Создаём новый Blowfish cipher
	newCipher, err := crypto.NewBlowfishCipher(decryptedKey)
	if err != nil {
		return 0, false, fmt.Errorf("creating new Blowfish cipher: %w", err)
	}

	// Переключаем cipher
	conn.SetBlowfishCipher(newCipher)

	// Переключаем состояние: CONNECTED → BF_CONNECTED
	conn.SetState(gameserver.GSStateBFConnected)

	slog.Info("BlowFishKey processed successfully", "ip", conn.IP(), "state", "BF_CONNECTED")

	// Не отправляем ответ, просто продолжаем
	return 0, true, nil
}

func handleGameServerAuth(_ context.Context, h *Handler, conn *GSConnection, body []byte, buf []byte) (int, bool, error) {
	// Парсим пакет
	var pkt clientpackets.GameServerAuth
	if err := pkt.Parse(body); err != nil {
		return 0, false, fmt.Errorf("parsing GameServerAuth packet: %w", err)
	}

	// Реализация handleRegProcess из Java (GameServerAuth.java:86-163)
	requestedID := int(pkt.ID)

	// Проверяем, зарегистрирован ли ID в БД
	existingInfo, exists := h.gsTable.GetByID(requestedID)

	if exists {
		// ID существует — проверяем hexID
		if bytes.Equal(existingInfo.HexID(), pkt.HexID) {
			// HexID совпадает — проверяем, не подключен ли уже
			if existingInfo.IsAuthed() {
				// Уже подключен — отказываем
				slog.Warn("GameServer already authenticated", "id", requestedID, "ip", conn.IP())
				n := serverpackets.LoginServerFail(buf, gameserver.ReasonAlreadyLoggedIn)
				return n, false, nil // close connection
			}

			// HexID совпадает и не подключен — регистрируем
			return finalizeRegistration(h, conn, existingInfo, pkt, buf)
		}

		// HexID не совпадает — пробуем альтернативный ID если разрешено
		if pkt.AcceptAlternate {
			// Пытаемся найти свободный ID
			newInfo := gameserver.NewGameServerInfo(0, pkt.HexID)
			assignedID, ok := h.gsTable.RegisterWithFirstAvailableID(newInfo, 127)
			if !ok {
				// Нет свободных ID
				slog.Warn("no free server ID available", "requested_id", requestedID, "ip", conn.IP())
				n := serverpackets.LoginServerFail(buf, gameserver.ReasonNoFreeID)
				return n, false, nil
			}

			slog.Info("registered GameServer with alternative ID", "requested_id", requestedID, "assigned_id", assignedID, "ip", conn.IP())
			return finalizeRegistration(h, conn, newInfo, pkt, buf)
		}

		// HexID не совпадает и альтернативный ID не разрешён — отказываем
		slog.Warn("wrong hexID", "id", requestedID, "ip", conn.IP())
		n := serverpackets.LoginServerFail(buf, gameserver.ReasonWrongHexID)
		return n, false, nil
	}

	// ID не существует — проверяем, разрешена ли регистрация новых серверов
	// TODO: добавить конфиг ACCEPT_NEW_GAMESERVER (пока всегда true)
	acceptNew := true

	if !acceptNew {
		slog.Warn("new GameServer registration not allowed", "id", requestedID, "ip", conn.IP())
		n := serverpackets.LoginServerFail(buf, gameserver.ReasonWrongHexID)
		return n, false, nil
	}

	// Регистрируем новый сервер
	newInfo := gameserver.NewGameServerInfo(requestedID, pkt.HexID)
	if !h.gsTable.Register(requestedID, newInfo) {
		// ID занят (race condition)
		slog.Warn("server ID reserved (race condition)", "id", requestedID, "ip", conn.IP())
		n := serverpackets.LoginServerFail(buf, gameserver.ReasonIDReserved)
		return n, false, nil
	}

	slog.Info("registered new GameServer", "id", requestedID, "ip", conn.IP())
	return finalizeRegistration(h, conn, newInfo, pkt, buf)
}

// finalizeRegistration завершает регистрацию GameServer: обновляет info, отправляет AuthResponse.
func finalizeRegistration(h *Handler, conn *GSConnection, info *gameserver.GameServerInfo, pkt clientpackets.GameServerAuth, buf []byte) (int, bool, error) {
	// Обновляем информацию о сервере
	info.SetPort(int(pkt.Port))
	info.SetMaxPlayers(int(pkt.MaxPlayers))

	// Конвертируем hosts
	hosts := make([]string, len(pkt.Hosts))
	for i, host := range pkt.Hosts {
		hosts[i] = host.Host
	}
	info.SetHosts(hosts)

	// Помечаем как аутентифицированный
	info.SetAuthed(true)

	// Привязываем к соединению
	conn.AttachGameServerInfo(info)

	// Переключаем состояние
	conn.SetState(gameserver.GSStateAuthed)

	// Отправляем AuthResponse
	serverID := byte(info.ID())
	serverName := fmt.Sprintf("Server %d", info.ID()) // TODO: загружать из конфига/БД
	n := serverpackets.AuthResponse(buf, serverID, serverName)

	slog.Info("GameServer authenticated successfully",
		"id", info.ID(),
		"port", info.Port(),
		"maxPlayers", info.MaxPlayers(),
		"ip", conn.IP())

	return n, true, nil
}

func handlePlayerInGame(_ context.Context, _ *Handler, conn *GSConnection, body []byte, _ []byte) (int, bool, error) {
	// Парсим пакет
	var pkt clientpackets.PlayerInGame
	if err := pkt.Parse(body); err != nil {
		return 0, false, fmt.Errorf("parsing PlayerInGame packet: %w", err)
	}

	// Добавляем всех игроков в список онлайн
	for _, account := range pkt.Accounts {
		conn.AddAccount(account)
	}

	gsInfo := conn.GameServerInfo()
	if gsInfo != nil {
		slog.Info("players registered as online",
			"count", len(pkt.Accounts),
			"server_id", gsInfo.ID(),
			"ip", conn.IP())
	}

	// Не отправляем ответ
	return 0, true, nil
}

func handlePlayerLogout(_ context.Context, _ *Handler, conn *GSConnection, body []byte, _ []byte) (int, bool, error) {
	// Парсим пакет
	var pkt clientpackets.PlayerLogout
	if err := pkt.Parse(body); err != nil {
		return 0, false, fmt.Errorf("parsing PlayerLogout packet: %w", err)
	}

	// Удаляем игрока из списка онлайн
	conn.RemoveAccount(pkt.Account)

	gsInfo := conn.GameServerInfo()
	if gsInfo != nil {
		slog.Info("player logged out", "account", pkt.Account, "server_id", gsInfo.ID(), "ip", conn.IP())
	}

	// Не отправляем ответ
	return 0, true, nil
}

func handlePlayerAuthRequest(_ context.Context, h *Handler, _ *GSConnection, body []byte, buf []byte) (int, bool, error) {
	// Парсим пакет
	var pkt clientpackets.PlayerAuthRequest
	if err := pkt.Parse(body); err != nil {
		return 0, false, fmt.Errorf("parsing PlayerAuthRequest packet: %w", err)
	}

	// TODO: добавить флаг showLicence из конфига (пока используем false)
	showLicence := false

	// Валидируем SessionKey через SessionManager
	valid := h.sessionManager.Validate(pkt.Account, pkt.SessionKey, showLicence)

	if valid {
		// Удаляем сессию (игрок переходит на GS)
		h.sessionManager.Remove(pkt.Account)
		slog.Info("player session validated successfully", "account", pkt.Account)
	} else {
		slog.Warn("player session validation failed", "account", pkt.Account)
	}

	// Отправляем PlayerAuthResponse
	n := serverpackets.PlayerAuthResponse(buf, pkt.Account, valid)
	return n, true, nil
}

func handleServerStatus(_ context.Context, _ *Handler, conn *GSConnection, body []byte, _ []byte) (int, bool, error) {
	// Парсим пакет
	var pkt clientpackets.ServerStatus
	if err := pkt.Parse(body); err != nil {
		return 0, false, fmt.Errorf("parsing ServerStatus packet: %w", err)
	}

	gsInfo := conn.GameServerInfo()
	if gsInfo == nil {
		return 0, false, fmt.Errorf("ServerStatus received but GameServer not authenticated")
	}

	// Обновляем атрибуты сервера
	// Согласно ServerStatus.java:66 и gameserver/types.go:64-71
	for _, attr := range pkt.Attributes {
		switch attr.ID {
		case 0: // showingBrackets
			gsInfo.SetShowingBrackets(attr.Value != 0)
		case 1: // serverType
			gsInfo.SetServerType(int(attr.Value))
		case 2: // status
			gsInfo.SetStatus(int(attr.Value))
		case 3: // ageLimit
			gsInfo.SetAgeLimit(int(attr.Value))
		case 4: // maxPlayers
			gsInfo.SetMaxPlayers(int(attr.Value))
		default:
			slog.Warn("unknown ServerStatus attribute", "id", attr.ID, "value", attr.Value)
		}
	}

	slog.Info("server status updated",
		"server_id", gsInfo.ID(),
		"status", gsInfo.Status(),
		"maxPlayers", gsInfo.MaxPlayers(),
		"ip", conn.IP())

	// Не отправляем ответ
	return 0, true, nil
}
