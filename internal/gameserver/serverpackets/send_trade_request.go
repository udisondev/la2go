package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeSendTradeRequest is the opcode for SendTradeRequest packet (S2C 0x5E).
// Notifies target player of an incoming trade request.
//
// Java reference: SendTradeRequest.java
const OpcodeSendTradeRequest = 0x5E

// SendTradeRequest packet (S2C 0x5E) sends a trade request to the target.
//
// Packet structure:
//   - opcode (byte) — 0x5E
//   - senderObjectID (int32) — ObjectID of the player requesting trade
type SendTradeRequest struct {
	SenderObjectID int32
}

// NewSendTradeRequest creates SendTradeRequest packet.
func NewSendTradeRequest(senderObjectID int32) *SendTradeRequest {
	return &SendTradeRequest{SenderObjectID: senderObjectID}
}

// Write serializes SendTradeRequest packet to bytes.
func (p *SendTradeRequest) Write() ([]byte, error) {
	w := packet.NewWriter(5)
	w.WriteByte(OpcodeSendTradeRequest)
	w.WriteInt(p.SenderObjectID)
	return w.Bytes(), nil
}
