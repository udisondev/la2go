package model

import (
	"fmt"
	"sync"
)

const (
	// MaxPartyMembers is the maximum party size (leader + 8 members).
	// Java reference: Party.java PARTY_MEMBER_COUNT_LIMIT = 9.
	MaxPartyMembers = 9

	// LootRuleFinders — finders keepers loot distribution.
	LootRuleFinders = 0
	// LootRuleRandom — random loot distribution.
	LootRuleRandom = 1
	// LootRuleRandomSpoil — random loot distribution including spoil.
	LootRuleRandomSpoil = 2
	// LootRuleOrder — round-robin loot distribution.
	LootRuleOrder = 3
	// LootRuleOrderSpoil — round-robin loot distribution including spoil.
	LootRuleOrderSpoil = 4
)

// Party represents a group of players cooperating together.
// Thread-safe: all methods acquire internal mutex.
//
// Java reference: Party.java (L2J Mobius CT 0 Interlude).
type Party struct {
	mu       sync.RWMutex
	id       int32
	leader   *Player
	members  []*Player // leader всегда первый элемент
	lootRule int32
}

// NewParty creates a party with the given leader and loot rule.
// Leader is automatically added as first member.
func NewParty(id int32, leader *Player, lootRule int32) *Party {
	p := &Party{
		id:       id,
		leader:   leader,
		members:  make([]*Player, 0, MaxPartyMembers),
		lootRule: lootRule,
	}
	p.members = append(p.members, leader)
	return p
}

// ID returns immutable party ID.
func (p *Party) ID() int32 {
	return p.id
}

// Leader returns current party leader.
func (p *Party) Leader() *Player {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.leader
}

// SetLeader changes party leader to the given player.
// The new leader is swapped to index 0 in the members slice (Java: Party.java setLeader).
// Caller must ensure the player is already a party member.
func (p *Party) SetLeader(player *Player) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.leader = player
	// Swap the new leader to index 0 (Java behavior)
	for i, m := range p.members {
		if m.ObjectID() == player.ObjectID() {
			p.members[0], p.members[i] = p.members[i], p.members[0]
			break
		}
	}
}

// LootRule returns current loot distribution rule.
func (p *Party) LootRule() int32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lootRule
}

// SetLootRule changes loot distribution rule.
func (p *Party) SetLootRule(rule int32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lootRule = rule
}

// Members returns a snapshot copy of party members slice.
// Safe to iterate without holding the lock.
func (p *Party) Members() []*Player {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]*Player, len(p.members))
	copy(result, p.members)
	return result
}

// MemberCount returns the number of members in party.
func (p *Party) MemberCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.members)
}

// IsMember checks if a player with given objectID is in this party.
func (p *Party) IsMember(objectID uint32) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, m := range p.members {
		if m.ObjectID() == objectID {
			return true
		}
	}
	return false
}

// IsLeader checks if a player with given objectID is the party leader.
func (p *Party) IsLeader(objectID uint32) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.leader.ObjectID() == objectID
}

// AddMember adds a player to the party.
// Returns error if party is full or player is already a member.
func (p *Party) AddMember(player *Player) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.members) >= MaxPartyMembers {
		return fmt.Errorf("party full (max %d members)", MaxPartyMembers)
	}

	for _, m := range p.members {
		if m.ObjectID() == player.ObjectID() {
			return fmt.Errorf("player %s already in party", player.Name())
		}
	}

	p.members = append(p.members, player)
	return nil
}

// RemoveMember removes a player from the party by objectID.
// If the leader leaves, the next member becomes leader.
// Returns true if the party should be disbanded (fewer than 2 members remaining).
func (p *Party) RemoveMember(objectID uint32) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := -1
	for i, m := range p.members {
		if m.ObjectID() == objectID {
			idx = i
			break
		}
	}

	if idx < 0 {
		return false
	}

	// Удаляем без сохранения порядка не подходит -- нужен стабильный порядок для лута
	p.members = append(p.members[:idx], p.members[idx+1:]...)

	// Лидер ушел -- передаем лидерство следующему
	if p.leader.ObjectID() == objectID && len(p.members) > 0 {
		p.leader = p.members[0]
	}

	return len(p.members) < 2
}

// GetMember returns a member by objectID (nil if not found).
func (p *Party) GetMember(objectID uint32) *Player {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, m := range p.members {
		if m.ObjectID() == objectID {
			return m
		}
	}
	return nil
}

// bonusExpSP is the party XP/SP bonus multiplier table from Java Party.java.
// Index = memberCount - 1. Java: BONUS_EXP_SP = {1.0, 1.10, 1.20, 1.30, 1.40, 1.50, 2.0, 2.10, 2.20}.
var bonusExpSP = [...]float64{1.0, 1.10, 1.20, 1.30, 1.40, 1.50, 2.0, 2.10, 2.20}

// GetXPBonus returns party XP bonus multiplier based on member count.
// Java reference: Party.java getExpBonus() / getBaseExpSpBonus().
func (p *Party) GetXPBonus() float64 {
	p.mu.RLock()
	count := len(p.members)
	p.mu.RUnlock()

	return baseExpSpBonus(count)
}

// baseExpSpBonus returns the base XP/SP bonus for given member count.
// Java: getBaseExpSpBonus(int membersCount).
func baseExpSpBonus(count int) float64 {
	i := count - 1
	if i < 1 {
		return 1.0
	}
	if i >= len(bonusExpSP) {
		i = len(bonusExpSP) - 1
	}
	return bonusExpSP[i]
}

// MembersInRange returns members within squared distance from (x, y, z).
// Used for XP distribution -- only nearby members get their share.
// distSq is the squared distance threshold (no sqrt for performance).
func (p *Party) MembersInRange(x, y, z int32, distSq int64) []*Player {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*Player, 0, len(p.members))
	for _, m := range p.members {
		loc := m.Location()
		dx := int64(loc.X) - int64(x)
		dy := int64(loc.Y) - int64(y)
		dz := int64(loc.Z) - int64(z)
		if dx*dx+dy*dy+dz*dz <= distSq {
			result = append(result, m)
		}
	}
	return result
}
