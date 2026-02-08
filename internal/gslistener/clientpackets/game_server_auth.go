package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gslistener/packet"
)

// GameServerAuth [0x01] — GS → LS запрос регистрации
//
// Format (после удаления opcode):
//   [id byte]                   // желаемый server ID
//   [acceptAlternate byte]      // 0x01 = принимать альтернативный ID, 0x00 = нет
//   [reserved byte]             // зарезервировано
//   [maxPlayers uint16]         // максимум игроков (2 bytes)
//   [port uint16]               // порт для клиентов (2 bytes)
//   [gameHosts string]          // null-terminated UTF-16LE список хостов
//   [hexId byte[32]]            // уникальный hex ID сервера (фиксированный размер 32 байта)
type GameServerAuth struct {
	ID              byte
	AcceptAlternate bool
	ReserveHost     bool
	Port            int16
	MaxPlayers      int32
	HexID           []byte
	Hosts           []HostEntry
}

// HostEntry представляет пару subnet/host.
type HostEntry struct {
	Subnet string
	Host   string
}

// Parse парсит пакет GameServerAuth из body (без opcode).
func (p *GameServerAuth) Parse(body []byte) error {
	r := packet.NewReader(body)

	// Читаем ID
	id, err := r.ReadByte()
	if err != nil {
		return fmt.Errorf("reading id: %w", err)
	}
	p.ID = id

	// Читаем acceptAlternate
	acceptAlt, err := r.ReadByte()
	if err != nil {
		return fmt.Errorf("reading acceptAlternate: %w", err)
	}
	p.AcceptAlternate = acceptAlt != 0

	// Читаем reserved
	_, err = r.ReadByte()
	if err != nil {
		return fmt.Errorf("reading reserved: %w", err)
	}

	// Читаем maxPlayers (uint16, НЕ int32!)
	maxPlayersShort, err := r.ReadShort()
	if err != nil {
		return fmt.Errorf("reading maxPlayers: %w", err)
	}
	p.MaxPlayers = int32(maxPlayersShort)

	// Читаем port (uint16)
	port, err := r.ReadShort()
	if err != nil {
		return fmt.Errorf("reading port: %w", err)
	}
	p.Port = port

	// Читаем gameHosts (простой null terminator для пустого списка)
	// TODO: proper parsing если hosts не пустые
	gameHostsByte, err := r.ReadByte()
	if err != nil {
		return fmt.Errorf("reading gameHosts terminator: %w", err)
	}
	_ = gameHostsByte // для пустого списка это 0x00

	// Читаем hexId (фиксированный размер 32 байта)
	const hexIdSize = 32
	hexId, err := r.ReadBytes(hexIdSize)
	if err != nil {
		return fmt.Errorf("reading hexId: %w", err)
	}
	p.HexID = hexId

	return nil
}
