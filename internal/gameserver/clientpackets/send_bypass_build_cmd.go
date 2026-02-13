package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeSendBypassBuildCmd is the C2S opcode 0x5B.
// GM //command bypass from client (e.g., //admin, //spawn).
const OpcodeSendBypassBuildCmd byte = 0x5B

// SendBypassBuildCmd contains the GM command string.
type SendBypassBuildCmd struct {
	Command string
}

// ParseSendBypassBuildCmd parses the C2S SendBypassBuildCmd packet.
func ParseSendBypassBuildCmd(data []byte) (*SendBypassBuildCmd, error) {
	r := packet.NewReader(data)
	cmd, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading command: %w", err)
	}
	return &SendBypassBuildCmd{Command: cmd}, nil
}
