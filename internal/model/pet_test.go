package model

import (
	"math"
	"sync"
	"testing"
)

// newTestPetSummon creates a Summon with SummonTypePet for pet tests.
func newTestPetSummon() *Summon {
	tmpl := newTestNpcTemplate()
	return NewSummon(
		6001,           // objectID
		1001,           // ownerID
		SummonTypePet,  // summonType
		12077,          // templateID
		tmpl,           // template
		"Wolf",         // name
		40,             // level
		3000, 1500,     // maxHP, maxMP
		200, 150, 100, 80, // pAtk, pDef, mAtk, mDef
	)
}

func TestNewPet(t *testing.T) {
	summon := newTestPetSummon()
	p := NewPet(summon, 7001, 12077, 50000, 70, 10000, 2.5)

	if p == nil {
		t.Fatal("NewPet returned nil")
	}
	if p.ControlItemID() != 7001 {
		t.Errorf("ControlItemID() = %d; want 7001", p.ControlItemID())
	}
	if p.NpcTemplateID() != 12077 {
		t.Errorf("NpcTemplateID() = %d; want 12077", p.NpcTemplateID())
	}
	if p.Experience() != 50000 {
		t.Errorf("Experience() = %d; want 50000", p.Experience())
	}
	if p.MaxLevel() != 70 {
		t.Errorf("MaxLevel() = %d; want 70", p.MaxLevel())
	}
	if p.CurrentFed() != 10000 {
		t.Errorf("CurrentFed() = %d; want 10000 (start full)", p.CurrentFed())
	}
	if p.MaxFed() != 10000 {
		t.Errorf("MaxFed() = %d; want 10000", p.MaxFed())
	}
	if p.FeedRate() != 2.5 {
		t.Errorf("FeedRate() = %f; want 2.5", p.FeedRate())
	}
	if p.Inventory() == nil {
		t.Errorf("Inventory() = nil; want non-nil")
	}
}

func TestPetAccessors(t *testing.T) {
	summon := newTestPetSummon()
	p := NewPet(summon, 8001, 999, 100000, 80, 5000, 3.0)

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"ControlItemID", p.ControlItemID(), uint32(8001)},
		{"NpcTemplateID", p.NpcTemplateID(), int32(999)},
		{"MaxLevel", p.MaxLevel(), int32(80)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v; want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestPetExperience(t *testing.T) {
	summon := newTestPetSummon()
	p := NewPet(summon, 7001, 12077, 1000, 70, 10000, 2.0)

	t.Run("initial", func(t *testing.T) {
		if p.Experience() != 1000 {
			t.Errorf("Experience() = %d; want 1000", p.Experience())
		}
	})

	t.Run("set", func(t *testing.T) {
		p.SetExperience(5000)
		if p.Experience() != 5000 {
			t.Errorf("Experience() = %d after Set(5000); want 5000", p.Experience())
		}
	})

	t.Run("set_negative_clamps_to_zero", func(t *testing.T) {
		p.SetExperience(-100)
		if p.Experience() != 0 {
			t.Errorf("Experience() = %d after Set(-100); want 0", p.Experience())
		}
	})

	t.Run("add_positive", func(t *testing.T) {
		p.SetExperience(1000)
		p.AddExperience(500)
		if p.Experience() != 1500 {
			t.Errorf("Experience() = %d after Add(500); want 1500", p.Experience())
		}
	})

	t.Run("add_negative", func(t *testing.T) {
		p.SetExperience(1000)
		p.AddExperience(-500)
		if p.Experience() != 500 {
			t.Errorf("Experience() = %d after Add(-500); want 500", p.Experience())
		}
	})

	t.Run("add_negative_below_zero_clamps", func(t *testing.T) {
		p.SetExperience(100)
		p.AddExperience(-500)
		if p.Experience() != 0 {
			t.Errorf("Experience() = %d after Add(-500) from 100; want 0", p.Experience())
		}
	})
}

func TestPetFeedSystem(t *testing.T) {
	t.Run("set_current_fed", func(t *testing.T) {
		summon := newTestPetSummon()
		p := NewPet(summon, 7001, 12077, 0, 70, 1000, 1.0)

		p.SetCurrentFed(500)
		if p.CurrentFed() != 500 {
			t.Errorf("CurrentFed() = %d after Set(500); want 500", p.CurrentFed())
		}
	})

	t.Run("set_current_fed_clamps_negative", func(t *testing.T) {
		summon := newTestPetSummon()
		p := NewPet(summon, 7001, 12077, 0, 70, 1000, 1.0)

		p.SetCurrentFed(-100)
		if p.CurrentFed() != 0 {
			t.Errorf("CurrentFed() = %d after Set(-100); want 0", p.CurrentFed())
		}
	})

	t.Run("set_current_fed_clamps_over_max", func(t *testing.T) {
		summon := newTestPetSummon()
		p := NewPet(summon, 7001, 12077, 0, 70, 1000, 1.0)

		p.SetCurrentFed(9999)
		if p.CurrentFed() != 1000 {
			t.Errorf("CurrentFed() = %d after Set(9999) with maxFed=1000; want 1000", p.CurrentFed())
		}
	})

	t.Run("consume_feed_not_hungry", func(t *testing.T) {
		summon := newTestPetSummon()
		p := NewPet(summon, 7001, 12077, 0, 70, 100, 10.0)

		hungry := p.ConsumeFeed()
		if hungry {
			t.Errorf("ConsumeFeed() = true; want false (currentFed should be %d)", p.CurrentFed())
		}
		if p.CurrentFed() != 90 {
			t.Errorf("CurrentFed() = %d after ConsumeFeed(); want 90", p.CurrentFed())
		}
	})

	t.Run("consume_feed_until_hungry", func(t *testing.T) {
		summon := newTestPetSummon()
		p := NewPet(summon, 7001, 12077, 0, 70, 100, 50.0)

		// Первый раз: 100 - 50 = 50, не голоден
		hungry := p.ConsumeFeed()
		if hungry {
			t.Errorf("first ConsumeFeed() = true; want false")
		}

		// Второй раз: 50 - 50 = 0, голоден
		hungry = p.ConsumeFeed()
		if !hungry {
			t.Errorf("second ConsumeFeed() = false; want true (fed=0)")
		}
		if p.CurrentFed() != 0 {
			t.Errorf("CurrentFed() = %d; want 0", p.CurrentFed())
		}
	})

	t.Run("consume_feed_clamps_to_zero", func(t *testing.T) {
		summon := newTestPetSummon()
		p := NewPet(summon, 7001, 12077, 0, 70, 100, 200.0)

		// Первый раз: 100 - 200 = -100 -> clamp 0
		hungry := p.ConsumeFeed()
		if !hungry {
			t.Errorf("ConsumeFeed() = false; want true (rate > current)")
		}
		if p.CurrentFed() != 0 {
			t.Errorf("CurrentFed() = %d; want 0 (clamped)", p.CurrentFed())
		}
	})
}

func TestPetFedPercentage(t *testing.T) {
	tests := []struct {
		name       string
		currentFed int32
		maxFed     int32
		want       float64
	}{
		{"full", 1000, 1000, 1.0},
		{"half", 500, 1000, 0.5},
		{"empty", 0, 1000, 0.0},
		{"quarter", 250, 1000, 0.25},
		{"zero_max", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summon := newTestPetSummon()
			p := NewPet(summon, 7001, 12077, 0, 70, tt.maxFed, 1.0)
			p.SetCurrentFed(tt.currentFed)

			got := p.FedPercentage()
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("FedPercentage() = %f; want %f", got, tt.want)
			}
		})
	}
}

func TestPetUpdateFeedStats(t *testing.T) {
	summon := newTestPetSummon()
	p := NewPet(summon, 7001, 12077, 0, 70, 1000, 2.0)

	// currentFed = maxFed = 1000 после создания.
	// Увеличиваем maxFed
	p.UpdateFeedStats(2000, 5.0)
	if p.MaxFed() != 2000 {
		t.Errorf("MaxFed() = %d after UpdateFeedStats(2000, 5.0); want 2000", p.MaxFed())
	}
	if p.FeedRate() != 5.0 {
		t.Errorf("FeedRate() = %f; want 5.0", p.FeedRate())
	}
	// currentFed не должен измениться (1000 < 2000)
	if p.CurrentFed() != 1000 {
		t.Errorf("CurrentFed() = %d; want 1000 (unchanged, within new max)", p.CurrentFed())
	}

	// Уменьшаем maxFed ниже текущего currentFed
	p.UpdateFeedStats(500, 1.0)
	if p.MaxFed() != 500 {
		t.Errorf("MaxFed() = %d after UpdateFeedStats(500, 1.0); want 500", p.MaxFed())
	}
	if p.CurrentFed() != 500 {
		t.Errorf("CurrentFed() = %d; want 500 (clamped to new maxFed)", p.CurrentFed())
	}
}

func TestPetInventory(t *testing.T) {
	summon := newTestPetSummon()
	p := NewPet(summon, 7001, 12077, 0, 70, 1000, 1.0)

	inv := p.Inventory()
	if inv == nil {
		t.Fatal("Inventory() = nil; want non-nil")
	}

	// Вызываем ещё раз -- тот же указатель
	inv2 := p.Inventory()
	if inv != inv2 {
		t.Errorf("Inventory() returned different pointer on second call")
	}
}

func TestPetName(t *testing.T) {
	summon := newTestPetSummon()
	p := NewPet(summon, 7001, 12077, 0, 70, 1000, 1.0)

	// Имя из summon
	if p.PetName() != "Wolf" {
		t.Errorf("PetName() = %q; want %q", p.PetName(), "Wolf")
	}

	// Установка нового имени
	p.SetPetName("Fluffy")
	if p.PetName() != "Fluffy" {
		t.Errorf("PetName() = %q after SetPetName(%q); want %q", p.PetName(), "Fluffy", "Fluffy")
	}

	// Name() тоже должен возвращать то же самое (делегация к Character)
	if p.Name() != "Fluffy" {
		t.Errorf("Name() = %q; want %q (should match PetName)", p.Name(), "Fluffy")
	}
}

func TestPetIsRespawned(t *testing.T) {
	summon := newTestPetSummon()
	p := NewPet(summon, 7001, 12077, 0, 70, 1000, 1.0)

	if p.IsRespawned() {
		t.Errorf("IsRespawned() = true; want false (default)")
	}

	p.SetRespawned(true)
	if !p.IsRespawned() {
		t.Errorf("IsRespawned() = false after SetRespawned(true); want true")
	}

	p.SetRespawned(false)
	if p.IsRespawned() {
		t.Errorf("IsRespawned() = true after SetRespawned(false); want false")
	}
}

func TestPetWorldObjectData(t *testing.T) {
	summon := newTestPetSummon()
	p := NewPet(summon, 7001, 12077, 0, 70, 1000, 1.0)

	data := summon.WorldObject.Data
	if data == nil {
		t.Fatal("WorldObject.Data = nil; want *Pet")
	}

	// Data должен указывать на Pet, не на Summon
	petData, ok := data.(*Pet)
	if !ok {
		t.Fatalf("WorldObject.Data type = %T; want *Pet", data)
	}
	if petData != p {
		t.Errorf("WorldObject.Data pointer = %p; want %p (same pet)", petData, p)
	}

	// Не должен быть *Summon
	_, isSummon := data.(*Summon)
	if isSummon {
		t.Errorf("WorldObject.Data asserts as *Summon; want only *Pet")
	}
}

func TestPetInheritsSummonMethods(t *testing.T) {
	summon := newTestPetSummon()
	p := NewPet(summon, 7001, 12077, 0, 70, 1000, 1.0)

	// IsPet наследуется от Summon
	if !p.IsPet() {
		t.Errorf("IsPet() = false; want true (pet type summon)")
	}
	if p.IsServitor() {
		t.Errorf("IsServitor() = true; want false (pet type summon)")
	}

	// OwnerID наследуется
	if p.OwnerID() != 1001 {
		t.Errorf("OwnerID() = %d; want 1001", p.OwnerID())
	}

	// Follow наследуется
	if !p.IsFollowing() {
		t.Errorf("IsFollowing() = false; want true (default)")
	}
	p.SetFollow(false)
	if p.IsFollowing() {
		t.Errorf("IsFollowing() = true after SetFollow(false); want false")
	}

	// Target наследуется
	p.SetTarget(99)
	if p.Target() != 99 {
		t.Errorf("Target() = %d; want 99", p.Target())
	}
	p.ClearTarget()
	if p.Target() != 0 {
		t.Errorf("Target() = %d after ClearTarget(); want 0", p.Target())
	}

	// Intention наследуется
	p.SetIntention(IntentionAttack)
	if p.Intention() != IntentionAttack {
		t.Errorf("Intention() = %v; want IntentionAttack", p.Intention())
	}
}

func TestPetConcurrentExperience(t *testing.T) {
	summon := newTestPetSummon()
	p := NewPet(summon, 7001, 12077, 0, 70, 10000, 1.0)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for range goroutines {
		go func() {
			defer wg.Done()
			p.AddExperience(100)
		}()
	}

	for range goroutines {
		go func() {
			defer wg.Done()
			_ = p.Experience()
		}()
	}

	wg.Wait()

	// 100 горутин добавили по 100 опыта
	got := p.Experience()
	want := int64(goroutines * 100)
	if got != want {
		t.Errorf("Experience() = %d after %d concurrent AddExperience(100); want %d", got, goroutines, want)
	}
}

func TestPetConcurrentFeed(t *testing.T) {
	summon := newTestPetSummon()
	p := NewPet(summon, 7001, 12077, 0, 70, 100000, 1.0)

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	for range goroutines {
		go func() {
			defer wg.Done()
			p.ConsumeFeed()
		}()
	}

	for range goroutines {
		go func() {
			defer wg.Done()
			_ = p.CurrentFed()
			_ = p.MaxFed()
			_ = p.FedPercentage()
		}()
	}

	for range goroutines {
		go func() {
			defer wg.Done()
			p.SetCurrentFed(500)
		}()
	}

	wg.Wait()
	// Если нет data race — тест пройден.
}

func TestPetConcurrentRespawned(t *testing.T) {
	summon := newTestPetSummon()
	p := NewPet(summon, 7001, 12077, 0, 70, 1000, 1.0)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for range goroutines {
		go func() {
			defer wg.Done()
			p.SetRespawned(true)
			p.SetRespawned(false)
		}()
	}

	for range goroutines {
		go func() {
			defer wg.Done()
			_ = p.IsRespawned()
		}()
	}

	wg.Wait()
}

func TestPetFeedRateAccessor(t *testing.T) {
	tests := []struct {
		name     string
		feedRate float64
	}{
		{"zero", 0.0},
		{"small", 0.5},
		{"normal", 2.0},
		{"large", 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summon := newTestPetSummon()
			p := NewPet(summon, 7001, 12077, 0, 70, 1000, tt.feedRate)

			if got := p.FeedRate(); math.Abs(got-tt.feedRate) > 1e-9 {
				t.Errorf("FeedRate() = %f; want %f", got, tt.feedRate)
			}
		})
	}
}

func TestPetMaxLevel(t *testing.T) {
	tests := []struct {
		name     string
		maxLevel int32
	}{
		{"wolf", 55},
		{"hatchling", 70},
		{"strider", 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summon := newTestPetSummon()
			p := NewPet(summon, 7001, 12077, 0, tt.maxLevel, 1000, 1.0)

			if got := p.MaxLevel(); got != tt.maxLevel {
				t.Errorf("MaxLevel() = %d; want %d", got, tt.maxLevel)
			}
		})
	}
}
