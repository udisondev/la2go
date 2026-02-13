package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func TestWareHouseDepositList_Write(t *testing.T) {
	tmpl := &model.ItemTemplate{ItemID: 100, Name: "Sword", Type: model.ItemTypeWeapon}
	item, err := model.NewItem(50001, 100, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}

	pkt := NewWareHouseDepositList(WarehouseTypePrivate, 10000, []*model.Item{item})
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// Check opcode
	if data[0] != OpcodeWareHouseDepositList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeWareHouseDepositList)
	}

	// Check minimum size: 1 opcode + 2 whType + 4 adena + 2 count + 38 per item = 47
	if len(data) < 47 {
		t.Errorf("packet size = %d, want >= 47", len(data))
	}
}

func TestWareHouseDepositList_EmptyItems(t *testing.T) {
	pkt := NewWareHouseDepositList(WarehouseTypePrivate, 5000, nil)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != OpcodeWareHouseDepositList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeWareHouseDepositList)
	}

	// 1 opcode + 2 whType + 4 adena + 2 count = 9
	if len(data) != 9 {
		t.Errorf("packet size = %d, want 9", len(data))
	}
}

func TestWareHouseWithdrawalList_Write(t *testing.T) {
	tmpl := &model.ItemTemplate{ItemID: 200, Name: "Armor", Type: model.ItemTypeArmor, BodyPart: model.ArmorSlotChest}
	item, err := model.NewItem(50002, 200, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}

	pkt := NewWareHouseWithdrawalList(WarehouseTypePrivate, 8000, []*model.Item{item})
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != OpcodeWareHouseWithdrawalList {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], OpcodeWareHouseWithdrawalList)
	}

	if len(data) < 47 {
		t.Errorf("packet size = %d, want >= 47", len(data))
	}
}
