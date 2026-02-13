package data

// MultisellEntry — exported view of a multisell entry for use outside the data package.
// Phase 8: Trade System Foundation.
type MultisellEntry struct {
	EntryID     int32
	Ingredients []MultisellIngredient
	Productions []MultisellIngredient
}

// MultisellIngredient — exported view of an ingredient or product in a multisell entry.
type MultisellIngredient struct {
	ItemID int32
	Count  int64
}

// GetMultisellEntries returns all entries in a multisell list as exported structs.
// Returns nil if multisell not found.
// Phase 8: Trade System Foundation.
func GetMultisellEntries(listID int32) []MultisellEntry {
	ms := MultisellTable[listID]
	if ms == nil {
		return nil
	}

	entries := make([]MultisellEntry, len(ms.items))
	for i, item := range ms.items {
		ings := make([]MultisellIngredient, len(item.ingredients))
		for j, ing := range item.ingredients {
			ings[j] = MultisellIngredient{
				ItemID: ing.itemID,
				Count:  ing.count,
			}
		}

		prods := make([]MultisellIngredient, len(item.productions))
		for j, prod := range item.productions {
			prods[j] = MultisellIngredient{
				ItemID: prod.itemID,
				Count:  prod.count,
			}
		}

		entries[i] = MultisellEntry{
			EntryID:     int32(i + 1), // 1-indexed
			Ingredients: ings,
			Productions: prods,
		}
	}
	return entries
}

// FindMultisellEntry finds a specific entry by entryID in a multisell list.
// Returns nil if not found. EntryID is 1-indexed.
func FindMultisellEntry(listID int32, entryID int32) *MultisellEntry {
	ms := MultisellTable[listID]
	if ms == nil {
		return nil
	}

	idx := int(entryID - 1)
	if idx < 0 || idx >= len(ms.items) {
		return nil
	}

	item := ms.items[idx]
	ings := make([]MultisellIngredient, len(item.ingredients))
	for j, ing := range item.ingredients {
		ings[j] = MultisellIngredient{
			ItemID: ing.itemID,
			Count:  ing.count,
		}
	}

	prods := make([]MultisellIngredient, len(item.productions))
	for j, prod := range item.productions {
		prods[j] = MultisellIngredient{
			ItemID: prod.itemID,
			Count:  prod.count,
		}
	}

	return &MultisellEntry{
		EntryID:     entryID,
		Ingredients: ings,
		Productions: prods,
	}
}
