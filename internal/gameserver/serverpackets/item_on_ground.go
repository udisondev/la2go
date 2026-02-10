package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeItemOnGround is the opcode for ItemOnGround packet (S2C 0x0B)
	// Also known as DropItem in some L2J versions
	OpcodeItemOnGround = 0x0B
)

// ItemOnGround packet (S2C 0x0B) sends information about item on the ground.
// Sent when item enters visibility range or is dropped.
// Phase 4.10 Part 3: Basic implementation for visible items.
type ItemOnGround struct {
	DroppedItem *model.DroppedItem
}

// NewItemOnGround creates ItemOnGround packet from DroppedItem model.
func NewItemOnGround(droppedItem *model.DroppedItem) *ItemOnGround {
	return &ItemOnGround{
		DroppedItem: droppedItem,
	}
}

// Write serializes ItemOnGround packet to binary format.
// ItemOnGround is simpler than NpcInfo (just item appearance + position).
func (p *ItemOnGround) Write() ([]byte, error) {
	w := packet.NewWriter(128)

	loc := p.DroppedItem.Location()
	item := p.DroppedItem.Item()

	w.WriteByte(OpcodeItemOnGround)

	// Object ID (item instance on ground)
	w.WriteInt(int32(p.DroppedItem.ObjectID()))

	// Item Type ID (from items table)
	w.WriteInt(item.ItemType())

	// Position
	w.WriteInt(loc.X)
	w.WriteInt(loc.Y)
	w.WriteInt(loc.Z)

	// Stackable (1=yes, 0=no)
	// TODO Phase 5: read from item template
	w.WriteInt(1) // default stackable

	// Count (how many items in stack)
	w.WriteInt(item.Count())

	return w.Bytes(), nil
}
