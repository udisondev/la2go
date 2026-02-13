package party

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/udisondev/la2go/internal/model"
)

func newTestPlayer(t *testing.T, objectID uint32, name string) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(objectID, int64(objectID), 1, name, 20, 0, 0)
	require.NoError(t, err, "NewPlayer(%d, %s)", objectID, name)
	return p
}

func TestManager_CreateParty(t *testing.T) {
	mgr := NewManager()
	leader := newTestPlayer(t, 1, "Leader")

	party := mgr.CreateParty(leader, model.LootRuleFinders)

	require.NotNil(t, party)
	assert.Equal(t, int32(1), party.ID())
	assert.Equal(t, leader, party.Leader())
	assert.Equal(t, int32(model.LootRuleFinders), party.LootRule())
	assert.Equal(t, 1, party.MemberCount())
	assert.Equal(t, 1, mgr.PartyCount())
}

func TestManager_GetParty(t *testing.T) {
	mgr := NewManager()
	leader := newTestPlayer(t, 1, "Leader")

	party := mgr.CreateParty(leader, model.LootRuleRandom)

	got := mgr.GetParty(party.ID())
	assert.Equal(t, party, got)

	// Несуществующая партия
	assert.Nil(t, mgr.GetParty(999))
}

func TestManager_DisbandParty(t *testing.T) {
	mgr := NewManager()
	leader := newTestPlayer(t, 1, "Leader")

	party := mgr.CreateParty(leader, model.LootRuleFinders)
	assert.Equal(t, 1, mgr.PartyCount())

	mgr.DisbandParty(party.ID())
	assert.Equal(t, 0, mgr.PartyCount())
	assert.Nil(t, mgr.GetParty(party.ID()))
}

func TestManager_DisbandParty_NonExistent(t *testing.T) {
	mgr := NewManager()

	// Удаление несуществующей партии не вызывает ошибку
	mgr.DisbandParty(999)
	assert.Equal(t, 0, mgr.PartyCount())
}

func TestManager_MultipleParties(t *testing.T) {
	mgr := NewManager()

	leader1 := newTestPlayer(t, 1, "Leader1")
	leader2 := newTestPlayer(t, 2, "Leader2")
	leader3 := newTestPlayer(t, 3, "Leader3")

	p1 := mgr.CreateParty(leader1, model.LootRuleFinders)
	p2 := mgr.CreateParty(leader2, model.LootRuleRandom)
	p3 := mgr.CreateParty(leader3, model.LootRuleOrderSpoil)

	assert.Equal(t, 3, mgr.PartyCount())

	// Каждая партия имеет уникальный ID
	assert.NotEqual(t, p1.ID(), p2.ID())
	assert.NotEqual(t, p2.ID(), p3.ID())
	assert.NotEqual(t, p1.ID(), p3.ID())

	// Получение каждой партии
	assert.Equal(t, p1, mgr.GetParty(p1.ID()))
	assert.Equal(t, p2, mgr.GetParty(p2.ID()))
	assert.Equal(t, p3, mgr.GetParty(p3.ID()))

	// Удаляем одну -- другие остаются
	mgr.DisbandParty(p2.ID())
	assert.Equal(t, 2, mgr.PartyCount())
	assert.Nil(t, mgr.GetParty(p2.ID()))
	assert.NotNil(t, mgr.GetParty(p1.ID()))
	assert.NotNil(t, mgr.GetParty(p3.ID()))
}

func TestManager_UniqueIDs(t *testing.T) {
	mgr := NewManager()
	ids := make(map[int32]struct{})

	for i := range 100 {
		leader := newTestPlayer(t, uint32(i+100), "Leader")
		party := mgr.CreateParty(leader, model.LootRuleFinders)

		_, exists := ids[party.ID()]
		assert.False(t, exists, "duplicate party ID: %d", party.ID())
		ids[party.ID()] = struct{}{}
	}
}

func TestManager_ConcurrentCreateParty(t *testing.T) {
	mgr := NewManager()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	parties := make([]*model.Party, goroutines)
	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()
			leader := newTestPlayer(t, uint32(idx+1000), "ConcurrentLeader")
			parties[idx] = mgr.CreateParty(leader, model.LootRuleFinders)
		}(i)
	}

	wg.Wait()

	assert.Equal(t, goroutines, mgr.PartyCount())

	// Все ID уникальны
	ids := make(map[int32]struct{})
	for _, p := range parties {
		require.NotNil(t, p)
		_, exists := ids[p.ID()]
		assert.False(t, exists, "duplicate party ID: %d", p.ID())
		ids[p.ID()] = struct{}{}
	}
}

func TestManager_ConcurrentCreateAndDisband(t *testing.T) {
	mgr := NewManager()

	const goroutines = 30
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Конкурентное создание
	var mu sync.Mutex
	var partyIDs []int32
	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()
			leader := newTestPlayer(t, uint32(idx+2000), "Leader")
			p := mgr.CreateParty(leader, model.LootRuleFinders)
			mu.Lock()
			partyIDs = append(partyIDs, p.ID())
			mu.Unlock()
		}(i)
	}

	// Конкурентное удаление (может удалять несуществующие -- допустимо)
	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()
			mgr.DisbandParty(int32(idx + 1))
		}(i)
	}

	wg.Wait()

	// Количество партий не отрицательное
	assert.GreaterOrEqual(t, mgr.PartyCount(), 0)
}
