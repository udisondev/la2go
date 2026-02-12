package serverpackets

import (
	"encoding/binary"
	"testing"
)

func TestBuyList_Write(t *testing.T) {
	products := []BuyListProduct{
		{ItemID: 57, Price: 100, Count: -1, Type1: 5, Type2: 5},
		{ItemID: 736, Price: 1000, Count: 10, Type1: 5, Type2: 5},
	}

	pkt := NewBuyList(50000, 7, products)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != OpcodeBuyList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeBuyList)
	}

	// Player Adena
	adena := int32(binary.LittleEndian.Uint32(data[1:5]))
	if adena != 50000 {
		t.Errorf("adena = %d, want 50000", adena)
	}

	// ListID
	listID := int32(binary.LittleEndian.Uint32(data[5:9]))
	if listID != 7 {
		t.Errorf("listID = %d, want 7", listID)
	}

	// Item count
	itemCount := int16(binary.LittleEndian.Uint16(data[9:11]))
	if itemCount != 2 {
		t.Errorf("itemCount = %d, want 2", itemCount)
	}
}

func TestBuyList_Write_Empty(t *testing.T) {
	pkt := NewBuyList(0, 1, nil)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != OpcodeBuyList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeBuyList)
	}

	itemCount := int16(binary.LittleEndian.Uint16(data[9:11]))
	if itemCount != 0 {
		t.Errorf("itemCount = %d, want 0", itemCount)
	}
}
