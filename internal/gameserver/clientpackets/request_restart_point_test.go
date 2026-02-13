package clientpackets

import "testing"

func TestParseRequestRestartPoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		data      []byte
		wantType  int32
		wantErr   bool
	}{
		{"village", makeInt32Bytes(0), RestartPointVillage, false},
		{"clan hall", makeInt32Bytes(1), RestartPointClanHall, false},
		{"castle", makeInt32Bytes(2), RestartPointCastle, false},
		{"siege HQ", makeInt32Bytes(3), RestartPointSiegeHQ, false},
		{"fixed", makeInt32Bytes(4), RestartPointFixed, false},
		{"too short", []byte{1, 2}, 0, true},
		{"empty", nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pkt, err := ParseRequestRestartPoint(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pkt.PointType != tt.wantType {
				t.Errorf("PointType = %d; want %d", pkt.PointType, tt.wantType)
			}
		})
	}
}

// makeInt32Bytes is defined in augmentation_test.go
