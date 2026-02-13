package manor

import (
	"sync"
	"testing"
)

func TestCropProcure_Basic(t *testing.T) {
	t.Parallel()

	cp := NewCropProcure(5073, 200, 1, 200, 500)

	if cp.CropID() != 5073 {
		t.Errorf("CropID() = %d; want 5073", cp.CropID())
	}
	if cp.Amount() != 200 {
		t.Errorf("Amount() = %d; want 200", cp.Amount())
	}
	if cp.Price() != 500 {
		t.Errorf("Price() = %d; want 500", cp.Price())
	}
	if cp.StartAmount() != 200 {
		t.Errorf("StartAmount() = %d; want 200", cp.StartAmount())
	}
	if cp.RewardType() != 1 {
		t.Errorf("RewardType() = %d; want 1", cp.RewardType())
	}
}

func TestCropProcure_SetAmount(t *testing.T) {
	t.Parallel()

	cp := NewCropProcure(5073, 200, 1, 200, 500)
	cp.SetAmount(100)

	if cp.Amount() != 100 {
		t.Errorf("Amount() = %d; want 100", cp.Amount())
	}
}

func TestCropProcure_DecreaseAmount(t *testing.T) {
	t.Parallel()

	cp := NewCropProcure(5073, 50, 2, 50, 500)

	if !cp.DecreaseAmount(20) {
		t.Error("DecreaseAmount(20) = false; want true")
	}
	if cp.Amount() != 30 {
		t.Errorf("Amount() = %d; want 30", cp.Amount())
	}

	if cp.DecreaseAmount(31) {
		t.Error("DecreaseAmount(31) = true; want false (amount=30)")
	}
	if cp.Amount() != 30 {
		t.Errorf("Amount() = %d; want 30 (unchanged)", cp.Amount())
	}
}

func TestCropProcure_DecreaseAmount_Concurrent(t *testing.T) {
	t.Parallel()

	cp := NewCropProcure(5073, 500, 1, 500, 100)

	var wg sync.WaitGroup
	successCount := make(chan int, 5)

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count := 0
			for range 200 {
				if cp.DecreaseAmount(1) {
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

	if total != 500 {
		t.Errorf("total successful decreases = %d; want 500", total)
	}
	if cp.Amount() != 0 {
		t.Errorf("Amount() = %d; want 0", cp.Amount())
	}
}
