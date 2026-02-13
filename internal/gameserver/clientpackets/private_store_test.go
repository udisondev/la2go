package clientpackets

import (
	"encoding/binary"
	"testing"
)

func TestParseSetPrivateStoreListSell(t *testing.T) {
	// Build packet: packageSale=0, count=2, items: {objID=100, count=5, price=1000}, {objID=200, count=3, price=500}
	data := make([]byte, 4+4+12*2) // 4+4+24 = 32 bytes
	binary.LittleEndian.PutUint32(data[0:], 0)      // packageSale=false
	binary.LittleEndian.PutUint32(data[4:], 2)      // count=2
	binary.LittleEndian.PutUint32(data[8:], 100)    // item[0].objectID
	binary.LittleEndian.PutUint32(data[12:], 5)     // item[0].count
	binary.LittleEndian.PutUint32(data[16:], 1000)  // item[0].price
	binary.LittleEndian.PutUint32(data[20:], 200)   // item[1].objectID
	binary.LittleEndian.PutUint32(data[24:], 3)     // item[1].count
	binary.LittleEndian.PutUint32(data[28:], 500)   // item[1].price

	pkt, err := ParseSetPrivateStoreListSell(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if pkt.PackageSale {
		t.Error("PackageSale should be false")
	}
	if len(pkt.Items) != 2 {
		t.Fatalf("Items count = %d, want 2", len(pkt.Items))
	}
	if pkt.Items[0].ObjectID != 100 || pkt.Items[0].Count != 5 || pkt.Items[0].Price != 1000 {
		t.Errorf("item[0] = %+v, want {100, 5, 1000}", pkt.Items[0])
	}
	if pkt.Items[1].ObjectID != 200 || pkt.Items[1].Count != 3 || pkt.Items[1].Price != 500 {
		t.Errorf("item[1] = %+v, want {200, 3, 500}", pkt.Items[1])
	}
}

func TestParseSetPrivateStoreListSell_PackageSale(t *testing.T) {
	data := make([]byte, 4+4+12) // 1 item
	binary.LittleEndian.PutUint32(data[0:], 1)   // packageSale=true
	binary.LittleEndian.PutUint32(data[4:], 1)   // count=1
	binary.LittleEndian.PutUint32(data[8:], 50)  // objectID
	binary.LittleEndian.PutUint32(data[12:], 1)  // count
	binary.LittleEndian.PutUint32(data[16:], 99) // price

	pkt, err := ParseSetPrivateStoreListSell(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if !pkt.PackageSale {
		t.Error("PackageSale should be true")
	}
}

func TestParseSetPrivateStoreListSell_EmptyList(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data[0:], 0) // packageSale
	binary.LittleEndian.PutUint32(data[4:], 0) // count=0

	pkt, err := ParseSetPrivateStoreListSell(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(pkt.Items) != 0 {
		t.Errorf("Items count = %d, want 0", len(pkt.Items))
	}
}

func TestParseSetPrivateStoreListSell_InvalidCount(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data[0:], 0)
	// count > 100
	binary.LittleEndian.PutUint32(data[4:], 101)

	_, err := ParseSetPrivateStoreListSell(data)
	if err == nil {
		t.Error("expected error for count > 100")
	}
}

func TestParseSetPrivateStoreListBuy(t *testing.T) {
	// Each item: itemID(4) + enchant(2) + reserved(2) + count(4) + price(4) = 16 bytes
	data := make([]byte, 4+16*2)
	binary.LittleEndian.PutUint32(data[0:], 2)     // count=2
	binary.LittleEndian.PutUint32(data[4:], 57)    // item[0].itemID (Adena)
	binary.LittleEndian.PutUint16(data[8:], 0)     // item[0].enchant
	binary.LittleEndian.PutUint16(data[10:], 0)    // item[0].reserved
	binary.LittleEndian.PutUint32(data[12:], 100)  // item[0].count
	binary.LittleEndian.PutUint32(data[16:], 1)    // item[0].price
	binary.LittleEndian.PutUint32(data[20:], 2381) // item[1].itemID
	binary.LittleEndian.PutUint16(data[24:], 0)    // item[1].enchant
	binary.LittleEndian.PutUint16(data[26:], 0)    // item[1].reserved
	binary.LittleEndian.PutUint32(data[28:], 50)   // item[1].count
	binary.LittleEndian.PutUint32(data[32:], 5000) // item[1].price

	pkt, err := ParseSetPrivateStoreListBuy(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(pkt.Items) != 2 {
		t.Fatalf("Items count = %d, want 2", len(pkt.Items))
	}
	if pkt.Items[0].ItemID != 57 {
		t.Errorf("item[0].ItemID = %d, want 57", pkt.Items[0].ItemID)
	}
	if pkt.Items[0].Count != 100 {
		t.Errorf("item[0].Count = %d, want 100", pkt.Items[0].Count)
	}
	if pkt.Items[1].Price != 5000 {
		t.Errorf("item[1].Price = %d, want 5000", pkt.Items[1].Price)
	}
}

func TestParseRequestPrivateStoreBuy(t *testing.T) {
	data := make([]byte, 4+4+12*1)
	binary.LittleEndian.PutUint32(data[0:], 42)   // storePlayerID
	binary.LittleEndian.PutUint32(data[4:], 1)    // count
	binary.LittleEndian.PutUint32(data[8:], 777)  // objectID
	binary.LittleEndian.PutUint32(data[12:], 10)  // count
	binary.LittleEndian.PutUint32(data[16:], 500) // price

	pkt, err := ParseRequestPrivateStoreBuy(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if pkt.StorePlayerID != 42 {
		t.Errorf("StorePlayerID = %d, want 42", pkt.StorePlayerID)
	}
	if len(pkt.Items) != 1 {
		t.Fatalf("Items count = %d, want 1", len(pkt.Items))
	}
	if pkt.Items[0].ObjectID != 777 {
		t.Errorf("item[0].ObjectID = %d, want 777", pkt.Items[0].ObjectID)
	}
}

func TestParseRequestPrivateStoreSell(t *testing.T) {
	// storePlayerID(4) + count(4) + items(20 each)
	data := make([]byte, 4+4+20)
	binary.LittleEndian.PutUint32(data[0:], 99)   // storePlayerID
	binary.LittleEndian.PutUint32(data[4:], 1)    // count
	binary.LittleEndian.PutUint32(data[8:], 555)  // objectID
	binary.LittleEndian.PutUint32(data[12:], 100) // itemID
	// 2 reserved shorts (4 bytes)
	binary.LittleEndian.PutUint16(data[16:], 0)
	binary.LittleEndian.PutUint16(data[18:], 0)
	binary.LittleEndian.PutUint32(data[20:], 25) // count
	binary.LittleEndian.PutUint32(data[24:], 300) // price

	pkt, err := ParseRequestPrivateStoreSell(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if pkt.StorePlayerID != 99 {
		t.Errorf("StorePlayerID = %d, want 99", pkt.StorePlayerID)
	}
	if pkt.Items[0].ObjectID != 555 || pkt.Items[0].ItemID != 100 {
		t.Errorf("item = %+v", pkt.Items[0])
	}
	if pkt.Items[0].Count != 25 || pkt.Items[0].Price != 300 {
		t.Errorf("item count/price = %d/%d, want 25/300", pkt.Items[0].Count, pkt.Items[0].Price)
	}
}

func TestParseSetPrivateStoreMsgSell(t *testing.T) {
	// UTF-16LE "Hi" + null terminator
	msg := []byte{0x48, 0x00, 0x69, 0x00, 0x00, 0x00} // "Hi\0"

	pkt, err := ParseSetPrivateStoreMsgSell(msg)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if pkt.Message != "Hi" {
		t.Errorf("Message = %q, want %q", pkt.Message, "Hi")
	}
}

func TestParseSetPrivateStoreMsgBuy(t *testing.T) {
	msg := []byte{0x42, 0x00, 0x75, 0x00, 0x79, 0x00, 0x00, 0x00} // "Buy\0"

	pkt, err := ParseSetPrivateStoreMsgBuy(msg)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if pkt.Message != "Buy" {
		t.Errorf("Message = %q, want %q", pkt.Message, "Buy")
	}
}
