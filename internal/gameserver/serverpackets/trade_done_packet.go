package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeTradeDone is the opcode for TradeDone packet (S2C 0x22).
// Notifies player that the trade has finished.
//
// Java reference: TradeDone.java (server packet)
const OpcodeTradeDone = 0x22

// TradeDonePacket packet (S2C 0x22) signals trade completion or cancellation.
//
// Packet structure:
//   - opcode (byte) — 0x22
//   - result (int32) — 0=cancelled, 1=success
type TradeDonePacket struct {
	Result int32
}

// NewTradeDonePacket creates TradeDone S2C packet.
func NewTradeDonePacket(result int32) *TradeDonePacket {
	return &TradeDonePacket{Result: result}
}

// Write serializes TradeDone packet to bytes.
func (p *TradeDonePacket) Write() ([]byte, error) {
	w := packet.NewWriter(5)
	w.WriteByte(OpcodeTradeDone)
	w.WriteInt(p.Result)
	return w.Bytes(), nil
}
