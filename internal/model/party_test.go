package model

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestPartyPlayer -- хелпер для создания тестового игрока с заданными координатами.
func newTestPartyPlayer(t *testing.T, objectID uint32, name string) *Player {
	t.Helper()
	p, err := NewPlayer(objectID, int64(objectID), 1, name, 20, 0, 0)
	require.NoError(t, err, "NewPlayer(%d, %s)", objectID, name)
	return p
}

func TestNewParty(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	party := NewParty(100, leader, LootRuleFinders)

	assert.Equal(t, int32(100), party.ID())
	assert.Equal(t, leader, party.Leader())
	assert.Equal(t, int32(LootRuleFinders), party.LootRule())
	assert.Equal(t, 1, party.MemberCount())
	assert.True(t, party.IsMember(1))
	assert.False(t, party.IsMember(999))
}

func TestParty_AddMember(t *testing.T) {
	tests := []struct {
		name      string
		addCount  int
		wantErr   bool
		wantCount int
	}{
		{
			name:      "add one member",
			addCount:  1,
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "fill to max",
			addCount:  MaxPartyMembers - 1, // leader + 8 = 9
			wantErr:   false,
			wantCount: MaxPartyMembers,
		},
		{
			name:      "overflow by one",
			addCount:  MaxPartyMembers, // leader + 9 = 10, overflow
			wantErr:   true,
			wantCount: MaxPartyMembers,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leader := newTestPartyPlayer(t, 1, "Leader")
			party := NewParty(1, leader, LootRuleFinders)

			var lastErr error
			for i := range tt.addCount {
				member := newTestPartyPlayer(t, uint32(i+10), "Member"+string(rune('A'+i)))
				if err := party.AddMember(member); err != nil {
					lastErr = err
				}
			}

			if tt.wantErr {
				assert.Error(t, lastErr, "expected error when adding too many members")
			} else {
				assert.NoError(t, lastErr)
			}
			assert.Equal(t, tt.wantCount, party.MemberCount())
		})
	}
}

func TestParty_AddMember_Duplicate(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	member := newTestPartyPlayer(t, 2, "Member")
	party := NewParty(1, leader, LootRuleFinders)

	require.NoError(t, party.AddMember(member))
	err := party.AddMember(member)

	assert.Error(t, err, "duplicate member should return error")
	assert.Equal(t, 2, party.MemberCount(), "member count should not increase on duplicate")
}

func TestParty_RemoveMember(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	member1 := newTestPartyPlayer(t, 2, "Member1")
	member2 := newTestPartyPlayer(t, 3, "Member2")
	party := NewParty(1, leader, LootRuleFinders)
	require.NoError(t, party.AddMember(member1))
	require.NoError(t, party.AddMember(member2))

	// Удаляем обычного участника -- партия не распадается
	shouldDisband := party.RemoveMember(2)
	assert.False(t, shouldDisband, "party should not disband with 2 members remaining")
	assert.Equal(t, 2, party.MemberCount())
	assert.False(t, party.IsMember(2))
}

func TestParty_RemoveMember_LeaderReassignment(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	member1 := newTestPartyPlayer(t, 2, "Member1")
	member2 := newTestPartyPlayer(t, 3, "Member2")
	party := NewParty(1, leader, LootRuleFinders)
	require.NoError(t, party.AddMember(member1))
	require.NoError(t, party.AddMember(member2))

	// Лидер уходит -- лидерство передается следующему
	shouldDisband := party.RemoveMember(1)
	assert.False(t, shouldDisband)
	assert.Equal(t, member1.ObjectID(), party.Leader().ObjectID(),
		"leader should be reassigned to next member")
}

func TestParty_RemoveMember_Disband(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	member := newTestPartyPlayer(t, 2, "Member")
	party := NewParty(1, leader, LootRuleFinders)
	require.NoError(t, party.AddMember(member))

	// Удаляем участника -- остается 1, надо распустить
	shouldDisband := party.RemoveMember(2)
	assert.True(t, shouldDisband, "party should disband with <2 members")
}

func TestParty_RemoveMember_NonExistent(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	party := NewParty(1, leader, LootRuleFinders)

	shouldDisband := party.RemoveMember(999)
	assert.False(t, shouldDisband, "removing non-existent member should return false")
	assert.Equal(t, 1, party.MemberCount())
}

func TestParty_IsMember(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	member := newTestPartyPlayer(t, 2, "Member")
	party := NewParty(1, leader, LootRuleFinders)
	require.NoError(t, party.AddMember(member))

	tests := []struct {
		name     string
		objectID uint32
		want     bool
	}{
		{"leader is member", 1, true},
		{"added member", 2, true},
		{"not a member", 999, false},
		{"zero objectID", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, party.IsMember(tt.objectID))
		})
	}
}

func TestParty_IsLeader(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	member := newTestPartyPlayer(t, 2, "Member")
	party := NewParty(1, leader, LootRuleFinders)
	require.NoError(t, party.AddMember(member))

	assert.True(t, party.IsLeader(1))
	assert.False(t, party.IsLeader(2))
	assert.False(t, party.IsLeader(999))
}

func TestParty_GetMember(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	member := newTestPartyPlayer(t, 2, "Member")
	party := NewParty(1, leader, LootRuleFinders)
	require.NoError(t, party.AddMember(member))

	got := party.GetMember(2)
	assert.Equal(t, member, got)

	assert.Nil(t, party.GetMember(999))
}

func TestParty_Members_ReturnsCopy(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	member := newTestPartyPlayer(t, 2, "Member")
	party := NewParty(1, leader, LootRuleFinders)
	require.NoError(t, party.AddMember(member))

	members1 := party.Members()
	members2 := party.Members()

	// Разные слайсы
	assert.Equal(t, len(members1), len(members2))
	members1[0] = nil // мутация не должна влиять на оригинал
	assert.NotNil(t, party.Members()[0], "mutating returned slice should not affect party")
}

func TestParty_SetLeader(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	member := newTestPartyPlayer(t, 2, "NewLeader")
	party := NewParty(1, leader, LootRuleFinders)
	require.NoError(t, party.AddMember(member))

	party.SetLeader(member)
	assert.Equal(t, member.ObjectID(), party.Leader().ObjectID())
}

func TestParty_SetLootRule(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	party := NewParty(1, leader, LootRuleFinders)

	assert.Equal(t, int32(LootRuleFinders), party.LootRule())

	party.SetLootRule(LootRuleRandom)
	assert.Equal(t, int32(LootRuleRandom), party.LootRule())

	party.SetLootRule(LootRuleOrderSpoil)
	assert.Equal(t, int32(LootRuleOrderSpoil), party.LootRule())
}

func TestParty_GetXPBonus(t *testing.T) {
	tests := []struct {
		name       string
		memberCnt  int // сколько добавить помимо лидера
		wantBonus  float64
	}{
		// Java: BONUS_EXP_SP = {1.0, 1.10, 1.20, 1.30, 1.40, 1.50, 2.0, 2.10, 2.20}
		// Index = memberCount - 1
		{"solo (1 member)", 0, 1.0},
		{"2 members", 1, 1.10},
		{"3 members", 2, 1.20},
		{"4 members", 3, 1.30},
		{"5 members", 4, 1.40},
		{"6 members", 5, 1.50},
		{"7 members", 6, 2.0},
		{"8 members", 7, 2.10},
		{"9 members (max)", 8, 2.20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leader := newTestPartyPlayer(t, 1, "Leader")
			party := NewParty(1, leader, LootRuleFinders)

			for i := range tt.memberCnt {
				m := newTestPartyPlayer(t, uint32(i+10), "M"+string(rune('A'+i)))
				require.NoError(t, party.AddMember(m))
			}

			assert.InDelta(t, tt.wantBonus, party.GetXPBonus(), 0.001,
				"XP bonus for %d members", tt.memberCnt+1)
		})
	}
}

func TestParty_MembersInRange(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	leader.SetLocation(NewLocation(0, 0, 0, 0))

	nearby := newTestPartyPlayer(t, 2, "Nearby")
	nearby.SetLocation(NewLocation(100, 100, 0, 0))

	far := newTestPartyPlayer(t, 3, "Far")
	far.SetLocation(NewLocation(10000, 10000, 0, 0))

	party := NewParty(1, leader, LootRuleFinders)
	require.NoError(t, party.AddMember(nearby))
	require.NoError(t, party.AddMember(far))

	// distSq = 1000*1000 = 1000000; nearby at (100,100) -> distSq = 20000 < 1000000
	// far at (10000,10000) -> distSq = 200000000 > 1000000
	inRange := party.MembersInRange(0, 0, 0, 1000*1000)

	assert.Len(t, inRange, 2, "leader and nearby should be in range")

	ids := make([]uint32, len(inRange))
	for i, m := range inRange {
		ids[i] = m.ObjectID()
	}
	assert.Contains(t, ids, uint32(1), "leader should be in range")
	assert.Contains(t, ids, uint32(2), "nearby player should be in range")
}

func TestParty_MembersInRange_AllFar(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	leader.SetLocation(NewLocation(0, 0, 0, 0))

	far := newTestPartyPlayer(t, 2, "Far")
	far.SetLocation(NewLocation(50000, 50000, 0, 0))

	party := NewParty(1, leader, LootRuleFinders)
	require.NoError(t, party.AddMember(far))

	// Только лидер в радиусе 100 от (0,0,0)
	inRange := party.MembersInRange(0, 0, 0, 100*100)
	assert.Len(t, inRange, 1)
	assert.Equal(t, uint32(1), inRange[0].ObjectID())
}

func TestParty_ConcurrentAddRemove(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	party := NewParty(1, leader, LootRuleFinders)

	const goroutines = 20
	const iterations = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // add + remove goroutines

	// Конкурентное добавление
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for i := range iterations {
				objID := uint32(1000 + id*100 + i)
				m := newTestPartyPlayer(t, objID, "Concurrent")
				// Ошибки ожидаемы (полная группа, дубликаты)
				_ = party.AddMember(m)
			}
		}(g)
	}

	// Конкурентное удаление
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for i := range iterations {
				objID := uint32(1000 + id*100 + i)
				party.RemoveMember(objID)
			}
		}(g)
	}

	wg.Wait()

	// Лидер всегда должен быть валидным
	assert.NotNil(t, party.Leader())
	count := party.MemberCount()
	assert.GreaterOrEqual(t, count, 0, "member count should not be negative")
	assert.LessOrEqual(t, count, MaxPartyMembers, "member count should not exceed max")
}

func TestParty_ConcurrentReads(t *testing.T) {
	leader := newTestPartyPlayer(t, 1, "Leader")
	party := NewParty(1, leader, LootRuleFinders)

	// Добавляем несколько участников
	for i := range 5 {
		m := newTestPartyPlayer(t, uint32(i+10), "Member")
		require.NoError(t, party.AddMember(m))
	}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			for range 100 {
				_ = party.Leader()
				_ = party.MemberCount()
				_ = party.Members()
				_ = party.IsMember(1)
				_ = party.LootRule()
				_ = party.GetXPBonus()
				_ = party.MembersInRange(0, 0, 0, 1000*1000)
			}
		}()
	}

	wg.Wait()
}
