package data

import (
	"testing"
)

func init() {
	// Инициализируем таблицу для тестов
	if TeleporterTable == nil {
		TeleporterTable = make(map[int32]*teleporterDef)
	}
}

func setupTestTeleporterTable(t *testing.T) {
	t.Helper()
	TeleporterTable = map[int32]*teleporterDef{
		30006: {
			npcID: 30006,
			teleports: []teleportGroupDef{
				{
					teleType: "NORMAL",
					locations: []teleportLocDef{
						{name: "Gludin Village", x: -80749, y: 149834, z: -3043, feeCount: 18000, feeId: 0, castleId: 0},
						{name: "Dark Elf Village", x: 9716, y: 15502, z: -4500, feeCount: 24000, feeId: 0, castleId: 0},
					},
				},
				{
					teleType: "NOBLES_TOKEN",
					locations: []teleportLocDef{
						{name: "Gludin Arena", x: -87328, y: 142266, z: -3640, feeCount: 1, feeId: 6651, castleId: 0},
					},
				},
				{
					teleType: "NOBLES_ADENA",
					locations: []teleportLocDef{
						{name: "Gludin Arena", x: -87328, y: 142266, z: -3640, feeCount: 1000, feeId: 0, castleId: 0},
					},
				},
			},
		},
		30256: {
			npcID: 30256,
			teleports: []teleportGroupDef{
				{
					teleType: "NORMAL",
					locations: []teleportLocDef{
						{name: "Town of Gludio", x: -12694, y: 122776, z: -3114, feeCount: 32000, feeId: 0, castleId: 1},
					},
				},
			},
		},
	}
}

func TestGetTeleportGroups(t *testing.T) {
	setupTestTeleporterTable(t)

	tests := []struct {
		name       string
		npcID      int32
		wantGroups int
	}{
		{name: "NPC with 3 groups", npcID: 30006, wantGroups: 3},
		{name: "NPC with 1 group", npcID: 30256, wantGroups: 1},
		{name: "NPC not found", npcID: 99999, wantGroups: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			groups := GetTeleportGroups(tc.npcID)
			if tc.wantGroups == 0 {
				if groups != nil {
					t.Errorf("GetTeleportGroups(%d) = %v, want nil", tc.npcID, groups)
				}
				return
			}
			if len(groups) != tc.wantGroups {
				t.Errorf("GetTeleportGroups(%d) returned %d groups, want %d", tc.npcID, len(groups), tc.wantGroups)
			}
		})
	}
}

func TestGetTeleportGroups_LocationData(t *testing.T) {
	setupTestTeleporterTable(t)

	groups := GetTeleportGroups(30006)
	if len(groups) == 0 {
		t.Fatal("expected groups for NPC 30006")
	}

	normal := groups[0]
	if normal.Type != "NORMAL" {
		t.Errorf("first group type = %q, want NORMAL", normal.Type)
	}
	if len(normal.Locations) != 2 {
		t.Fatalf("NORMAL group has %d locations, want 2", len(normal.Locations))
	}

	loc := normal.Locations[0]
	if loc.Name != "Gludin Village" {
		t.Errorf("loc.Name = %q, want %q", loc.Name, "Gludin Village")
	}
	if loc.X != -80749 || loc.Y != 149834 || loc.Z != -3043 {
		t.Errorf("loc coords = (%d,%d,%d), want (-80749,149834,-3043)", loc.X, loc.Y, loc.Z)
	}
	if loc.FeeCount != 18000 {
		t.Errorf("loc.FeeCount = %d, want 18000", loc.FeeCount)
	}
	if loc.Index != 0 {
		t.Errorf("loc.Index = %d, want 0", loc.Index)
	}
}

func TestGetTeleportGroupByType(t *testing.T) {
	setupTestTeleporterTable(t)

	tests := []struct {
		name      string
		npcID     int32
		teleType  string
		wantNil   bool
		wantLocs  int
	}{
		{name: "NORMAL exists", npcID: 30006, teleType: "NORMAL", wantLocs: 2},
		{name: "NOBLES_TOKEN exists", npcID: 30006, teleType: "NOBLES_TOKEN", wantLocs: 1},
		{name: "NOBLES_ADENA exists", npcID: 30006, teleType: "NOBLES_ADENA", wantLocs: 1},
		{name: "type not found", npcID: 30006, teleType: "OTHER", wantNil: true},
		{name: "NPC not found", npcID: 99999, teleType: "NORMAL", wantNil: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			group := GetTeleportGroupByType(tc.npcID, tc.teleType)
			if tc.wantNil {
				if group != nil {
					t.Errorf("expected nil, got group with %d locations", len(group.Locations))
				}
				return
			}
			if group == nil {
				t.Fatal("expected non-nil group")
			}
			if len(group.Locations) != tc.wantLocs {
				t.Errorf("locations count = %d, want %d", len(group.Locations), tc.wantLocs)
			}
		})
	}
}

func TestGetTeleportGroupByType_NoblesTokenFeeID(t *testing.T) {
	setupTestTeleporterTable(t)

	group := GetTeleportGroupByType(30006, "NOBLES_TOKEN")
	if group == nil {
		t.Fatal("NOBLES_TOKEN group not found")
	}

	loc := group.Locations[0]
	if loc.FeeID != 6651 {
		t.Errorf("FeeID = %d, want 6651 (Noblesse Token)", loc.FeeID)
	}
	if loc.FeeCount != 1 {
		t.Errorf("FeeCount = %d, want 1", loc.FeeCount)
	}
}

func TestGetTeleportLocation(t *testing.T) {
	setupTestTeleporterTable(t)

	tests := []struct {
		name     string
		npcID    int32
		teleType string
		index    int
		wantNil  bool
		wantName string
	}{
		{name: "valid location 0", npcID: 30006, teleType: "NORMAL", index: 0, wantName: "Gludin Village"},
		{name: "valid location 1", npcID: 30006, teleType: "NORMAL", index: 1, wantName: "Dark Elf Village"},
		{name: "index out of range", npcID: 30006, teleType: "NORMAL", index: 99, wantNil: true},
		{name: "negative index", npcID: 30006, teleType: "NORMAL", index: -1, wantNil: true},
		{name: "wrong type", npcID: 30006, teleType: "OTHER", index: 0, wantNil: true},
		{name: "NPC not found", npcID: 99999, teleType: "NORMAL", index: 0, wantNil: true},
		{name: "castle location", npcID: 30256, teleType: "NORMAL", index: 0, wantName: "Town of Gludio"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			loc := GetTeleportLocation(tc.npcID, tc.teleType, tc.index)
			if tc.wantNil {
				if loc != nil {
					t.Errorf("expected nil, got location %q", loc.Name)
				}
				return
			}
			if loc == nil {
				t.Fatal("expected non-nil location")
			}
			if loc.Name != tc.wantName {
				t.Errorf("Name = %q, want %q", loc.Name, tc.wantName)
			}
		})
	}
}

func TestGetTeleportLocation_CastleID(t *testing.T) {
	setupTestTeleporterTable(t)

	loc := GetTeleportLocation(30256, "NORMAL", 0)
	if loc == nil {
		t.Fatal("expected non-nil location")
	}
	if loc.CastleID != 1 {
		t.Errorf("CastleID = %d, want 1", loc.CastleID)
	}
}

func TestHasTeleporter(t *testing.T) {
	setupTestTeleporterTable(t)

	if !HasTeleporter(30006) {
		t.Error("HasTeleporter(30006) = false, want true")
	}
	if !HasTeleporter(30256) {
		t.Error("HasTeleporter(30256) = false, want true")
	}
	if HasTeleporter(99999) {
		t.Error("HasTeleporter(99999) = true, want false")
	}
}
