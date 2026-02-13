package serverpackets

import (
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/game/manor"
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeExShowCropInfo is the sub-opcode for ExShowCropInfo (S2C 0xFE:0x1D).
const SubOpcodeExShowCropInfo int16 = 0x1D

// ExShowCropInfo sends the list of crops available for procurement at a manor.
//
// Packet structure:
//   - opcode (byte) -- 0xFE
//   - subOpcode (short) -- 0x1D
//   - hideButtons (byte) -- 0 or 1
//   - manorID (int32) -- castle ID
//   - unknown (int32) -- always 0
//   - count (int32) -- number of crops
//   - per crop:
//   - cropID (int32)
//   - amount (int32)
//   - startAmount (int32)
//   - price (int32)
//   - rewardType (byte) -- 1 or 2
//   - level (int32) -- from SeedTemplate
//   - reward1Flag (byte) -- always 1
//   - reward1 (int32)
//   - reward2Flag (byte) -- always 1
//   - reward2 (int32)
type ExShowCropInfo struct {
	ManorID     int32
	HideButtons bool
	Crops       []*manor.CropProcure
}

// Write serializes ExShowCropInfo to binary.
func (p *ExShowCropInfo) Write() ([]byte, error) {
	// 1 + 2 + 1 + 4 + 4 + 4 + len(Crops)*(4+4+4+4+1+4+1+4+1+4) = 16 + 31*N
	w := packet.NewWriter(16 + len(p.Crops)*31)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExShowCropInfo)

	if p.HideButtons {
		w.WriteByte(1)
	} else {
		w.WriteByte(0)
	}

	w.WriteInt(p.ManorID)
	w.WriteInt(0) // unknown
	w.WriteInt(int32(len(p.Crops)))

	for _, cp := range p.Crops {
		w.WriteInt(cp.CropID())
		w.WriteInt(cp.Amount())
		w.WriteInt(cp.StartAmount())
		w.WriteInt(int32(cp.Price()))
		w.WriteByte(byte(cp.RewardType()))

		tmpl := data.GetSeedByCropID(cp.CropID())
		if tmpl != nil {
			w.WriteInt(tmpl.Level)
			w.WriteByte(1)
			w.WriteInt(tmpl.Reward1)
			w.WriteByte(1)
			w.WriteInt(tmpl.Reward2)
		} else {
			w.WriteInt(0)
			w.WriteByte(1)
			w.WriteInt(0)
			w.WriteByte(1)
			w.WriteInt(0)
		}
	}

	return w.Bytes(), nil
}
