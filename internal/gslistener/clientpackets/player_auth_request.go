package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gslistener/packet"
	"github.com/udisondev/la2go/internal/login"
)

// PlayerAuthRequest [0x05] — GS → LS запрос валидации сессии игрока
//
// Format:
//   [opcode 0x05]
//   [account UTF-16LE null-terminated]
//   [playOkID1 int32]
//   [playOkID2 int32]
//   [loginOkID1 int32]
//   [loginOkID2 int32]
type PlayerAuthRequest struct {
	Account    string
	SessionKey login.SessionKey
}

// Parse парсит пакет PlayerAuthRequest из body (без opcode).
func (p *PlayerAuthRequest) Parse(body []byte) error {
	r := packet.NewReader(body)

	// Читаем account
	account, err := r.ReadString()
	if err != nil {
		return fmt.Errorf("reading account: %w", err)
	}
	p.Account = account

	// Читаем playOkID1
	playOkID1, err := r.ReadInt()
	if err != nil {
		return fmt.Errorf("reading playOkID1: %w", err)
	}

	// Читаем playOkID2
	playOkID2, err := r.ReadInt()
	if err != nil {
		return fmt.Errorf("reading playOkID2: %w", err)
	}

	// Читаем loginOkID1
	loginOkID1, err := r.ReadInt()
	if err != nil {
		return fmt.Errorf("reading loginOkID1: %w", err)
	}

	// Читаем loginOkID2
	loginOkID2, err := r.ReadInt()
	if err != nil {
		return fmt.Errorf("reading loginOkID2: %w", err)
	}

	p.SessionKey = login.SessionKey{
		LoginOkID1: loginOkID1,
		LoginOkID2: loginOkID2,
		PlayOkID1:  playOkID1,
		PlayOkID2:  playOkID2,
	}

	return nil
}
