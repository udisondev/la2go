package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gslistener/packet"
)

// GameServerAuth [0x01] — GS → LS запрос регистрации
//
// Format (после удаления opcode, совпадает с Java GameServerAuth.java):
//
//	[id byte]                   // желаемый server ID
//	[acceptAlternate byte]      // 0x01 = принимать альтернативный ID, 0x00 = нет
//	[reserved byte]             // зарезервировано
//	[port int16]                // порт для клиентов (readShort, 2 bytes)
//	[maxPlayers int32]          // максимум игроков (readInt, 4 bytes)
//	[hexIdSize int32]           // размер hexId (readInt, 4 bytes)
//	[hexId byte[hexIdSize]]     // уникальный hex ID сервера (переменная длина)
//	[hostPairs int32]           // количество пар subnet/host (readInt, 4 bytes)
//	[hosts string[2*hostPairs]] // subnet и host строки (readString, UTF-16LE)
type GameServerAuth struct {
	ID              byte
	AcceptAlternate bool
	ReserveHost     bool
	Port            int16
	MaxPlayers      int32
	HexID           []byte
	Hosts           []string
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

	// Читаем port (readShort — 2 bytes, как в Java)
	port, err := r.ReadShort()
	if err != nil {
		return fmt.Errorf("reading port: %w", err)
	}
	p.Port = port

	// Читаем maxPlayers (readInt — 4 bytes, как в Java)
	maxPlayers, err := r.ReadInt()
	if err != nil {
		return fmt.Errorf("reading maxPlayers: %w", err)
	}
	p.MaxPlayers = maxPlayers

	// Читаем hexId (переменная длина: сначала size как int32, потом bytes)
	hexIdSize, err := r.ReadInt()
	if err != nil {
		return fmt.Errorf("reading hexId size: %w", err)
	}
	if hexIdSize < 0 || hexIdSize > 256 {
		return fmt.Errorf("invalid hexId size: %d", hexIdSize)
	}
	hexId, err := r.ReadBytes(int(hexIdSize))
	if err != nil {
		return fmt.Errorf("reading hexId: %w", err)
	}
	p.HexID = hexId

	// Читаем hosts (Java: size = 2 * readInt(), _hosts = new String[size])
	hostPairs, err := r.ReadInt()
	if err != nil {
		return fmt.Errorf("reading host pairs count: %w", err)
	}
	hostCount := int(2 * hostPairs)
	if hostCount < 0 || hostCount > 100 {
		return fmt.Errorf("invalid host count: %d", hostCount)
	}
	p.Hosts = make([]string, 0, hostCount)
	for range hostCount {
		s, err := r.ReadString()
		if err != nil {
			return fmt.Errorf("reading host string: %w", err)
		}
		p.Hosts = append(p.Hosts, s)
	}

	return nil
}
