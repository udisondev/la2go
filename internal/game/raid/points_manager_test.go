package raid

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"testing"
)

// mockRaidPointsStore implements RaidPointsStore for testing.
type mockRaidPointsStore struct {
	mu     sync.Mutex
	points map[string]int32 // key: "charID:bossID"
}

func newMockRaidPointsStore() *mockRaidPointsStore {
	return &mockRaidPointsStore{points: make(map[string]int32)}
}

func (s *mockRaidPointsStore) AddRaidPoints(_ context.Context, characterID, bossID, points int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := fmt.Sprintf("%d:%d", characterID, bossID)
	s.points[key] += points
	return nil
}

func (s *mockRaidPointsStore) GetTotalRaidPoints(_ context.Context, characterID int32) (int32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var total int32
	prefix := fmt.Sprintf("%d:", characterID)
	for key, pts := range s.points {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			total += pts
		}
	}
	return total, nil
}

func (s *mockRaidPointsStore) GetTopRaidPointPlayers(_ context.Context, limit int) ([]RaidPointsEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Aggregate per character
	charPts := make(map[int32]int32)
	for key, pts := range s.points {
		var charID, bossID int32
		if _, err := fmt.Sscanf(key, "%d:%d", &charID, &bossID); err != nil {
			continue
		}
		charPts[charID] += pts
	}

	// Sort descending
	entries := make([]RaidPointsEntry, 0, len(charPts))
	for charID, pts := range charPts {
		entries = append(entries, RaidPointsEntry{CharacterID: charID, Points: pts})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Points > entries[j].Points
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}
	return entries, nil
}

func (s *mockRaidPointsStore) ResetAllRaidPoints(_ context.Context) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := int64(len(s.points))
	s.points = make(map[string]int32)
	return count, nil
}

func TestPointsManager_AddPoints(t *testing.T) {
	t.Parallel()

	store := newMockRaidPointsStore()
	mgr := NewPointsManager(store)

	ctx := context.Background()

	// Boss level 50 → points = 50/2 = 25
	if err := mgr.AddPoints(ctx, 1001, 25001, 50); err != nil {
		t.Fatalf("AddPoints: %v", err)
	}

	pts, err := mgr.GetPoints(ctx, 1001)
	if err != nil {
		t.Fatalf("GetPoints: %v", err)
	}
	if pts != 25 {
		t.Errorf("GetPoints(1001) = %d; want 25", pts)
	}
}

func TestPointsManager_AddPoints_MinOne(t *testing.T) {
	t.Parallel()

	store := newMockRaidPointsStore()
	mgr := NewPointsManager(store)
	ctx := context.Background()

	// Boss level 1 → points = max(1, 1/2) = max(1, 0) = 1
	if err := mgr.AddPoints(ctx, 1001, 25001, 1); err != nil {
		t.Fatalf("AddPoints: %v", err)
	}

	pts, err := mgr.GetPoints(ctx, 1001)
	if err != nil {
		t.Fatalf("GetPoints: %v", err)
	}
	if pts != 1 {
		t.Errorf("GetPoints(1001) = %d; want 1 (minimum)", pts)
	}
}

func TestPointsManager_AddPoints_Accumulates(t *testing.T) {
	t.Parallel()

	store := newMockRaidPointsStore()
	mgr := NewPointsManager(store)
	ctx := context.Background()

	// Kill two bosses
	if err := mgr.AddPoints(ctx, 1001, 25001, 50); err != nil { // 25 points
		t.Fatalf("AddPoints: %v", err)
	}
	if err := mgr.AddPoints(ctx, 1001, 25002, 70); err != nil { // 35 points
		t.Fatalf("AddPoints: %v", err)
	}

	pts, err := mgr.GetPoints(ctx, 1001)
	if err != nil {
		t.Fatalf("GetPoints: %v", err)
	}
	if pts != 60 { // 25 + 35
		t.Errorf("GetPoints(1001) = %d; want 60 (25+35)", pts)
	}
}

func TestPointsManager_GetRanking(t *testing.T) {
	t.Parallel()

	store := newMockRaidPointsStore()
	mgr := NewPointsManager(store)
	ctx := context.Background()

	// Player 1: 25 pts, Player 2: 50 pts, Player 3: 10 pts
	if err := mgr.AddPoints(ctx, 1001, 25001, 50); err != nil {
		t.Fatalf("AddPoints: %v", err)
	}
	if err := mgr.AddPoints(ctx, 1002, 25001, 100); err != nil {
		t.Fatalf("AddPoints: %v", err)
	}
	if err := mgr.AddPoints(ctx, 1003, 25001, 20); err != nil {
		t.Fatalf("AddPoints: %v", err)
	}

	ranking, err := mgr.GetRanking(ctx, 2) // top 2
	if err != nil {
		t.Fatalf("GetRanking: %v", err)
	}
	if len(ranking) != 2 {
		t.Fatalf("ranking length = %d; want 2", len(ranking))
	}
	if ranking[0].CharacterID != 1002 {
		t.Errorf("top1 CharacterID = %d; want 1002", ranking[0].CharacterID)
	}
	if ranking[0].Points != 50 { // 100/2 = 50
		t.Errorf("top1 Points = %d; want 50", ranking[0].Points)
	}
}

func TestPointsManager_WeeklyReset(t *testing.T) {
	t.Parallel()

	store := newMockRaidPointsStore()
	mgr := NewPointsManager(store)
	ctx := context.Background()

	// Add some points
	if err := mgr.AddPoints(ctx, 1001, 25001, 50); err != nil {
		t.Fatalf("AddPoints: %v", err)
	}
	if err := mgr.AddPoints(ctx, 1002, 25001, 70); err != nil {
		t.Fatalf("AddPoints: %v", err)
	}

	// Reset
	deleted, err := mgr.WeeklyReset(ctx)
	if err != nil {
		t.Fatalf("WeeklyReset: %v", err)
	}
	if deleted != 2 {
		t.Errorf("deleted = %d; want 2", deleted)
	}

	// Verify cache cleared
	pts, err := mgr.GetPoints(ctx, 1001)
	if err != nil {
		t.Fatalf("GetPoints after reset: %v", err)
	}
	if pts != 0 {
		t.Errorf("GetPoints(1001) = %d after reset; want 0", pts)
	}
}

func TestCalculatePoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level int32
		want  int32
	}{
		{0, 1},   // min 1
		{1, 1},   // 1/2=0 → min 1
		{2, 1},   // 2/2=1
		{10, 5},  // 10/2=5
		{50, 25}, // 50/2=25
		{79, 39}, // 79/2=39
		{80, 40}, // 80/2=40
	}

	for _, tt := range tests {
		if got := CalculatePoints(tt.level); got != tt.want {
			t.Errorf("CalculatePoints(%d) = %d; want %d", tt.level, got, tt.want)
		}
	}
}

func TestPointsManager_GetPoints_CacheHit(t *testing.T) {
	t.Parallel()

	store := newMockRaidPointsStore()
	mgr := NewPointsManager(store)
	ctx := context.Background()

	// Add points (populates cache)
	if err := mgr.AddPoints(ctx, 1001, 25001, 50); err != nil {
		t.Fatalf("AddPoints: %v", err)
	}

	// First call: cache hit
	pts1, err := mgr.GetPoints(ctx, 1001)
	if err != nil {
		t.Fatalf("GetPoints: %v", err)
	}

	// Second call: still cache
	pts2, err := mgr.GetPoints(ctx, 1001)
	if err != nil {
		t.Fatalf("GetPoints: %v", err)
	}

	if pts1 != pts2 {
		t.Errorf("cache inconsistency: %d != %d", pts1, pts2)
	}
}
