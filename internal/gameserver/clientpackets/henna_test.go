package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestParseRequestHennaItemList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		buildData func() []byte
		wantVal   int32
		wantErr   bool
	}{
		{
			name: "valid packet",
			buildData: func() []byte {
				w := packet.NewWriter(4)
				w.WriteInt(42)
				return w.Bytes()
			},
			wantVal: 42,
		},
		{
			name: "empty data",
			buildData: func() []byte {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseRequestHennaItemList(tt.buildData())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if got.Unknown != tt.wantVal {
				t.Errorf("Unknown = %d; want %d", got.Unknown, tt.wantVal)
			}
		})
	}
}

func TestParseRequestHennaItemInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		buildData func() []byte
		wantID  int32
		wantErr bool
	}{
		{
			name: "valid symbolID",
			buildData: func() []byte {
				w := packet.NewWriter(4)
				w.WriteInt(101)
				return w.Bytes()
			},
			wantID: 101,
		},
		{
			name: "empty data",
			buildData: func() []byte {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseRequestHennaItemInfo(tt.buildData())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if got.SymbolID != tt.wantID {
				t.Errorf("SymbolID = %d; want %d", got.SymbolID, tt.wantID)
			}
		})
	}
}

func TestParseRequestHennaEquip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		buildData func() []byte
		wantID  int32
		wantErr bool
	}{
		{
			name: "valid symbolID",
			buildData: func() []byte {
				w := packet.NewWriter(4)
				w.WriteInt(55)
				return w.Bytes()
			},
			wantID: 55,
		},
		{
			name: "empty data",
			buildData: func() []byte {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseRequestHennaEquip(tt.buildData())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if got.SymbolID != tt.wantID {
				t.Errorf("SymbolID = %d; want %d", got.SymbolID, tt.wantID)
			}
		})
	}
}

func TestParseRequestHennaRemoveList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		buildData func() []byte
		wantVal   int32
		wantErr   bool
	}{
		{
			name: "valid packet",
			buildData: func() []byte {
				w := packet.NewWriter(4)
				w.WriteInt(7)
				return w.Bytes()
			},
			wantVal: 7,
		},
		{
			name: "empty data",
			buildData: func() []byte {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseRequestHennaRemoveList(tt.buildData())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if got.Unknown != tt.wantVal {
				t.Errorf("Unknown = %d; want %d", got.Unknown, tt.wantVal)
			}
		})
	}
}

func TestParseRequestHennaItemRemoveInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		buildData func() []byte
		wantID  int32
		wantErr bool
	}{
		{
			name: "valid symbolID",
			buildData: func() []byte {
				w := packet.NewWriter(4)
				w.WriteInt(200)
				return w.Bytes()
			},
			wantID: 200,
		},
		{
			name: "empty data",
			buildData: func() []byte {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseRequestHennaItemRemoveInfo(tt.buildData())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if got.SymbolID != tt.wantID {
				t.Errorf("SymbolID = %d; want %d", got.SymbolID, tt.wantID)
			}
		})
	}
}

func TestParseRequestHennaRemove(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		buildData func() []byte
		wantID  int32
		wantErr bool
	}{
		{
			name: "valid symbolID",
			buildData: func() []byte {
				w := packet.NewWriter(4)
				w.WriteInt(33)
				return w.Bytes()
			},
			wantID: 33,
		},
		{
			name: "empty data",
			buildData: func() []byte {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseRequestHennaRemove(tt.buildData())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if got.SymbolID != tt.wantID {
				t.Errorf("SymbolID = %d; want %d", got.SymbolID, tt.wantID)
			}
		})
	}
}
