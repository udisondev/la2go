package data

import "testing"

func TestIsRaidBoss(t *testing.T) {
	t.Parallel()

	// NPC table loaded in TestMain (recipe_accessors_test.go)
	raidBossIDs := GetAllRaidBossIDs()
	if len(raidBossIDs) == 0 {
		t.Fatal("GetAllRaidBossIDs returned empty list")
	}

	firstRB := raidBossIDs[0]
	if !IsRaidBoss(firstRB) {
		t.Errorf("IsRaidBoss(%d) = false; want true", firstRB)
	}
	if IsGrandBoss(firstRB) {
		t.Errorf("IsGrandBoss(%d) = true for raid_boss; want false", firstRB)
	}
	if !IsAnyBoss(firstRB) {
		t.Errorf("IsAnyBoss(%d) = false for raid_boss; want true", firstRB)
	}
}

func TestIsGrandBoss(t *testing.T) {
	t.Parallel()

	grandBossIDs := GetAllGrandBossIDs()
	if len(grandBossIDs) == 0 {
		t.Fatal("GetAllGrandBossIDs returned empty list")
	}

	firstGB := grandBossIDs[0]
	if !IsGrandBoss(firstGB) {
		t.Errorf("IsGrandBoss(%d) = false; want true", firstGB)
	}
	if IsRaidBoss(firstGB) {
		t.Errorf("IsRaidBoss(%d) = true for grand_boss; want false", firstGB)
	}
	if !IsAnyBoss(firstGB) {
		t.Errorf("IsAnyBoss(%d) = false for grand_boss; want true", firstGB)
	}
}

func TestIsAnyBoss_NonBoss(t *testing.T) {
	t.Parallel()

	// Find a regular monster (not boss)
	for id, def := range NpcTable {
		if def.npcType == "monster" {
			if IsAnyBoss(id) {
				t.Errorf("IsAnyBoss(%d) = true for monster; want false", id)
			}
			if IsRaidBoss(id) {
				t.Errorf("IsRaidBoss(%d) = true for monster; want false", id)
			}
			if IsGrandBoss(id) {
				t.Errorf("IsGrandBoss(%d) = true for monster; want false", id)
			}
			return
		}
	}
}

func TestIsRaidBoss_Unknown(t *testing.T) {
	t.Parallel()

	if IsRaidBoss(999999) {
		t.Error("IsRaidBoss(999999) = true; want false for unknown NPC")
	}
	if IsGrandBoss(999999) {
		t.Error("IsGrandBoss(999999) = true; want false for unknown NPC")
	}
	if IsAnyBoss(999999) {
		t.Error("IsAnyBoss(999999) = true; want false for unknown NPC")
	}
}

func TestGetAllRaidBossIDs(t *testing.T) {
	t.Parallel()

	ids := GetAllRaidBossIDs()
	if len(ids) == 0 {
		t.Fatal("GetAllRaidBossIDs returned empty; expected >0 raid bosses")
	}

	for _, id := range ids {
		def := GetNpcDef(id)
		if def == nil {
			t.Errorf("GetNpcDef(%d) = nil for raid boss ID", id)
			continue
		}
		if def.npcType != "raid_boss" {
			t.Errorf("raid boss ID %d has npcType %q; want %q", id, def.npcType, "raid_boss")
		}
	}

	t.Logf("found %d raid bosses", len(ids))
}

func TestGetAllGrandBossIDs(t *testing.T) {
	t.Parallel()

	ids := GetAllGrandBossIDs()
	if len(ids) == 0 {
		t.Fatal("GetAllGrandBossIDs returned empty; expected >0 grand bosses")
	}

	for _, id := range ids {
		def := GetNpcDef(id)
		if def == nil {
			t.Errorf("GetNpcDef(%d) = nil for grand boss ID", id)
			continue
		}
		if def.npcType != "grand_boss" {
			t.Errorf("grand boss ID %d has npcType %q; want %q", id, def.npcType, "grand_boss")
		}
	}

	t.Logf("found %d grand bosses", len(ids))
}
