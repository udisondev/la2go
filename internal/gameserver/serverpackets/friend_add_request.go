package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeFriendAddRequest is the opcode for the friend invite dialog (S2C 0x7D).
const OpcodeFriendAddRequest = 0x7D

// FriendAddRequestPacket asks the target player to accept a friend invite.
type FriendAddRequestPacket struct {
	RequestorName string
}

// NewFriendAddRequest creates a new FriendAddRequest packet.
func NewFriendAddRequest(name string) *FriendAddRequestPacket {
	return &FriendAddRequestPacket{RequestorName: name}
}

// Write serializes the FriendAddRequest packet.
func (p *FriendAddRequestPacket) Write() ([]byte, error) {
	w := packet.NewWriter(32)
	_ = w.WriteByte(OpcodeFriendAddRequest)
	w.WriteString(p.RequestorName)
	w.WriteInt(0) // unknown / reserved
	return w.Bytes(), nil
}
