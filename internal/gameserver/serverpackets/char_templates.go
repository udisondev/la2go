package serverpackets

import (
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeCharTemplates is the opcode for CharTemplates (S2C 0x17).
// Java reference: ServerPackets.CHAR_TEMPLATES(0x17).
const OpcodeCharTemplates = 0x17

// baseClassIDs — the 9 base classes available for character creation.
var baseClassIDs = []uint8{0, 10, 18, 25, 31, 38, 44, 49, 53}

// CharTemplates sends available character creation templates to the client.
//
// Packet structure:
//   - opcode (byte) — 0x17
//   - count (int32) — number of templates
//   - for each template:
//   - race (int32)
//   - classID (int32)
//   - 6× stat block: [0x46 (int32), value (int32), 0x0A (int32)]
type CharTemplates struct{}

// Write serializes CharTemplates packet.
func (p *CharTemplates) Write() ([]byte, error) {
	// 1 opcode + 4 count + 9 templates * (8 + 6*12) = 1 + 4 + 9*80 = 725 bytes
	w := packet.NewWriter(728)
	w.WriteByte(OpcodeCharTemplates)

	// Collect available templates
	type tmplEntry struct {
		race    int32
		classID int32
		str     int32
		dex     int32
		con     int32
		int_    int32
		wit     int32
		men     int32
	}

	var entries []tmplEntry
	for _, cid := range baseClassIDs {
		tmpl := data.GetTemplate(cid)
		if tmpl == nil {
			continue
		}
		info := data.GetClassInfo(int32(cid))
		race := int32(0)
		if info != nil {
			race = info.Race
		}
		entries = append(entries, tmplEntry{
			race:    race,
			classID: int32(cid),
			str:     int32(tmpl.BaseSTR),
			dex:     int32(tmpl.BaseDEX),
			con:     int32(tmpl.BaseCON),
			int_:    int32(tmpl.BaseINT),
			wit:     int32(tmpl.BaseWIT),
			men:     int32(tmpl.BaseMEN),
		})
	}

	w.WriteInt(int32(len(entries)))
	for _, e := range entries {
		w.WriteInt(e.race)
		w.WriteInt(e.classID)
		// STR
		w.WriteInt(0x46)
		w.WriteInt(e.str)
		w.WriteInt(0x0A)
		// DEX
		w.WriteInt(0x46)
		w.WriteInt(e.dex)
		w.WriteInt(0x0A)
		// CON
		w.WriteInt(0x46)
		w.WriteInt(e.con)
		w.WriteInt(0x0A)
		// INT
		w.WriteInt(0x46)
		w.WriteInt(e.int_)
		w.WriteInt(0x0A)
		// WIT
		w.WriteInt(0x46)
		w.WriteInt(e.wit)
		w.WriteInt(0x0A)
		// MEN
		w.WriteInt(0x46)
		w.WriteInt(e.men)
		w.WriteInt(0x0A)
	}

	return w.Bytes(), nil
}
