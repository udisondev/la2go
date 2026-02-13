package zone

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// newTestCharacter creates a Character at the given position for testing.
func newTestCharacter(objectID uint32, x, y, z int32) *model.Character {
	loc := model.NewLocation(x, y, z, 0)
	return model.NewCharacter(objectID, "test", loc, 1, 1000, 500, 800)
}

func TestZoneIDSetAndCheck(t *testing.T) {
	ch := newTestCharacter(1, 0, 0, 0)

	// Initially no zones set
	if ch.IsInsideZone(model.ZoneIDPeace) {
		t.Fatal("expected not in PEACE zone initially")
	}

	// Set PEACE
	ch.SetInsideZone(model.ZoneIDPeace, true)
	if !ch.IsInsideZone(model.ZoneIDPeace) {
		t.Fatal("expected in PEACE zone after SetInsideZone(true)")
	}

	// Set WATER too
	ch.SetInsideZone(model.ZoneIDWater, true)
	if !ch.IsInsideZone(model.ZoneIDWater) {
		t.Fatal("expected in WATER zone")
	}
	if !ch.IsInsideZone(model.ZoneIDPeace) {
		t.Fatal("PEACE should still be set")
	}

	// Clear PEACE
	ch.SetInsideZone(model.ZoneIDPeace, false)
	if ch.IsInsideZone(model.ZoneIDPeace) {
		t.Fatal("expected not in PEACE zone after clear")
	}
	if !ch.IsInsideZone(model.ZoneIDWater) {
		t.Fatal("WATER should still be set")
	}
}

func TestClearAllZoneFlags(t *testing.T) {
	ch := newTestCharacter(1, 0, 0, 0)
	ch.SetInsideZone(model.ZoneIDPeace, true)
	ch.SetInsideZone(model.ZoneIDTown, true)
	ch.SetInsideZone(model.ZoneIDWater, true)

	ch.ClearAllZoneFlags()

	if ch.IsInsideZone(model.ZoneIDPeace) || ch.IsInsideZone(model.ZoneIDTown) || ch.IsInsideZone(model.ZoneIDWater) {
		t.Fatal("expected all zone flags cleared")
	}
}

func TestAllZoneIDs(t *testing.T) {
	ch := newTestCharacter(1, 0, 0, 0)

	// Set all 22 zones
	for i := model.ZoneID(0); i < model.ZoneIDCount; i++ {
		ch.SetInsideZone(i, true)
	}

	// Verify all set
	for i := model.ZoneID(0); i < model.ZoneIDCount; i++ {
		if !ch.IsInsideZone(i) {
			t.Fatalf("ZoneID %d should be set", i)
		}
	}

	// Clear odd zones
	for i := model.ZoneID(1); i < model.ZoneIDCount; i += 2 {
		ch.SetInsideZone(i, false)
	}

	// Verify even still set, odd cleared
	for i := model.ZoneID(0); i < model.ZoneIDCount; i++ {
		expected := i%2 == 0
		if ch.IsInsideZone(i) != expected {
			t.Fatalf("ZoneID %d: expected %v, got %v", i, expected, !expected)
		}
	}
}

func TestPeaceZoneOnEnterOnExit(t *testing.T) {
	base := &BaseZone{
		id:       100,
		name:     "TestPeace",
		zoneType: TypePeace,
		shape:    "Cuboid",
		minZ:     -1000,
		maxZ:     1000,
		nodesX:   []int32{0, 100},
		nodesY:   []int32{0, 100},
	}
	z := NewPeaceZone(base)

	ch := newTestCharacter(1, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDPeace) {
		t.Fatal("expected PEACE flag set on enter")
	}
	if z.CharacterCount() != 1 {
		t.Fatalf("expected 1 character tracked, got %d", z.CharacterCount())
	}

	// Move character outside
	ch.SetLocation(model.NewLocation(200, 200, 0, 0))
	z.RevalidateInZone(ch)

	if ch.IsInsideZone(model.ZoneIDPeace) {
		t.Fatal("expected PEACE flag cleared on exit")
	}
	if z.CharacterCount() != 0 {
		t.Fatalf("expected 0 characters tracked, got %d", z.CharacterCount())
	}
}

func TestTownZoneOnEnterOnExit(t *testing.T) {
	base := &BaseZone{
		id:       101,
		name:     "TestTown",
		zoneType: TypeTown,
		shape:    "Cuboid",
		minZ:     -1000,
		maxZ:     1000,
		nodesX:   []int32{0, 100},
		nodesY:   []int32{0, 100},
	}
	z := NewTownZone(base)

	ch := newTestCharacter(1, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDTown) {
		t.Fatal("expected TOWN flag set on enter")
	}
	if ch.IsInsideZone(model.ZoneIDPeace) {
		t.Fatal("expected PEACE flag NOT set (TownZone does not set PEACE)")
	}

	ch.SetLocation(model.NewLocation(200, 200, 0, 0))
	z.RevalidateInZone(ch)

	if ch.IsInsideZone(model.ZoneIDTown) {
		t.Fatal("expected TOWN flag cleared on exit")
	}
}

func TestCastleZoneOnEnterOnExit(t *testing.T) {
	base := &BaseZone{
		id:       200,
		name:     "TestCastle",
		zoneType: TypeCastle,
		shape:    "Cuboid",
		minZ:     -1000,
		maxZ:     1000,
		nodesX:   []int32{0, 100},
		nodesY:   []int32{0, 100},
	}
	z := NewCastleZone(base)

	ch := newTestCharacter(1, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDCastle) {
		t.Fatal("expected CASTLE flag set")
	}
	if ch.IsInsideZone(model.ZoneIDPeace) {
		t.Fatal("expected PEACE flag NOT set (CastleZone does not set PEACE)")
	}
	if ch.IsInsideZone(model.ZoneIDTown) {
		t.Fatal("expected TOWN flag NOT set for castle zone")
	}
}

func TestPvPZoneOnEnterOnExit(t *testing.T) {
	base := &BaseZone{
		id:       300,
		name:     "TestSiege",
		zoneType: TypeSiege,
		shape:    "Cuboid",
		minZ:     -1000,
		maxZ:     1000,
		nodesX:   []int32{0, 100},
		nodesY:   []int32{0, 100},
	}
	z := NewPvPZone(base)

	ch := newTestCharacter(1, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDPVP) {
		t.Fatal("expected PVP flag set on siege zone enter")
	}
	if !ch.IsInsideZone(model.ZoneIDSiege) {
		t.Fatal("expected SIEGE flag set")
	}
	if !ch.IsInsideZone(model.ZoneIDNoSummonFriend) {
		t.Fatal("expected NO_SUMMON_FRIEND flag set")
	}

	// Exit
	ch.SetLocation(model.NewLocation(200, 200, 0, 0))
	z.RevalidateInZone(ch)

	if ch.IsInsideZone(model.ZoneIDPVP) {
		t.Fatal("expected PVP flag cleared on exit")
	}
	if ch.IsInsideZone(model.ZoneIDSiege) {
		t.Fatal("expected SIEGE flag cleared")
	}
	if ch.IsInsideZone(model.ZoneIDNoSummonFriend) {
		t.Fatal("expected NO_SUMMON_FRIEND flag cleared")
	}
}

func TestWaterZoneOnEnterOnExit(t *testing.T) {
	base := &BaseZone{
		id:       400,
		name:     "TestLake",
		zoneType: TypeWater,
		shape:    "Cuboid",
		minZ:     -1000,
		maxZ:     1000,
		nodesX:   []int32{0, 100},
		nodesY:   []int32{0, 100},
	}
	z := NewWaterZone(base)

	ch := newTestCharacter(1, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDWater) {
		t.Fatal("expected WATER flag set")
	}

	ch.SetLocation(model.NewLocation(200, 200, 0, 0))
	z.RevalidateInZone(ch)

	if ch.IsInsideZone(model.ZoneIDWater) {
		t.Fatal("expected WATER flag cleared")
	}
}

func TestDamageZoneNoFlags(t *testing.T) {
	base := &BaseZone{
		id:       500,
		name:     "TestPoison",
		zoneType: TypeDamage,
		shape:    "Cuboid",
		minZ:     -1000,
		maxZ:     1000,
		nodesX:   []int32{0, 100},
		nodesY:   []int32{0, 100},
	}
	z := NewDamageZone(base)

	ch := newTestCharacter(1, 50, 50, 0)
	z.RevalidateInZone(ch)

	// DamageZone should NOT set any zone flags (Java behavior)
	for i := model.ZoneID(0); i < model.ZoneIDCount; i++ {
		if ch.IsInsideZone(i) {
			t.Fatalf("DamageZone should not set any flags, but ZoneID %d is set", i)
		}
	}

	// But character should still be tracked
	if z.CharacterCount() != 1 {
		t.Fatalf("expected 1 character tracked in damage zone, got %d", z.CharacterCount())
	}
}

func TestRevalidateIdempotent(t *testing.T) {
	base := &BaseZone{
		id:       600,
		name:     "TestIdempotent",
		zoneType: TypeTown,
		shape:    "Cuboid",
		minZ:     -1000,
		maxZ:     1000,
		nodesX:   []int32{0, 100},
		nodesY:   []int32{0, 100},
	}
	z := NewTownZone(base)
	ch := newTestCharacter(1, 50, 50, 0)

	// Call RevalidateInZone multiple times — should only enter once
	z.RevalidateInZone(ch)
	z.RevalidateInZone(ch)
	z.RevalidateInZone(ch)

	if z.CharacterCount() != 1 {
		t.Fatalf("expected 1 character after multiple revalidates, got %d", z.CharacterCount())
	}
}

func TestMultipleCharactersInZone(t *testing.T) {
	base := &BaseZone{
		id:       700,
		name:     "TestMulti",
		zoneType: TypeTown,
		shape:    "Cuboid",
		minZ:     -1000,
		maxZ:     1000,
		nodesX:   []int32{0, 100},
		nodesY:   []int32{0, 100},
	}
	z := NewTownZone(base)

	ch1 := newTestCharacter(1, 50, 50, 0)
	ch2 := newTestCharacter(2, 60, 60, 0)
	ch3 := newTestCharacter(3, 200, 200, 0) // outside

	z.RevalidateInZone(ch1)
	z.RevalidateInZone(ch2)
	z.RevalidateInZone(ch3)

	if z.CharacterCount() != 2 {
		t.Fatalf("expected 2 characters tracked, got %d", z.CharacterCount())
	}

	chars := z.GetCharactersInside()
	if len(chars) != 2 {
		t.Fatalf("GetCharactersInside: expected 2, got %d", len(chars))
	}
}

func TestRemoveCharacter(t *testing.T) {
	base := &BaseZone{
		id:       800,
		name:     "TestRemove",
		zoneType: TypeTown,
		shape:    "Cuboid",
		minZ:     -1000,
		maxZ:     1000,
		nodesX:   []int32{0, 100},
		nodesY:   []int32{0, 100},
	}
	z := NewTownZone(base)

	ch := newTestCharacter(1, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDTown) {
		t.Fatal("expected TOWN flag after enter")
	}

	// Force remove (e.g., logout)
	z.RemoveCharacter(ch)

	if ch.IsInsideZone(model.ZoneIDTown) {
		t.Fatal("expected TOWN flag cleared after RemoveCharacter")
	}
	if z.CharacterCount() != 0 {
		t.Fatalf("expected 0 characters after remove, got %d", z.CharacterCount())
	}

	// Double remove should be no-op
	z.RemoveCharacter(ch)
}

func TestManagerRemoveFromAllZones(t *testing.T) {
	mgr := NewManager()

	base1 := &BaseZone{
		id: 900, name: "Town1", zoneType: TypeTown, shape: "Cuboid",
		minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
	}
	base2 := &BaseZone{
		id: 901, name: "Water1", zoneType: TypeWater, shape: "Cuboid",
		minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
	}

	z1 := NewTownZone(base1)
	z2 := NewWaterZone(base2)
	mgr.zones = append(mgr.zones, z1, z2)

	ch := newTestCharacter(1, 50, 50, 0)
	z1.RevalidateInZone(ch)
	z2.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDTown) || !ch.IsInsideZone(model.ZoneIDWater) {
		t.Fatal("expected TOWN + WATER flags set")
	}

	mgr.RemoveFromAllZones(ch)

	if ch.IsInsideZone(model.ZoneIDTown) || ch.IsInsideZone(model.ZoneIDWater) {
		t.Fatal("expected all zone flags cleared after RemoveFromAllZones")
	}

	if z1.CharacterCount() != 0 || z2.CharacterCount() != 0 {
		t.Fatal("expected 0 characters in both zones after RemoveFromAllZones")
	}
}
