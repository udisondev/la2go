package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestParseRequestSetSeed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		manorID   int32
		seeds     []SeedEntry
	}{
		{
			name:    "empty list",
			manorID: 1,
			seeds:   nil,
		},
		{
			name:    "single seed",
			manorID: 2,
			seeds: []SeedEntry{
				{SeedID: 5016, StartAmount: 100, Price: 3000},
			},
		},
		{
			name:    "multiple seeds",
			manorID: 1,
			seeds: []SeedEntry{
				{SeedID: 5016, StartAmount: 100, Price: 3000},
				{SeedID: 5017, StartAmount: 200, Price: 5000},
				{SeedID: 5018, StartAmount: 50, Price: 1500},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := packet.NewWriter(64)
			w.WriteInt(tt.manorID)
			w.WriteInt(int32(len(tt.seeds)))
			for _, s := range tt.seeds {
				w.WriteInt(s.SeedID)
				w.WriteInt(s.StartAmount)
				w.WriteInt(s.Price)
			}

			pkt, err := ParseRequestSetSeed(w.Bytes())
			if err != nil {
				t.Fatalf("ParseRequestSetSeed: %v", err)
			}

			if pkt.ManorID != tt.manorID {
				t.Errorf("ManorID = %d; want %d", pkt.ManorID, tt.manorID)
			}

			if len(pkt.Seeds) != len(tt.seeds) {
				t.Fatalf("len(Seeds) = %d; want %d", len(pkt.Seeds), len(tt.seeds))
			}

			for i, want := range tt.seeds {
				got := pkt.Seeds[i]
				if got.SeedID != want.SeedID {
					t.Errorf("Seeds[%d].SeedID = %d; want %d", i, got.SeedID, want.SeedID)
				}
				if got.StartAmount != want.StartAmount {
					t.Errorf("Seeds[%d].StartAmount = %d; want %d", i, got.StartAmount, want.StartAmount)
				}
				if got.Price != want.Price {
					t.Errorf("Seeds[%d].Price = %d; want %d", i, got.Price, want.Price)
				}
			}
		})
	}
}

func TestParseRequestSetSeed_ShortData(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestSetSeed(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
}

func TestParseRequestSetSeed_InvalidCount(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(8)
	w.WriteInt(1) // manorID
	w.WriteInt(-1) // invalid count

	_, err := ParseRequestSetSeed(w.Bytes())
	if err == nil {
		t.Error("expected error for negative count")
	}
}

func TestParseRequestSetCrop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		manorID int32
		crops   []CropEntry
	}{
		{
			name:    "empty list",
			manorID: 1,
			crops:   nil,
		},
		{
			name:    "single crop",
			manorID: 3,
			crops: []CropEntry{
				{CropID: 5078, StartAmount: 100, Price: 2000, RewardType: 1},
			},
		},
		{
			name:    "multiple crops",
			manorID: 2,
			crops: []CropEntry{
				{CropID: 5078, StartAmount: 100, Price: 2000, RewardType: 1},
				{CropID: 5079, StartAmount: 50, Price: 3000, RewardType: 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := packet.NewWriter(64)
			w.WriteInt(tt.manorID)
			w.WriteInt(int32(len(tt.crops)))
			for _, c := range tt.crops {
				w.WriteInt(c.CropID)
				w.WriteInt(c.StartAmount)
				w.WriteInt(c.Price)
				w.WriteByte(c.RewardType)
			}

			pkt, err := ParseRequestSetCrop(w.Bytes())
			if err != nil {
				t.Fatalf("ParseRequestSetCrop: %v", err)
			}

			if pkt.ManorID != tt.manorID {
				t.Errorf("ManorID = %d; want %d", pkt.ManorID, tt.manorID)
			}

			if len(pkt.Crops) != len(tt.crops) {
				t.Fatalf("len(Crops) = %d; want %d", len(pkt.Crops), len(tt.crops))
			}

			for i, want := range tt.crops {
				got := pkt.Crops[i]
				if got.CropID != want.CropID {
					t.Errorf("Crops[%d].CropID = %d; want %d", i, got.CropID, want.CropID)
				}
				if got.StartAmount != want.StartAmount {
					t.Errorf("Crops[%d].StartAmount = %d; want %d", i, got.StartAmount, want.StartAmount)
				}
				if got.Price != want.Price {
					t.Errorf("Crops[%d].Price = %d; want %d", i, got.Price, want.Price)
				}
				if got.RewardType != want.RewardType {
					t.Errorf("Crops[%d].RewardType = %d; want %d", i, got.RewardType, want.RewardType)
				}
			}
		})
	}
}

func TestParseRequestSetCrop_ShortData(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestSetCrop(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
}

func TestParseRequestSetCrop_InvalidCount(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(8)
	w.WriteInt(1)   // manorID
	w.WriteInt(501) // exceeds max

	_, err := ParseRequestSetCrop(w.Bytes())
	if err == nil {
		t.Error("expected error for count > 500")
	}
}

func TestParseRequestSetCrop_TruncatedEntry(t *testing.T) {
	t.Parallel()

	w := packet.NewWriter(16)
	w.WriteInt(1) // manorID
	w.WriteInt(1) // count = 1
	w.WriteInt(5078) // cropID
	// Missing amount, price, rewardType

	_, err := ParseRequestSetCrop(w.Bytes())
	if err == nil {
		t.Error("expected error for truncated entry")
	}
}
