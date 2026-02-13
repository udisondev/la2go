package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeEnchantResult is the opcode for EnchantResult packet (S2C 0x81).
// Notifies player of the enchantment result.
//
// Java reference: EnchantResult.java
const OpcodeEnchantResult = 0x81

// EnchantResult packet (S2C 0x81) sends enchantment result.
//
// Packet structure:
//   - opcode (byte) — 0x81
//   - result (int32) — 0=failed, >0=success (new enchant level)
type EnchantResult struct {
	Result int32
}

// NewEnchantResult creates EnchantResult packet.
func NewEnchantResult(result int32) *EnchantResult {
	return &EnchantResult{Result: result}
}

// Write serializes EnchantResult packet to bytes.
func (p *EnchantResult) Write() ([]byte, error) {
	w := packet.NewWriter(5)
	w.WriteByte(OpcodeEnchantResult)
	w.WriteInt(p.Result)
	return w.Bytes(), nil
}
