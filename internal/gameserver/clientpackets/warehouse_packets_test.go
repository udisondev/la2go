package clientpackets

import (
	"encoding/binary"
	"testing"
)

func TestParseSendWareHouseDepositList(t *testing.T) {
	// Build packet: count=2, item1(objID=100, count=5), item2(objID=200, count=3)
	data := make([]byte, 4+2*8)
	binary.LittleEndian.PutUint32(data[0:], 2)    // count
	binary.LittleEndian.PutUint32(data[4:], 100)   // item1 objectID
	binary.LittleEndian.PutUint32(data[8:], 5)     // item1 count
	binary.LittleEndian.PutUint32(data[12:], 200)  // item2 objectID
	binary.LittleEndian.PutUint32(data[16:], 3)    // item2 count

	pkt, err := ParseSendWareHouseDepositList(data)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(pkt.Items) != 2 {
		t.Fatalf("items count = %d, want 2", len(pkt.Items))
	}

	if pkt.Items[0].ObjectID != 100 || pkt.Items[0].Count != 5 {
		t.Errorf("item[0] = {%d, %d}, want {100, 5}", pkt.Items[0].ObjectID, pkt.Items[0].Count)
	}
	if pkt.Items[1].ObjectID != 200 || pkt.Items[1].Count != 3 {
		t.Errorf("item[1] = {%d, %d}, want {200, 3}", pkt.Items[1].ObjectID, pkt.Items[1].Count)
	}
}

func TestParseSendWareHouseDepositList_InvalidCount(t *testing.T) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data[0:], 0) // count=0

	_, err := ParseSendWareHouseDepositList(data)
	if err == nil {
		t.Error("Parse() should fail with count=0")
	}
}

func TestParseSendWareHouseDepositList_ZeroQuantity(t *testing.T) {
	data := make([]byte, 4+8)
	binary.LittleEndian.PutUint32(data[0:], 1) // count=1
	binary.LittleEndian.PutUint32(data[4:], 100) // objectID
	binary.LittleEndian.PutUint32(data[8:], 0)   // count=0 (invalid)

	_, err := ParseSendWareHouseDepositList(data)
	if err == nil {
		t.Error("Parse() should fail with item count=0")
	}
}

func TestParseSendWareHouseWithDrawList(t *testing.T) {
	// Build packet: count=1, item(objID=500, count=10)
	data := make([]byte, 4+8)
	binary.LittleEndian.PutUint32(data[0:], 1)   // count
	binary.LittleEndian.PutUint32(data[4:], 500)  // objectID
	binary.LittleEndian.PutUint32(data[8:], 10)   // count

	pkt, err := ParseSendWareHouseWithDrawList(data)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(pkt.Items) != 1 {
		t.Fatalf("items count = %d, want 1", len(pkt.Items))
	}
	if pkt.Items[0].ObjectID != 500 || pkt.Items[0].Count != 10 {
		t.Errorf("item[0] = {%d, %d}, want {500, 10}", pkt.Items[0].ObjectID, pkt.Items[0].Count)
	}
}

func TestParseMultiSellChoose(t *testing.T) {
	data := make([]byte, 12)
	binary.LittleEndian.PutUint32(data[0:], 42)  // listId
	binary.LittleEndian.PutUint32(data[4:], 3)   // entryId
	binary.LittleEndian.PutUint32(data[8:], 5)   // amount

	pkt, err := ParseMultiSellChoose(data)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if pkt.ListID != 42 {
		t.Errorf("ListID = %d, want 42", pkt.ListID)
	}
	if pkt.EntryID != 3 {
		t.Errorf("EntryID = %d, want 3", pkt.EntryID)
	}
	if pkt.Amount != 5 {
		t.Errorf("Amount = %d, want 5", pkt.Amount)
	}
}

func TestParseMultiSellChoose_InvalidAmount(t *testing.T) {
	data := make([]byte, 12)
	binary.LittleEndian.PutUint32(data[0:], 42)
	binary.LittleEndian.PutUint32(data[4:], 3)
	binary.LittleEndian.PutUint32(data[8:], 0) // amount=0 (invalid)

	_, err := ParseMultiSellChoose(data)
	if err == nil {
		t.Error("Parse() should fail with amount=0")
	}
}

func TestParseMultiSellChoose_TooHighAmount(t *testing.T) {
	data := make([]byte, 12)
	binary.LittleEndian.PutUint32(data[0:], 42)
	binary.LittleEndian.PutUint32(data[4:], 3)
	binary.LittleEndian.PutUint32(data[8:], 10000) // amount>5000 (invalid)

	_, err := ParseMultiSellChoose(data)
	if err == nil {
		t.Error("Parse() should fail with amount>5000")
	}
}
