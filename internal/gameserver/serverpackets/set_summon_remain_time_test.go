package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestSetSummonRemainTime_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		maxTime       int32
		remainingTime int32
	}{
		{
			name:          "half remaining",
			maxTime:       600,
			remainingTime: 300,
		},
		{
			name:          "full remaining",
			maxTime:       1200,
			remainingTime: 1200,
		},
		{
			name:          "zero remaining",
			maxTime:       600,
			remainingTime: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pkt := &SetSummonRemainTime{
				MaxTime:       tt.maxTime,
				RemainingTime: tt.remainingTime,
			}
			data, err := pkt.Write()
			if err != nil {
				t.Fatalf("Write() error: %v", err)
			}

			// opcode(1) + maxTime(4) + remainingTime(4) = 9
			if len(data) != 9 {
				t.Fatalf("len(data) = %d; want 9", len(data))
			}

			if data[0] != OpcodeSetSummonRemainTime {
				t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeSetSummonRemainTime)
			}

			r := packet.NewReader(data[1:])

			maxTime, err := r.ReadInt()
			if err != nil {
				t.Fatalf("ReadInt maxTime: %v", err)
			}
			if maxTime != tt.maxTime {
				t.Errorf("maxTime = %d; want %d", maxTime, tt.maxTime)
			}

			remainTime, err := r.ReadInt()
			if err != nil {
				t.Fatalf("ReadInt remainTime: %v", err)
			}
			if remainTime != tt.remainingTime {
				t.Errorf("remainTime = %d; want %d", remainTime, tt.remainingTime)
			}

			if r.Remaining() != 0 {
				t.Errorf("remaining bytes: got %d, want 0", r.Remaining())
			}
		})
	}
}

func TestSetSummonRemainTime_Opcode(t *testing.T) {
	t.Parallel()
	if OpcodeSetSummonRemainTime != 0xD1 {
		t.Errorf("opcode = 0x%02X; want 0xD1", OpcodeSetSummonRemainTime)
	}
}
