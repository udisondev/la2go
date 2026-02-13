package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeL2FriendSay is the S2C opcode 0xFD.
// Delivers a friend PM to the receiving client.
const OpcodeL2FriendSay byte = 0xFD

// L2FriendSay delivers a friend PM to a player.
type L2FriendSay struct {
	Sender   string
	Receiver string
	Message  string
}

// Write serializes the packet.
func (p *L2FriendSay) Write() ([]byte, error) {
	w := packet.NewWriter(128)
	w.WriteByte(OpcodeL2FriendSay)
	w.WriteInt(0) // unknown field (Java reference)
	w.WriteString(p.Receiver)
	w.WriteString(p.Sender)
	w.WriteString(p.Message)
	return w.Bytes(), nil
}
