package model

import (
	"sync"
	"testing"
)

func TestAggroList_AddHate(t *testing.T) {
	list := NewAggroList()

	list.AddHate(1001, 50)
	list.AddHate(1001, 30)

	info := list.Get(1001)
	if info == nil {
		t.Fatal("expected AggroInfo for objectID 1001")
	}
	if info.Hate() != 80 {
		t.Errorf("Hate() = %d, want 80", info.Hate())
	}
}

func TestAggroList_AddDamage(t *testing.T) {
	list := NewAggroList()

	list.AddDamage(2001, 100)
	list.AddDamage(2001, 50)

	info := list.Get(2001)
	if info == nil {
		t.Fatal("expected AggroInfo for objectID 2001")
	}
	if info.Damage() != 150 {
		t.Errorf("Damage() = %d, want 150", info.Damage())
	}
}

func TestAggroList_GetMostHated(t *testing.T) {
	list := NewAggroList()

	list.AddHate(1001, 50)
	list.AddHate(1002, 100)
	list.AddHate(1003, 30)

	mostHated := list.GetMostHated()
	if mostHated != 1002 {
		t.Errorf("GetMostHated() = %d, want 1002", mostHated)
	}
}

func TestAggroList_GetMostHated_Empty(t *testing.T) {
	list := NewAggroList()

	mostHated := list.GetMostHated()
	if mostHated != 0 {
		t.Errorf("GetMostHated() on empty list = %d, want 0", mostHated)
	}
}

func TestAggroList_Remove(t *testing.T) {
	list := NewAggroList()

	list.AddHate(1001, 50)
	list.AddHate(1002, 100)

	list.Remove(1002)

	mostHated := list.GetMostHated()
	if mostHated != 1001 {
		t.Errorf("after Remove(1002), GetMostHated() = %d, want 1001", mostHated)
	}

	if list.Get(1002) != nil {
		t.Error("Get(1002) should return nil after Remove")
	}
}

func TestAggroList_Clear(t *testing.T) {
	list := NewAggroList()

	list.AddHate(1001, 50)
	list.AddHate(1002, 100)

	list.Clear()

	if !list.IsEmpty() {
		t.Error("IsEmpty() should return true after Clear")
	}
}

func TestAggroList_IsEmpty(t *testing.T) {
	list := NewAggroList()

	if !list.IsEmpty() {
		t.Error("new AggroList should be empty")
	}

	list.AddHate(1001, 1)
	if list.IsEmpty() {
		t.Error("AggroList with entry should not be empty")
	}
}

func TestAggroList_Concurrent(t *testing.T) {
	list := NewAggroList()
	var wg sync.WaitGroup

	// 10 goroutines adding hate concurrently
	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for range 100 {
				list.AddHate(uint32(id+1), 1)
			}
		}(i)
	}

	wg.Wait()

	// All 10 entries should exist
	for i := range 10 {
		info := list.Get(uint32(i + 1))
		if info == nil {
			t.Errorf("missing entry for objectID %d", i+1)
			continue
		}
		if info.Hate() != 100 {
			t.Errorf("objectID %d: Hate() = %d, want 100", i+1, info.Hate())
		}
	}
}

func TestCalcHateValue(t *testing.T) {
	tests := []struct {
		damage   int32
		npcLevel int32
		want     int64
	}{
		{100, 10, 588},   // (100*100)/(10+7) = 588
		{50, 1, 625},     // (50*100)/(1+7) = 625
		{200, 80, 229},   // (200*100)/(80+7) = 229
		{0, 10, 0},       // zero damage = zero hate
		{100, 0, 1250},   // level 0 → clamped to 1: (100*100)/(1+7) = 1250
	}

	for _, tt := range tests {
		got := CalcHateValue(tt.damage, tt.npcLevel)
		if got != tt.want {
			t.Errorf("CalcHateValue(%d, %d) = %d, want %d",
				tt.damage, tt.npcLevel, got, tt.want)
		}
	}
}
