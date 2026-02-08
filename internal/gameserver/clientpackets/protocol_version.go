package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/constants"
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const OpcodeProtocolVersion = 0x0E

// ProtocolVersion is the first packet sent by the client after receiving KeyPacket.
// Contains the protocol revision (0x0106 for Interlude).
//
// Structure:
// - int32: protocol revision (0x0106 for Interlude)
type ProtocolVersion struct {
	ProtocolRevision int32
}

// Parse parses a ProtocolVersion packet from the given data (without opcode).
func ParseProtocolVersion(data []byte) (*ProtocolVersion, error) {
	r := packet.NewReader(data)

	revision, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading protocol revision: %w", err)
	}

	return &ProtocolVersion{
		ProtocolRevision: revision,
	}, nil
}

// IsValid checks if the protocol revision is valid for Interlude.
func (p *ProtocolVersion) IsValid() bool {
	return p.ProtocolRevision == constants.ProtocolRevisionInterlude
}
