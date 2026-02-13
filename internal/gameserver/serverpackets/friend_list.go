package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeFriendList is the opcode for FriendList packet (S2C 0xFA).
const OpcodeFriendList = 0xFA

// FriendInfo contains friend data for the FriendList packet.
type FriendInfo struct {
	ObjectID int32
	Name     string
	IsOnline bool
}

// FriendListPacket sends the full friends list to the client.
type FriendListPacket struct {
	Friends []FriendInfo
}

// NewFriendListPacket creates a new FriendList packet.
func NewFriendListPacket(friends []FriendInfo) *FriendListPacket {
	return &FriendListPacket{Friends: friends}
}

// Write serializes the FriendList packet.
func (p *FriendListPacket) Write() ([]byte, error) {
	w := packet.NewWriter(128)
	_ = w.WriteByte(OpcodeFriendList)
	w.WriteInt(int32(len(p.Friends)))

	for _, f := range p.Friends {
		w.WriteInt(f.ObjectID)
		w.WriteString(f.Name)

		online := int32(0)
		if f.IsOnline {
			online = 1
		}
		w.WriteInt(online)

		// objectID if online, 0 otherwise
		if f.IsOnline {
			w.WriteInt(f.ObjectID)
		} else {
			w.WriteInt(0)
		}
	}

	return w.Bytes(), nil
}
