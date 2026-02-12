package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeCharMoveToLocation is the opcode for CharMoveToLocation packet (S2C 0x01)
	OpcodeCharMoveToLocation = 0x01
)

// CharMoveToLocation packet (S2C 0x01) broadcasts character movement to other players.
// Sent when player moves (click-to-move). Other players in visibility range see the movement.
type CharMoveToLocation struct {
	ObjectID int32 // Character object ID
	TargetX  int32 // Target X coordinate
	TargetY  int32 // Target Y coordinate
	TargetZ  int32 // Target Z coordinate
	OriginX  int32 // Origin X coordinate (current position)
	OriginY  int32 // Origin Y coordinate (current position)
	OriginZ  int32 // Origin Z coordinate (current position)
}

// NewCharMoveToLocation creates CharMoveToLocation packet from Player and target location.
func NewCharMoveToLocation(player *model.Player, targetX, targetY, targetZ int32) CharMoveToLocation {
	loc := player.Location()
	return CharMoveToLocation{
		ObjectID: int32(player.CharacterID()),
		TargetX:  targetX,
		TargetY:  targetY,
		TargetZ:  targetZ,
		OriginX:  loc.X,
		OriginY:  loc.Y,
		OriginZ:  loc.Z,
	}
}

// Write serializes CharMoveToLocation packet to binary format.
func (p *CharMoveToLocation) Write() ([]byte, error) {
	w := packet.NewWriter(32)

	w.WriteByte(OpcodeCharMoveToLocation)
	w.WriteInt(p.ObjectID)
	w.WriteInt(p.TargetX)
	w.WriteInt(p.TargetY)
	w.WriteInt(p.TargetZ)
	w.WriteInt(p.OriginX)
	w.WriteInt(p.OriginY)
	w.WriteInt(p.OriginZ)

	return w.Bytes(), nil
}
