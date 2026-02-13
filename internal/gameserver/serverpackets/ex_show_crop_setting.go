package serverpackets

import (
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/game/manor"
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeExShowCropSetting is the sub-opcode for ExShowCropSetting (S2C 0xFE:0x20).
const SubOpcodeExShowCropSetting int16 = 0x20

// ExShowCropSetting sends crop procurement settings for a manor castle.
// Used by clan leaders to configure crop procurement for the next period.
//
// Packet structure:
//   - opcode (byte) -- 0xFE
//   - subOpcode (short) -- 0x20
//   - manorID (int32)
//   - count (int32) -- number of seed templates (crops keyed by seed)
//   - per seed template:
//   - cropID (int32)
//   - level (int32)
//   - reward1Flag (byte=1), reward1 (int32)
//   - reward2Flag (byte=1), reward2 (int32)
//   - limitCrops (int32)
//   - unknown (int32) -- always 0
//   - minPrice (int32)
//   - maxPrice (int32)
//   - currentStartAmount (int32)
//   - currentPrice (int32)
//   - currentRewardType (byte)
//   - nextStartAmount (int32)
//   - nextPrice (int32)
//   - nextRewardType (byte)
type ExShowCropSetting struct {
	ManorID       int32
	Seeds         []*data.SeedTemplate
	CurrentPeriod []*manor.CropProcure
	NextPeriod    []*manor.CropProcure
}

// Write serializes ExShowCropSetting to binary.
func (p *ExShowCropSetting) Write() ([]byte, error) {
	// 1 + 2 + 4 + 4 + len*(4+4+1+4+1+4+4+4+4+4+4+4+1+4+4+1) = 11 + 52*N
	w := packet.NewWriter(11 + len(p.Seeds)*52)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExShowCropSetting)
	w.WriteInt(p.ManorID)
	w.WriteInt(int32(len(p.Seeds)))

	for _, tmpl := range p.Seeds {
		w.WriteInt(tmpl.CropID)
		w.WriteInt(tmpl.Level)
		w.WriteByte(1)
		w.WriteInt(tmpl.Reward1)
		w.WriteByte(1)
		w.WriteInt(tmpl.Reward2)
		w.WriteInt(tmpl.LimitCrops)
		w.WriteInt(0) // unknown, always 0
		w.WriteInt(int32(data.CropMinPrice(tmpl.CropID)))
		w.WriteInt(int32(data.CropMaxPrice(tmpl.CropID)))

		// Current period
		curCP := findCropProcure(p.CurrentPeriod, tmpl.CropID)
		if curCP != nil {
			w.WriteInt(curCP.StartAmount())
			w.WriteInt(int32(curCP.Price()))
			w.WriteByte(byte(curCP.RewardType()))
		} else {
			w.WriteInt(0)
			w.WriteInt(0)
			w.WriteByte(0)
		}

		// Next period
		nextCP := findCropProcure(p.NextPeriod, tmpl.CropID)
		if nextCP != nil {
			w.WriteInt(nextCP.StartAmount())
			w.WriteInt(int32(nextCP.Price()))
			w.WriteByte(byte(nextCP.RewardType()))
		} else {
			w.WriteInt(0)
			w.WriteInt(0)
			w.WriteByte(0)
		}
	}

	return w.Bytes(), nil
}

func findCropProcure(list []*manor.CropProcure, cropID int32) *manor.CropProcure {
	for _, cp := range list {
		if cp.CropID() == cropID {
			return cp
		}
	}
	return nil
}
