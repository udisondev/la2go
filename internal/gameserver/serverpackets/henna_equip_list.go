package serverpackets

import (
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeHennaEquipList is the opcode for HennaEquipList packet (S2C 0xE2).
// Lists available hennas for the player's class.
const OpcodeHennaEquipList = 0xE2

// HennaEquipList packet (S2C 0xE7) — list of hennas available for equip.
// Sent when player opens the henna equip dialog at a Symbol Maker NPC.
//
// Java reference: HennaEquipList.java
type HennaEquipList struct {
	Player *model.Player
}

// NewHennaEquipList creates a HennaEquipList packet.
func NewHennaEquipList(player *model.Player) HennaEquipList {
	return HennaEquipList{Player: player}
}

// Write serializes HennaEquipList to binary format.
func (p *HennaEquipList) Write() ([]byte, error) {
	classID := p.Player.ClassID()
	hennas := data.GetHennaInfoListForClass(classID)

	// Filter: only show hennas where player has at least 1 dye item in inventory
	available := make([]*data.HennaInfo, 0, len(hennas))
	for _, h := range hennas {
		if p.Player.Inventory().FindItemByItemID(h.DyeItemID) != nil {
			available = append(available, h)
		}
	}

	w := packet.NewWriter(64 + len(available)*20)
	w.WriteByte(OpcodeHennaEquipList)

	w.WriteInt(int32(p.Player.Inventory().GetAdena())) // Current adena
	w.WriteInt(int32(p.Player.GetHennaEmptySlots()))   // Available slots
	w.WriteInt(int32(len(available)))                   // Henna count

	for _, h := range available {
		w.WriteInt(h.DyeID)
		w.WriteInt(h.DyeItemID)
		w.WriteInt(h.WearCount)
		w.WriteInt(int32(h.WearFee))
		if h.IsAllowedClass(classID) {
			w.WriteInt(1)
		} else {
			w.WriteInt(0)
		}
	}

	return w.Bytes(), nil
}
