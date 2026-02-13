package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeL2Friend is the opcode for friend add/remove notification (S2C 0xFB).
const OpcodeL2Friend = 0xFB

// L2Friend action types.
const (
	FriendActionAdd    = 1
	FriendActionModify = 2
	FriendActionRemove = 3
)

// L2FriendPacket notifies the client about a friend list change.
type L2FriendPacket struct {
	Action   int32
	Name     string
	IsOnline bool
	ObjectID int32
}

// NewL2FriendPacket creates a new L2Friend notification packet.
func NewL2FriendPacket(action int32, name string, isOnline bool, objectID int32) *L2FriendPacket {
	return &L2FriendPacket{
		Action:   action,
		Name:     name,
		IsOnline: isOnline,
		ObjectID: objectID,
	}
}

// Write serializes the L2Friend packet.
func (p *L2FriendPacket) Write() ([]byte, error) {
	w := packet.NewWriter(64)
	_ = w.WriteByte(OpcodeL2Friend)
	w.WriteInt(p.Action)
	w.WriteInt(0) // unknown / reserved
	w.WriteString(p.Name)

	online := int32(0)
	if p.IsOnline {
		online = 1
	}
	w.WriteInt(online)
	w.WriteInt(p.ObjectID)

	return w.Bytes(), nil
}
