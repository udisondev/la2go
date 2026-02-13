package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestDeleteMacro is the C2S opcode 0xC2.
const OpcodeRequestDeleteMacro byte = 0xC2

// RequestDeleteMacro represents a request to delete a macro.
type RequestDeleteMacro struct {
	MacroID int32
}

// ParseRequestDeleteMacro parses the packet from raw bytes.
func ParseRequestDeleteMacro(data []byte) (*RequestDeleteMacro, error) {
	r := packet.NewReader(data)

	id, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading MacroID: %w", err)
	}

	return &RequestDeleteMacro{
		MacroID: id,
	}, nil
}
