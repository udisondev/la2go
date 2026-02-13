package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeWareHouseDepositList is the opcode for WareHouseDepositList packet (S2C 0x41).
// Sends list of items that player can deposit into warehouse.
//
// Phase 8: Trade System Foundation.
// Java reference: WareHouseDepositList.java
const OpcodeWareHouseDepositList = 0x41

// Warehouse types matching L2J Mobius constants.
const (
	WarehouseTypePrivate  int16 = 1
	WarehouseTypeClan     int16 = 2
	WarehouseTypeCastle   int16 = 3
	WarehouseTypeFreight  int16 = 4
)

// WareHouseDepositList packet (S2C 0x41) shows warehouse deposit dialog.
//
// Packet structure:
//   - opcode (byte) — 0x41
//   - whType (short) — warehouse type (1=PRIVATE, 2=CLAN, 3=CASTLE, 4=FREIGHT)
//   - playerAdena (int32) — player's current Adena
//   - itemCount (short) — number of depositable items
//   - for each item:
//   - type1 (short)
//   - objectId (int32)
//   - itemId (int32) — template ID
//   - count (int32)
//   - type2 (short)
//   - customType1 (short) — 0
//   - bodyPart (int32)
//   - enchantLevel (short)
//   - unknown1 (short) — 0
//   - customType2 (short) — 0
//   - objectId2 (int32) — repeat objectId
//   - augmentation1 (int32) — 0
//   - augmentation2 (int32) — 0
//
// Phase 8: Trade System Foundation.
type WareHouseDepositList struct {
	WhType      int16
	PlayerAdena int64
	Items       []*model.Item
}

// NewWareHouseDepositList creates WareHouseDepositList packet.
func NewWareHouseDepositList(whType int16, playerAdena int64, items []*model.Item) *WareHouseDepositList {
	return &WareHouseDepositList{
		WhType:      whType,
		PlayerAdena: playerAdena,
		Items:       items,
	}
}

// Write serializes WareHouseDepositList packet to bytes.
func (p *WareHouseDepositList) Write() ([]byte, error) {
	// 1 opcode + 2 whType + 4 adena + 2 count + 38 bytes per item
	w := packet.NewWriter(9 + len(p.Items)*38)

	w.WriteByte(OpcodeWareHouseDepositList)
	w.WriteShort(p.WhType)
	w.WriteInt(int32(p.PlayerAdena))
	w.WriteShort(int16(len(p.Items)))

	for _, item := range p.Items {
		writeWarehouseItem(w, item)
	}

	return w.Bytes(), nil
}

// writeWarehouseItem writes a single item entry for warehouse packets.
// Shared between WareHouseDepositList and WareHouseWithdrawalList.
func writeWarehouseItem(w *packet.Writer, item *model.Item) {
	tmpl := item.Template()

	var type1, type2 int16
	var bodyPart int32
	if tmpl != nil {
		type1 = tmpl.Type1
		type2 = tmpl.Type2
		bodyPart = tmpl.BodyPartMask
	}

	w.WriteShort(type1)                    // type1
	w.WriteInt(int32(item.ObjectID()))     // objectId
	w.WriteInt(item.ItemID())              // itemId (template ID)
	w.WriteInt(item.Count())               // count
	w.WriteShort(type2)                    // type2
	w.WriteShort(0)                        // customType1
	w.WriteInt(bodyPart)                   // bodyPart
	w.WriteShort(int16(item.Enchant()))    // enchantLevel
	w.WriteShort(0)                        // unknown1
	w.WriteShort(0)                        // customType2
	w.WriteInt(int32(item.ObjectID()))     // objectId2 (repeat)
	w.WriteInt(0)                          // augmentation1
	w.WriteInt(0)                          // augmentation2
}

