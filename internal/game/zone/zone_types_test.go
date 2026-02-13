package zone

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func TestArenaZoneFlags(t *testing.T) {
	base := &BaseZone{
		id: 1001, name: "Arena", zoneType: TypeArena, shape: "Cuboid",
		minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
	}
	z := NewArenaZone(base)

	if z.IsPeace() {
		t.Fatal("ArenaZone should not be peace")
	}
	if !z.AllowsPvP() {
		t.Fatal("ArenaZone should allow PvP")
	}

	ch := newTestCharacter(10, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDPVP) {
		t.Fatal("expected PVP flag set")
	}

	ch.SetLocation(model.NewLocation(200, 200, 0, 0))
	z.RevalidateInZone(ch)

	if ch.IsInsideZone(model.ZoneIDPVP) {
		t.Fatal("expected PVP flag cleared on exit")
	}
}

func TestJailZoneFlags(t *testing.T) {
	base := &BaseZone{
		id: 1002, name: "Jail", zoneType: TypeJail, shape: "Cuboid",
		minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
	}
	z := NewJailZone(base)

	if z.AllowsPvP() {
		t.Fatal("JailZone should not allow PvP")
	}

	ch := newTestCharacter(11, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDJail) {
		t.Fatal("expected JAIL flag set")
	}
	if !ch.IsInsideZone(model.ZoneIDNoSummonFriend) {
		t.Fatal("expected NO_SUMMON_FRIEND flag set in jail")
	}

	ch.SetLocation(model.NewLocation(200, 200, 0, 0))
	z.RevalidateInZone(ch)

	if ch.IsInsideZone(model.ZoneIDJail) {
		t.Fatal("expected JAIL flag cleared on exit")
	}
	if ch.IsInsideZone(model.ZoneIDNoSummonFriend) {
		t.Fatal("expected NO_SUMMON_FRIEND cleared on exit")
	}
}

func TestOlympiadStadiumFlags(t *testing.T) {
	base := &BaseZone{
		id: 1003, name: "OlyArena", zoneType: TypeOlympiadStadium, shape: "Cuboid",
		minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
		params: map[string]string{"stadiumId": "3"},
	}
	z := NewOlympiadStadiumZone(base)

	if z.StadiumID() != 3 {
		t.Fatalf("expected stadiumID 3, got %d", z.StadiumID())
	}

	ch := newTestCharacter(12, 50, 50, 0)
	z.RevalidateInZone(ch)

	for _, zoneID := range []model.ZoneID{
		model.ZoneIDPVP,
		model.ZoneIDNoSummonFriend,
		model.ZoneIDNoLanding,
		model.ZoneIDNoRestart,
		model.ZoneIDNoBookmark,
	} {
		if !ch.IsInsideZone(zoneID) {
			t.Fatalf("expected ZoneID %d set in olympiad stadium", zoneID)
		}
	}

	ch.SetLocation(model.NewLocation(200, 200, 0, 0))
	z.RevalidateInZone(ch)

	for _, zoneID := range []model.ZoneID{
		model.ZoneIDPVP,
		model.ZoneIDNoSummonFriend,
		model.ZoneIDNoLanding,
		model.ZoneIDNoRestart,
		model.ZoneIDNoBookmark,
	} {
		if ch.IsInsideZone(zoneID) {
			t.Fatalf("expected ZoneID %d cleared on olympiad exit", zoneID)
		}
	}
}

func TestClanHallZoneFlags(t *testing.T) {
	base := &BaseZone{
		id: 1004, name: "ClanHall", zoneType: TypeClanHall, shape: "Cuboid",
		minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
	}
	z := NewClanHallZone(base)

	ch := newTestCharacter(13, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDClanHall) {
		t.Fatal("expected CLAN_HALL flag set")
	}
}

func TestNoStoreZoneFlags(t *testing.T) {
	base := &BaseZone{
		id: 1005, name: "NoStore", zoneType: TypeNoStore, shape: "Cuboid",
		minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
	}
	z := NewNoStoreZone(base)

	ch := newTestCharacter(14, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDNoStore) {
		t.Fatal("expected NO_STORE flag set")
	}
}

func TestNoLandingZoneFlags(t *testing.T) {
	base := &BaseZone{
		id: 1006, name: "NoLand", zoneType: TypeNoLanding, shape: "Cuboid",
		minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
	}
	z := NewNoLandingZone(base)

	ch := newTestCharacter(15, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDNoLanding) {
		t.Fatal("expected NO_LANDING flag set")
	}
}

func TestMotherTreeZoneFlags(t *testing.T) {
	base := &BaseZone{
		id: 1007, name: "MotherTree", zoneType: TypeMotherTree, shape: "Cuboid",
		minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
		params: map[string]string{"HpRegenBonus": "3", "MpRegenBonus": "2"},
	}
	z := NewMotherTreeZone(base)

	if z.HpRegenBonus() != 3 {
		t.Fatalf("expected HpRegenBonus 3, got %d", z.HpRegenBonus())
	}
	if z.MpRegenBonus() != 2 {
		t.Fatalf("expected MpRegenBonus 2, got %d", z.MpRegenBonus())
	}

	ch := newTestCharacter(16, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDMotherTree) {
		t.Fatal("expected MOTHER_TREE flag set")
	}
}

func TestSwampZoneFlags(t *testing.T) {
	base := &BaseZone{
		id: 1008, name: "Swamp", zoneType: TypeSwamp, shape: "Cuboid",
		minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
		params: map[string]string{"move_bonus": "0.3"},
	}
	z := NewSwampZone(base)

	if z.MoveBonus() < 0.29 || z.MoveBonus() > 0.31 {
		t.Fatalf("expected MoveBonus ~0.3, got %f", z.MoveBonus())
	}

	ch := newTestCharacter(17, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDSwamp) {
		t.Fatal("expected SWAMP flag set")
	}
}

func TestBossZoneWhitelist(t *testing.T) {
	base := &BaseZone{
		id: 1009, name: "Boss", zoneType: TypeBoss, shape: "Cuboid",
		minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
		params: map[string]string{"oustX": "100", "oustY": "200", "oustZ": "300"},
	}
	z := NewBossZone(base)

	if z.IsPlayerAllowed(1) {
		t.Fatal("expected player 1 not allowed initially")
	}

	z.AllowPlayerEntry(1, 60_000_000_000) // 60s
	if !z.IsPlayerAllowed(1) {
		t.Fatal("expected player 1 allowed after AllowPlayerEntry")
	}

	z.RemovePlayer(1)
	if z.IsPlayerAllowed(1) {
		t.Fatal("expected player 1 not allowed after RemovePlayer")
	}

	ox, oy, oz := z.OustLocation()
	if ox != 100 || oy != 200 || oz != 300 {
		t.Fatalf("expected oust 100,200,300 got %d,%d,%d", ox, oy, oz)
	}
}

func TestFishingZoneNoFlags(t *testing.T) {
	base := &BaseZone{
		id: 1010, name: "Fishing", zoneType: TypeFishing, shape: "Cuboid",
		minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
	}
	z := NewFishingZone(base)

	ch := newTestCharacter(18, 50, 50, 0)
	z.RevalidateInZone(ch)

	// FishingZone sets no flags (data-only)
	for i := model.ZoneID(0); i < model.ZoneIDCount; i++ {
		if ch.IsInsideZone(i) {
			t.Fatalf("FishingZone should not set any flags, but ZoneID %d is set", i)
		}
	}

	if z.CharacterCount() != 1 {
		t.Fatalf("expected 1 character tracked, got %d", z.CharacterCount())
	}
}

func TestStubZonesNoFlags(t *testing.T) {
	tests := []struct {
		name     string
		zoneType string
		newFn    func(*BaseZone) Zone
	}{
		{"Condition", TypeCondition, func(b *BaseZone) Zone { return NewConditionZone(b) }},
		{"SiegableHall", TypeSiegableHall, func(b *BaseZone) Zone { return NewSiegableHallZone(b) }},
		{"ResidenceTeleport", TypeResidenceTeleport, func(b *BaseZone) Zone { return NewResidenceTeleportZone(b) }},
		{"ResidenceHallTeleport", TypeResidenceHallTeleport, func(b *BaseZone) Zone { return NewResidenceHallTeleportZone(b) }},
		{"Respawn", TypeRespawn, func(b *BaseZone) Zone { return NewRespawnZone(b) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := &BaseZone{
				id: 2000, name: tt.name, zoneType: tt.zoneType, shape: "Cuboid",
				minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
			}
			z := tt.newFn(base)

			if z.IsPeace() {
				t.Fatalf("%s should not be peace", tt.name)
			}

			ch := newTestCharacter(99, 50, 50, 0)
			z.RevalidateInZone(ch)

			for i := model.ZoneID(0); i < model.ZoneIDCount; i++ {
				if ch.IsInsideZone(i) {
					t.Fatalf("%s should not set any flags, but ZoneID %d is set", tt.name, i)
				}
			}
		})
	}
}

func TestNoPvPZoneFlags(t *testing.T) {
	base := &BaseZone{
		id: 1011, name: "NoPvP", zoneType: TypeNoPvP, shape: "Cuboid",
		minZ: -1000, maxZ: 1000, nodesX: []int32{0, 100}, nodesY: []int32{0, 100},
	}
	z := NewNoPvPZone(base)

	if z.AllowsPvP() {
		t.Fatal("NoPvPZone should not allow PvP")
	}

	ch := newTestCharacter(19, 50, 50, 0)
	z.RevalidateInZone(ch)

	if !ch.IsInsideZone(model.ZoneIDNoPVP) {
		t.Fatal("expected NO_PVP flag set")
	}
}

func TestCylinderShapeContains(t *testing.T) {
	base := &BaseZone{
		id:       3000,
		name:     "CylinderTest",
		zoneType: TypeEffect,
		shape:    "Cylinder",
		minZ:     -500,
		maxZ:     500,
		rad:      100,
		nodesX:   []int32{1000}, // center X
		nodesY:   []int32{2000}, // center Y
	}

	// Inside cylinder — at center
	if !base.Contains(1000, 2000, 0) {
		t.Fatal("center should be inside cylinder")
	}

	// Inside cylinder — within radius
	if !base.Contains(1050, 2050, 0) {
		t.Fatal("point within radius should be inside cylinder")
	}

	// Outside cylinder — beyond radius
	if base.Contains(1100, 2100, 0) {
		t.Fatal("point beyond radius should be outside cylinder")
	}

	// Outside cylinder — Z out of range
	if base.Contains(1000, 2000, 600) {
		t.Fatal("point above maxZ should be outside cylinder")
	}
	if base.Contains(1000, 2000, -600) {
		t.Fatal("point below minZ should be outside cylinder")
	}

	// Edge case — exactly at radius boundary
	if !base.Contains(1100, 2000, 0) {
		t.Fatal("point at exactly radius should be inside")
	}

	// Edge case — just beyond radius
	if base.Contains(1101, 2000, 0) {
		t.Fatal("point at radius+1 should be outside")
	}
}

func TestCylinderZeroRadius(t *testing.T) {
	base := &BaseZone{
		id: 3001, name: "ZeroRad", zoneType: TypeEffect, shape: "Cylinder",
		minZ: -100, maxZ: 100, rad: 0,
		nodesX: []int32{0}, nodesY: []int32{0},
	}

	if base.Contains(0, 0, 0) {
		t.Fatal("cylinder with zero radius should not contain any point")
	}
}
