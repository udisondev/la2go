package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeTradeOtherDone is the opcode for TradeOtherDone packet (S2C 0x7C).
// Notifies player that the trade partner has confirmed.
//
// Java reference: TradeOtherDone.java
const OpcodeTradeOtherDone = 0x7C

// TradeOtherDone packet (S2C 0x7C) signals that the trade partner confirmed.
//
// Packet structure:
//   - opcode (byte) — 0x7C
type TradeOtherDone struct{}

// NewTradeOtherDone creates TradeOtherDone packet.
func NewTradeOtherDone() *TradeOtherDone {
	return &TradeOtherDone{}
}

// Write serializes TradeOtherDone packet to bytes.
func (p *TradeOtherDone) Write() ([]byte, error) {
	w := packet.NewWriter(1)
	w.WriteByte(OpcodeTradeOtherDone)
	return w.Bytes(), nil
}
