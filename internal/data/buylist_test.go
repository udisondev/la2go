package data

import (
	"testing"
)

func TestLoadBuylists(t *testing.T) {
	if err := LoadBuylists(); err != nil {
		t.Fatalf("LoadBuylists() error: %v", err)
	}

	if len(BuylistTable) == 0 {
		t.Skip("no buylists loaded (generated data may be empty)")
	}

	// Verify NPC index was built
	if NpcBuylistIndex == nil {
		t.Fatal("NpcBuylistIndex is nil after LoadBuylists()")
	}
}

func TestGetBuylistProducts(t *testing.T) {
	if err := LoadBuylists(); err != nil {
		t.Fatalf("LoadBuylists() error: %v", err)
	}

	if len(BuylistTable) == 0 {
		t.Skip("no buylists loaded")
	}

	// Get first available listID
	var listID int32
	for id := range BuylistTable {
		listID = id
		break
	}

	products := GetBuylistProducts(listID)
	if products == nil {
		t.Fatalf("GetBuylistProducts(%d) = nil", listID)
	}

	// Verify product fields are populated
	for i, p := range products {
		if p.ItemID <= 0 {
			t.Errorf("products[%d].ItemID = %d, want > 0", i, p.ItemID)
		}
	}
}

func TestGetBuylistProducts_NotFound(t *testing.T) {
	if err := LoadBuylists(); err != nil {
		t.Fatalf("LoadBuylists() error: %v", err)
	}

	products := GetBuylistProducts(-1)
	if products != nil {
		t.Errorf("GetBuylistProducts(-1) = %v, want nil", products)
	}
}

func TestFindProductInBuylist(t *testing.T) {
	if err := LoadBuylists(); err != nil {
		t.Fatalf("LoadBuylists() error: %v", err)
	}

	if len(BuylistTable) == 0 {
		t.Skip("no buylists loaded")
	}

	// Find a buylist with at least one item
	var listID int32
	var firstItemID int32
	for id, bl := range BuylistTable {
		if len(bl.items) > 0 {
			listID = id
			firstItemID = bl.items[0].itemID
			break
		}
	}
	if firstItemID == 0 {
		t.Skip("no buylists with items")
	}

	// Should find the item
	product := FindProductInBuylist(listID, firstItemID)
	if product == nil {
		t.Fatalf("FindProductInBuylist(%d, %d) = nil", listID, firstItemID)
	}
	if product.ItemID != firstItemID {
		t.Errorf("product.ItemID = %d, want %d", product.ItemID, firstItemID)
	}

	// Should not find non-existent item
	notFound := FindProductInBuylist(listID, -1)
	if notFound != nil {
		t.Errorf("FindProductInBuylist(%d, -1) = %v, want nil", listID, notFound)
	}
}

func TestGetBuylistsByNpc(t *testing.T) {
	if err := LoadBuylists(); err != nil {
		t.Fatalf("LoadBuylists() error: %v", err)
	}

	// Non-existent NPC should return nil
	lists := GetBuylistsByNpc(-1)
	if lists != nil {
		t.Errorf("GetBuylistsByNpc(-1) = %v, want nil", lists)
	}

	// If there are NPCs with buylists, verify they have valid list IDs
	for npcID, listIDs := range NpcBuylistIndex {
		if len(listIDs) == 0 {
			t.Errorf("NPC %d has empty buylist list", npcID)
		}
		for _, id := range listIDs {
			if BuylistTable[id] == nil {
				t.Errorf("NPC %d references invalid buylist %d", npcID, id)
			}
		}
	}
}
