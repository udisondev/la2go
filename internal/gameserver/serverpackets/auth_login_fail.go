package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeAuthLoginFail is the opcode for AuthLoginFail (S2C 0x14).
// Sent when GameServer rejects client's session key validation.
//
// Phase 8.1: Used by handler.go for session rejection.
// Java reference: LoginFail.java (gameserver network, NOT loginserver).
const OpcodeAuthLoginFail = 0x14

// AuthLoginFail reasons.
const (
	AuthFailReasonSystemError       = 0x01
	AuthFailReasonAccessDenied      = 0x10
	AuthFailReasonAccountInUse      = 0x07
	AuthFailReasonAccountBanned     = 0x09
	AuthFailReasonServerMaintenance = 0x10
)

// AuthLoginFail sends login failure to game client.
//
// Packet structure:
//   - opcode (byte) — 0x12
//   - reason (int32) — failure reason code
type AuthLoginFail struct {
	Reason int32
}

// NewAuthLoginFail creates a new AuthLoginFail packet.
func NewAuthLoginFail(reason int32) *AuthLoginFail {
	return &AuthLoginFail{Reason: reason}
}

// Write serializes AuthLoginFail packet.
func (p *AuthLoginFail) Write() ([]byte, error) {
	w := packet.NewWriter(5)
	w.WriteByte(OpcodeAuthLoginFail)
	w.WriteInt(p.Reason)
	return w.Bytes(), nil
}
