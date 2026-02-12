package combat

import (
	"math/rand/v2"

	"github.com/udisondev/la2go/internal/config"
	"github.com/udisondev/la2go/internal/data"
)

// DropResult represents a single item that dropped from NPC death.
// Phase 5.10: DROP/LOOT System.
type DropResult struct {
	ItemID int32
	Count  int32
}

// CalculateDrops computes which items drop when NPC with given templateID dies.
//
// Algorithm (simplified, no level gap / premium):
//  1. Look up npcDef drops (grouped drop lists)
//  2. For each drop group:
//     a) Roll group chance (group.Chance × rates.DeathDropChanceMultiplier)
//     b) If group passes, for each item in group:
//     c) Roll item chance (item.Chance × rates.DeathDropChanceMultiplier)
//     d) Count = random(min..max) × rates.DeathDropAmountMultiplier
//     e) Append DropResult if count > 0
//
// Phase 5.10: DROP/LOOT System.
func CalculateDrops(templateID int32, rates *config.Rates) []DropResult {
	def := data.GetNpcDef(templateID)
	if def == nil {
		return nil
	}

	drops := def.Drops()
	if len(drops) == 0 {
		return nil
	}

	chanceMultiplier := 1.0
	amountMultiplier := 1.0
	if rates != nil {
		chanceMultiplier = rates.DeathDropChanceMultiplier
		amountMultiplier = rates.DeathDropAmountMultiplier
	}

	var results []DropResult

	for _, group := range drops {
		// Roll group chance
		groupChance := group.Chance() * chanceMultiplier
		if groupChance <= 0 {
			continue
		}
		if groupChance < 100 && rand.Float64()*100.0 >= groupChance {
			continue
		}

		// Group passed — roll each item
		for _, item := range group.Items() {
			itemChance := item.Chance() * chanceMultiplier
			if itemChance <= 0 {
				continue
			}
			if itemChance < 100 && rand.Float64()*100.0 >= itemChance {
				continue
			}

			// Calculate count
			minCount := item.Min()
			maxCount := item.Max()
			if minCount <= 0 {
				minCount = 1
			}
			if maxCount < minCount {
				maxCount = minCount
			}

			count := minCount
			if maxCount > minCount {
				count = int32(rand.IntN(int(maxCount-minCount+1))) + minCount
			}

			// Apply amount multiplier
			count = int32(float64(count) * amountMultiplier)
			if count <= 0 {
				count = 1
			}

			results = append(results, DropResult{
				ItemID: item.ItemID(),
				Count:  count,
			})
		}
	}

	return results
}
