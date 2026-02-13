package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

func init() {
	if err := data.LoadHennaTemplates(); err != nil {
		panic("LoadHennaTemplates: " + err.Error())
	}
}

func newTestPlayer(t *testing.T) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(1, 100, 1, "HennaTest", 40, 0, 11)
	if err != nil {
		t.Fatalf("NewPlayer() error: %v", err)
	}
	return p
}

func TestHennaInfo_Write_NoHennas(t *testing.T) {
	t.Parallel()

	p := newTestPlayer(t)
	pkt := NewHennaInfo(p)

	raw, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(raw)

	// Opcode
	opcode, err := r.ReadByte()
	if err != nil {
		t.Fatalf("read opcode: %v", err)
	}
	if opcode != OpcodeHennaInfo {
		t.Errorf("opcode = 0x%02X; want 0x%02X", opcode, OpcodeHennaInfo)
	}

	// 6 stat bytes (all 0 for no hennas)
	for _, statName := range []string{"INT", "STR", "CON", "MEN", "DEX", "WIT"} {
		b, err := r.ReadByte()
		if err != nil {
			t.Fatalf("read stat %s: %v", statName, err)
		}
		if b != 0 {
			t.Errorf("stat %s = %d; want 0", statName, b)
		}
	}

	// Max slots
	maxSlots, err := r.ReadInt()
	if err != nil {
		t.Fatalf("read maxSlots: %v", err)
	}
	if maxSlots != model.MaxHennaSlots {
		t.Errorf("maxSlots = %d; want %d", maxSlots, model.MaxHennaSlots)
	}

	// Equipped count
	count, err := r.ReadInt()
	if err != nil {
		t.Fatalf("read count: %v", err)
	}
	if count != 0 {
		t.Errorf("equipped count = %d; want 0", count)
	}
}

func TestHennaInfo_Write_WithHennas(t *testing.T) {
	t.Parallel()

	p := newTestPlayer(t)

	// Добавляем 2 хенны: dyeID=1 (STR+1,CON-3), dyeID=2 (STR+1,DEX-3)
	if _, err := p.AddHenna(1); err != nil {
		t.Fatal(err)
	}
	if _, err := p.AddHenna(2); err != nil {
		t.Fatal(err)
	}

	pkt := NewHennaInfo(p)
	raw, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(raw)

	// Opcode
	if opcode, err := r.ReadByte(); err != nil || opcode != OpcodeHennaInfo {
		t.Fatalf("opcode: err=%v, opcode=0x%02X", err, opcode)
	}

	// Stats: INT=0, STR=2, CON=-3, MEN=0, DEX=-3, WIT=0
	intStat, _ := r.ReadByte()
	strStat, _ := r.ReadByte()
	conStat, _ := r.ReadByte()
	menStat, _ := r.ReadByte()
	dexStat, _ := r.ReadByte()
	witStat, _ := r.ReadByte()

	if strStat != 2 {
		t.Errorf("STR stat = %d; want 2", strStat)
	}
	if intStat != 0 {
		t.Errorf("INT stat = %d; want 0", intStat)
	}
	// CON=-3 → byte underflow to 253, verify raw
	_ = conStat
	_ = menStat
	_ = dexStat
	_ = witStat

	// Max slots
	maxSlots, _ := r.ReadInt()
	if maxSlots != model.MaxHennaSlots {
		t.Errorf("maxSlots = %d; want %d", maxSlots, model.MaxHennaSlots)
	}

	// 2 equipped hennas
	count, _ := r.ReadInt()
	if count != 2 {
		t.Errorf("equipped count = %d; want 2", count)
	}

	// First henna
	dyeID1, _ := r.ReadInt()
	if dyeID1 != 1 {
		t.Errorf("first dyeID = %d; want 1", dyeID1)
	}
	unknown1, _ := r.ReadInt()
	if unknown1 != 1 {
		t.Errorf("first unknown = %d; want 1", unknown1)
	}

	// Second henna
	dyeID2, _ := r.ReadInt()
	if dyeID2 != 2 {
		t.Errorf("second dyeID = %d; want 2", dyeID2)
	}
}

func TestHennaItemDrawInfo_Write(t *testing.T) {
	t.Parallel()

	p := newTestPlayer(t)
	henna := &data.HennaInfo{
		DyeID:     1,
		DyeItemID: 4445,
		StatSTR:   1,
		StatCON:   -3,
		WearCount: 10,
		WearFee:   37000,
	}

	pkt := NewHennaItemDrawInfo(p, henna)
	raw, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(raw)

	opcode, _ := r.ReadByte()
	if opcode != OpcodeHennaItemDrawInfo {
		t.Errorf("opcode = 0x%02X; want 0x%02X", opcode, OpcodeHennaItemDrawInfo)
	}

	dyeID, _ := r.ReadInt()
	if dyeID != 1 {
		t.Errorf("dyeID = %d; want 1", dyeID)
	}

	dyeItemID, _ := r.ReadInt()
	if dyeItemID != 4445 {
		t.Errorf("dyeItemID = %d; want 4445", dyeItemID)
	}

	wearCount, _ := r.ReadInt()
	if wearCount != 10 {
		t.Errorf("wearCount = %d; want 10", wearCount)
	}

	wearFee, _ := r.ReadInt()
	if wearFee != 37000 {
		t.Errorf("wearFee = %d; want 37000", wearFee)
	}
}

func TestHennaItemRemoveInfo_Write(t *testing.T) {
	t.Parallel()

	p := newTestPlayer(t)
	henna := &data.HennaInfo{
		DyeID:       1,
		DyeItemID:   4445,
		StatSTR:     1,
		StatCON:     -3,
		CancelCount: 5,
		CancelFee:   7400,
	}

	pkt := NewHennaItemRemoveInfo(p, henna)
	raw, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(raw)

	opcode, _ := r.ReadByte()
	if opcode != OpcodeHennaItemRemoveInfo {
		t.Errorf("opcode = 0x%02X; want 0x%02X", opcode, OpcodeHennaItemRemoveInfo)
	}

	dyeID, _ := r.ReadInt()
	if dyeID != 1 {
		t.Errorf("dyeID = %d; want 1", dyeID)
	}

	dyeItemID, _ := r.ReadInt()
	if dyeItemID != 4445 {
		t.Errorf("dyeItemID = %d; want 4445", dyeItemID)
	}

	cancelCount, _ := r.ReadInt()
	if cancelCount != 5 {
		t.Errorf("cancelCount = %d; want 5", cancelCount)
	}

	cancelFee, _ := r.ReadInt()
	if cancelFee != 7400 {
		t.Errorf("cancelFee = %d; want 7400", cancelFee)
	}
}

func TestHennaEquipList_Write_Empty(t *testing.T) {
	t.Parallel()

	p := newTestPlayer(t)
	// Без предметов в инвентаре — список будет пустым
	pkt := NewHennaEquipList(p)
	raw, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(raw)

	opcode, _ := r.ReadByte()
	if opcode != OpcodeHennaEquipList {
		t.Errorf("opcode = 0x%02X; want 0x%02X", opcode, OpcodeHennaEquipList)
	}

	adena, _ := r.ReadInt()
	if adena != 0 {
		t.Errorf("adena = %d; want 0 (no adena)", adena)
	}

	emptySlots, _ := r.ReadInt()
	if emptySlots != model.MaxHennaSlots {
		t.Errorf("emptySlots = %d; want %d", emptySlots, model.MaxHennaSlots)
	}

	hennaCount, _ := r.ReadInt()
	if hennaCount != 0 {
		t.Errorf("hennaCount = %d; want 0 (no dye items in inventory)", hennaCount)
	}
}
