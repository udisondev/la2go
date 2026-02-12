package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestParseRequestBuyItem(t *testing.T) {
	w := packet.NewWriter(64)
	w.WriteInt(100)         // listID
	w.WriteInt(2)           // item count
	w.WriteInt(57)          // itemID 1 (Adena)
	w.WriteInt(1000)        // count 1
	w.WriteInt(736)         // itemID 2 (Scroll of Escape)
	w.WriteInt(5)           // count 2

	pkt, err := ParseRequestBuyItem(w.Bytes())
	if err != nil {
		t.Fatalf("ParseRequestBuyItem() error: %v", err)
	}

	if pkt.ListID != 100 {
		t.Errorf("ListID = %d, want 100", pkt.ListID)
	}

	if len(pkt.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(pkt.Items))
	}

	if pkt.Items[0].ItemID != 57 {
		t.Errorf("Items[0].ItemID = %d, want 57", pkt.Items[0].ItemID)
	}
	if pkt.Items[0].Count != 1000 {
		t.Errorf("Items[0].Count = %d, want 1000", pkt.Items[0].Count)
	}
	if pkt.Items[1].ItemID != 736 {
		t.Errorf("Items[1].ItemID = %d, want 736", pkt.Items[1].ItemID)
	}
	if pkt.Items[1].Count != 5 {
		t.Errorf("Items[1].Count = %d, want 5", pkt.Items[1].Count)
	}
}

func TestParseRequestBuyItem_InvalidCount(t *testing.T) {
	w := packet.NewWriter(16)
	w.WriteInt(1)   // listID
	w.WriteInt(200) // invalid count (> 100)

	_, err := ParseRequestBuyItem(w.Bytes())
	if err == nil {
		t.Error("expected error for count > 100, got nil")
	}
}

func TestParseRequestBuyItem_ZeroQuantity(t *testing.T) {
	w := packet.NewWriter(32)
	w.WriteInt(1) // listID
	w.WriteInt(1) // 1 item
	w.WriteInt(57) // itemID
	w.WriteInt(0)  // zero quantity

	_, err := ParseRequestBuyItem(w.Bytes())
	if err == nil {
		t.Error("expected error for zero quantity, got nil")
	}
}
