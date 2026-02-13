package serverpackets

import (
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeHennaItemDrawInfo is the opcode for HennaItemDrawInfo packet (S2C 0xE3).
// Detailed info about a henna before equipping.
// Java reference: HennaItemInfo.java
const OpcodeHennaItemDrawInfo = 0xE3

// HennaItemDrawInfo packet (S2C 0xE4) — info about henna before equip.
//
// Java reference: HennaItemDrawInfo.java
type HennaItemDrawInfo struct {
	Player *model.Player
	Henna  *data.HennaInfo
}

// NewHennaItemDrawInfo creates a HennaItemDrawInfo packet.
func NewHennaItemDrawInfo(player *model.Player, henna *data.HennaInfo) HennaItemDrawInfo {
	return HennaItemDrawInfo{Player: player, Henna: henna}
}

// Write serializes HennaItemDrawInfo to binary format.
func (p *HennaItemDrawInfo) Write() ([]byte, error) {
	w := packet.NewWriter(64)
	w.WriteByte(OpcodeHennaItemDrawInfo)

	w.WriteInt(p.Henna.DyeID)
	w.WriteInt(p.Henna.DyeItemID)
	w.WriteInt(p.Henna.WearCount)
	w.WriteInt(int32(p.Henna.WearFee))
	if p.Henna.IsAllowedClass(p.Player.ClassID()) {
		w.WriteInt(1)
	} else {
		w.WriteInt(0)
	}
	w.WriteInt(int32(p.Player.Inventory().GetAdena()))

	// Current stats + expected after equip (int32 for current, byte for projected)
	w.WriteInt(int32(p.Player.GetINT()))
	w.WriteByte(byte(int32(p.Player.GetINT()) + p.Henna.StatINT))
	w.WriteInt(int32(p.Player.GetSTR()))
	w.WriteByte(byte(int32(p.Player.GetSTR()) + p.Henna.StatSTR))
	w.WriteInt(int32(p.Player.GetCON()))
	w.WriteByte(byte(int32(p.Player.GetCON()) + p.Henna.StatCON))
	w.WriteInt(int32(p.Player.GetMEN()))
	w.WriteByte(byte(int32(p.Player.GetMEN()) + p.Henna.StatMEN))
	w.WriteInt(int32(p.Player.GetDEX()))
	w.WriteByte(byte(int32(p.Player.GetDEX()) + p.Henna.StatDEX))
	w.WriteInt(int32(p.Player.GetWIT()))
	w.WriteByte(byte(int32(p.Player.GetWIT()) + p.Henna.StatWIT))

	return w.Bytes(), nil
}
