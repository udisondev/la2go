package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestParseRequestRecipeBookOpen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		buildData      func() []byte
		wantDwarven    bool
		wantErr        bool
	}{
		{
			name: "isDwarven=0 means dwarven craft",
			buildData: func() []byte {
				w := packet.NewWriter(4)
				w.WriteInt(0)
				return w.Bytes()
			},
			wantDwarven: true,
		},
		{
			name: "isDwarven=1 means common craft",
			buildData: func() []byte {
				w := packet.NewWriter(4)
				w.WriteInt(1)
				return w.Bytes()
			},
			wantDwarven: false,
		},
		{
			name: "empty data returns error",
			buildData: func() []byte {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := tt.buildData()
			got, err := ParseRequestRecipeBookOpen(data)

			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseRequestRecipeBookOpen() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseRequestRecipeBookOpen() error: %v", err)
			}

			if got.IsDwarvenCraft != tt.wantDwarven {
				t.Errorf("ParseRequestRecipeBookOpen() IsDwarvenCraft = %v; want %v", got.IsDwarvenCraft, tt.wantDwarven)
			}
		})
	}
}

func TestParseRequestRecipeItemMakeInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		buildData func() []byte
		wantID  int32
		wantErr bool
	}{
		{
			name: "valid recipe list id",
			buildData: func() []byte {
				w := packet.NewWriter(4)
				w.WriteInt(1042)
				return w.Bytes()
			},
			wantID: 1042,
		},
		{
			name: "empty data returns error",
			buildData: func() []byte {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := tt.buildData()
			got, err := ParseRequestRecipeItemMakeInfo(data)

			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseRequestRecipeItemMakeInfo() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseRequestRecipeItemMakeInfo() error: %v", err)
			}

			if got.RecipeListID != tt.wantID {
				t.Errorf("ParseRequestRecipeItemMakeInfo() RecipeListID = %d; want %d", got.RecipeListID, tt.wantID)
			}
		})
	}
}

func TestParseRequestRecipeItemMakeSelf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		buildData func() []byte
		wantID  int32
		wantErr bool
	}{
		{
			name: "valid recipe list id",
			buildData: func() []byte {
				w := packet.NewWriter(4)
				w.WriteInt(2077)
				return w.Bytes()
			},
			wantID: 2077,
		},
		{
			name: "empty data returns error",
			buildData: func() []byte {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := tt.buildData()
			got, err := ParseRequestRecipeItemMakeSelf(data)

			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseRequestRecipeItemMakeSelf() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseRequestRecipeItemMakeSelf() error: %v", err)
			}

			if got.RecipeListID != tt.wantID {
				t.Errorf("ParseRequestRecipeItemMakeSelf() RecipeListID = %d; want %d", got.RecipeListID, tt.wantID)
			}
		})
	}
}
