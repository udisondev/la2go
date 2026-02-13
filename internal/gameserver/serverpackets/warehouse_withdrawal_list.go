package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeWareHouseWithdrawalList is the opcode for WareHouseWithdrawalList packet (S2C 0x42).
// Sends list of items that player can withdraw from warehouse.
//
// Phase 8: Trade System Foundation.
// Java reference: WareHouseWithdrawalList.java
const OpcodeWareHouseWithdrawalList = 0x42

// WareHouseWithdrawalList packet (S2C 0x42) shows warehouse withdraw dialog.
//
// Packet structure is identical to WareHouseDepositList (0x41)
// except the opcode and data source (warehouse items vs inventory items).
//
// Phase 8: Trade System Foundation.
type WareHouseWithdrawalList struct {
	WhType      int16
	PlayerAdena int64
	Items       []*model.Item
}

// NewWareHouseWithdrawalList creates WareHouseWithdrawalList packet.
func NewWareHouseWithdrawalList(whType int16, playerAdena int64, items []*model.Item) *WareHouseWithdrawalList {
	return &WareHouseWithdrawalList{
		WhType:      whType,
		PlayerAdena: playerAdena,
		Items:       items,
	}
}

// Write serializes WareHouseWithdrawalList packet to bytes.
func (p *WareHouseWithdrawalList) Write() ([]byte, error) {
	// 1 opcode + 2 whType + 4 adena + 2 count + 38 bytes per item
	w := packet.NewWriter(9 + len(p.Items)*38)

	w.WriteByte(OpcodeWareHouseWithdrawalList)
	w.WriteShort(p.WhType)
	w.WriteInt(int32(p.PlayerAdena))
	w.WriteShort(int16(len(p.Items)))

	for _, item := range p.Items {
		writeWarehouseItem(w, item)
	}

	return w.Bytes(), nil
}
