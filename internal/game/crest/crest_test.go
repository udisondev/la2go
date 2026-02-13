package crest

import (
	"testing"
)

func TestCreateCrestPledge(t *testing.T) {
	tbl := NewTable()
	data := make([]byte, 128)
	c, err := tbl.CreateCrest(data, Pledge)
	if err != nil {
		t.Fatalf("CreateCrest(Pledge) error: %v", err)
	}
	if c.Type() != Pledge {
		t.Errorf("Type() = %d; want %d", c.Type(), Pledge)
	}
	if len(c.Data()) != 128 {
		t.Errorf("Data() len = %d; want 128", len(c.Data()))
	}
}

func TestCreateCrestPledgeLarge(t *testing.T) {
	tbl := NewTable()
	data := make([]byte, 2000)
	c, err := tbl.CreateCrest(data, PledgeLarge)
	if err != nil {
		t.Fatalf("CreateCrest(PledgeLarge) error: %v", err)
	}
	if c.Type() != PledgeLarge {
		t.Errorf("Type() = %d; want %d", c.Type(), PledgeLarge)
	}
}

func TestCreateCrestAlly(t *testing.T) {
	tbl := NewTable()
	data := make([]byte, 100)
	c, err := tbl.CreateCrest(data, Ally)
	if err != nil {
		t.Fatalf("CreateCrest(Ally) error: %v", err)
	}
	if c.Type() != Ally {
		t.Errorf("Type() = %d; want %d", c.Type(), Ally)
	}
}

func TestCreateCrestPledgeTooLarge(t *testing.T) {
	tbl := NewTable()
	data := make([]byte, MaxPledgeSize+1)
	_, err := tbl.CreateCrest(data, Pledge)
	if err == nil {
		t.Fatal("CreateCrest(Pledge) expected error for oversized data")
	}
}

func TestCreateCrestPledgeLargeTooLarge(t *testing.T) {
	tbl := NewTable()
	data := make([]byte, MaxPledgeLargeSize+1)
	_, err := tbl.CreateCrest(data, PledgeLarge)
	if err == nil {
		t.Fatal("CreateCrest(PledgeLarge) expected error for oversized data")
	}
}

func TestCreateCrestAllyTooLarge(t *testing.T) {
	tbl := NewTable()
	data := make([]byte, MaxAllySize+1)
	_, err := tbl.CreateCrest(data, Ally)
	if err == nil {
		t.Fatal("CreateCrest(Ally) expected error for oversized data")
	}
}

func TestCreateCrestInvalidType(t *testing.T) {
	tbl := NewTable()
	_, err := tbl.CreateCrest([]byte{1}, CrestType(99))
	if err == nil {
		t.Fatal("CreateCrest(invalid) expected error")
	}
}

func TestGetCrestExisting(t *testing.T) {
	tbl := NewTable()
	data := []byte{0xDE, 0xAD}
	created, err := tbl.CreateCrest(data, Pledge)
	if err != nil {
		t.Fatalf("CreateCrest error: %v", err)
	}

	got := tbl.Crest(created.ID())
	if got == nil {
		t.Fatal("Crest() returned nil for existing crest")
	}
	if got.ID() != created.ID() {
		t.Errorf("Crest().ID() = %d; want %d", got.ID(), created.ID())
	}
}

func TestGetCrestNonExisting(t *testing.T) {
	tbl := NewTable()
	got := tbl.Crest(999)
	if got != nil {
		t.Errorf("Crest(999) = %v; want nil", got)
	}
}

func TestRemoveCrest(t *testing.T) {
	tbl := NewTable()
	c, err := tbl.CreateCrest([]byte{1, 2, 3}, Pledge)
	if err != nil {
		t.Fatalf("CreateCrest error: %v", err)
	}

	tbl.RemoveCrest(c.ID())

	got := tbl.Crest(c.ID())
	if got != nil {
		t.Errorf("Crest(%d) after remove = %v; want nil", c.ID(), got)
	}
}

func TestInitFromRows(t *testing.T) {
	tbl := NewTable()
	rows := []CrestRow{
		{CrestID: 10, Data: []byte{1}, Type: int32(Pledge)},
		{CrestID: 20, Data: []byte{2, 3}, Type: int32(Ally)},
		{CrestID: 15, Data: []byte{4, 5, 6}, Type: int32(PledgeLarge)},
	}
	tbl.Init(rows)

	// Все три герба загружены.
	for _, row := range rows {
		c := tbl.Crest(row.CrestID)
		if c == nil {
			t.Errorf("Crest(%d) = nil after Init", row.CrestID)
			continue
		}
		if c.Type() != CrestType(row.Type) {
			t.Errorf("Crest(%d).Type() = %d; want %d", row.CrestID, c.Type(), row.Type)
		}
		if len(c.Data()) != len(row.Data) {
			t.Errorf("Crest(%d).Data() len = %d; want %d", row.CrestID, len(c.Data()), len(row.Data))
		}
	}

	// nextID должен быть max(IDs) + 1 = 21.
	next, err := tbl.CreateCrest([]byte{0xFF}, Pledge)
	if err != nil {
		t.Fatalf("CreateCrest after Init error: %v", err)
	}
	if next.ID() != 21 {
		t.Errorf("ID after Init = %d; want 21", next.ID())
	}
}

func TestAtomicIDGeneration(t *testing.T) {
	tbl := NewTable()

	ids := make(map[int32]struct{}, 100)
	for range 100 {
		c, err := tbl.CreateCrest([]byte{1}, Pledge)
		if err != nil {
			t.Fatalf("CreateCrest error: %v", err)
		}
		if _, dup := ids[c.ID()]; dup {
			t.Fatalf("duplicate crest ID: %d", c.ID())
		}
		ids[c.ID()] = struct{}{}
	}

	if len(ids) != 100 {
		t.Errorf("unique IDs = %d; want 100", len(ids))
	}
}

func TestCreateCrestMaxBoundary(t *testing.T) {
	tests := []struct {
		name string
		typ  CrestType
		size int
	}{
		{"Pledge exact max", Pledge, MaxPledgeSize},
		{"PledgeLarge exact max", PledgeLarge, MaxPledgeLargeSize},
		{"Ally exact max", Ally, MaxAllySize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := NewTable()
			data := make([]byte, tt.size)
			c, err := tbl.CreateCrest(data, tt.typ)
			if err != nil {
				t.Fatalf("CreateCrest(%d bytes, type %d) error: %v", tt.size, tt.typ, err)
			}
			if len(c.Data()) != tt.size {
				t.Errorf("Data() len = %d; want %d", len(c.Data()), tt.size)
			}
		})
	}
}

func TestRemoveNonExisting(t *testing.T) {
	tbl := NewTable()
	// Не должно паниковать.
	tbl.RemoveCrest(999)
}
