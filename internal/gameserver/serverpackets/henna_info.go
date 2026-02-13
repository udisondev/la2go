package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeHennaInfo is the opcode for HennaInfo packet (S2C 0xE4).
// Sends current henna stat bonuses and equipped hennas to the client.
const OpcodeHennaInfo = 0xE4

// HennaInfo packet (S2C 0xE5) — current hennas on player.
// Sent when player equips/removes henna or enters world.
//
// Java reference: HennaInfo.java
type HennaInfo struct {
	Player *model.Player
}

// NewHennaInfo creates a HennaInfo packet.
func NewHennaInfo(player *model.Player) HennaInfo {
	return HennaInfo{Player: player}
}

// Write serializes HennaInfo to binary format.
func (p *HennaInfo) Write() ([]byte, error) {
	w := packet.NewWriter(64)
	w.WriteByte(OpcodeHennaInfo)

	// Stat bonuses (byte each)
	w.WriteByte(byte(p.Player.HennaStatINT()))
	w.WriteByte(byte(p.Player.HennaStatSTR()))
	w.WriteByte(byte(p.Player.HennaStatCON()))
	w.WriteByte(byte(p.Player.HennaStatMEN()))
	w.WriteByte(byte(p.Player.HennaStatDEX()))
	w.WriteByte(byte(p.Player.HennaStatWIT()))

	// Max slots
	w.WriteInt(model.MaxHennaSlots)

	// Count equipped hennas
	hennaList := p.Player.GetHennaList()
	count := int32(0)
	for _, h := range hennaList {
		if h != nil {
			count++
		}
	}
	w.WriteInt(count)

	// Equipped hennas
	for _, h := range hennaList {
		if h != nil {
			w.WriteInt(h.DyeID)
			w.WriteInt(1) // unknown (always 1 in Java)
		}
	}

	return w.Bytes(), nil
}
