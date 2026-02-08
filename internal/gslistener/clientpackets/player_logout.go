package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gslistener/packet"
)

// PlayerLogout [0x03] — GS → LS игрок вышел
//
// Format:
//   [opcode 0x03]
//   [account UTF-16LE null-terminated]
type PlayerLogout struct {
	Account string
}

// Parse парсит пакет PlayerLogout из body (без opcode).
func (p *PlayerLogout) Parse(body []byte) error {
	r := packet.NewReader(body)

	// Читаем account
	account, err := r.ReadString()
	if err != nil {
		return fmt.Errorf("reading account: %w", err)
	}

	p.Account = account
	return nil
}
