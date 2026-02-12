package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

const (
	// OpcodeRestartResponse is the opcode for RestartResponse packet (S2C 0x5F)
	OpcodeRestartResponse = 0x5F
)

// RestartResponse is sent in response to RequestRestart packet (0x46).
// Confirms or denies player's request to return to character selection screen.
//
// Packet structure:
//   - opcode: byte (0x5F)
//   - result: int32 (1 = success/allowed, 0 = denied)
//
// Reference: L2J_Mobius RestartResponse.java
type RestartResponse struct {
	Result bool
}

// NewRestartResponse creates a RestartResponse packet.
// result=true: restart allowed, client returns to character selection
// result=false: restart denied (client remains in game)
func NewRestartResponse(result bool) RestartResponse {
	return RestartResponse{Result: result}
}

// Write serializes the RestartResponse packet to bytes.
func (p *RestartResponse) Write() ([]byte, error) {
	w := packet.NewWriter(5) // opcode(1) + result(4)

	if err := w.WriteByte(OpcodeRestartResponse); err != nil {
		return nil, err
	}

	// Convert bool to int32 (1 = true, 0 = false)
	resultInt := int32(0)
	if p.Result {
		resultInt = 1
	}
	w.WriteInt(resultInt)

	return w.Bytes(), nil
}
