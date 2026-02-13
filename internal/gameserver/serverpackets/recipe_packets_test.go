package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestRecipeBookItemList_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		pkt            *RecipeBookItemList
		wantOpcode     byte
		wantIsDwarven  int32 // 0=dwarven, 1=common
		wantMaxMP      int32
		wantCount      int32
		wantRecipeIDs  []int32
		wantDisplayIdx []int32
	}{
		{
			name: "empty recipe list",
			pkt: &RecipeBookItemList{
				IsDwarvenCraft: true,
				MaxMP:          500,
				RecipeIDs:      nil,
			},
			wantOpcode:    0xD6,
			wantIsDwarven: 0,
			wantMaxMP:     500,
			wantCount:     0,
		},
		{
			name: "three recipes dwarven",
			pkt: &RecipeBookItemList{
				IsDwarvenCraft: true,
				MaxMP:          1200,
				RecipeIDs:      []int32{101, 202, 303},
			},
			wantOpcode:     0xD6,
			wantIsDwarven:  0,
			wantMaxMP:      1200,
			wantCount:      3,
			wantRecipeIDs:  []int32{101, 202, 303},
			wantDisplayIdx: []int32{1, 2, 3},
		},
		{
			name: "common craft encoding",
			pkt: &RecipeBookItemList{
				IsDwarvenCraft: false,
				MaxMP:          800,
				RecipeIDs:      []int32{55},
			},
			wantOpcode:     0xD6,
			wantIsDwarven:  1,
			wantMaxMP:      800,
			wantCount:      1,
			wantRecipeIDs:  []int32{55},
			wantDisplayIdx: []int32{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := tt.pkt.Write()
			if err != nil {
				t.Fatalf("RecipeBookItemList.Write() error: %v", err)
			}

			r := packet.NewReader(data)

			opcode, err := r.ReadByte()
			if err != nil {
				t.Fatalf("read opcode: %v", err)
			}
			if opcode != tt.wantOpcode {
				t.Errorf("RecipeBookItemList.Write() opcode = 0x%02X; want 0x%02X", opcode, tt.wantOpcode)
			}

			isDwarven, err := r.ReadInt()
			if err != nil {
				t.Fatalf("read isDwarvenCraft: %v", err)
			}
			if isDwarven != tt.wantIsDwarven {
				t.Errorf("RecipeBookItemList.Write() isDwarvenCraft = %d; want %d", isDwarven, tt.wantIsDwarven)
			}

			maxMP, err := r.ReadInt()
			if err != nil {
				t.Fatalf("read maxMP: %v", err)
			}
			if maxMP != tt.wantMaxMP {
				t.Errorf("RecipeBookItemList.Write() maxMP = %d; want %d", maxMP, tt.wantMaxMP)
			}

			count, err := r.ReadInt()
			if err != nil {
				t.Fatalf("read recipeCount: %v", err)
			}
			if count != tt.wantCount {
				t.Errorf("RecipeBookItemList.Write() recipeCount = %d; want %d", count, tt.wantCount)
			}

			for i := range int(count) {
				recipeID, err := r.ReadInt()
				if err != nil {
					t.Fatalf("read recipeIDs[%d]: %v", i, err)
				}
				if recipeID != tt.wantRecipeIDs[i] {
					t.Errorf("RecipeBookItemList.Write() recipeIDs[%d] = %d; want %d", i, recipeID, tt.wantRecipeIDs[i])
				}

				displayIdx, err := r.ReadInt()
				if err != nil {
					t.Fatalf("read displayIndex[%d]: %v", i, err)
				}
				if displayIdx != tt.wantDisplayIdx[i] {
					t.Errorf("RecipeBookItemList.Write() displayIndex[%d] = %d; want %d", i, displayIdx, tt.wantDisplayIdx[i])
				}
			}

			if remaining := r.Remaining(); remaining != 0 {
				t.Errorf("RecipeBookItemList.Write() %d unexpected trailing bytes", remaining)
			}
		})
	}
}

func TestRecipeItemMakeInfo_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		pkt           *RecipeItemMakeInfo
		wantOpcode    byte
		wantRecipeID  int32
		wantIsDwarven int32
		wantCurrentMP int32
		wantMaxMP     int32
		wantSuccess   int32
	}{
		{
			name: "success dwarven craft",
			pkt: &RecipeItemMakeInfo{
				RecipeListID:   42,
				IsDwarvenCraft: true,
				CurrentMP:      350,
				MaxMP:          500,
				Success:        true,
			},
			wantOpcode:    0xD7,
			wantRecipeID:  42,
			wantIsDwarven: 0,
			wantCurrentMP: 350,
			wantMaxMP:     500,
			wantSuccess:   1,
		},
		{
			name: "failure common craft",
			pkt: &RecipeItemMakeInfo{
				RecipeListID:   99,
				IsDwarvenCraft: false,
				CurrentMP:      100,
				MaxMP:          1000,
				Success:        false,
			},
			wantOpcode:    0xD7,
			wantRecipeID:  99,
			wantIsDwarven: 1,
			wantCurrentMP: 100,
			wantMaxMP:     1000,
			wantSuccess:   0,
		},
		{
			name: "success common craft",
			pkt: &RecipeItemMakeInfo{
				RecipeListID:   7,
				IsDwarvenCraft: false,
				CurrentMP:      0,
				MaxMP:          200,
				Success:        true,
			},
			wantOpcode:    0xD7,
			wantRecipeID:  7,
			wantIsDwarven: 1,
			wantCurrentMP: 0,
			wantMaxMP:     200,
			wantSuccess:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := tt.pkt.Write()
			if err != nil {
				t.Fatalf("RecipeItemMakeInfo.Write() error: %v", err)
			}

			r := packet.NewReader(data)

			opcode, err := r.ReadByte()
			if err != nil {
				t.Fatalf("read opcode: %v", err)
			}
			if opcode != tt.wantOpcode {
				t.Errorf("RecipeItemMakeInfo.Write() opcode = 0x%02X; want 0x%02X", opcode, tt.wantOpcode)
			}

			recipeID, err := r.ReadInt()
			if err != nil {
				t.Fatalf("read recipeListId: %v", err)
			}
			if recipeID != tt.wantRecipeID {
				t.Errorf("RecipeItemMakeInfo.Write() recipeListId = %d; want %d", recipeID, tt.wantRecipeID)
			}

			isDwarven, err := r.ReadInt()
			if err != nil {
				t.Fatalf("read isDwarvenCraft: %v", err)
			}
			if isDwarven != tt.wantIsDwarven {
				t.Errorf("RecipeItemMakeInfo.Write() isDwarvenCraft = %d; want %d", isDwarven, tt.wantIsDwarven)
			}

			currentMP, err := r.ReadInt()
			if err != nil {
				t.Fatalf("read currentMp: %v", err)
			}
			if currentMP != tt.wantCurrentMP {
				t.Errorf("RecipeItemMakeInfo.Write() currentMp = %d; want %d", currentMP, tt.wantCurrentMP)
			}

			maxMP, err := r.ReadInt()
			if err != nil {
				t.Fatalf("read maxMp: %v", err)
			}
			if maxMP != tt.wantMaxMP {
				t.Errorf("RecipeItemMakeInfo.Write() maxMp = %d; want %d", maxMP, tt.wantMaxMP)
			}

			success, err := r.ReadInt()
			if err != nil {
				t.Fatalf("read success: %v", err)
			}
			if success != tt.wantSuccess {
				t.Errorf("RecipeItemMakeInfo.Write() success = %d; want %d", success, tt.wantSuccess)
			}

			if remaining := r.Remaining(); remaining != 0 {
				t.Errorf("RecipeItemMakeInfo.Write() %d unexpected trailing bytes", remaining)
			}
		})
	}
}

func TestRecipeShopItemInfo_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		pkt               *RecipeShopItemInfo
		wantOpcode        byte
		wantCrafterObjID  int32
		wantRecipeID      int32
		wantCurrentMP     int32
		wantMaxMP         int32
		wantUnknown       int32
	}{
		{
			name: "basic packet",
			pkt: &RecipeShopItemInfo{
				CrafterObjectID: 12345,
				RecipeID:        800,
				CurrentMP:       300,
				MaxMP:           600,
			},
			wantOpcode:       0xDA,
			wantCrafterObjID: 12345,
			wantRecipeID:     800,
			wantCurrentMP:    300,
			wantMaxMP:        600,
			wantUnknown:      -1,
		},
		{
			name: "zero mp values",
			pkt: &RecipeShopItemInfo{
				CrafterObjectID: 1,
				RecipeID:        55,
				CurrentMP:       0,
				MaxMP:           0,
			},
			wantOpcode:       0xDA,
			wantCrafterObjID: 1,
			wantRecipeID:     55,
			wantCurrentMP:    0,
			wantMaxMP:        0,
			wantUnknown:      -1,
		},
		{
			name: "large crafter object id",
			pkt: &RecipeShopItemInfo{
				CrafterObjectID: 0x7FFFFFFF,
				RecipeID:        999,
				CurrentMP:       1000,
				MaxMP:           2000,
			},
			wantOpcode:       0xDA,
			wantCrafterObjID: 0x7FFFFFFF,
			wantRecipeID:     999,
			wantCurrentMP:    1000,
			wantMaxMP:        2000,
			wantUnknown:      -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := tt.pkt.Write()
			if err != nil {
				t.Fatalf("RecipeShopItemInfo.Write() error: %v", err)
			}

			r := packet.NewReader(data)

			opcode, err := r.ReadByte()
			if err != nil {
				t.Fatalf("read opcode: %v", err)
			}
			if opcode != tt.wantOpcode {
				t.Errorf("RecipeShopItemInfo.Write() opcode = 0x%02X; want 0x%02X", opcode, tt.wantOpcode)
			}

			crafterObjID, err := r.ReadInt()
			if err != nil {
				t.Fatalf("read crafterObjectId: %v", err)
			}
			if crafterObjID != tt.wantCrafterObjID {
				t.Errorf("RecipeShopItemInfo.Write() crafterObjectId = %d; want %d", crafterObjID, tt.wantCrafterObjID)
			}

			recipeID, err := r.ReadInt()
			if err != nil {
				t.Fatalf("read recipeId: %v", err)
			}
			if recipeID != tt.wantRecipeID {
				t.Errorf("RecipeShopItemInfo.Write() recipeId = %d; want %d", recipeID, tt.wantRecipeID)
			}

			currentMP, err := r.ReadInt()
			if err != nil {
				t.Fatalf("read currentMp: %v", err)
			}
			if currentMP != tt.wantCurrentMP {
				t.Errorf("RecipeShopItemInfo.Write() currentMp = %d; want %d", currentMP, tt.wantCurrentMP)
			}

			maxMP, err := r.ReadInt()
			if err != nil {
				t.Fatalf("read maxMp: %v", err)
			}
			if maxMP != tt.wantMaxMP {
				t.Errorf("RecipeShopItemInfo.Write() maxMp = %d; want %d", maxMP, tt.wantMaxMP)
			}

			unknown, err := r.ReadInt()
			if err != nil {
				t.Fatalf("read unknown: %v", err)
			}
			if unknown != tt.wantUnknown {
				t.Errorf("RecipeShopItemInfo.Write() unknown = %d; want %d (0xFFFFFFFF)", unknown, tt.wantUnknown)
			}

			if remaining := r.Remaining(); remaining != 0 {
				t.Errorf("RecipeShopItemInfo.Write() %d unexpected trailing bytes", remaining)
			}
		})
	}
}
