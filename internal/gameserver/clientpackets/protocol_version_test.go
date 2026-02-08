package clientpackets

import (
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/constants"
)

func TestParseProtocolVersion(t *testing.T) {
	tests := []struct {
		name     string
		revision int32
		isValid  bool
	}{
		{
			name:     "valid Interlude revision",
			revision: constants.ProtocolRevisionInterlude,
			isValid:  true,
		},
		{
			name:     "invalid revision",
			revision: 0x9999,
			isValid:  false,
		},
		{
			name:     "zero revision",
			revision: 0x0000,
			isValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode packet (without opcode)
			data := make([]byte, 4)
			binary.LittleEndian.PutUint32(data, uint32(tt.revision))

			pkt, err := ParseProtocolVersion(data)
			if err != nil {
				t.Fatalf("ParseProtocolVersion failed: %v", err)
			}

			if pkt.ProtocolRevision != tt.revision {
				t.Errorf("expected revision 0x%04X, got 0x%04X", tt.revision, pkt.ProtocolRevision)
			}

			if pkt.IsValid() != tt.isValid {
				t.Errorf("expected IsValid=%v, got %v", tt.isValid, pkt.IsValid())
			}
		})
	}
}

func TestParseProtocolVersion_NotEnoughData(t *testing.T) {
	data := []byte{0x01, 0x02} // only 2 bytes instead of 4

	_, err := ParseProtocolVersion(data)
	if err == nil {
		t.Error("expected error when parsing incomplete ProtocolVersion packet")
	}
}
