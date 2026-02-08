package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gslistener/packet"
)

// PlayerInGame [0x02] — GS → LS список онлайн игроков
//
// Format:
//   [opcode 0x02]
//   [count short]
//   for each count:
//     [account UTF-16LE null-terminated]
type PlayerInGame struct {
	Accounts []string
}

// Parse парсит пакет PlayerInGame из body (без opcode).
func (p *PlayerInGame) Parse(body []byte) error {
	r := packet.NewReader(body)

	// Читаем count
	count, err := r.ReadShort()
	if err != nil {
		return fmt.Errorf("reading count: %w", err)
	}

	if count < 0 || count > 10000 {
		return fmt.Errorf("invalid count: %d", count)
	}

	// Читаем аккаунты
	p.Accounts = make([]string, 0, count)
	for range count {
		account, err := r.ReadString()
		if err != nil {
			return fmt.Errorf("reading account: %w", err)
		}
		p.Accounts = append(p.Accounts, account)
	}

	return nil
}
