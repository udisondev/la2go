package wedding

import (
	"context"
	"testing"
)

func TestCanEngage_Success(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")

	if err := CanEngage(p1, p2, mgr); err != nil {
		t.Errorf("CanEngage() error = %v; want nil", err)
	}
}

func TestCanEngage_Self(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")

	if err := CanEngage(p1, p1, mgr); err != ErrSelfEngage {
		t.Errorf("CanEngage(self) error = %v; want ErrSelfEngage", err)
	}
}

func TestCanEngage_RequesterEngaged(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")
	p3 := makeTestPlayer(t, 300, "Carol")

	if _, err := mgr.Engage(context.Background(), p1, p2); err != nil {
		t.Fatalf("Engage() error: %v", err)
	}

	if err := CanEngage(p1, p3, mgr); err != ErrAlreadyEngaged {
		t.Errorf("CanEngage(engaged, free) error = %v; want ErrAlreadyEngaged", err)
	}
}

func TestCanEngage_TargetEngaged(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")
	p3 := makeTestPlayer(t, 300, "Carol")

	if _, err := mgr.Engage(context.Background(), p1, p2); err != nil {
		t.Fatalf("Engage() error: %v", err)
	}

	if err := CanEngage(p3, p2, mgr); err != ErrAlreadyEngaged {
		t.Errorf("CanEngage(free, engaged) error = %v; want ErrAlreadyEngaged", err)
	}
}

func TestCanEngage_TargetHasPendingRequest(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")

	// Simulate pending request on p2.
	p2.SetEngageRequest(true, 999)

	if err := CanEngage(p1, p2, mgr); err != ErrAlreadyEngaged {
		t.Errorf("CanEngage(free, pendingRequest) error = %v; want ErrAlreadyEngaged", err)
	}
}

func TestCanDivorce_Success(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")

	if _, err := mgr.Engage(context.Background(), p1, p2); err != nil {
		t.Fatalf("Engage() error: %v", err)
	}

	c, err := CanDivorce(p1, mgr)
	if err != nil {
		t.Errorf("CanDivorce() error = %v; want nil", err)
	}
	if c == nil {
		t.Fatal("CanDivorce() couple = nil")
	}
}

func TestCanDivorce_NotEngaged(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")

	_, err := CanDivorce(p1, mgr)
	if err != ErrNotEngaged {
		t.Errorf("CanDivorce() error = %v; want ErrNotEngaged", err)
	}
}

func TestCalcDivorcePenalty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		partnerAdena int64
		want         int64
	}{
		{"standard", 1_000_000, 200_000},
		{"zero", 0, 0},
		{"large", 250_000_000, 50_000_000},
		{"small", 100, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CalcDivorcePenalty(tt.partnerAdena)
			if got != tt.want {
				t.Errorf("CalcDivorcePenalty(%d) = %d; want %d", tt.partnerAdena, got, tt.want)
			}
		})
	}
}

func TestCanTeleportToPartner_Success(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")

	c, _ := mgr.Engage(context.Background(), p1, p2)
	if err := mgr.Marry(context.Background(), c.ID); err != nil {
		t.Fatalf("Marry() error: %v", err)
	}

	partnerID, err := CanTeleportToPartner(p1, mgr)
	if err != nil {
		t.Errorf("CanTeleportToPartner() error = %v; want nil", err)
	}
	if partnerID != 200 {
		t.Errorf("partnerID = %d; want 200", partnerID)
	}
}

func TestCanTeleportToPartner_NotMarried(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")
	p2 := makeTestPlayer(t, 200, "Bob")

	// Only engaged, not married.
	if _, err := mgr.Engage(context.Background(), p1, p2); err != nil {
		t.Fatalf("Engage() error: %v", err)
	}

	_, err := CanTeleportToPartner(p1, mgr)
	if err != ErrNotMarried {
		t.Errorf("CanTeleportToPartner() error = %v; want ErrNotMarried", err)
	}
}

func TestCanTeleportToPartner_NotEngaged(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	mgr := NewManager(store)

	p1 := makeTestPlayer(t, 100, "Alice")

	_, err := CanTeleportToPartner(p1, mgr)
	if err != ErrNotEngaged {
		t.Errorf("CanTeleportToPartner() error = %v; want ErrNotEngaged", err)
	}
}
