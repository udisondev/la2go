package clientpackets

import (
	"encoding/binary"
	"testing"
)

// --- ParseRequestExMagicSkillUseGround ---

func TestParseRequestExMagicSkillUseGround(t *testing.T) {
	t.Parallel()

	// 4 int32 (x, y, z, skillID) + 1 int32 (ctrl) + 1 byte (shift) = 21 bytes
	data := make([]byte, 21)
	binary.LittleEndian.PutUint32(data[0:], uint32(1000))
	binary.LittleEndian.PutUint32(data[4:], uint32(2000))
	neg100 := int32(-100)
	binary.LittleEndian.PutUint32(data[8:], uint32(neg100)) // -100 as int32
	binary.LittleEndian.PutUint32(data[12:], uint32(261))
	binary.LittleEndian.PutUint32(data[16:], 1) // ctrl = true
	data[20] = 0                                 // shift = false

	pkt, err := ParseRequestExMagicSkillUseGround(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.X != 1000 {
		t.Errorf("X = %d; want 1000", pkt.X)
	}
	if pkt.Y != 2000 {
		t.Errorf("Y = %d; want 2000", pkt.Y)
	}
	if pkt.Z != -100 {
		t.Errorf("Z = %d; want -100", pkt.Z)
	}
	if pkt.SkillID != 261 {
		t.Errorf("SkillID = %d; want 261", pkt.SkillID)
	}
	if !pkt.CtrlPressed {
		t.Error("CtrlPressed = false; want true")
	}
	if pkt.ShiftPressed {
		t.Error("ShiftPressed = true; want false")
	}
}

func TestParseRequestExMagicSkillUseGround_TooShort(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestExMagicSkillUseGround([]byte{0x01, 0x00, 0x00, 0x00})
	if err == nil {
		t.Error("expected error for short data")
	}
}

// --- ParseRequestBuySeed ---

func TestParseRequestBuySeed(t *testing.T) {
	t.Parallel()

	// manorID(4) + count(4) + 2 items * (itemID(4) + count(4)) = 24 bytes
	data := make([]byte, 24)
	binary.LittleEndian.PutUint32(data[0:], 1)     // ManorID = 1
	binary.LittleEndian.PutUint32(data[4:], 2)     // count = 2
	binary.LittleEndian.PutUint32(data[8:], 5016)  // item[0].ItemID
	binary.LittleEndian.PutUint32(data[12:], 10)   // item[0].Count
	binary.LittleEndian.PutUint32(data[16:], 5017) // item[1].ItemID
	binary.LittleEndian.PutUint32(data[20:], 20)   // item[1].Count

	pkt, err := ParseRequestBuySeed(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.ManorID != 1 {
		t.Errorf("ManorID = %d; want 1", pkt.ManorID)
	}
	if len(pkt.Items) != 2 {
		t.Fatalf("len(Items) = %d; want 2", len(pkt.Items))
	}
	if pkt.Items[0].ItemID != 5016 {
		t.Errorf("Items[0].ItemID = %d; want 5016", pkt.Items[0].ItemID)
	}
	if pkt.Items[0].Count != 10 {
		t.Errorf("Items[0].Count = %d; want 10", pkt.Items[0].Count)
	}
	if pkt.Items[1].ItemID != 5017 {
		t.Errorf("Items[1].ItemID = %d; want 5017", pkt.Items[1].ItemID)
	}
	if pkt.Items[1].Count != 20 {
		t.Errorf("Items[1].Count = %d; want 20", pkt.Items[1].Count)
	}
}

func TestParseRequestBuySeed_ZeroItems(t *testing.T) {
	t.Parallel()

	// manorID(4) + count(4) = 8 bytes
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data[0:], 3) // ManorID = 3
	binary.LittleEndian.PutUint32(data[4:], 0) // count = 0

	pkt, err := ParseRequestBuySeed(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.ManorID != 3 {
		t.Errorf("ManorID = %d; want 3", pkt.ManorID)
	}
	if len(pkt.Items) != 0 {
		t.Errorf("len(Items) = %d; want 0", len(pkt.Items))
	}
}

func TestParseRequestBuySeed_TooShort(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestBuySeed([]byte{0x01, 0x00})
	if err == nil {
		t.Error("expected error for short data")
	}
}

// --- ParseRequestProcureCropList ---

func TestParseRequestProcureCropList(t *testing.T) {
	t.Parallel()

	// count(4) + 2 entries * (objectID(4) + itemID(4) + manorID(4) + count(4)) = 36 bytes
	data := make([]byte, 36)
	binary.LittleEndian.PutUint32(data[0:], 2)      // count = 2
	binary.LittleEndian.PutUint32(data[4:], 100001)  // entry[0].ObjectID
	binary.LittleEndian.PutUint32(data[8:], 5501)    // entry[0].ItemID
	binary.LittleEndian.PutUint32(data[12:], 1)      // entry[0].ManorID
	binary.LittleEndian.PutUint32(data[16:], 50)     // entry[0].Count
	binary.LittleEndian.PutUint32(data[20:], 100002) // entry[1].ObjectID
	binary.LittleEndian.PutUint32(data[24:], 5502)   // entry[1].ItemID
	binary.LittleEndian.PutUint32(data[28:], 2)      // entry[1].ManorID
	binary.LittleEndian.PutUint32(data[32:], 30)     // entry[1].Count

	pkt, err := ParseRequestProcureCropList(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkt.Items) != 2 {
		t.Fatalf("len(Items) = %d; want 2", len(pkt.Items))
	}
	if pkt.Items[0].ObjectID != 100001 {
		t.Errorf("Items[0].ObjectID = %d; want 100001", pkt.Items[0].ObjectID)
	}
	if pkt.Items[0].ItemID != 5501 {
		t.Errorf("Items[0].ItemID = %d; want 5501", pkt.Items[0].ItemID)
	}
	if pkt.Items[0].ManorID != 1 {
		t.Errorf("Items[0].ManorID = %d; want 1", pkt.Items[0].ManorID)
	}
	if pkt.Items[0].Count != 50 {
		t.Errorf("Items[0].Count = %d; want 50", pkt.Items[0].Count)
	}
	if pkt.Items[1].ObjectID != 100002 {
		t.Errorf("Items[1].ObjectID = %d; want 100002", pkt.Items[1].ObjectID)
	}
	if pkt.Items[1].ItemID != 5502 {
		t.Errorf("Items[1].ItemID = %d; want 5502", pkt.Items[1].ItemID)
	}
	if pkt.Items[1].ManorID != 2 {
		t.Errorf("Items[1].ManorID = %d; want 2", pkt.Items[1].ManorID)
	}
	if pkt.Items[1].Count != 30 {
		t.Errorf("Items[1].Count = %d; want 30", pkt.Items[1].Count)
	}
}

func TestParseRequestProcureCropList_Empty(t *testing.T) {
	t.Parallel()

	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data[0:], 0) // count = 0

	pkt, err := ParseRequestProcureCropList(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkt.Items) != 0 {
		t.Errorf("len(Items) = %d; want 0", len(pkt.Items))
	}
}

func TestParseRequestProcureCropList_TooShort(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestProcureCropList([]byte{0x01})
	if err == nil {
		t.Error("expected error for short data")
	}
}

// --- ParseRequestExEnchantSkillInfo ---

func TestParseRequestExEnchantSkillInfo(t *testing.T) {
	t.Parallel()

	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data[0:], 56)  // SkillID
	binary.LittleEndian.PutUint32(data[4:], 101) // SkillLevel

	pkt, err := ParseRequestExEnchantSkillInfo(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.SkillID != 56 {
		t.Errorf("SkillID = %d; want 56", pkt.SkillID)
	}
	if pkt.SkillLevel != 101 {
		t.Errorf("SkillLevel = %d; want 101", pkt.SkillLevel)
	}
}

func TestParseRequestExEnchantSkillInfo_TooShort(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestExEnchantSkillInfo([]byte{0x01, 0x00, 0x00, 0x00})
	if err == nil {
		t.Error("expected error for short data")
	}
}

// --- ParseRequestExEnchantSkill ---

func TestParseRequestExEnchantSkill(t *testing.T) {
	t.Parallel()

	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data[0:], 56)  // SkillID
	binary.LittleEndian.PutUint32(data[4:], 101) // SkillLevel

	pkt, err := ParseRequestExEnchantSkill(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.SkillID != 56 {
		t.Errorf("SkillID = %d; want 56", pkt.SkillID)
	}
	if pkt.SkillLevel != 101 {
		t.Errorf("SkillLevel = %d; want 101", pkt.SkillLevel)
	}
}

func TestParseRequestExEnchantSkill_TooShort(t *testing.T) {
	t.Parallel()

	_, err := ParseRequestExEnchantSkill([]byte{0x01, 0x00})
	if err == nil {
		t.Error("expected error for short data")
	}
}
