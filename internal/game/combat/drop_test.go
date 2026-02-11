package combat

import (
	"testing"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/data"
)

func init() {
	// Load NPC and item templates for tests
	if err := data.LoadNpcTemplates(); err != nil {
		panic("load NPC templates: " + err.Error())
	}
	if err := data.LoadItemTemplates(); err != nil {
		panic("load item templates: " + err.Error())
	}
}

func TestCalculateDrops_UnknownNpc(t *testing.T) {
	rates := &config.Rates{
		DeathDropChanceMultiplier: 1.0,
		DeathDropAmountMultiplier: 1.0,
	}

	drops := CalculateDrops(999999, rates)
	if drops != nil {
		t.Errorf("expected nil drops for unknown NPC, got %v", drops)
	}
}

func TestCalculateDrops_NpcWithoutDrops(t *testing.T) {
	// NPC 18002 "Blood Queen" has no drops in generated data
	rates := &config.Rates{
		DeathDropChanceMultiplier: 1.0,
		DeathDropAmountMultiplier: 1.0,
	}

	drops := CalculateDrops(18002, rates)
	if drops != nil {
		t.Errorf("expected nil drops for NPC without drop table, got %v", drops)
	}
}

func TestCalculateDrops_NilRates(t *testing.T) {
	// NPC 13031 "Huge Pig": 100% group, 100% item chance, itemID=9142, min=1, max=2
	drops := CalculateDrops(13031, nil)
	if len(drops) == 0 {
		t.Fatal("expected drops with nil rates (should use default 1.0)")
	}

	if drops[0].ItemID != 9142 {
		t.Errorf("expected itemID 9142, got %d", drops[0].ItemID)
	}
	if drops[0].Count < 1 || drops[0].Count > 2 {
		t.Errorf("expected count 1-2, got %d", drops[0].Count)
	}
}

func TestCalculateDrops_100PercentChance(t *testing.T) {
	// NPC 13031 "Huge Pig": 100% group, 100% item chance, itemID=9142
	rates := &config.Rates{
		DeathDropChanceMultiplier: 1.0,
		DeathDropAmountMultiplier: 1.0,
	}

	// Run multiple times — should always drop
	for range 100 {
		drops := CalculateDrops(13031, rates)
		if len(drops) != 1 {
			t.Fatalf("expected exactly 1 drop for 100%% chance NPC, got %d", len(drops))
		}
		if drops[0].ItemID != 9142 {
			t.Errorf("expected itemID 9142, got %d", drops[0].ItemID)
		}
		if drops[0].Count < 1 || drops[0].Count > 2 {
			t.Errorf("expected count 1-2, got %d", drops[0].Count)
		}
	}
}

func TestCalculateDrops_MinMaxCount(t *testing.T) {
	// NPC 13031: min=1, max=2
	rates := &config.Rates{
		DeathDropChanceMultiplier: 1.0,
		DeathDropAmountMultiplier: 1.0,
	}

	gotMin := false
	gotMax := false
	for range 1000 {
		drops := CalculateDrops(13031, rates)
		if len(drops) != 1 {
			t.Fatalf("expected 1 drop, got %d", len(drops))
		}
		switch drops[0].Count {
		case 1:
			gotMin = true
		case 2:
			gotMax = true
		default:
			t.Fatalf("unexpected count %d (want 1 or 2)", drops[0].Count)
		}
		if gotMin && gotMax {
			break
		}
	}

	if !gotMin {
		t.Error("never got min count=1 in 1000 iterations")
	}
	if !gotMax {
		t.Error("never got max count=2 in 1000 iterations")
	}
}

func TestCalculateDrops_AmountMultiplier(t *testing.T) {
	// NPC 13031: min=1, max=2, with 3x amount multiplier
	rates := &config.Rates{
		DeathDropChanceMultiplier: 1.0,
		DeathDropAmountMultiplier: 3.0,
	}

	for range 100 {
		drops := CalculateDrops(13031, rates)
		if len(drops) != 1 {
			t.Fatalf("expected 1 drop, got %d", len(drops))
		}
		// min=1*3=3, max=2*3=6
		if drops[0].Count < 3 || drops[0].Count > 6 {
			t.Errorf("expected count 3-6 with 3x multiplier, got %d", drops[0].Count)
		}
	}
}

func TestCalculateDrops_ZeroChanceMultiplier(t *testing.T) {
	// With 0x chance multiplier, nothing should drop
	rates := &config.Rates{
		DeathDropChanceMultiplier: 0.0,
		DeathDropAmountMultiplier: 1.0,
	}

	for range 100 {
		drops := CalculateDrops(13031, rates)
		if len(drops) != 0 {
			t.Fatalf("expected 0 drops with 0x chance, got %d", len(drops))
		}
	}
}

func TestCalculateDrops_MultipleGroups(t *testing.T) {
	// NPC 18003 "Bearded Keltir" has 5+ drop groups
	// With high chance multiplier, we should get items from multiple groups
	rates := &config.Rates{
		DeathDropChanceMultiplier: 100.0, // Boost all chances to near 100%
		DeathDropAmountMultiplier: 1.0,
	}

	totalDrops := 0
	const iterations = 100
	for range iterations {
		drops := CalculateDrops(18003, rates)
		totalDrops += len(drops)
	}

	// With 100x multiplier and 5 groups, we should get many drops
	avgDrops := float64(totalDrops) / float64(iterations)
	if avgDrops < 2.0 {
		t.Errorf("expected avg >2 drops with 100x chance, got %.1f", avgDrops)
	}
}

func TestCalculateDrops_AdenaDropStatistics(t *testing.T) {
	// NPC 18003 "Bearded Keltir": group chance=100, itemID=57 (Adena), chance=70, min=32, max=48
	rates := &config.Rates{
		DeathDropChanceMultiplier: 1.0,
		DeathDropAmountMultiplier: 1.0,
	}

	adenaCount := 0
	const iterations = 1000
	for range iterations {
		drops := CalculateDrops(18003, rates)
		for _, d := range drops {
			if d.ItemID == 57 {
				adenaCount++
				if d.Count < 32 || d.Count > 48 {
					t.Errorf("Adena count out of range: got %d, want 32-48", d.Count)
				}
			}
		}
	}

	// Adena has 70% drop chance, expect ~700 drops ± margin
	rate := float64(adenaCount) / float64(iterations) * 100.0
	if rate < 55.0 || rate > 85.0 {
		t.Errorf("Adena drop rate = %.1f%%, expected ~70%% (55-85%%)", rate)
	}
}
