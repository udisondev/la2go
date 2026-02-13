package olympiad

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func makeTestPlayers(t *testing.T, count int, classID int32) []*model.Player {
	t.Helper()
	players := make([]*model.Player, count)
	for i := range count {
		p, err := model.NewPlayer(uint32(100+i), int64(100+i), 1, "Player"+string(rune('A'+i)), 40, 0, classID)
		if err != nil {
			t.Fatalf("NewPlayer: %v", err)
		}
		p.SetLocation(model.NewLocation(0, 0, 0, 0))
		players[i] = p
	}
	return players
}

func TestNewManager(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)

	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Nobles() != nobles {
		t.Error("Nobles() mismatch")
	}
	if m.BattleStarted() {
		t.Error("BattleStarted should be false initially")
	}
	if m.ActiveGameCount() != 0 {
		t.Errorf("ActiveGameCount() = %d; want 0", m.ActiveGameCount())
	}
}

func TestManager_RegisterNonClassBased(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)
	players := makeTestPlayers(t, 3, 88)

	for _, p := range players {
		m.RegisterNonClassBased(p)
	}

	if m.NonClassedCount() != 3 {
		t.Errorf("NonClassedCount() = %d; want 3", m.NonClassedCount())
	}
}

func TestManager_RegisterClassBased(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)
	players := makeTestPlayers(t, 5, 88)

	for _, p := range players {
		m.RegisterClassBased(p, 88)
	}

	if m.ClassedCount(88) != 5 {
		t.Errorf("ClassedCount(88) = %d; want 5", m.ClassedCount(88))
	}
}

func TestManager_IsRegistered(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)
	players := makeTestPlayers(t, 2, 88)

	m.RegisterNonClassBased(players[0])
	m.RegisterClassBased(players[1], 88)

	if !m.IsRegistered(players[0].ObjectID()) {
		t.Error("player[0] should be registered (non-classed)")
	}
	if !m.IsRegistered(players[1].ObjectID()) {
		t.Error("player[1] should be registered (classed)")
	}
	if m.IsRegistered(999) {
		t.Error("player 999 should not be registered")
	}
}

func TestManager_UnregisterPlayer_NonClassed(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)
	players := makeTestPlayers(t, 3, 88)

	for _, p := range players {
		m.RegisterNonClassBased(p)
	}

	if !m.UnregisterPlayer(players[1].ObjectID()) {
		t.Error("UnregisterPlayer should return true")
	}
	if m.NonClassedCount() != 2 {
		t.Errorf("NonClassedCount() = %d; want 2", m.NonClassedCount())
	}
	if m.IsRegistered(players[1].ObjectID()) {
		t.Error("player[1] should no longer be registered")
	}
}

func TestManager_UnregisterPlayer_Classed(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)
	players := makeTestPlayers(t, 3, 88)

	for _, p := range players {
		m.RegisterClassBased(p, 88)
	}

	if !m.UnregisterPlayer(players[0].ObjectID()) {
		t.Error("UnregisterPlayer should return true")
	}
	if m.ClassedCount(88) != 2 {
		t.Errorf("ClassedCount(88) = %d; want 2", m.ClassedCount(88))
	}
}

func TestManager_UnregisterPlayer_NotFound(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)

	if m.UnregisterPlayer(999) {
		t.Error("UnregisterPlayer(999) should return false for empty manager")
	}
}

func TestManager_HasEnoughNonClassed(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)

	if m.HasEnoughNonClassed() {
		t.Error("should not have enough with 0 players")
	}

	players := makeTestPlayers(t, MinNonClassedParticipants, 88)
	for _, p := range players {
		m.RegisterNonClassBased(p)
	}

	if !m.HasEnoughNonClassed() {
		t.Errorf("should have enough with %d players", MinNonClassedParticipants)
	}
}

func TestManager_HasEnoughClassed(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)

	classes := m.HasEnoughClassed()
	if len(classes) != 0 {
		t.Errorf("HasEnoughClassed() = %v; want empty", classes)
	}

	players := makeTestPlayers(t, MinClassedParticipants, 88)
	for _, p := range players {
		m.RegisterClassBased(p, 88)
	}

	classes = m.HasEnoughClassed()
	if len(classes) != 1 {
		t.Fatalf("HasEnoughClassed() count = %d; want 1", len(classes))
	}
	if classes[0] != 88 {
		t.Errorf("HasEnoughClassed()[0] = %d; want 88", classes[0])
	}
}

func TestManager_ClearRegistered(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)

	players := makeTestPlayers(t, 5, 88)
	for _, p := range players {
		m.RegisterNonClassBased(p)
	}

	m.ClearRegistered()

	if m.NonClassedCount() != 0 {
		t.Errorf("NonClassedCount() after clear = %d; want 0", m.NonClassedCount())
	}
}

func TestManager_CreateMatches_NonClassed(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)

	// Зарегистрировать nobles + players
	players := makeTestPlayers(t, MinNonClassedParticipants, 88)
	for _, p := range players {
		nobles.Register(p.CharacterID(), 88)
		m.RegisterNonClassBased(p)
	}

	games := m.CreateMatches()

	if len(games) == 0 {
		t.Fatal("CreateMatches should create at least 1 game")
	}

	// Проверить что стадион занят
	g := games[0]
	if !g.Stadium().InUse() {
		t.Error("Stadium should be InUse after match creation")
	}
	if g.CompType() != CompNonClassed {
		t.Errorf("CompType() = %v; want NON_CLASSED", g.CompType())
	}
	if g.Player1() == nil || g.Player2() == nil {
		t.Error("players should not be nil")
	}

	if !m.BattleStarted() {
		t.Error("BattleStarted should be true after creating matches")
	}
}

func TestManager_CreateMatches_Classed(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)

	players := makeTestPlayers(t, MinClassedParticipants, 88)
	for _, p := range players {
		nobles.Register(p.CharacterID(), 88)
		m.RegisterClassBased(p, 88)
	}

	games := m.CreateMatches()

	if len(games) == 0 {
		t.Fatal("CreateMatches should create at least 1 classed game")
	}

	g := games[0]
	if g.CompType() != CompClassed {
		t.Errorf("CompType() = %v; want CLASSED", g.CompType())
	}
}

func TestManager_CreateMatches_NotEnough(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)

	// Только 2 игрока — недостаточно для обоих типов
	players := makeTestPlayers(t, 2, 88)
	for _, p := range players {
		m.RegisterNonClassBased(p)
	}

	games := m.CreateMatches()

	if len(games) != 0 {
		t.Errorf("CreateMatches() count = %d; want 0 (not enough players)", len(games))
	}
}

func TestManager_RemoveGame(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)

	players := makeTestPlayers(t, MinNonClassedParticipants, 88)
	for _, p := range players {
		nobles.Register(p.CharacterID(), 88)
		m.RegisterNonClassBased(p)
	}

	games := m.CreateMatches()
	if len(games) == 0 {
		t.Fatal("no games created")
	}

	stadiumID := games[0].Stadium().ID()
	m.RemoveGame(stadiumID)

	if m.GetGame(stadiumID) != nil {
		t.Error("game should be removed")
	}
	if m.ActiveGameCount() != 0 {
		t.Errorf("ActiveGameCount() = %d; want 0", m.ActiveGameCount())
	}
	if m.BattleStarted() {
		t.Error("BattleStarted should be false when no games active")
	}
}

func TestManager_FindGameByPlayer(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)

	players := makeTestPlayers(t, MinNonClassedParticipants, 88)
	for _, p := range players {
		nobles.Register(p.CharacterID(), 88)
		m.RegisterNonClassBased(p)
	}

	games := m.CreateMatches()
	if len(games) == 0 {
		t.Fatal("no games created")
	}

	p1ID := games[0].Player1().ObjectID
	found := m.FindGameByPlayer(p1ID)
	if found == nil {
		t.Error("FindGameByPlayer should find the game")
	}

	if m.FindGameByPlayer(999) != nil {
		t.Error("FindGameByPlayer(999) should return nil")
	}
}

func TestManager_Stadium(t *testing.T) {
	nobles := NewNobleTable()
	m := NewManager(nobles)

	s := m.Stadium(0)
	if s == nil {
		t.Fatal("Stadium(0) returned nil")
	}
	if s.ID() != 0 {
		t.Errorf("Stadium(0).ID() = %d; want 0", s.ID())
	}

	if m.Stadium(-1) != nil {
		t.Error("Stadium(-1) should return nil")
	}
	if m.Stadium(22) != nil {
		t.Error("Stadium(22) should return nil")
	}
}

func TestNextOpponents(t *testing.T) {
	players := makeTestPlayers(t, 4, 88)

	opponents, remaining := nextOpponents(players)

	if opponents == nil {
		t.Fatal("nextOpponents returned nil")
	}
	if len(opponents) != 2 {
		t.Fatalf("opponents count = %d; want 2", len(opponents))
	}
	if len(remaining) != 2 {
		t.Errorf("remaining count = %d; want 2", len(remaining))
	}

	// Проверить что выбранные отличаются от оставшихся
	selectedIDs := map[uint32]bool{
		opponents[0].ObjectID(): true,
		opponents[1].ObjectID(): true,
	}
	for _, r := range remaining {
		if selectedIDs[r.ObjectID()] {
			t.Error("selected player found in remaining list")
		}
	}
}

func TestNextOpponents_NotEnough(t *testing.T) {
	players := makeTestPlayers(t, 1, 88)

	opponents, remaining := nextOpponents(players)

	if opponents != nil {
		t.Error("nextOpponents should return nil for < 2 players")
	}
	if len(remaining) != 1 {
		t.Errorf("remaining count = %d; want 1", len(remaining))
	}
}

func TestNextOpponents_Empty(t *testing.T) {
	opponents, remaining := nextOpponents(nil)

	if opponents != nil {
		t.Error("nextOpponents(nil) should return nil")
	}
	if remaining != nil {
		t.Error("remaining should be nil")
	}
}
