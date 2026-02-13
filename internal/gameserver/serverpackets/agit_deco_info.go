package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeAgitDecoInfo is the server packet for clan hall decoration info (S2C 0xF7).
// This is a REGULAR packet, NOT an extended 0xFE packet.
//
// Java reference: AgitDecoInfo.java, ServerPackets.AGIT_DECO_INFO(0xF7).
const OpcodeAgitDecoInfo = 0xF7

// AgitDecoInfo sends clan hall function/decoration levels.
//
// Packet structure (S2C 0xF7):
//   - opcode      byte   0xF7
//   - hallID      int32
//   - hpLevel     int32  HP restoration level (0 = none)
//   - mpLevel     int32  MP restoration level
//   - expLevel    int32  EXP restoration level
//   - spLevel     int32  SP restoration level (always 0 in Interlude)
//   - teleLevel   int32  Teleport level
//   - curtainLvl  int32  Curtain decoration level
//   - frontLvl    int32  Front platform decoration level
//   - itemLevel   int32  Item creation level
//   - supportLvl  int32  Support buff level
type AgitDecoInfo struct {
	HallID       int32
	HPLevel      int32
	MPLevel      int32
	ExpLevel     int32
	SPLevel      int32 // Always 0 in Interlude
	TeleLevel    int32
	CurtainLevel int32
	FrontLevel   int32
	ItemLevel    int32
	SupportLevel int32
}

// Write serializes the AgitDecoInfo packet.
func (p *AgitDecoInfo) Write() ([]byte, error) {
	w := packet.NewWriter(64)

	w.WriteByte(OpcodeAgitDecoInfo)
	w.WriteInt(p.HallID)

	// Function levels (each encoded as: level > 0 ? 1 : 0, then level).
	writeFuncLevel(w, p.HPLevel)
	writeFuncLevel(w, p.MPLevel)
	writeFuncLevel(w, p.ExpLevel)
	writeFuncLevel(w, p.SPLevel)
	writeFuncLevel(w, p.TeleLevel)
	writeFuncLevel(w, p.CurtainLevel)
	writeFuncLevel(w, p.FrontLevel)
	writeFuncLevel(w, p.ItemLevel)
	writeFuncLevel(w, p.SupportLevel)

	return w.Bytes(), nil
}

// writeFuncLevel writes a function active flag + level.
func writeFuncLevel(w *packet.Writer, level int32) {
	if level > 0 {
		w.WriteInt(1)
		w.WriteInt(level)
	} else {
		w.WriteInt(0)
		w.WriteInt(0)
	}
}
