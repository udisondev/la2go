package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestBlock is the opcode for block/unblock/list request (C2S 0xA0).
const OpcodeRequestBlock = 0xA0

// Block operation types.
const (
	BlockTypeBlock      = 0
	BlockTypeUnblock    = 1
	BlockTypeList       = 2
	BlockTypeAllBlock   = 3
	BlockTypeAllUnblock = 4
)

// RequestBlock represents a client block management request.
type RequestBlock struct {
	Type int32  // 0=block, 1=unblock, 2=list, 3=allblock, 4=allunblock
	Name string // target name (only for type 0 and 1)
}

// ParseRequestBlock parses a block request packet.
func ParseRequestBlock(data []byte) (*RequestBlock, error) {
	r := packet.NewReader(data)
	blockType, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading type: %w", err)
	}

	var name string
	if blockType == BlockTypeBlock || blockType == BlockTypeUnblock {
		name, err = r.ReadString()
		if err != nil {
			return nil, fmt.Errorf("reading name: %w", err)
		}
	}

	return &RequestBlock{Type: blockType, Name: name}, nil
}
