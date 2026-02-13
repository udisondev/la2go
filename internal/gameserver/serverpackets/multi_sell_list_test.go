package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/data"
)

func TestMultiSellList_Write(t *testing.T) {
	entries := []data.MultisellEntry{
		{
			EntryID: 1,
			Ingredients: []data.MultisellIngredient{
				{ItemID: 57, Count: 1000},
			},
			Productions: []data.MultisellIngredient{
				{ItemID: 100, Count: 1},
			},
		},
	}

	pkt := NewMultiSellList(42, entries, 1)
	pktData, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if pktData[0] != OpcodeMultiSellList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", pktData[0], OpcodeMultiSellList)
	}

	// 1 opcode + 4*5 header + 17 entry + 22 prod + 18 ing = 78
	expectedMinSize := 1 + 20 + 17 + 22 + 18
	if len(pktData) < expectedMinSize {
		t.Errorf("packet size = %d, want >= %d", len(pktData), expectedMinSize)
	}
}

func TestMultiSellList_Pagination(t *testing.T) {
	// Create 50 entries (should span 2 pages)
	entries := make([]data.MultisellEntry, 50)
	for i := range entries {
		entries[i] = data.MultisellEntry{
			EntryID: int32(i + 1),
			Ingredients: []data.MultisellIngredient{
				{ItemID: 57, Count: 100},
			},
			Productions: []data.MultisellIngredient{
				{ItemID: int32(100 + i), Count: 1},
			},
		}
	}

	// Page 1 should have 40 entries
	pkt1 := NewMultiSellList(1, entries, 1)
	data1, err := pkt1.Write()
	if err != nil {
		t.Fatalf("Write(page1) error: %v", err)
	}
	if data1[0] != OpcodeMultiSellList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data1[0], OpcodeMultiSellList)
	}

	// Page 2 should have 10 entries
	pkt2 := NewMultiSellList(1, entries, 2)
	data2, err := pkt2.Write()
	if err != nil {
		t.Fatalf("Write(page2) error: %v", err)
	}

	// Page 2 should be smaller than page 1 (10 vs 40 entries)
	if len(data2) >= len(data1) {
		t.Errorf("page2 size (%d) should be < page1 size (%d)", len(data2), len(data1))
	}
}

func TestMultiSellList_EmptyEntries(t *testing.T) {
	pkt := NewMultiSellList(1, nil, 1)
	pktData, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if pktData[0] != OpcodeMultiSellList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", pktData[0], OpcodeMultiSellList)
	}

	// 1 opcode + 4*5 header = 21
	if len(pktData) != 21 {
		t.Errorf("packet size = %d, want 21", len(pktData))
	}
}
