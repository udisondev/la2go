package serverpackets

import (
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// Wait type constants for ChangeWaitType packet.
// Java reference: ChangeWaitType.java (WT_SITTING=0, WT_STANDING=1).
const (
	WaitTypeSitting      int32 = 0
	WaitTypeStanding     int32 = 1
	WaitTypeFakeDeathOn  int32 = 2
	WaitTypeFakeDeathOff int32 = 3
)

// ChangeWaitType notifies the client about a sit/stand/fake-death state change.
// Opcode: 0x2F.
// Java reference: ChangeWaitType.java.
type ChangeWaitType struct {
	ObjectID int32
	WaitType int32
	X, Y, Z  int32
}

// NewChangeWaitType creates a ChangeWaitType packet from player and wait type.
func NewChangeWaitType(player *model.Player, waitType int32) *ChangeWaitType {
	loc := player.Location()
	return &ChangeWaitType{
		ObjectID: int32(player.ObjectID()),
		WaitType: waitType,
		X:        loc.X,
		Y:        loc.Y,
		Z:        loc.Z,
	}
}

// Write serializes ChangeWaitType to bytes.
func (p *ChangeWaitType) Write() ([]byte, error) {
	w := packet.NewWriter(64)
	w.WriteByte(0x2F)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.WaitType)
	w.WriteInt(p.X)
	w.WriteInt(p.Y)
	w.WriteInt(p.Z)
	return w.Bytes(), nil
}
