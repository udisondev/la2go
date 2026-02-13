package commands

import (
	"strings"
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/admin"
	"github.com/udisondev/la2go/internal/model"
)

// mockClientMgr implements ClientManager for testing.
type mockClientMgr struct {
	players    map[string]*model.Player
	kicked     []string
	count      int
}

func newMockClientMgr() *mockClientMgr {
	return &mockClientMgr{
		players: make(map[string]*model.Player),
	}
}

func (m *mockClientMgr) FindPlayerByName(name string) *model.Player {
	for n, p := range m.players {
		if strings.EqualFold(n, name) {
			return p
		}
	}
	return nil
}

func (m *mockClientMgr) ForEachPlayer(fn func(*model.Player) bool) {
	for _, p := range m.players {
		if !fn(p) {
			break
		}
	}
}

func (m *mockClientMgr) PlayerCount() int {
	if m.count > 0 {
		return m.count
	}
	return len(m.players)
}

func (m *mockClientMgr) KickPlayer(name string) bool {
	if _, ok := m.players[name]; ok {
		m.kicked = append(m.kicked, name)
		return true
	}
	return false
}

func newTestPlayer(t *testing.T, name string) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(1, 100, 200, name, 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	p.SetAccessLevel(100) // admin by default
	return p
}

// --- Heal tests ---

func TestHeal_SelfHeal(t *testing.T) {
	player := newTestPlayer(t, "AdminGM")
	player.Character.SetCurrentHP(1)
	player.Character.SetCurrentMP(1)
	player.Character.SetCurrentCP(0)

	cmd := NewHeal(newMockClientMgr())
	if err := cmd.Handle(player, []string{"heal"}); err != nil {
		t.Fatalf("Heal.Handle: %v", err)
	}

	if player.CurrentHP() != player.MaxHP() {
		t.Errorf("HP = %d, want %d", player.CurrentHP(), player.MaxHP())
	}
	if player.CurrentMP() != player.MaxMP() {
		t.Errorf("MP = %d, want %d", player.CurrentMP(), player.MaxMP())
	}

	msg := player.ClearLastAdminMessage()
	if !strings.Contains(msg, "self") && !strings.Contains(msg, "Self") {
		t.Errorf("message = %q, want mention of self heal", msg)
	}
}

func TestHeal_TargetByName(t *testing.T) {
	mgr := newMockClientMgr()
	target, err := model.NewPlayer(2, 101, 200, "Target", 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	target.Character.SetCurrentHP(1)
	mgr.players["Target"] = target

	player := newTestPlayer(t, "AdminGM")
	cmd := NewHeal(mgr)
	if err := cmd.Handle(player, []string{"heal", "Target"}); err != nil {
		t.Fatalf("Heal.Handle: %v", err)
	}

	if target.CurrentHP() != target.MaxHP() {
		t.Errorf("target HP = %d, want %d", target.CurrentHP(), target.MaxHP())
	}

	msg := player.ClearLastAdminMessage()
	if !strings.Contains(msg, "Target") {
		t.Errorf("message = %q, want mention of Target", msg)
	}
}

func TestHeal_TargetNotFound(t *testing.T) {
	mgr := newMockClientMgr()
	player := newTestPlayer(t, "AdminGM")
	cmd := NewHeal(mgr)

	err := cmd.Handle(player, []string{"heal", "NoSuchPlayer"})
	if err == nil {
		t.Error("expected error for missing target")
	}
}

// --- Kick tests ---

func TestKick_Success(t *testing.T) {
	mgr := newMockClientMgr()
	target := newTestPlayer(t, "Victim")
	mgr.players["Victim"] = target

	player := newTestPlayer(t, "AdminGM")
	cmd := NewKick(mgr)

	if err := cmd.Handle(player, []string{"kick", "Victim"}); err != nil {
		t.Fatalf("Kick.Handle: %v", err)
	}

	if len(mgr.kicked) != 1 || mgr.kicked[0] != "Victim" {
		t.Errorf("kicked = %v, want [Victim]", mgr.kicked)
	}

	msg := player.ClearLastAdminMessage()
	if !strings.Contains(msg, "Victim") {
		t.Errorf("message = %q, want mention of Victim", msg)
	}
}

func TestKick_MissingArgs(t *testing.T) {
	cmd := NewKick(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"kick"})
	if err == nil {
		t.Error("expected error for missing args")
	}
}

func TestKick_NotFound(t *testing.T) {
	cmd := NewKick(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"kick", "NoSuchPlayer"})
	if err == nil {
		t.Error("expected error for missing player")
	}
}

// --- Ban tests ---

func TestBan_Success(t *testing.T) {
	mgr := newMockClientMgr()
	target, err := model.NewPlayer(2, 101, 200, "Cheater", 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	mgr.players["Cheater"] = target

	player := newTestPlayer(t, "AdminGM")
	cmd := NewBan(mgr)

	if err := cmd.Handle(player, []string{"ban", "Cheater"}); err != nil {
		t.Fatalf("Ban.Handle: %v", err)
	}

	if target.AccessLevel() != -100 {
		t.Errorf("target access level = %d, want -100", target.AccessLevel())
	}

	if len(mgr.kicked) != 1 {
		t.Errorf("kicked count = %d, want 1", len(mgr.kicked))
	}
}

func TestBan_MissingArgs(t *testing.T) {
	cmd := NewBan(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"ban"})
	if err == nil {
		t.Error("expected error for missing args")
	}
}

func TestBan_TargetOffline(t *testing.T) {
	cmd := NewBan(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"ban", "OfflinePlayer"})
	if err == nil {
		t.Error("expected error for offline target")
	}
}

// --- Invisible / Invul tests ---

func TestInvisible_Toggle(t *testing.T) {
	cmd := &Invisible{}
	player := newTestPlayer(t, "AdminGM")

	if player.IsInvisible() {
		t.Fatal("player should start visible")
	}

	if err := cmd.Handle(player, []string{"invisible"}); err != nil {
		t.Fatalf("Invisible.Handle: %v", err)
	}
	if !player.IsInvisible() {
		t.Error("player should be invisible after first toggle")
	}

	if err := cmd.Handle(player, []string{"invisible"}); err != nil {
		t.Fatalf("Invisible.Handle: %v", err)
	}
	if player.IsInvisible() {
		t.Error("player should be visible after second toggle")
	}
}

func TestInvul_Toggle(t *testing.T) {
	cmd := &Invul{}
	player := newTestPlayer(t, "AdminGM")

	if player.IsInvulnerable() {
		t.Fatal("player should start vulnerable")
	}

	if err := cmd.Handle(player, []string{"invul"}); err != nil {
		t.Fatalf("Invul.Handle: %v", err)
	}
	if !player.IsInvulnerable() {
		t.Error("player should be invulnerable after first toggle")
	}

	if err := cmd.Handle(player, []string{"invul"}); err != nil {
		t.Fatalf("Invul.Handle: %v", err)
	}
	if player.IsInvulnerable() {
		t.Error("player should be vulnerable after second toggle")
	}
}

// --- Announce tests ---

func TestAnnounce_Success(t *testing.T) {
	cmd := NewAnnounce(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	if err := cmd.Handle(player, []string{"announce", "Server", "restart"}); err != nil {
		t.Fatalf("Announce.Handle: %v", err)
	}

	msg := player.ClearLastAdminMessage()
	if !strings.HasPrefix(msg, "ANNOUNCE:") {
		t.Errorf("message = %q, want ANNOUNCE: prefix", msg)
	}
	if !strings.Contains(msg, "Server restart") {
		t.Errorf("message = %q, want to contain 'Server restart'", msg)
	}
}

func TestAnnounce_MissingText(t *testing.T) {
	cmd := NewAnnounce(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"announce"})
	if err == nil {
		t.Error("expected error for missing text")
	}
}

// --- SetLevel tests ---

func TestSetLevel_Self(t *testing.T) {
	cmd := &SetLevel{}
	player := newTestPlayer(t, "AdminGM")

	if err := cmd.Handle(player, []string{"setlevel", "60"}); err != nil {
		t.Fatalf("SetLevel.Handle: %v", err)
	}

	if player.Level() != 60 {
		t.Errorf("level = %d, want 60", player.Level())
	}
}

func TestSetLevel_InvalidRange(t *testing.T) {
	cmd := &SetLevel{}
	player := newTestPlayer(t, "AdminGM")

	if err := cmd.Handle(player, []string{"setlevel", "0"}); err == nil {
		t.Error("expected error for level 0")
	}
	if err := cmd.Handle(player, []string{"setlevel", "81"}); err == nil {
		t.Error("expected error for level 81")
	}
}

func TestSetLevel_MissingArgs(t *testing.T) {
	cmd := &SetLevel{}
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"setlevel"})
	if err == nil {
		t.Error("expected error for missing level arg")
	}
}

// --- User commands tests ---

func TestLoc(t *testing.T) {
	cmd := &Loc{}
	player := newTestPlayer(t, "TestUser")
	player.SetLocation(model.NewLocation(100, 200, -300, 0))

	if err := cmd.Handle(player, ""); err != nil {
		t.Fatalf("Loc.Handle: %v", err)
	}

	msg := player.ClearLastAdminMessage()
	if !strings.Contains(msg, "100") || !strings.Contains(msg, "200") || !strings.Contains(msg, "-300") {
		t.Errorf("message = %q, want coordinates 100 200 -300", msg)
	}
}

func TestGameTime(t *testing.T) {
	cmd := &GameTime{}
	player := newTestPlayer(t, "TestUser")

	if err := cmd.Handle(player, ""); err != nil {
		t.Fatalf("GameTime.Handle: %v", err)
	}

	msg := player.ClearLastAdminMessage()
	if !strings.Contains(msg, "Game time") || !strings.Contains(msg, "Server time") {
		t.Errorf("message = %q, want 'Game time' and 'Server time'", msg)
	}
}

func TestUnstuck_Normal(t *testing.T) {
	cmd := &Unstuck{}
	player := newTestPlayer(t, "TestUser")
	player.SetLocation(model.NewLocation(5000, 5000, 5000, 0))

	if err := cmd.Handle(player, ""); err != nil {
		t.Fatalf("Unstuck.Handle: %v", err)
	}

	loc := player.Location()
	if loc.X != unstuckX || loc.Y != unstuckY || loc.Z != unstuckZ {
		t.Errorf("location = (%d, %d, %d), want (%d, %d, %d)",
			loc.X, loc.Y, loc.Z, unstuckX, unstuckY, unstuckZ)
	}
}

func TestUnstuck_WhileDead(t *testing.T) {
	cmd := &Unstuck{}
	player := newTestPlayer(t, "TestUser")
	player.Character.DoDie(nil)

	err := cmd.Handle(player, "")
	if err == nil {
		t.Error("expected error when using /unstuck while dead")
	}
}

func TestOnline(t *testing.T) {
	mgr := newMockClientMgr()
	mgr.count = 42

	cmd := NewOnline(mgr)
	player := newTestPlayer(t, "TestUser")

	if err := cmd.Handle(player, ""); err != nil {
		t.Fatalf("Online.Handle: %v", err)
	}

	msg := player.ClearLastAdminMessage()
	if !strings.Contains(msg, "42") {
		t.Errorf("message = %q, want to contain '42'", msg)
	}
}

// --- Kill tests ---

func TestKill_ByName(t *testing.T) {
	mgr := newMockClientMgr()
	target, err := model.NewPlayer(2, 101, 200, "Victim", 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	mgr.players["Victim"] = target

	player := newTestPlayer(t, "AdminGM")
	cmd := NewKill(mgr)

	if err := cmd.Handle(player, []string{"kill", "Victim"}); err != nil {
		t.Fatalf("Kill.Handle: %v", err)
	}

	if !target.IsDead() {
		t.Error("target should be dead after //kill")
	}
	if target.CurrentHP() != 0 {
		t.Errorf("target HP = %d, want 0", target.CurrentHP())
	}
}

func TestKill_NoTarget(t *testing.T) {
	cmd := NewKill(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"kill"})
	if err == nil {
		t.Error("expected error when no target selected")
	}
}

func TestKill_ByNameNotFound(t *testing.T) {
	cmd := NewKill(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"kill", "NoSuchPlayer"})
	if err == nil {
		t.Error("expected error for missing target")
	}
}

// --- Res tests ---

func TestRes_ByName(t *testing.T) {
	mgr := newMockClientMgr()
	target, err := model.NewPlayer(2, 101, 200, "DeadGuy", 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	target.Character.DoDie(nil) // kill them first
	mgr.players["DeadGuy"] = target

	player := newTestPlayer(t, "AdminGM")
	cmd := NewRes(mgr)

	if err := cmd.Handle(player, []string{"res", "DeadGuy"}); err != nil {
		t.Fatalf("Res.Handle: %v", err)
	}

	if target.IsDead() {
		t.Error("target should be alive after //res")
	}
	if target.CurrentHP() != target.MaxHP() {
		t.Errorf("target HP = %d, want %d", target.CurrentHP(), target.MaxHP())
	}
}

func TestRes_NoTarget(t *testing.T) {
	cmd := NewRes(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"res"})
	if err == nil {
		t.Error("expected error when no target selected")
	}
}

func TestRes_ByNameNotFound(t *testing.T) {
	cmd := NewRes(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"res", "NoSuchPlayer"})
	if err == nil {
		t.Error("expected error for missing target")
	}
}

// --- Speed tests ---

func TestSpeed_Valid(t *testing.T) {
	cmd := &Speed{}
	player := newTestPlayer(t, "AdminGM")

	if err := cmd.Handle(player, []string{"speed", "2.0"}); err != nil {
		t.Fatalf("Speed.Handle: %v", err)
	}

	msg := player.ClearLastAdminMessage()
	if !strings.Contains(msg, "2.0") {
		t.Errorf("message = %q, want mention of 2.0", msg)
	}
}

func TestSpeed_MissingArgs(t *testing.T) {
	cmd := &Speed{}
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"speed"})
	if err == nil {
		t.Error("expected error for missing speed arg")
	}
}

func TestSpeed_InvalidValue(t *testing.T) {
	cmd := &Speed{}
	player := newTestPlayer(t, "AdminGM")

	if err := cmd.Handle(player, []string{"speed", "abc"}); err == nil {
		t.Error("expected error for non-numeric speed")
	}
}

func TestSpeed_OutOfRange(t *testing.T) {
	cmd := &Speed{}
	player := newTestPlayer(t, "AdminGM")

	if err := cmd.Handle(player, []string{"speed", "0.01"}); err == nil {
		t.Error("expected error for speed < 0.1")
	}
	if err := cmd.Handle(player, []string{"speed", "100.0"}); err == nil {
		t.Error("expected error for speed > 50")
	}
}

// --- Info tests ---

func TestInfo_ServerStatus(t *testing.T) {
	mgr := newMockClientMgr()
	mgr.count = 10
	cmd := NewInfo(mgr)
	player := newTestPlayer(t, "AdminGM")

	// No target, no name arg → server status
	if err := cmd.Handle(player, []string{"info"}); err != nil {
		t.Fatalf("Info.Handle: %v", err)
	}

	msg := player.ClearLastAdminMessage()
	if !strings.Contains(msg, "Server Status") {
		t.Errorf("message = %q, want 'Server Status'", msg)
	}
	if !strings.Contains(msg, "10") {
		t.Errorf("message = %q, want '10' players", msg)
	}
}

func TestInfo_PlayerByName(t *testing.T) {
	mgr := newMockClientMgr()
	target, err := model.NewPlayer(2, 101, 200, "InfoTarget", 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	mgr.players["InfoTarget"] = target

	cmd := NewInfo(mgr)
	player := newTestPlayer(t, "AdminGM")

	if err := cmd.Handle(player, []string{"info", "InfoTarget"}); err != nil {
		t.Fatalf("Info.Handle: %v", err)
	}

	msg := player.ClearLastAdminMessage()
	if !strings.Contains(msg, "InfoTarget") {
		t.Errorf("message = %q, want mention of InfoTarget", msg)
	}
	if !strings.Contains(msg, "Level: 40") {
		t.Errorf("message = %q, want 'Level: 40'", msg)
	}
}

func TestInfo_PlayerNotFound(t *testing.T) {
	cmd := NewInfo(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"info", "NoSuch"})
	if err == nil {
		t.Error("expected error for missing player")
	}
}

// --- SetLevel edge cases ---

func TestSetLevel_NonNumeric(t *testing.T) {
	cmd := &SetLevel{}
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"setlevel", "abc"})
	if err == nil {
		t.Error("expected error for non-numeric level")
	}
}

// --- Heal edge case: heal by name not found ---

func TestHeal_SelfWhenNoTarget(t *testing.T) {
	cmd := NewHeal(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")
	player.Character.SetCurrentHP(10)

	// No args and no target → self heal
	if err := cmd.Handle(player, []string{"heal"}); err != nil {
		t.Fatalf("Heal.Handle: %v", err)
	}

	if player.CurrentHP() != player.MaxHP() {
		t.Errorf("HP = %d, want %d (self heal)", player.CurrentHP(), player.MaxHP())
	}
}

// --- Jail missing args ---

func TestJail_MissingArgs(t *testing.T) {
	cmd := NewJail(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"jail"})
	if err == nil {
		t.Error("expected error for missing args")
	}
	err = cmd.Handle(player, []string{"jail", "player"})
	if err == nil {
		t.Error("expected error for missing minutes")
	}
}

func TestJail_InvalidMinutes(t *testing.T) {
	cmd := NewJail(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"jail", "player", "abc"})
	if err == nil {
		t.Error("expected error for non-numeric minutes")
	}
	err = cmd.Handle(player, []string{"jail", "player", "0"})
	if err == nil {
		t.Error("expected error for 0 minutes")
	}
}

func TestJail_TargetNotFound(t *testing.T) {
	cmd := NewJail(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"jail", "NoSuch", "10"})
	if err == nil {
		t.Error("expected error for missing target")
	}
}

// --- GiveItem error cases ---

func TestGiveItem_MissingArgs(t *testing.T) {
	cmd := &GiveItem{}
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"give_item"})
	if err == nil {
		t.Error("expected error for missing args")
	}
	err = cmd.Handle(player, []string{"give_item", "57"})
	if err == nil {
		t.Error("expected error for missing count")
	}
}

func TestGiveItem_InvalidItemID(t *testing.T) {
	cmd := &GiveItem{}
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"give_item", "abc", "10"})
	if err == nil {
		t.Error("expected error for non-numeric itemID")
	}
}

func TestGiveItem_InvalidCount(t *testing.T) {
	cmd := &GiveItem{}
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"give_item", "57", "abc"})
	if err == nil {
		t.Error("expected error for non-numeric count")
	}
	err = cmd.Handle(player, []string{"give_item", "57", "0"})
	if err == nil {
		t.Error("expected error for count 0")
	}
	err = cmd.Handle(player, []string{"give_item", "57", "9999999"})
	if err == nil {
		t.Error("expected error for count > 999999")
	}
}

func TestGiveItem_ItemNotFound(t *testing.T) {
	cmd := &GiveItem{}
	player := newTestPlayer(t, "AdminGM")

	// Valid numeric args, but data not loaded → item template not found
	err := cmd.Handle(player, []string{"give_item", "999999", "1"})
	if err == nil {
		t.Error("expected error for non-existent item template")
	}
}

// --- Teleport error cases ---

func TestTeleport_MissingArgs(t *testing.T) {
	cmd := NewTeleport(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"teleport"})
	if err == nil {
		t.Error("expected error for missing coords")
	}
}

func TestTeleport_GotoNotFound(t *testing.T) {
	cmd := NewTeleport(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"goto", "NoSuch"})
	if err == nil {
		t.Error("expected error for missing target")
	}
}

func TestTeleport_RecallNotFound(t *testing.T) {
	cmd := NewTeleport(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"recall", "NoSuch"})
	if err == nil {
		t.Error("expected error for missing target")
	}
}

func TestTeleport_InvalidCoordinates(t *testing.T) {
	cmd := NewTeleport(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"teleport", "abc", "0", "0"})
	if err == nil {
		t.Error("expected error for non-numeric x")
	}
	err = cmd.Handle(player, []string{"teleport", "0", "abc", "0"})
	if err == nil {
		t.Error("expected error for non-numeric y")
	}
	err = cmd.Handle(player, []string{"teleport", "0", "0", "abc"})
	if err == nil {
		t.Error("expected error for non-numeric z")
	}
}

func TestTeleport_TooFewCoords(t *testing.T) {
	cmd := NewTeleport(newMockClientMgr())
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"teleport", "100", "200"})
	if err == nil {
		t.Error("expected error for only 2 coordinates")
	}
}

// --- Delete error cases ---

func TestDelete_NoTarget(t *testing.T) {
	cmd := &Delete{}
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"delete"})
	if err == nil {
		t.Error("expected error when no target selected")
	}
}

// --- Spawn error cases ---

func TestSpawn_MissingArgs(t *testing.T) {
	cmd := &Spawn{}
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"spawn"})
	if err == nil {
		t.Error("expected error for missing npcID")
	}
}

func TestSpawn_InvalidNpcID(t *testing.T) {
	cmd := &Spawn{}
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"spawn", "abc"})
	if err == nil {
		t.Error("expected error for non-numeric npcID")
	}
}

func TestSpawn_NpcNotFound(t *testing.T) {
	cmd := &Spawn{}
	player := newTestPlayer(t, "AdminGM")

	// Valid numeric ID, but data not loaded → NPC template not found
	err := cmd.Handle(player, []string{"spawn", "999999"})
	if err == nil {
		t.Error("expected error for non-existent NPC template")
	}
}

func TestSpawn_InvalidCount(t *testing.T) {
	cmd := &Spawn{}
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"spawn", "100", "abc"})
	if err == nil {
		t.Error("expected error for non-numeric count")
	}
}

func TestSpawn_CountOutOfRange(t *testing.T) {
	cmd := &Spawn{}
	player := newTestPlayer(t, "AdminGM")

	err := cmd.Handle(player, []string{"spawn", "999999", "0"})
	if err == nil {
		t.Error("expected error for count 0")
	}
	err = cmd.Handle(player, []string{"spawn", "999999", "101"})
	if err == nil {
		t.Error("expected error for count > 100")
	}
}

// --- Registry tests ---

func TestRegisterAll(t *testing.T) {
	mgr := newMockClientMgr()
	h := admin.NewHandler()

	RegisterAll(h, mgr)

	// Verify admin commands registered
	adminCount := h.AdminCommandCount()
	if adminCount < 20 {
		t.Errorf("AdminCommandCount = %d, want >= 20", adminCount)
	}

	// Verify user commands registered
	userCount := h.UserCommandCount()
	if userCount < 4 {
		t.Errorf("UserCommandCount = %d, want >= 4", userCount)
	}
}

// TestAllRequiredAccessLevels verifies all commands return valid access levels.
func TestAllRequiredAccessLevels(t *testing.T) {
	mgr := newMockClientMgr()

	adminCmds := []struct {
		name string
		cmd  interface {
			RequiredAccessLevel() int32
			Names() []string
		}
	}{
		{"Teleport", NewTeleport(mgr)},
		{"Kill", NewKill(mgr)},
		{"Heal", NewHeal(mgr)},
		{"Res", NewRes(mgr)},
		{"Announce", NewAnnounce(mgr)},
		{"Kick", NewKick(mgr)},
		{"Ban", NewBan(mgr)},
		{"Jail", NewJail(mgr)},
		{"Info", NewInfo(mgr)},
		{"Spawn", &Spawn{}},
		{"Delete", &Delete{}},
		{"SetLevel", &SetLevel{}},
		{"GiveItem", &GiveItem{}},
		{"Invisible", &Invisible{}},
		{"Invul", &Invul{}},
		{"Speed", &Speed{}},
	}

	for _, tt := range adminCmds {
		level := tt.cmd.RequiredAccessLevel()
		if level < 1 {
			t.Errorf("%s.RequiredAccessLevel() = %d, want >= 1", tt.name, level)
		}
		names := tt.cmd.Names()
		if len(names) == 0 {
			t.Errorf("%s.Names() returned empty", tt.name)
		}
	}
}

