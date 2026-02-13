package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/login"
)

const OpcodeAuthLogin = 0x08

// AuthLogin is sent by the client to authenticate with the GameServer.
// Contains the account name and SessionKey (obtained from LoginServer).
//
// Structure:
// - string: account name (UTF-16LE null-terminated)
// - int32[4]: SessionKey (playOkID2, playOkID1, loginOkID1, loginOkID2)
// - int32: unknown (seems to be always 0)
// - int32: unknown (seems to be always 0)
// - int32: unknown (seems to be always 0)
// - int32: unknown (seems to be always 0)
type AuthLogin struct {
	AccountName string
	SessionKey  login.SessionKey
}

// ParseAuthLogin parses an AuthLogin packet from the given data (without opcode).
func ParseAuthLogin(data []byte) (*AuthLogin, error) {
	r := packet.NewReader(data)

	accountName, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading account name: %w", err)
	}

	// Read SessionKey (4×int32)
	// Java order: playKey2 first, playKey1 second (see AuthLogin.java readImpl)
	playOkID2, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading playOkID2: %w", err)
	}

	playOkID1, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading playOkID1: %w", err)
	}

	loginOkID1, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading loginOkID1: %w", err)
	}

	loginOkID2, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading loginOkID2: %w", err)
	}

	// Skip 4 unknown int32 fields (seems to be always 0)
	for range 4 {
		if _, err := r.ReadInt(); err != nil {
			return nil, fmt.Errorf("reading unknown field: %w", err)
		}
	}

	return &AuthLogin{
		AccountName: accountName,
		SessionKey: login.SessionKey{
			PlayOkID1:  playOkID1,
			PlayOkID2:  playOkID2,
			LoginOkID1: loginOkID1,
			LoginOkID2: loginOkID2,
		},
	}, nil
}
