package cursed

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func TestNewWeapon(t *testing.T) {
	t.Parallel()

	w := NewWeapon(ZaricheItemID, "Zariche", ZaricheSkillID)

	if w.ItemID() != ZaricheItemID {
		t.Errorf("ItemID() = %d, want %d", w.ItemID(), ZaricheItemID)
	}
	if w.Name() != "Zariche" {
		t.Errorf("Name() = %q, want %q", w.Name(), "Zariche")
	}
	if w.SkillID() != ZaricheSkillID {
		t.Errorf("SkillID() = %d, want %d", w.SkillID(), ZaricheSkillID)
	}
	if w.State() != StateInactive {
		t.Errorf("State() = %d, want %d", w.State(), StateInactive)
	}
	if w.IsActive() {
		t.Error("IsActive() = true, want false")
	}
}

func TestNewManager(t *testing.T) {
	t.Parallel()

	m := NewManager()

	// Проверяем что оба оружия зарегистрированы
	zariche := m.Weapon(ZaricheItemID)
	if zariche == nil {
		t.Fatal("Zariche not found")
	}
	if zariche.ItemID() != ZaricheItemID {
		t.Errorf("Zariche.ItemID() = %d, want %d", zariche.ItemID(), ZaricheItemID)
	}

	akamanah := m.Weapon(AkamanahItemID)
	if akamanah == nil {
		t.Fatal("Akamanah not found")
	}
	if akamanah.ItemID() != AkamanahItemID {
		t.Errorf("Akamanah.ItemID() = %d, want %d", akamanah.ItemID(), AkamanahItemID)
	}

	// Несуществующее оружие
	if w := m.Weapon(9999); w != nil {
		t.Error("Weapon(9999) should be nil")
	}
}

func TestIsCursedWeapon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		itemID int32
		want   bool
	}{
		{"Zariche", ZaricheItemID, true},
		{"Akamanah", AkamanahItemID, true},
		{"random item", 57, false},
		{"zero", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsCursedWeapon(tt.itemID); got != tt.want {
				t.Errorf("IsCursedWeapon(%d) = %v, want %v", tt.itemID, got, tt.want)
			}
		})
	}
}

func TestManagerIsCursed(t *testing.T) {
	t.Parallel()

	m := NewManager()

	if !m.IsCursed(ZaricheItemID) {
		t.Error("IsCursed(Zariche) = false, want true")
	}
	if !m.IsCursed(AkamanahItemID) {
		t.Error("IsCursed(Akamanah) = false, want true")
	}
	if m.IsCursed(9999) {
		t.Error("IsCursed(9999) = true, want false")
	}
}

func TestCursedWeaponIDs(t *testing.T) {
	t.Parallel()

	m := NewManager()
	ids := m.CursedWeaponIDs()

	if len(ids) != 2 {
		t.Fatalf("CursedWeaponIDs() len = %d, want 2", len(ids))
	}

	found := map[int32]bool{}
	for _, id := range ids {
		found[id] = true
	}
	if !found[ZaricheItemID] {
		t.Error("Zariche not in CursedWeaponIDs()")
	}
	if !found[AkamanahItemID] {
		t.Error("Akamanah not in CursedWeaponIDs()")
	}
}

func TestWeapons(t *testing.T) {
	t.Parallel()

	m := NewManager()
	weapons := m.Weapons()

	if len(weapons) != 2 {
		t.Fatalf("Weapons() len = %d, want 2", len(weapons))
	}
}

func TestSkillLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		nbKills  int32
		want     int32
	}{
		{"zero kills", 0, 1},
		{"9 kills", 9, 1},
		{"10 kills", 10, 2},
		{"25 kills", 25, 3},
		{"99 kills", 99, 10},
		{"100 kills", 100, 10},    // cap at max level
		{"999 kills", 999, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := NewWeapon(ZaricheItemID, "Zariche", ZaricheSkillID)
			w.nbKills = tt.nbKills

			if got := w.SkillLevel(); got != tt.want {
				t.Errorf("SkillLevel() with %d kills = %d, want %d", tt.nbKills, got, tt.want)
			}
		})
	}
}

func TestActivateAndEndOfLife(t *testing.T) {
	t.Parallel()

	w := NewWeapon(ZaricheItemID, "Zariche", ZaricheSkillID)
	player := makeTestPlayer(t, 1001, 1002, "TestPlayer")
	player.SetKarma(500)
	player.SetPKKills(10)

	if err := w.Activate(player); err != nil {
		t.Fatalf("Activate() error = %v", err)
	}

	// Проверяем состояние оружия
	if w.State() != StateActivated {
		t.Errorf("State() = %d, want %d", w.State(), StateActivated)
	}
	if w.PlayerID() != 1001 {
		t.Errorf("PlayerID() = %d, want 1001", w.PlayerID())
	}

	// Проверяем что karma/pkKills игрока изменились
	if player.Karma() != MaxCursedKarma {
		t.Errorf("player.Karma() = %d, want %d", player.Karma(), MaxCursedKarma)
	}
	if player.PKKills() != 0 {
		t.Errorf("player.PKKills() = %d, want 0", player.PKKills())
	}
	if player.CursedWeaponEquippedID() != ZaricheItemID {
		t.Errorf("player.CursedWeaponEquippedID() = %d, want %d", player.CursedWeaponEquippedID(), ZaricheItemID)
	}

	// Проверяем сохранённые оригинальные значения
	if w.PlayerKarma() != 500 {
		t.Errorf("PlayerKarma() = %d, want 500", w.PlayerKarma())
	}
	if w.PlayerPKKills() != 10 {
		t.Errorf("PlayerPKKills() = %d, want 10", w.PlayerPKKills())
	}

	// EndOfLife — восстанавливает значения
	w.EndOfLife()

	if w.State() != StateInactive {
		t.Errorf("State() after EndOfLife = %d, want %d", w.State(), StateInactive)
	}
	if player.Karma() != 500 {
		t.Errorf("player.Karma() after EndOfLife = %d, want 500", player.Karma())
	}
	if player.PKKills() != 10 {
		t.Errorf("player.PKKills() after EndOfLife = %d, want 10", player.PKKills())
	}
	if player.CursedWeaponEquippedID() != 0 {
		t.Errorf("player.CursedWeaponEquippedID() after EndOfLife = %d, want 0", player.CursedWeaponEquippedID())
	}
}

func TestActivateAlreadyActive(t *testing.T) {
	t.Parallel()

	w := NewWeapon(ZaricheItemID, "Zariche", ZaricheSkillID)
	player := makeTestPlayer(t, 1001, 1002, "TestPlayer")

	if err := w.Activate(player); err != nil {
		t.Fatalf("first Activate() error = %v", err)
	}
	defer w.Stop()

	player2 := makeTestPlayer(t, 2001, 2002, "TestPlayer2")
	if err := w.Activate(player2); err != ErrAlreadyActive {
		t.Errorf("second Activate() error = %v, want ErrAlreadyActive", err)
	}
}

func TestIncreaseKills(t *testing.T) {
	t.Parallel()

	w := NewWeapon(ZaricheItemID, "Zariche", ZaricheSkillID)
	player := makeTestPlayer(t, 1001, 1002, "TestPlayer")

	if err := w.Activate(player); err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	defer w.Stop()

	initialEnd := w.EndTime()

	w.IncreaseKills()

	if w.NBKills() != 1 {
		t.Errorf("NBKills() = %d, want 1", w.NBKills())
	}

	// Duration should decrease by durationLost*60000 ms
	expectedEnd := initialEnd - DefaultDurationLost*60000
	if w.EndTime() != expectedEnd {
		t.Errorf("EndTime() = %d, want %d (decreased by %d ms)", w.EndTime(), expectedEnd, DefaultDurationLost*60000)
	}

	// Player PKKills = nbKills
	if player.PKKills() != 1 {
		t.Errorf("player.PKKills() = %d, want 1", player.PKKills())
	}
}

func TestRestore(t *testing.T) {
	t.Parallel()

	w := NewWeapon(AkamanahItemID, "Akamanah", AkamanahSkillID)
	w.Restore(5001, 1000, 50, 25, 99999999999)
	defer w.Stop()

	if w.State() != StateActivated {
		t.Errorf("State() = %d, want %d", w.State(), StateActivated)
	}
	if w.PlayerID() != 5001 {
		t.Errorf("PlayerID() = %d, want 5001", w.PlayerID())
	}
	if w.PlayerKarma() != 1000 {
		t.Errorf("PlayerKarma() = %d, want 1000", w.PlayerKarma())
	}
	if w.PlayerPKKills() != 50 {
		t.Errorf("PlayerPKKills() = %d, want 50", w.PlayerPKKills())
	}
	if w.NBKills() != 25 {
		t.Errorf("NBKills() = %d, want 25", w.NBKills())
	}
}

func TestSaveData(t *testing.T) {
	t.Parallel()

	w := NewWeapon(ZaricheItemID, "Zariche", ZaricheSkillID)
	w.Restore(3001, 777, 33, 15, 888888)
	defer w.Stop()

	itemID, charID, karma, pk, kills, end := w.SaveData()

	if itemID != ZaricheItemID {
		t.Errorf("itemID = %d, want %d", itemID, ZaricheItemID)
	}
	if charID != 3001 {
		t.Errorf("charID = %d, want 3001", charID)
	}
	if karma != 777 {
		t.Errorf("karma = %d, want 777", karma)
	}
	if pk != 33 {
		t.Errorf("pk = %d, want 33", pk)
	}
	if kills != 15 {
		t.Errorf("kills = %d, want 15", kills)
	}
	if end != 888888 {
		t.Errorf("endTime = %d, want 888888", end)
	}
}

func TestManagerCheckOwnsWeapon(t *testing.T) {
	t.Parallel()

	m := NewManager()
	zariche := m.Weapon(ZaricheItemID)

	player := makeTestPlayer(t, 1001, 1002, "TestPlayer")
	if err := zariche.Activate(player); err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	defer zariche.Stop()

	if got := m.CheckOwnsWeapon(1001); got != ZaricheItemID {
		t.Errorf("CheckOwnsWeapon(1001) = %d, want %d", got, ZaricheItemID)
	}
	if got := m.CheckOwnsWeapon(9999); got != 0 {
		t.Errorf("CheckOwnsWeapon(9999) = %d, want 0", got)
	}
}

func TestManagerCheckPlayer(t *testing.T) {
	t.Parallel()

	m := NewManager()
	zariche := m.Weapon(ZaricheItemID)

	player := makeTestPlayer(t, 1001, 1002, "TestPlayer")
	if err := zariche.Activate(player); err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	defer zariche.Stop()

	// Создаём нового игрока с тем же ObjectID (как при логине)
	loginPlayer := makeTestPlayer(t, 1001, 1002, "TestPlayer")
	found := m.CheckPlayer(loginPlayer)
	if found == nil {
		t.Fatal("CheckPlayer() returned nil, want weapon")
	}
	if found.ItemID() != ZaricheItemID {
		t.Errorf("found weapon = %d, want %d", found.ItemID(), ZaricheItemID)
	}
}

func TestLocationInfoEmpty(t *testing.T) {
	t.Parallel()

	m := NewManager()
	infos := m.LocationInfo()

	if len(infos) != 0 {
		t.Errorf("LocationInfo() len = %d, want 0 (all inactive)", len(infos))
	}
}

func TestLocationInfoActivated(t *testing.T) {
	t.Parallel()

	m := NewManager()
	zariche := m.Weapon(ZaricheItemID)

	player := makeTestPlayer(t, 1001, 1002, "TestPlayer")
	if err := zariche.Activate(player); err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	defer zariche.Stop()

	infos := m.LocationInfo()
	if len(infos) != 1 {
		t.Fatalf("LocationInfo() len = %d, want 1", len(infos))
	}

	if infos[0].ItemID != ZaricheItemID {
		t.Errorf("ItemID = %d, want %d", infos[0].ItemID, ZaricheItemID)
	}
	if infos[0].Activated != 1 {
		t.Errorf("Activated = %d, want 1", infos[0].Activated)
	}
}

func TestConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  int32
		want int32
	}{
		{"ZaricheItemID", ZaricheItemID, 8190},
		{"AkamanahItemID", AkamanahItemID, 8689},
		{"ZaricheSkillID", ZaricheSkillID, 3603},
		{"AkamanahSkillID", AkamanahSkillID, 3629},
		{"VoidBurstSkill", VoidBurstSkill, 3630},
		{"VoidFlowSkill", VoidFlowSkill, 3631},
		{"MaxCursedKarma", MaxCursedKarma, 9999999},
		{"DefaultDropRate", DefaultDropRate, 1},
		{"DefaultDisappearChance", DefaultDisappearChance, 50},
		{"DefaultStageKills", DefaultStageKills, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestManagerStop(t *testing.T) {
	t.Parallel()

	m := NewManager()
	zariche := m.Weapon(ZaricheItemID)
	player := makeTestPlayer(t, 1001, 1002, "TestPlayer")

	if err := zariche.Activate(player); err != nil {
		t.Fatalf("Activate() error = %v", err)
	}

	// Stop не должен паниковать
	m.Stop()

	// Повторный Stop тоже не должен паниковать
	m.Stop()
}

// makeTestPlayer создаёт тестового игрока с минимальными полями.
func makeTestPlayer(t *testing.T, objectID uint32, characterID int64, name string) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(objectID, characterID, 0, name, 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer() error = %v", err)
	}
	return p
}
