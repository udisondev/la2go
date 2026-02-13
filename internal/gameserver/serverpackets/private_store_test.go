package serverpackets

import (
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func TestPrivateStoreListSell_Write(t *testing.T) {
	items := []*model.TradeItem{
		{ObjectID: 1001, ItemID: 57, Count: 10, Price: 1000, Enchant: 3, Type2: 5, BodyPart: 0},
		{ObjectID: 1002, ItemID: 100, Count: 1, Price: 50000, Enchant: 0, Type2: 1, BodyPart: 64},
	}

	pkt := &PrivateStoreListSell{
		SellerObjectID: 42,
		PackageSale:    false,
		BuyerAdena:     500000,
		Items:          items,
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if data[0] != OpcodePrivateStoreListSell {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePrivateStoreListSell)
	}

	sellerObjID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if sellerObjID != 42 {
		t.Errorf("sellerObjectID = %d, want 42", sellerObjID)
	}

	pkgSale := int32(binary.LittleEndian.Uint32(data[5:9]))
	if pkgSale != 0 {
		t.Errorf("packageSale = %d, want 0", pkgSale)
	}

	buyerAdena := int32(binary.LittleEndian.Uint32(data[9:13]))
	if buyerAdena != 500000 {
		t.Errorf("buyerAdena = %d, want 500000", buyerAdena)
	}

	itemCount := int32(binary.LittleEndian.Uint32(data[13:17]))
	if itemCount != 2 {
		t.Errorf("itemCount = %d, want 2", itemCount)
	}
}

func TestPrivateStoreListSell_PackageSale(t *testing.T) {
	pkt := &PrivateStoreListSell{
		SellerObjectID: 1,
		PackageSale:    true,
		BuyerAdena:     0,
		Items:          nil,
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	pkgSale := int32(binary.LittleEndian.Uint32(data[5:9]))
	if pkgSale != 1 {
		t.Errorf("packageSale = %d, want 1", pkgSale)
	}
}

func TestPrivateStoreListBuy_Write(t *testing.T) {
	items := []*model.TradeItem{
		{ObjectID: 0, ItemID: 2381, Count: 50, StoreCount: 100, Price: 500, Type2: 5, BodyPart: 0},
	}

	pkt := &PrivateStoreListBuy{
		BuyerObjectID: 99,
		SellerAdena:   1000000,
		Items:         items,
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if data[0] != OpcodePrivateStoreListBuy {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePrivateStoreListBuy)
	}

	buyerObjID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if buyerObjID != 99 {
		t.Errorf("buyerObjectID = %d, want 99", buyerObjID)
	}
}

func TestPrivateStoreMsgSell_Write(t *testing.T) {
	pkt := &PrivateStoreMsgSell{
		ObjectID: 42,
		Message:  "WTS Swords",
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if data[0] != OpcodePrivateStoreMsgSell {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePrivateStoreMsgSell)
	}

	objID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if objID != 42 {
		t.Errorf("objectID = %d, want 42", objID)
	}

	// Message starts at offset 5, UTF-16LE
	// 'W' = 0x57, 0x00
	if data[5] != 0x57 || data[6] != 0x00 {
		t.Errorf("first char = 0x%02X%02X, want 0x5700 ('W')", data[5], data[6])
	}
}

func TestPrivateStoreMsgBuy_Write(t *testing.T) {
	pkt := &PrivateStoreMsgBuy{
		ObjectID: 99,
		Message:  "WTB Arrows",
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if data[0] != OpcodePrivateStoreMsgBuy {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePrivateStoreMsgBuy)
	}
}

func TestPrivateStoreManageListSell_Write(t *testing.T) {
	tmpl := &model.ItemTemplate{
		ItemID: 57,
		Name:   "Adena",
		Type:   model.ItemTypeEtcItem,
	}

	item, _ := model.NewItem(1001, 57, 100, 10, tmpl)

	pkt := &PrivateStoreManageListSell{
		ObjectID:      42,
		PackageSale:   false,
		PlayerAdena:   100000,
		SellableItems: []*model.Item{item},
		StoreItems:    nil,
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if data[0] != OpcodePrivateStoreManageListSell {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePrivateStoreManageListSell)
	}
}

func TestPrivateStoreBuyManageList_Write(t *testing.T) {
	pkt := &PrivateStoreBuyManageList{
		ObjectID:    99,
		PlayerAdena: 50000,
		StoreItems: []*model.TradeItem{
			{ItemID: 2381, Count: 100, Price: 500, Type2: 5},
		},
	}

	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if data[0] != OpcodePrivateStoreBuyManageList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodePrivateStoreBuyManageList)
	}

	objID := int32(binary.LittleEndian.Uint32(data[1:5]))
	if objID != 99 {
		t.Errorf("objectID = %d, want 99", objID)
	}
}
