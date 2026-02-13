package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// Sub-opcodes for cursed weapon extended packets (0xFE + sub).
const (
	SubOpcodeExCursedWeaponList     int16 = 0x45
	SubOpcodeExCursedWeaponLocation int16 = 0x46
)

// ExCursedWeaponList (S2C 0xFE:0x45) sends list of all cursed weapon item IDs.
// Java: ExCursedWeaponList.java
//
// Format: (ch)d[d]
//
//	opcode (byte) = 0xFE
//	subOpcode (short) = 0x45
//	count (int32) — number of weapons
//	for each: itemID (int32)
type ExCursedWeaponList struct {
	WeaponIDs []int32
}

// Write serializes ExCursedWeaponList to binary.
func (p *ExCursedWeaponList) Write() ([]byte, error) {
	w := packet.NewWriter(3 + 4 + len(p.WeaponIDs)*4)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExCursedWeaponList)
	w.WriteInt(int32(len(p.WeaponIDs)))
	for _, id := range p.WeaponIDs {
		w.WriteInt(id)
	}

	return w.Bytes(), nil
}

// CursedWeaponLocationInfo holds position data for one cursed weapon.
type CursedWeaponLocationInfo struct {
	ItemID    int32
	Activated int32 // 0=dropped on ground, 1=equipped by player
	X, Y, Z  int32
}

// ExCursedWeaponLocation (S2C 0xFE:0x46) sends positions of active cursed weapons.
// Java: ExCursedWeaponLocation.java
//
// Format: (ch)d[ddddd]
//
//	opcode (byte) = 0xFE
//	subOpcode (short) = 0x46
//	count (int32)
//	for each:
//	  itemID (int32)
//	  activated (int32) — 0=dropped, 1=equipped
//	  x (int32), y (int32), z (int32)
type ExCursedWeaponLocation struct {
	Weapons []CursedWeaponLocationInfo
}

// Write serializes ExCursedWeaponLocation to binary.
func (p *ExCursedWeaponLocation) Write() ([]byte, error) {
	w := packet.NewWriter(3 + 4 + len(p.Weapons)*20)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExCursedWeaponLocation)

	if len(p.Weapons) == 0 {
		w.WriteInt(0)
		w.WriteInt(0)
	} else {
		w.WriteInt(int32(len(p.Weapons)))
		for _, info := range p.Weapons {
			w.WriteInt(info.ItemID)
			w.WriteInt(info.Activated)
			w.WriteInt(info.X)
			w.WriteInt(info.Y)
			w.WriteInt(info.Z)
		}
	}

	return w.Bytes(), nil
}
