package clientpackets

import "testing"

func TestParseRequestSSQStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     []byte
		wantPage byte
		wantErr  bool
	}{
		{"page 1", []byte{1}, 1, false},
		{"page 2", []byte{2}, 2, false},
		{"page 3", []byte{3}, 3, false},
		{"page 4", []byte{4}, 4, false},
		{"empty data", []byte{}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pkt, err := ParseRequestSSQStatus(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("ParseRequestSSQStatus() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRequestSSQStatus() error = %v", err)
			}
			if pkt.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", pkt.Page, tt.wantPage)
			}
		})
	}
}

func TestOpcodeRequestSSQStatus(t *testing.T) {
	t.Parallel()

	if OpcodeRequestSSQStatus != 0xC7 {
		t.Errorf("OpcodeRequestSSQStatus = 0x%02X, want 0xC7", OpcodeRequestSSQStatus)
	}
}
