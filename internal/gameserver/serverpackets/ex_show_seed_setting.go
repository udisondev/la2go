package serverpackets

import (
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/game/manor"
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeExShowSeedSetting is the sub-opcode for ExShowSeedSetting (S2C 0xFE:0x1F).
const SubOpcodeExShowSeedSetting int16 = 0x1F

// ExShowSeedSetting sends seed production settings for a manor castle.
// Used by clan leaders to configure seed production for the next period.
//
// Packet structure:
//   - opcode (byte) -- 0xFE
//   - subOpcode (short) -- 0x1F
//   - manorID (int32)
//   - count (int32) -- number of seed templates
//   - per seed template:
//   - seedID (int32)
//   - level (int32)
//   - reward1Flag (byte=1), reward1 (int32)
//   - reward2Flag (byte=1), reward2 (int32)
//   - limitSeeds (int32)
//   - refPrice (int32)
//   - minPrice (int32)
//   - maxPrice (int32)
//   - currentStartAmount (int32)
//   - currentPrice (int32)
//   - nextStartAmount (int32)
//   - nextPrice (int32)
type ExShowSeedSetting struct {
	ManorID       int32
	Seeds         []*data.SeedTemplate
	CurrentPeriod []*manor.SeedProduction
	NextPeriod    []*manor.SeedProduction
}

// Write serializes ExShowSeedSetting to binary.
func (p *ExShowSeedSetting) Write() ([]byte, error) {
	// 1 + 2 + 4 + 4 + len*(4+4+1+4+1+4+4+4+4+4+4+4+4+4) = 11 + 50*N
	w := packet.NewWriter(11 + len(p.Seeds)*50)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExShowSeedSetting)
	w.WriteInt(p.ManorID)
	w.WriteInt(int32(len(p.Seeds)))

	for _, tmpl := range p.Seeds {
		w.WriteInt(tmpl.SeedID)
		w.WriteInt(tmpl.Level)
		w.WriteByte(1)
		w.WriteInt(tmpl.Reward1)
		w.WriteByte(1)
		w.WriteInt(tmpl.Reward2)
		w.WriteInt(tmpl.LimitSeeds)
		w.WriteInt(int32(data.SeedReferencePrice(tmpl.SeedID)))
		w.WriteInt(int32(data.SeedMinPrice(tmpl.SeedID)))
		w.WriteInt(int32(data.SeedMaxPrice(tmpl.SeedID)))

		// Current period
		curSP := findSeedProduction(p.CurrentPeriod, tmpl.SeedID)
		if curSP != nil {
			w.WriteInt(curSP.StartAmount())
			w.WriteInt(int32(curSP.Price()))
		} else {
			w.WriteInt(0)
			w.WriteInt(0)
		}

		// Next period
		nextSP := findSeedProduction(p.NextPeriod, tmpl.SeedID)
		if nextSP != nil {
			w.WriteInt(nextSP.StartAmount())
			w.WriteInt(int32(nextSP.Price()))
		} else {
			w.WriteInt(0)
			w.WriteInt(0)
		}
	}

	return w.Bytes(), nil
}

func findSeedProduction(list []*manor.SeedProduction, seedID int32) *manor.SeedProduction {
	for _, sp := range list {
		if sp.SeedID() == seedID {
			return sp
		}
	}
	return nil
}
