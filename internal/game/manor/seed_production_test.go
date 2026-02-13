package manor

import (
	"sync"
	"testing"
)

func TestSeedProduction_Basic(t *testing.T) {
	t.Parallel()

	sp := NewSeedProduction(5016, 100, 1000, 100)

	if sp.SeedID() != 5016 {
		t.Errorf("SeedID() = %d; want 5016", sp.SeedID())
	}
	if sp.Amount() != 100 {
		t.Errorf("Amount() = %d; want 100", sp.Amount())
	}
	if sp.Price() != 1000 {
		t.Errorf("Price() = %d; want 1000", sp.Price())
	}
	if sp.StartAmount() != 100 {
		t.Errorf("StartAmount() = %d; want 100", sp.StartAmount())
	}
}

func TestSeedProduction_SetAmount(t *testing.T) {
	t.Parallel()

	sp := NewSeedProduction(5016, 100, 1000, 100)
	sp.SetAmount(50)

	if sp.Amount() != 50 {
		t.Errorf("Amount() = %d; want 50", sp.Amount())
	}
}

func TestSeedProduction_DecreaseAmount(t *testing.T) {
	t.Parallel()

	sp := NewSeedProduction(5016, 100, 1000, 100)

	if !sp.DecreaseAmount(30) {
		t.Error("DecreaseAmount(30) = false; want true")
	}
	if sp.Amount() != 70 {
		t.Errorf("Amount() = %d; want 70", sp.Amount())
	}

	if !sp.DecreaseAmount(70) {
		t.Error("DecreaseAmount(70) = false; want true")
	}
	if sp.Amount() != 0 {
		t.Errorf("Amount() = %d; want 0", sp.Amount())
	}

	// Попытка уменьшить ниже нуля.
	if sp.DecreaseAmount(1) {
		t.Error("DecreaseAmount(1) = true; want false (amount=0)")
	}
}

func TestSeedProduction_DecreaseAmount_InsufficientAmount(t *testing.T) {
	t.Parallel()

	sp := NewSeedProduction(5016, 10, 1000, 100)

	if sp.DecreaseAmount(11) {
		t.Error("DecreaseAmount(11) = true; want false (amount=10)")
	}
	if sp.Amount() != 10 {
		t.Errorf("Amount() = %d; want 10 (unchanged)", sp.Amount())
	}
}

func TestSeedProduction_DecreaseAmount_Concurrent(t *testing.T) {
	t.Parallel()

	sp := NewSeedProduction(5016, 1000, 1000, 1000)

	var wg sync.WaitGroup
	successCount := make(chan int, 10)

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count := 0
			for range 200 {
				if sp.DecreaseAmount(1) {
					count++
				}
			}
			successCount <- count
		}()
	}

	wg.Wait()
	close(successCount)

	total := 0
	for c := range successCount {
		total += c
	}

	// 10 goroutines × 200 attempts = 2000, but only 1000 seeds.
	if total != 1000 {
		t.Errorf("total successful decreases = %d; want 1000", total)
	}
	if sp.Amount() != 0 {
		t.Errorf("Amount() = %d; want 0", sp.Amount())
	}
}
