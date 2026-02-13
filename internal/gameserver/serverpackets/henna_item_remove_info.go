package serverpackets

import (
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeHennaItemRemoveInfo is the opcode for HennaItemRemoveInfo packet (S2C 0xE6).
// Detailed info about a henna before removing.
const OpcodeHennaItemRemoveInfo = 0xE6

// HennaItemRemoveInfo packet (S2C 0xE6) — info about henna before removal.
// Stats are SUBTRACTED (shows what happens after removal).
//
// Java reference: HennaItemRemoveInfo.java
type HennaItemRemoveInfo struct {
	Player *model.Player
	Henna  *data.HennaInfo
}

// NewHennaItemRemoveInfo creates a HennaItemRemoveInfo packet.
func NewHennaItemRemoveInfo(player *model.Player, henna *data.HennaInfo) HennaItemRemoveInfo {
	return HennaItemRemoveInfo{Player: player, Henna: henna}
}

// Write serializes HennaItemRemoveInfo to binary format.
func (p *HennaItemRemoveInfo) Write() ([]byte, error) {
	w := packet.NewWriter(64)
	w.WriteByte(OpcodeHennaItemRemoveInfo)

	w.WriteInt(p.Henna.DyeID)
	w.WriteInt(p.Henna.DyeItemID)
	w.WriteInt(p.Henna.CancelCount) // returned dye count
	w.WriteInt(int32(p.Henna.CancelFee))
	if p.Henna.IsAllowedClass(p.Player.ClassID()) {
		w.WriteInt(1)
	} else {
		w.WriteInt(0)
	}
	w.WriteInt(int32(p.Player.Inventory().GetAdena()))

	// Current stats + expected after removal (SUBTRACT)
	w.WriteInt(int32(p.Player.GetINT()))
	w.WriteByte(byte(int32(p.Player.GetINT()) - p.Henna.StatINT))
	w.WriteInt(int32(p.Player.GetSTR()))
	w.WriteByte(byte(int32(p.Player.GetSTR()) - p.Henna.StatSTR))
	w.WriteInt(int32(p.Player.GetCON()))
	w.WriteByte(byte(int32(p.Player.GetCON()) - p.Henna.StatCON))
	w.WriteInt(int32(p.Player.GetMEN()))
	w.WriteByte(byte(int32(p.Player.GetMEN()) - p.Henna.StatMEN))
	w.WriteInt(int32(p.Player.GetDEX()))
	w.WriteByte(byte(int32(p.Player.GetDEX()) - p.Henna.StatDEX))
	w.WriteInt(int32(p.Player.GetWIT()))
	w.WriteByte(byte(int32(p.Player.GetWIT()) - p.Henna.StatWIT))

	return w.Bytes(), nil
}
