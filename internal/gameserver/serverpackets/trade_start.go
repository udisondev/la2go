package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeTradeStart is the opcode for TradeStart packet (S2C 0x1E).
// Opens the trade window and shows tradeable items from the player's inventory.
//
// Java reference: TradeStart.java
const OpcodeTradeStart = 0x1E

// TradeStart packet (S2C 0x1E) opens the trade window.
//
// Packet structure:
//   - opcode (byte) — 0x1E
//   - partnerObjectID (int32) — ObjectID of trade partner
//   - itemCount (short) — number of tradeable items
//   - for each item (30 bytes):
//   - type1 (short)
//   - objectID (int32)
//   - itemID (int32)
//   - count (int32)
//   - type2 (short)
//   - unknown1 (short) — 0
//   - bodyPart (int32)
//   - enchant (short)
//   - unknown2 (short) — 0
//   - customType2 (short) — 0
type TradeStart struct {
	PartnerObjectID int32
	Items           []*model.Item
}

// NewTradeStart creates TradeStart packet.
func NewTradeStart(partnerObjectID int32, items []*model.Item) *TradeStart {
	return &TradeStart{
		PartnerObjectID: partnerObjectID,
		Items:           items,
	}
}

// Write serializes TradeStart packet to bytes.
func (p *TradeStart) Write() ([]byte, error) {
	w := packet.NewWriter(7 + len(p.Items)*30)

	w.WriteByte(OpcodeTradeStart)
	w.WriteInt(p.PartnerObjectID)
	w.WriteShort(int16(len(p.Items)))

	for _, item := range p.Items {
		tmpl := item.Template()
		var type1, type2 int16
		var bodyPart int32
		if tmpl != nil {
			type1 = tmpl.Type1
			type2 = tmpl.Type2
			bodyPart = tmpl.BodyPartMask
		}

		w.WriteShort(type1)                    // type1
		w.WriteInt(int32(item.ObjectID()))     // objectID
		w.WriteInt(item.ItemID())              // itemID
		w.WriteInt(item.Count())               // count
		w.WriteShort(type2)                    // type2
		w.WriteShort(0)                        // unknown1
		w.WriteInt(bodyPart)                   // bodyPart
		w.WriteShort(int16(item.Enchant()))    // enchant
		w.WriteShort(0)                        // unknown2
		w.WriteShort(0)                        // customType2
	}

	return w.Bytes(), nil
}
