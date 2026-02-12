package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestParseRequestSellItem(t *testing.T) {
	w := packet.NewWriter(64)
	w.WriteInt(2)      // item count
	w.WriteInt(10001)  // objectID 1
	w.WriteInt(57)     // itemID 1
	w.WriteInt(500)    // count 1
	w.WriteInt(10002)  // objectID 2
	w.WriteInt(736)    // itemID 2
	w.WriteInt(1)      // count 2

	pkt, err := ParseRequestSellItem(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestSellItem() error: %v", err)
	}

	if len(pkt.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(pkt.Items))
	}

	if pkt.Items[0].ObjectID != 10001 {
		t.Errorf("Items[0].ObjectID = %d, want 10001", pkt.Items[0].ObjectID)
	}
	if pkt.Items[0].ItemID != 57 {
		t.Errorf("Items[0].ItemID = %d, want 57", pkt.Items[0].ItemID)
	}
	if pkt.Items[0].Count != 500 {
		t.Errorf("Items[0].Count = %d, want 500", pkt.Items[0].Count)
	}
}

func TestParseRequestSellItem_InvalidCount(t *testing.T) {
	w := packet.NewWriter(16)
	w.WriteInt(-1) // negative count

	_, err := ParseRequestSellItem(w.Bytes())
	if err == nil {
		t.Error("expected error for negative count, got nil")
	}
}

func TestParseRequestSellItem_ZeroQuantity(t *testing.T) {
	w := packet.NewWriter(32)
	w.WriteInt(1)      // 1 item
	w.WriteInt(10001)  // objectID
	w.WriteInt(57)     // itemID
	w.WriteInt(0)      // zero quantity

	_, err := ParseRequestSellItem(w.Bytes())
	if err == nil {
		t.Error("expected error for zero quantity, got nil")
	}
}
