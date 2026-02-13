package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// S2C extended sub-opcodes for augmentation (0xFE:0x50-0x57).
// Java reference: ServerPackets.java EX_SHOW_VARIATION_* / EX_PUT_* / EX_VARIATION_*.
const (
	SubOpcodeExShowVariationMakeWindow   int16 = 0x50 // Open augmentation UI
	SubOpcodeExShowVariationCancelWindow int16 = 0x51 // Open cancel augmentation UI
	SubOpcodeExPutItemResultForVariation int16 = 0x52 // Target weapon confirmed
	SubOpcodeExPutIntensiveResult        int16 = 0x53 // Life Stone + Gemstone confirmed
	SubOpcodeExPutCommissionResult       int16 = 0x54 // Gemstone commission confirmed
	SubOpcodeExVariationResult           int16 = 0x55 // Augmentation result
	SubOpcodeExPutItemForVariationCancel int16 = 0x56 // Cancel window item confirmed
	SubOpcodeExVariationCancelResult     int16 = 0x57 // Cancel result
)

// ExShowVariationMakeWindow opens the augmentation UI (S2C 0xFE:0x50).
// No data fields — just opcode + sub-opcode.
// Java reference: ExShowVariationMakeWindow.java (STATIC_PACKET).
type ExShowVariationMakeWindow struct{}

// Write serializes ExShowVariationMakeWindow.
func (p ExShowVariationMakeWindow) Write() ([]byte, error) {
	w := packet.NewWriter(3)
	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExShowVariationMakeWindow)
	return w.Bytes(), nil
}

// ExShowVariationCancelWindow opens the augmentation cancel UI (S2C 0xFE:0x51).
// No data fields — just opcode + sub-opcode.
// Java reference: ExShowVariationCancelWindow.java (STATIC_PACKET).
type ExShowVariationCancelWindow struct{}

// Write serializes ExShowVariationCancelWindow.
func (p ExShowVariationCancelWindow) Write() ([]byte, error) {
	w := packet.NewWriter(3)
	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExShowVariationCancelWindow)
	return w.Bytes(), nil
}

// ExPutItemResultForVariation shows target weapon info (S2C 0xFE:0x52).
// Java reference: ExPutItemResultForVariationMake.java.
type ExPutItemResultForVariation struct {
	ItemObjectID int32
	ItemID       int32
}

// Write serializes ExPutItemResultForVariation.
func (p ExPutItemResultForVariation) Write() ([]byte, error) {
	w := packet.NewWriter(15)
	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExPutItemResultForVariation)
	w.WriteInt(p.ItemObjectID)
	w.WriteInt(p.ItemID)
	w.WriteInt(1) // hardcoded success flag (Java)
	return w.Bytes(), nil
}

// ExPutIntensiveResult confirms Life Stone + Gemstone input (S2C 0xFE:0x53).
// Java reference: ExPutIntensiveResultForVariationMake.java.
type ExPutIntensiveResult struct {
	RefinerObjectID int32 // Life Stone ObjectID
	LifeStoneItemID int32 // Life Stone template ID
	GemstoneItemID  int32 // Gemstone template ID
	GemstoneCount   int32 // Required gemstone count
}

// Write serializes ExPutIntensiveResult.
func (p ExPutIntensiveResult) Write() ([]byte, error) {
	w := packet.NewWriter(23)
	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExPutIntensiveResult)
	w.WriteInt(p.RefinerObjectID)
	w.WriteInt(p.LifeStoneItemID)
	w.WriteInt(p.GemstoneItemID)
	w.WriteInt(p.GemstoneCount)
	w.WriteInt(1) // hardcoded success flag (Java)
	return w.Bytes(), nil
}

// ExPutCommissionResult confirms Gemstone commission (S2C 0xFE:0x54).
// Java reference: ExPutCommissionResultForVariationMake.java.
type ExPutCommissionResult struct {
	GemstoneObjectID int32 // Gemstone ObjectID
	GemstoneCount    int32 // Gemstone count
}

// Write serializes ExPutCommissionResult.
func (p ExPutCommissionResult) Write() ([]byte, error) {
	w := packet.NewWriter(23)
	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExPutCommissionResult)
	w.WriteInt(p.GemstoneObjectID)
	w.WriteInt(0) // unknown1 (Java hardcoded)
	w.WriteInt(p.GemstoneCount)
	w.WriteInt(0) // unknown2 (Java hardcoded)
	w.WriteInt(1) // success flag (Java hardcoded)
	return w.Bytes(), nil
}

// ExVariationResult sends augmentation result (S2C 0xFE:0x55).
// Java reference: ExVariationResult.java.
type ExVariationResult struct {
	Stat12 int32 // First augmentation stat pair (bits 0-15: stat1, 16-31: stat2)
	Stat34 int32 // Second augmentation stat pair (bits 0-15: stat3, 16-31: stat4)
	Result int32 // 1 = success, 0 = failure
}

// Write serializes ExVariationResult.
func (p ExVariationResult) Write() ([]byte, error) {
	w := packet.NewWriter(15)
	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExVariationResult)
	w.WriteInt(p.Stat12)
	w.WriteInt(p.Stat34)
	w.WriteInt(p.Result)
	return w.Bytes(), nil
}

// ExPutItemForVariationCancel shows item info in cancel window (S2C 0xFE:0x56).
// Java reference: ExPutItemResultForVariationCancel.java.
type ExPutItemForVariationCancel struct {
	ItemObjectID int32
	Price        int64 // Adena cost for cancellation
}

// Write serializes ExPutItemForVariationCancel.
func (p ExPutItemForVariationCancel) Write() ([]byte, error) {
	w := packet.NewWriter(27)
	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExPutItemForVariationCancel)
	w.WriteInt(0x40A97712) // magic constant (Java hardcoded)
	w.WriteInt(p.ItemObjectID)
	w.WriteInt(0x27)   // magic constant (Java hardcoded)
	w.WriteInt(0x2006) // magic constant (Java hardcoded)
	w.WriteInt(int32(p.Price & 0xFFFFFFFF))
	w.WriteInt(int32(p.Price >> 32))
	w.WriteInt(1) // success flag (Java hardcoded)
	return w.Bytes(), nil
}

// ExVariationCancelResult sends augmentation removal result (S2C 0xFE:0x57).
// Java reference: ExVariationCancelResult.java.
type ExVariationCancelResult struct {
	Result int32 // 1 = success, 0 = failure
}

// Write serializes ExVariationCancelResult.
func (p ExVariationCancelResult) Write() ([]byte, error) {
	w := packet.NewWriter(11)
	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExVariationCancelResult)
	w.WriteInt(1)        // closeWindow = 1 (Java hardcoded)
	w.WriteInt(p.Result) // success/failure
	return w.Bytes(), nil
}
