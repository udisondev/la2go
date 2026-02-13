package wedding

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// mockStore implements CoupleStore for tests.
type mockStore struct {
	mu      sync.Mutex
	couples map[int32]CoupleRow
	nextID  int32
	failOn  string // method name to fail on
}

func newMockStore() *mockStore {
	return &mockStore{
		couples: make(map[int32]CoupleRow),
		nextID:  1,
	}
}

func (s *mockStore) LoadAll(_ context.Context) ([]CoupleRow, error) {
	if s.failOn == "LoadAll" {
		return nil, fmt.Errorf("mock load error")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	rows := make([]CoupleRow, 0, len(s.couples))
	for _, row := range s.couples {
		rows = append(rows, row)
	}
	return rows, nil
}

func (s *mockStore) Create(_ context.Context, p1ID, p2ID int32) (int32, error) {
	if s.failOn == "Create" {
		return 0, fmt.Errorf("mock create error")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if p1ID > p2ID {
		p1ID, p2ID = p2ID, p1ID
	}
	id := s.nextID
	s.nextID++
	s.couples[id] = CoupleRow{
		ID:        id,
		Player1ID: p1ID,
		Player2ID: p2ID,
		Married:   false,
	}
	return id, nil
}

func (s *mockStore) UpdateMarried(_ context.Context, coupleID int32) error {
	if s.failOn == "UpdateMarried" {
		return fmt.Errorf("mock update error")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	row, ok := s.couples[coupleID]
	if !ok {
		return fmt.Errorf("couple %d not found", coupleID)
	}
	row.Married = true
	s.couples[coupleID] = row
	return nil
}

func (s *mockStore) Delete(_ context.Context, coupleID int32) error {
	if s.failOn == "Delete" {
		return fmt.Errorf("mock delete error")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.couples, coupleID)
	return nil
}

func (s *mockStore) FindByPlayer(_ context.Context, playerID int32) (*CoupleRow, error) {
	if s.failOn == "FindByPlayer" {
		return nil, fmt.Errorf("mock find error")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, row := range s.couples {
		if row.Player1ID == playerID || row.Player2ID == playerID {
			return &row, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (s *mockStore) addCouple(id, p1, p2 int32, married bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.couples[id] = CoupleRow{ID: id, Player1ID: p1, Player2ID: p2, Married: married}
	if s.nextID <= id {
		s.nextID = id + 1
	}
}

func makeTestPlayer(t *testing.T, objectID uint32, name string) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(objectID, int64(objectID)*10, int64(objectID)*100, name, 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer(%d, %q): %v", objectID, name, err)
	}
	return p
}

func TestNewManager(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	if mgr == nil {
		t.Fatal("NewManager() = nil")
	}
	if mgr.CoupleCount() != 0 {
		t.Errorf("CoupleCount() = %d; want 0", mgr.CoupleCount())
	}
}

func TestInit_LoadsCouples(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	store.addCouple(1, 100, 200, false)
	store.addCouple(2, 300, 400, true)

	mgr := NewManager(store)
	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	if mgr.CoupleCount() != 2 {
		t.Errorf("CoupleCount() = %d; want 2", mgr.CoupleCount())
	}

	c := mgr.CoupleByPlayer(100)
	if c == nil {
		t.Fatal("CoupleByPlayer(100) = nil")
	}
	if c.ID != 1 {
		t.Errorf("couple.ID = %d; want 1", c.ID)
	}

	c2 := mgr.CoupleByPlayer(400)
	if c2 == nil {
		t.Fatal("CoupleByPlayer(400) = nil")
	}
	if !c2.Married {
		t.Error("couple 2 should be married")
	}
}

func TestInit_Error(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	store.failOn = "LoadAll"

	mgr := NewManager(store)
	err := mgr.Init(context.Background())
	if err == nil {
		t.Fatal("Init() expected error")
	}
}

func TestEngage_Success(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")

	c, err := mgr.Engage(context.Background(), p1, p2)
	if err != nil {
		t.Fatalf("Engage() error: %v", err)
	}

	if c.ID == 0 {
		t.Error("couple.ID = 0; want non-zero")
	}
	if c.Married {
		t.Error("couple should not be married yet")
	}

	// Verify player state.
	if p1.PartnerID() != 200 {
		t.Errorf("p1.PartnerID() = %d; want 200", p1.PartnerID())
	}
	if p2.PartnerID() != 100 {
		t.Errorf("p2.PartnerID() = %d; want 100", p2.PartnerID())
	}
	if p1.CoupleID() != c.ID {
		t.Errorf("p1.CoupleID() = %d; want %d", p1.CoupleID(), c.ID)
	}

	// Verify manager indexes.
	if mgr.CoupleCount() != 1 {
		t.Errorf("CoupleCount() = %d; want 1", mgr.CoupleCount())
	}
	if mgr.CoupleByPlayer(100) == nil {
		t.Error("CoupleByPlayer(100) = nil")
	}
	if mgr.CoupleByPlayer(200) == nil {
		t.Error("CoupleByPlayer(200) = nil")
	}
}

func TestEngage_SelfEngage(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")

	_, err := mgr.Engage(context.Background(), p1, p1)
	if err != ErrSelfEngage {
		t.Errorf("Engage(self) error = %v; want ErrSelfEngage", err)
	}
}

func TestEngage_AlreadyEngaged(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")
	p3 := makeTestPlayer(t, 300, "Carol")

	if _, err := mgr.Engage(context.Background(), p1, p2); err != nil {
		t.Fatalf("Engage(p1,p2) error: %v", err)
	}

	// p1 already engaged — cannot engage again.
	_, err := mgr.Engage(context.Background(), p1, p3)
	if err != ErrAlreadyEngaged {
		t.Errorf("Engage(p1,p3) error = %v; want ErrAlreadyEngaged", err)
	}

	// p3 tries to engage with p2 who is already engaged.
	_, err = mgr.Engage(context.Background(), p3, p2)
	if err != ErrAlreadyEngaged {
		t.Errorf("Engage(p3,p2) error = %v; want ErrAlreadyEngaged", err)
	}
}

func TestEngage_DBError(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	store.failOn = "Create"
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")

	_, err := mgr.Engage(context.Background(), p1, p2)
	if err == nil {
		t.Fatal("Engage() expected error on DB failure")
	}

	// Manager state should be clean.
	if mgr.CoupleCount() != 0 {
		t.Errorf("CoupleCount() = %d; want 0 after DB error", mgr.CoupleCount())
	}
}

func TestEngage_NormalizesOrder(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	// p2 has lower ObjectID — should become Player1.
	p1 := makeTestPlayer(t, 500, "Alice")
	p2 := makeTestPlayer(t, 100, "Bob")

	c, err := mgr.Engage(context.Background(), p1, p2)
	if err != nil {
		t.Fatalf("Engage() error: %v", err)
	}

	if c.Player1ID != 100 || c.Player2ID != 500 {
		t.Errorf("couple = (%d, %d); want (100, 500)", c.Player1ID, c.Player2ID)
	}
}

func TestMarry_Success(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")

	c, err := mgr.Engage(context.Background(), p1, p2)
	if err != nil {
		t.Fatalf("Engage() error: %v", err)
	}

	if err := mgr.Marry(context.Background(), c.ID); err != nil {
		t.Fatalf("Marry() error: %v", err)
	}

	// Verify couple is married.
	updated := mgr.Couple(c.ID)
	if !updated.Married {
		t.Error("couple should be married")
	}
}

func TestMarry_NotFound(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	err := mgr.Marry(context.Background(), 999)
	if err != ErrCoupleNotFound {
		t.Errorf("Marry(999) error = %v; want ErrCoupleNotFound", err)
	}
}

func TestMarry_Idempotent(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")

	c, _ := mgr.Engage(context.Background(), p1, p2)
	if err := mgr.Marry(context.Background(), c.ID); err != nil {
		t.Fatalf("Marry() first call error: %v", err)
	}

	// Second call should be idempotent.
	if err := mgr.Marry(context.Background(), c.ID); err != nil {
		t.Errorf("Marry() second call error: %v", err)
	}
}

func TestMarry_DBError_Rollback(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")

	c, _ := mgr.Engage(context.Background(), p1, p2)

	store.failOn = "UpdateMarried"
	err := mgr.Marry(context.Background(), c.ID)
	if err == nil {
		t.Fatal("Marry() expected error on DB failure")
	}

	// In-memory state should be rolled back.
	updated := mgr.Couple(c.ID)
	if updated.Married {
		t.Error("couple should not be married after rollback")
	}
}

func TestDivorce_Success(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")

	c, _ := mgr.Engage(context.Background(), p1, p2)
	if err := mgr.Marry(context.Background(), c.ID); err != nil {
		t.Fatalf("Marry() error: %v", err)
	}

	if err := mgr.Divorce(context.Background(), c.ID, p1, p2); err != nil {
		t.Fatalf("Divorce() error: %v", err)
	}

	// Verify couple removed.
	if mgr.CoupleCount() != 0 {
		t.Errorf("CoupleCount() = %d; want 0", mgr.CoupleCount())
	}
	if mgr.CoupleByPlayer(100) != nil {
		t.Error("CoupleByPlayer(100) should be nil after divorce")
	}

	// Verify player state cleared.
	if p1.PartnerID() != 0 {
		t.Errorf("p1.PartnerID() = %d; want 0", p1.PartnerID())
	}
	if p2.IsMarried() {
		t.Error("p2 should not be married after divorce")
	}
}

func TestDivorce_NotFound(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	err := mgr.Divorce(context.Background(), 999, nil, nil)
	if err != ErrCoupleNotFound {
		t.Errorf("Divorce(999) error = %v; want ErrCoupleNotFound", err)
	}
}

func TestDivorce_PartnerOffline(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")

	c, _ := mgr.Engage(context.Background(), p1, p2)

	// p2 is offline (nil).
	if err := mgr.Divorce(context.Background(), c.ID, p1, nil); err != nil {
		t.Fatalf("Divorce() error: %v", err)
	}

	// p1 state should be cleared, p2 was nil so skipped.
	if p1.PartnerID() != 0 {
		t.Errorf("p1.PartnerID() = %d; want 0", p1.PartnerID())
	}
}

func TestDivorce_DBError_Rollback(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")

	c, _ := mgr.Engage(context.Background(), p1, p2)

	store.failOn = "Delete"
	err := mgr.Divorce(context.Background(), c.ID, p1, p2)
	if err == nil {
		t.Fatal("Divorce() expected error on DB failure")
	}

	// In-memory state should be rolled back.
	if mgr.CoupleCount() != 1 {
		t.Errorf("CoupleCount() = %d; want 1 after rollback", mgr.CoupleCount())
	}
	if mgr.CoupleByPlayer(100) == nil {
		t.Error("CoupleByPlayer(100) should still exist after rollback")
	}
}

func TestCoupleByPlayer_NotFound(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	if c := mgr.CoupleByPlayer(999); c != nil {
		t.Errorf("CoupleByPlayer(999) = %v; want nil", c)
	}
}

func TestRestorePlayerState(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	store.addCouple(1, 100, 200, true)

	mgr := NewManager(store)
	if err := mgr.Init(context.Background()); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	p := makeTestPlayer(t, 100, "Alice")

	mgr.RestorePlayerState(p)

	if p.CoupleID() != 1 {
		t.Errorf("CoupleID() = %d; want 1", p.CoupleID())
	}
	if p.PartnerID() != 200 {
		t.Errorf("PartnerID() = %d; want 200", p.PartnerID())
	}
	if !p.IsMarried() {
		t.Error("should be married")
	}
}

func TestRestorePlayerState_NoCoupleRecord(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p := makeTestPlayer(t, 999, "Nobody")
	mgr.RestorePlayerState(p)

	if p.CoupleID() != 0 {
		t.Errorf("CoupleID() = %d; want 0", p.CoupleID())
	}
}

func TestConcurrentEngageDivorce(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	var wg sync.WaitGroup

	// Create 50 couples concurrently.
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p1 := makeTestPlayer(t, uint32(i*2+1000), fmt.Sprintf("P%dA", i))
			p2 := makeTestPlayer(t, uint32(i*2+1001), fmt.Sprintf("P%dB", i))
			_, _ = mgr.Engage(context.Background(), p1, p2)
		}()
	}
	wg.Wait()

	count := mgr.CoupleCount()
	if count != 50 {
		t.Errorf("CoupleCount() = %d; want 50", count)
	}
}
