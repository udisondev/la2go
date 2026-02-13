package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeTradeOwnAdd is the opcode for TradeOwnAdd packet (S2C 0x20).
// Notifies player that their own item was added to trade window.
//
// Java reference: TradeOwnAdd.java
const OpcodeTradeOwnAdd = 0x20

// OpcodeTradeOtherAdd is the opcode for TradeOtherAdd packet (S2C 0x21).
// Notifies player that the trade partner added an item.
//
// Java reference: TradeOtherAdd.java
const OpcodeTradeOtherAdd = 0x21

// TradeItemAdd represents a trade item add notification (S2C 0x20 or 0x21).
//
// Packet structure (identical for both OwnAdd and OtherAdd):
//   - opcode (byte) — 0x20 (own) or 0x21 (other)
//   - count (short) — always 1
//   - type1 (short)
//   - objectID (int32)
//   - itemID (int32)
//   - itemCount (int32)
//   - type2 (short)
//   - customType1 (short) — 0
//   - bodyPart (int32)
//   - enchant (short)
//   - unknown (short) — 0
//   - customType2 (short) — 0
type TradeItemAdd struct {
	Opcode     byte
	Type1      int16
	ObjectID   int32
	ItemID     int32
	Count      int32
	Type2      int16
	BodyPart   int32
	Enchant    int16
}

// NewTradeOwnAdd creates TradeOwnAdd packet.
func NewTradeOwnAdd(type1 int16, objectID int32, itemID int32, count int32,
	type2 int16, bodyPart int32, enchant int16) *TradeItemAdd {
	return &TradeItemAdd{
		Opcode:   OpcodeTradeOwnAdd,
		Type1:    type1,
		ObjectID: objectID,
		ItemID:   itemID,
		Count:    count,
		Type2:    type2,
		BodyPart: bodyPart,
		Enchant:  enchant,
	}
}

// NewTradeOtherAdd creates TradeOtherAdd packet.
func NewTradeOtherAdd(type1 int16, objectID int32, itemID int32, count int32,
	type2 int16, bodyPart int32, enchant int16) *TradeItemAdd {
	return &TradeItemAdd{
		Opcode:   OpcodeTradeOtherAdd,
		Type1:    type1,
		ObjectID: objectID,
		ItemID:   itemID,
		Count:    count,
		Type2:    type2,
		BodyPart: bodyPart,
		Enchant:  enchant,
	}
}

// Write serializes TradeItemAdd packet to bytes.
func (p *TradeItemAdd) Write() ([]byte, error) {
	w := packet.NewWriter(31) // 1 + 2 + 28

	w.WriteByte(p.Opcode)
	w.WriteShort(1)          // always 1 item added
	w.WriteShort(p.Type1)    // type1
	w.WriteInt(p.ObjectID)   // objectID
	w.WriteInt(p.ItemID)     // itemID
	w.WriteInt(p.Count)      // count
	w.WriteShort(p.Type2)    // type2
	w.WriteShort(0)          // customType1
	w.WriteInt(p.BodyPart)   // bodyPart
	w.WriteShort(p.Enchant)  // enchant
	w.WriteShort(0)          // unknown
	w.WriteShort(0)          // customType2

	return w.Bytes(), nil
}
