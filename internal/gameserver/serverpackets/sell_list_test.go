package serverpackets

import (
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func TestSellList_Write(t *testing.T) {
	tmpl := &model.ItemTemplate{
		ItemID: 100,
		Name:   "Short Sword",
		Type:   model.ItemTypeWeapon,
		PAtk:   10,
	}
	item, err := model.NewItem(50001, 100, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}

	items := []SellListItem{
		{Item: item, SellPrice: 500},
	}

	pkt := NewSellList(10000, items)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != OpcodeSellList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSellList)
	}

	// Player Adena
	adena := int32(binary.LittleEndian.Uint32(data[1:5]))
	if adena != 10000 {
		t.Errorf("adena = %d, want 10000", adena)
	}

	// Item count
	itemCount := int16(binary.LittleEndian.Uint16(data[5:7]))
	if itemCount != 1 {
		t.Errorf("itemCount = %d, want 1", itemCount)
	}
}

func TestSellList_Write_Empty(t *testing.T) {
	pkt := NewSellList(0, nil)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != OpcodeSellList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeSellList)
	}

	itemCount := int16(binary.LittleEndian.Uint16(data[5:7]))
	if itemCount != 0 {
		t.Errorf("itemCount = %d, want 0", itemCount)
	}
}
