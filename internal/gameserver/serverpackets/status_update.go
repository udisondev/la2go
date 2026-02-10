package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeStatusUpdate is the opcode for StatusUpdate packet (S2C 0x0E)
	OpcodeStatusUpdate = 0x0E
)

// StatusUpdate attribute IDs for the attribute list.
const (
	AttrLevel      = 0x01
	AttrExp        = 0x02
	AttrSTR        = 0x03
	AttrDEX        = 0x04
	AttrCON        = 0x05
	AttrINT        = 0x06
	AttrWIT        = 0x07
	AttrMEN        = 0x08
	AttrCurrentHP  = 0x09
	AttrMaxHP      = 0x0A
	AttrCurrentMP  = 0x0B
	AttrMaxMP      = 0x0C
	AttrSP         = 0x0D
	AttrCurrentCP  = 0x21
	AttrMaxCP      = 0x22
	AttrKarma      = 0x1B
	AttrPvPFlag    = 0x1A
	AttrLoad       = 0x0E
	AttrMaxLoad    = 0x0F
)

// StatusUpdate packet (S2C 0x0E) sends character stat updates to client.
// Used to update HP/MP/CP bars and other stats without resending full UserInfo.
// Sent after UserInfo during spawn, and during gameplay when stats change.
type StatusUpdate struct {
	ObjectID   int32
	Attributes []StatusAttribute
}

// StatusAttribute represents a single stat update (ID + value).
type StatusAttribute struct {
	ID    int32
	Value int32
}

// NewStatusUpdate creates StatusUpdate packet with typical spawn attributes.
// Sends current/max HP/MP/CP which are critical for UI display.
// Phase 5.2: Fixed ObjectID bug (was CharacterID, now ObjectID).
func NewStatusUpdate(player *model.Player) *StatusUpdate {
	return &StatusUpdate{
		ObjectID: int32(player.ObjectID()), // Phase 5.2: Use ObjectID, not CharacterID
		Attributes: []StatusAttribute{
			{ID: AttrCurrentHP, Value: player.CurrentHP()},
			{ID: AttrMaxHP, Value: player.MaxHP()},
			{ID: AttrCurrentMP, Value: player.CurrentMP()},
			{ID: AttrMaxMP, Value: player.MaxMP()},
			{ID: AttrCurrentCP, Value: player.CurrentCP()},
			{ID: AttrMaxCP, Value: player.MaxCP()},
		},
	}
}

// NewStatusUpdateForTarget creates StatusUpdate packet for a target Character.
// Used when selecting a target to display HP/MP/CP bars.
// Phase 5.2: Target System support.
func NewStatusUpdateForTarget(character *model.Character) *StatusUpdate {
	return &StatusUpdate{
		ObjectID: int32(character.ObjectID()),
		Attributes: []StatusAttribute{
			{ID: AttrCurrentHP, Value: character.CurrentHP()},
			{ID: AttrMaxHP, Value: character.MaxHP()},
			{ID: AttrCurrentMP, Value: character.CurrentMP()},
			{ID: AttrMaxMP, Value: character.MaxMP()},
			{ID: AttrCurrentCP, Value: character.CurrentCP()},
			{ID: AttrMaxCP, Value: character.MaxCP()},
		},
	}
}

// Write serializes StatusUpdate packet to binary format.
func (p *StatusUpdate) Write() ([]byte, error) {
	// Buffer size: opcode(1) + objectID(4) + attrCount(4) + attrs(8 per attr)
	w := packet.NewWriter(9 + len(p.Attributes)*8)

	w.WriteByte(OpcodeStatusUpdate)
	w.WriteInt(p.ObjectID)
	w.WriteInt(int32(len(p.Attributes)))

	for _, attr := range p.Attributes {
		w.WriteInt(attr.ID)
		w.WriteInt(attr.Value)
	}

	return w.Bytes(), nil
}
