package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeInventoryUpdate is the opcode for InventoryUpdate packet (S2C 0x27).
	// Java reference: InventoryUpdate.java
	OpcodeInventoryUpdate = 0x27
)

// InventoryUpdate change types.
const (
	InvUpdateAdd    int16 = 1 // New item added
	InvUpdateModify int16 = 2 // Item modified (count, enchant, etc.)
	InvUpdateRemove int16 = 3 // Item removed
)

// InvUpdateEntry represents a single item change in InventoryUpdate.
type InvUpdateEntry struct {
	ChangeType int16      // 1=add, 2=modify, 3=remove
	Item       *model.Item
}

// InventoryUpdate packet (S2C 0x27) sends incremental inventory changes.
// Used after equip/unequip, item pickup, item drop, trade, etc.
// More efficient than resending the full InventoryItemList.
//
// Java reference: InventoryUpdate.java, AbstractItemPacket.writeItem()
type InventoryUpdate struct {
	Entries []InvUpdateEntry
}

// NewInventoryUpdate creates an InventoryUpdate packet with the given entries.
func NewInventoryUpdate(entries ...InvUpdateEntry) *InventoryUpdate {
	return &InventoryUpdate{Entries: entries}
}

// Write serializes InventoryUpdate packet to binary format.
//
// Packet structure:
//
//	opcode (byte) = 0x27
//	itemCount (short)
//	for each entry:
//	  changeType (short) — 1=add, 2=modify, 3=remove
//	  type1 (short) — item category
//	  objectID (int32)
//	  itemID (int32) — template ID
//	  count (int32)
//	  type2 (short) — sub-type
//	  customType1 (short) — 0
//	  equipped (short) — 0 or 1
//	  bodyPart (int32) — slot mask
//	  enchant (short)
//	  customType2 (short) — 0
//	  augmentation (int32) — 0
//	  mana (int32) — -1
//
// Java reference: InventoryUpdate.writeImpl()
func (p *InventoryUpdate) Write() ([]byte, error) {
	// 3 header bytes + 38 bytes per entry (2 for changeType + 36 for item)
	w := packet.NewWriter(3 + len(p.Entries)*38)

	w.WriteByte(OpcodeInventoryUpdate)
	w.WriteShort(int16(len(p.Entries)))

	for _, e := range p.Entries {
		w.WriteShort(e.ChangeType)
		writeItem(w, e.Item)
	}

	return w.Bytes(), nil
}
