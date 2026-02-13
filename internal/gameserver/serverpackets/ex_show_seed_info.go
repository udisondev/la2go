package serverpackets

import (
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/game/manor"
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeExShowSeedInfo is the sub-opcode for ExShowSeedInfo (S2C 0xFE:0x1C).
const SubOpcodeExShowSeedInfo int16 = 0x1C

// ExShowSeedInfo sends the list of seeds available for purchase at a manor.
//
// Packet structure:
//   - opcode (byte) -- 0xFE
//   - subOpcode (short) -- 0x1C
//   - hideButtons (byte) -- 0 or 1
//   - manorID (int32) -- castle ID
//   - unknown (int32) -- always 0
//   - count (int32) -- number of seeds
//   - per seed:
//   - seedID (int32)
//   - amount (int32)
//   - startAmount (int32)
//   - price (int32)
//   - level (int32)
//   - reward1Flag (byte) -- always 1
//   - reward1 (int32)
//   - reward2Flag (byte) -- always 1
//   - reward2 (int32)
type ExShowSeedInfo struct {
	ManorID     int32
	HideButtons bool
	Seeds       []*manor.SeedProduction
}

// Write serializes ExShowSeedInfo to binary.
func (p *ExShowSeedInfo) Write() ([]byte, error) {
	// 1 + 2 + 1 + 4 + 4 + 4 + len(Seeds)*(4+4+4+4+4+1+4+1+4) = 12 + 30*N
	w := packet.NewWriter(16 + len(p.Seeds)*30)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExShowSeedInfo)

	if p.HideButtons {
		w.WriteByte(1)
	} else {
		w.WriteByte(0)
	}

	w.WriteInt(p.ManorID)
	w.WriteInt(0) // unknown
	w.WriteInt(int32(len(p.Seeds)))

	for _, sp := range p.Seeds {
		w.WriteInt(sp.SeedID())
		w.WriteInt(sp.Amount())
		w.WriteInt(sp.StartAmount())
		w.WriteInt(int32(sp.Price()))
		tmpl := data.GetSeedTemplate(sp.SeedID())
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
