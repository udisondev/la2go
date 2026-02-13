package gameserver

import (
	"sync"
	"testing"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/game/hall"
	"github.com/udisondev/la2go/internal/game/siege"
	"github.com/udisondev/la2go/internal/model"
)

var loadZonesOnce sync.Once

func ensureZonesLoaded(t *testing.T) {
	t.Helper()
	loadZonesOnce.Do(func() {
		if err := data.LoadZones(); err != nil {
			t.Fatalf("LoadZones: %v", err)
		}
	})
}

func newRespawnHandler(t *testing.T, hallTbl *hall.Table, siegeMgr *siege.Manager) *Handler {
	t.Helper()
	cm := NewClientManager()
	return NewHandler(
		nil, cm, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, siegeMgr, nil,
		hallTbl, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
}

func newRespawnPlayer(t *testing.T, objectID uint32, clanID int32) *model.Player {
	t.Helper()
	player, err := model.NewPlayer(objectID, int64(objectID), int64(objectID), "Respawner", 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))
	if clanID > 0 {
		player.SetClanID(clanID)
	}
	return player
}

// --- getClanHallRespawn ---

func TestGetClanHallRespawn_NilTable(t *testing.T) {
	t.Parallel()
	h := newRespawnHandler(t, nil, nil)
	player := newRespawnPlayer(t, 1, 100)

	loc := h.getClanHallRespawn(player)
	if loc != defaultRespawnLoc {
		t.Errorf("getClanHallRespawn(nil table) = %v; want %v", loc, defaultRespawnLoc)
	}
}

func TestGetClanHallRespawn_NoClan(t *testing.T) {
	t.Parallel()
	ht := hall.NewTable()
	h := newRespawnHandler(t, ht, nil)
	player := newRespawnPlayer(t, 1, 0)

	loc := h.getClanHallRespawn(player)
	if loc != defaultRespawnLoc {
		t.Errorf("getClanHallRespawn(no clan) = %v; want %v", loc, defaultRespawnLoc)
	}
}

func TestGetClanHallRespawn_ClanNoHall(t *testing.T) {
	t.Parallel()
	ht := hall.NewTable()
	h := newRespawnHandler(t, ht, nil)
	player := newRespawnPlayer(t, 1, 999)

	loc := h.getClanHallRespawn(player)
	if loc != defaultRespawnLoc {
		t.Errorf("getClanHallRespawn(clan no hall) = %v; want %v", loc, defaultRespawnLoc)
	}
}

func TestGetClanHallRespawn_WithZoneData(t *testing.T) {
	t.Parallel()
	ensureZonesLoaded(t)

	ht := hall.NewTable()
	// Hall 36 = "The Golden Chamber" в Aden.
	// ClanHallZone для hallID=36 существует в zone_data_generated.go
	// (aden_castle_agit_001 с params clanHallId=36, спавн {149134, 23140, -2155}).
	if err := ht.SetOwner(36, 500); err != nil {
		t.Fatalf("SetOwner: %v", err)
	}

	h := newRespawnHandler(t, ht, nil)
	player := newRespawnPlayer(t, 1, 500)

	loc := h.getClanHallRespawn(player)
	if loc == defaultRespawnLoc {
		t.Error("getClanHallRespawn() = defaultRespawnLoc; want clan hall spawn")
	}

	// Проверяем конкретные координаты из zone_data_generated.go.
	if loc.X != 149134 || loc.Y != 23140 || loc.Z != -2155 {
		t.Errorf("getClanHallRespawn() = (%d, %d, %d); want (149134, 23140, -2155)",
			loc.X, loc.Y, loc.Z)
	}
}

// --- getCastleRespawn ---

func TestGetCastleRespawn_NilManager(t *testing.T) {
	t.Parallel()
	h := newRespawnHandler(t, nil, nil)
	player := newRespawnPlayer(t, 1, 100)

	loc := h.getCastleRespawn(player)
	if loc != defaultRespawnLoc {
		t.Errorf("getCastleRespawn(nil manager) = %v; want %v", loc, defaultRespawnLoc)
	}
}

func TestGetCastleRespawn_NoClan(t *testing.T) {
	t.Parallel()
	sm := siege.NewManager(siege.DefaultManagerConfig())
	h := newRespawnHandler(t, nil, sm)
	player := newRespawnPlayer(t, 1, 0)

	loc := h.getCastleRespawn(player)
	if loc != defaultRespawnLoc {
		t.Errorf("getCastleRespawn(no clan) = %v; want %v", loc, defaultRespawnLoc)
	}
}

func TestGetCastleRespawn_ClanNoCastle(t *testing.T) {
	t.Parallel()
	sm := siege.NewManager(siege.DefaultManagerConfig())
	h := newRespawnHandler(t, nil, sm)
	player := newRespawnPlayer(t, 1, 999)

	loc := h.getCastleRespawn(player)
	if loc != defaultRespawnLoc {
		t.Errorf("getCastleRespawn(clan no castle) = %v; want %v", loc, defaultRespawnLoc)
	}
}

func TestGetCastleRespawn_WithZoneData(t *testing.T) {
	t.Parallel()
	ensureZonesLoaded(t)

	sm := siege.NewManager(siege.DefaultManagerConfig())
	// Castle 5 = Aden. CastleZone с castleId=5 содержит спавн {147700, 4608, -2784}.
	castle := sm.Castle(5)
	if castle == nil {
		t.Fatal("Castle(5) = nil")
	}
	castle.SetOwnerClanID(600)

	h := newRespawnHandler(t, nil, sm)
	player := newRespawnPlayer(t, 1, 600)

	loc := h.getCastleRespawn(player)
	if loc == defaultRespawnLoc {
		t.Error("getCastleRespawn() = defaultRespawnLoc; want castle spawn")
	}

	// Первый спавн aden_castle без spawnType: {147700, 4608, -2784}.
	if loc.X != 147700 || loc.Y != 4608 || loc.Z != -2784 {
		t.Errorf("getCastleRespawn() = (%d, %d, %d); want (147700, 4608, -2784)",
			loc.X, loc.Y, loc.Z)
	}
}

// --- getSiegeHQRespawn ---

func TestGetSiegeHQRespawn_NilManager(t *testing.T) {
	t.Parallel()
	h := newRespawnHandler(t, nil, nil)
	player := newRespawnPlayer(t, 1, 100)

	loc := h.getSiegeHQRespawn(player)
	if loc != defaultRespawnLoc {
		t.Errorf("getSiegeHQRespawn(nil manager) = %v; want %v", loc, defaultRespawnLoc)
	}
}

func TestGetSiegeHQRespawn_NoClan(t *testing.T) {
	t.Parallel()
	sm := siege.NewManager(siege.DefaultManagerConfig())
	h := newRespawnHandler(t, nil, sm)
	player := newRespawnPlayer(t, 1, 0)

	loc := h.getSiegeHQRespawn(player)
	if loc != defaultRespawnLoc {
		t.Errorf("getSiegeHQRespawn(no clan) = %v; want %v", loc, defaultRespawnLoc)
	}
}

func TestGetSiegeHQRespawn_ClanNotRegistered(t *testing.T) {
	t.Parallel()
	sm := siege.NewManager(siege.DefaultManagerConfig())
	h := newRespawnHandler(t, nil, sm)
	player := newRespawnPlayer(t, 1, 999)

	loc := h.getSiegeHQRespawn(player)
	if loc != defaultRespawnLoc {
		t.Errorf("getSiegeHQRespawn(not registered) = %v; want %v", loc, defaultRespawnLoc)
	}
}

func TestGetSiegeHQRespawn_AttackerSiegeRunning(t *testing.T) {
	t.Parallel()
	sm := siege.NewManager(siege.DefaultManagerConfig())

	// Регистрируем клан 700 как атакующего замка Aden (5).
	if err := sm.RegisterAttacker(5, 700, "TestClan", 5); err != nil {
		t.Fatalf("RegisterAttacker: %v", err)
	}

	// Запускаем осаду.
	castle := sm.Castle(5)
	s := castle.Siege()
	s.StartSiege()
	defer s.EndSiege()

	h := newRespawnHandler(t, nil, sm)
	player := newRespawnPlayer(t, 1, 700)

	// Поскольку флаговая система не реализована, ожидаем defaultRespawnLoc.
	loc := h.getSiegeHQRespawn(player)
	if loc != defaultRespawnLoc {
		t.Errorf("getSiegeHQRespawn(attacker, running) = %v; want %v (flag system not implemented)",
			loc, defaultRespawnLoc)
	}
}

func TestGetSiegeHQRespawn_DefenderFallback(t *testing.T) {
	t.Parallel()
	sm := siege.NewManager(siege.DefaultManagerConfig())

	// Регистрируем клан как защитника -- HQ респаун не для защитников.
	if err := sm.RegisterDefender(5, 800, "DefClan", 5); err != nil {
		t.Fatalf("RegisterDefender: %v", err)
	}

	castle := sm.Castle(5)
	s := castle.Siege()
	s.StartSiege()
	defer s.EndSiege()

	h := newRespawnHandler(t, nil, sm)
	player := newRespawnPlayer(t, 1, 800)

	loc := h.getSiegeHQRespawn(player)
	if loc != defaultRespawnLoc {
		t.Errorf("getSiegeHQRespawn(defender) = %v; want %v", loc, defaultRespawnLoc)
	}
}

// --- findZoneSpawn ---

func TestFindZoneSpawn_NotFound(t *testing.T) {
	t.Parallel()
	ensureZonesLoaded(t)

	_, ok := findZoneSpawn("ClanHallZone", "clanHallId", 99999)
	if ok {
		t.Error("findZoneSpawn(nonexistent) = ok; want false")
	}
}

func TestFindZoneSpawn_ClanHall(t *testing.T) {
	t.Parallel()
	ensureZonesLoaded(t)

	loc, ok := findZoneSpawn("ClanHallZone", "clanHallId", 36)
	if !ok {
		t.Fatal("findZoneSpawn(ClanHallZone, 36) = !ok; want ok")
	}
	if loc.X != 149134 || loc.Y != 23140 || loc.Z != -2155 {
		t.Errorf("findZoneSpawn(ClanHallZone, 36) = (%d, %d, %d); want (149134, 23140, -2155)",
			loc.X, loc.Y, loc.Z)
	}
}

func TestFindZoneSpawn_Castle(t *testing.T) {
	t.Parallel()
	ensureZonesLoaded(t)

	loc, ok := findZoneSpawn("CastleZone", "castleId", 5)
	if !ok {
		t.Fatal("findZoneSpawn(CastleZone, 5) = !ok; want ok")
	}
	if loc.X != 147700 || loc.Y != 4608 || loc.Z != -2784 {
		t.Errorf("findZoneSpawn(CastleZone, 5) = (%d, %d, %d); want (147700, 4608, -2784)",
			loc.X, loc.Y, loc.Z)
	}
}
