package clan

import (
	"errors"
	"maps"
	"strings"
	"sync"
	"sync/atomic"
)

// Clan level limits.
const (
	MaxClanLevel = 8
	MinClanLevel = 0
)

// Max members by clan level.
var maxMembersByLevel = [MaxClanLevel + 1]int32{
	10, // Level 0
	15, // Level 1
	20, // Level 2
	30, // Level 3
	40, // Level 4
	40, // Level 5
	40, // Level 6
	40, // Level 7
	40, // Level 8
}

// Level upgrade requirements.
type LevelRequirement struct {
	SP         int64
	Adena      int64
	ItemID     int32 // 0 if no item needed
	ItemCount  int32
	Reputation int32 // For levels 5-8
	Members    int32 // Minimum member count for levels 5-8
}

// LevelRequirements maps clan level -> requirements to reach that level.
var LevelRequirements = map[int32]LevelRequirement{
	1: {SP: 20_000, Adena: 650_000},
	2: {SP: 100_000, Adena: 2_500_000},
	3: {SP: 350_000, Adena: 5_000_000, ItemID: 1419, ItemCount: 1}, // Blood Mark
	4: {SP: 1_000_000, Adena: 10_000_000, ItemID: 3874, ItemCount: 1}, // Alliance Manifesto
	5: {Reputation: 10_000, Members: 30},
	6: {Reputation: 5_000, Members: 30},
	7: {Reputation: 10_000, Members: 50},
	8: {Reputation: 20_000, Members: 80},
}

// Alliance penalty types.
const (
	AllyPenaltyNone         int32 = 0
	AllyPenaltyClanLeaved   int32 = 1
	AllyPenaltyClanDismissed int32 = 2
	AllyPenaltyDismissClan  int32 = 3
	AllyPenaltyDissolveAlly int32 = 4
)

// Alliance limits.
const (
	MaxClansInAlly = 3
	MaxAllyNameLen = 16
	MinAllyNameLen = 2
)

// Common clan errors.
var (
	ErrClanFull         = errors.New("clan is full")
	ErrAlreadyInClan    = errors.New("already in clan")
	ErrNotInClan        = errors.New("not in clan")
	ErrInsufficientRank = errors.New("insufficient rank")
	ErrSubPledgeFull    = errors.New("sub-pledge is full")
	ErrClanLevelTooLow  = errors.New("clan level too low")
	ErrAlreadyAtWar     = errors.New("already at war with this clan")
	ErrNotAtWar         = errors.New("not at war with this clan")
	ErrCannotDisband    = errors.New("cannot disband clan")
	ErrAlreadyInAlly    = errors.New("already in alliance")
	ErrNotInAlly        = errors.New("not in alliance")
	ErrNotAllyLeader    = errors.New("not alliance leader")
	ErrAllyFull         = errors.New("alliance is full")
	ErrAllyNameTaken    = errors.New("alliance name already taken")
	ErrAllyNameInvalid  = errors.New("invalid alliance name")
	ErrAllyPenalty      = errors.New("alliance penalty active")
)

// Clan represents a player clan.
// Thread-safe: all mutable fields protected by mu.
type Clan struct {
	mu sync.RWMutex

	id         int32
	name       string
	leaderID   int64 // Object ID of clan leader
	level      int32
	crestID    int32
	largeCrest int32
	allyID     int32
	allyCrest  int32
	allyName              string
	allyPenaltyExpiryTime int64 // Unix millis, 0 if no penalty
	allyPenaltyType       int32 // 0=none, 1=clan left, 2=clan dismissed, 3=dismiss clan, 4=dissolve ally

	reputation atomic.Int32

	// Members indexed by playerID.
	members map[int64]*Member

	// Sub-pledges (Academy, Royal Guard, Knights).
	subPledges map[int32]*SubPledge

	// Clan wars: set of clan IDs.
	atWarWith      map[int32]struct{} // Clans we declared war on
	atWarAttackers map[int32]struct{} // Clans that declared war on us

	// Clan skills: skillID -> level.
	skills map[int32]int32

	// Rank privileges: powerGrade -> Privilege mask.
	rankPrivileges map[int32]Privilege

	// Notices / announcements.
	notice          string
	noticeEnabled   bool
	introductionMsg string

	// Dissolution tracking.
	dissolutionTime int64 // Unix millis, 0 if not dissolving
}

// New creates a new clan.
func New(id int32, name string, leaderID int64) *Clan {
	c := &Clan{
		id:             id,
		name:           name,
		leaderID:       leaderID,
		members:        make(map[int64]*Member, 10),
		subPledges:     make(map[int32]*SubPledge, 4),
		atWarWith:      make(map[int32]struct{}, 4),
		atWarAttackers: make(map[int32]struct{}, 4),
		skills:         make(map[int32]int32, 8),
		rankPrivileges: make(map[int32]Privilege, 9),
	}
	// Default rank privileges.
	for grade := int32(1); grade <= 9; grade++ {
		c.rankPrivileges[grade] = DefaultRankPrivileges(grade)
	}
	return c
}

// ID returns the clan ID.
func (c *Clan) ID() int32 { return c.id }

// Name returns the clan name.
func (c *Clan) Name() string { return c.name }

// LeaderID returns the object ID of the clan leader.
func (c *Clan) LeaderID() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.leaderID
}

// SetLeaderID sets the clan leader.
func (c *Clan) SetLeaderID(id int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.leaderID = id
}

// Level returns the clan level (0-8).
func (c *Clan) Level() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.level
}

// SetLevel sets the clan level.
func (c *Clan) SetLevel(level int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.level = level
}

// CrestID returns the clan crest ID.
func (c *Clan) CrestID() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.crestID
}

// SetCrestID sets the clan crest ID.
func (c *Clan) SetCrestID(id int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.crestID = id
}

// LargeCrestID returns the large crest ID.
func (c *Clan) LargeCrestID() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.largeCrest
}

// SetLargeCrestID sets the large crest ID.
func (c *Clan) SetLargeCrestID(id int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.largeCrest = id
}

// AllyID returns the alliance ID.
func (c *Clan) AllyID() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.allyID
}

// SetAllyID sets the alliance ID.
func (c *Clan) SetAllyID(id int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.allyID = id
}

// AllyCrestID returns the alliance crest ID.
func (c *Clan) AllyCrestID() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.allyCrest
}

// SetAllyCrestID sets the alliance crest ID.
func (c *Clan) SetAllyCrestID(id int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.allyCrest = id
}

// AllyName returns the alliance name.
func (c *Clan) AllyName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.allyName
}

// SetAllyName sets the alliance name.
func (c *Clan) SetAllyName(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.allyName = name
}

// Reputation returns the clan reputation score.
func (c *Clan) Reputation() int32 {
	return c.reputation.Load()
}

// SetReputation sets the clan reputation score.
func (c *Clan) SetReputation(rep int32) {
	c.reputation.Store(rep)
}

// AddReputation atomically adds to the reputation (can be negative).
func (c *Clan) AddReputation(delta int32) int32 {
	return c.reputation.Add(delta)
}

// MaxMembers returns the maximum number of members for the current level.
func (c *Clan) MaxMembers() int32 {
	c.mu.RLock()
	lvl := c.level
	c.mu.RUnlock()
	if lvl < 0 || lvl > MaxClanLevel {
		return maxMembersByLevel[0]
	}
	return maxMembersByLevel[lvl]
}

// MemberCount returns the current number of members.
func (c *Clan) MemberCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.members)
}

// Member returns a clan member by player ID, or nil if not found.
func (c *Clan) Member(playerID int64) *Member {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.members[playerID]
}

// AddMember adds a member to the clan.
// Returns ErrClanFull if the clan has reached max members.
func (c *Clan) AddMember(m *Member) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if int32(len(c.members)) >= maxMembersByLevel[c.level] {
		return ErrClanFull
	}
	c.members[m.PlayerID()] = m
	return nil
}

// RemoveMember removes a member from the clan by player ID.
// Returns the removed member, or nil if not found.
func (c *Clan) RemoveMember(playerID int64) *Member {
	c.mu.Lock()
	defer c.mu.Unlock()

	m, ok := c.members[playerID]
	if !ok {
		return nil
	}
	delete(c.members, playerID)
	return m
}

// ForEachMember iterates over all members.
// The callback receives each member; return false to stop iteration.
func (c *Clan) ForEachMember(fn func(*Member) bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, m := range c.members {
		if !fn(m) {
			return
		}
	}
}

// Members returns a snapshot slice of all members.
func (c *Clan) Members() []*Member {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*Member, 0, len(c.members))
	for _, m := range c.members {
		result = append(result, m)
	}
	return result
}

// OnlineMemberCount returns the number of online members.
func (c *Clan) OnlineMemberCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, m := range c.members {
		if m.Online() {
			count++
		}
	}
	return count
}

// SubPledge returns the sub-pledge by type, or nil if not found.
func (c *Clan) SubPledge(pledgeType int32) *SubPledge {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.subPledges[pledgeType]
}

// AddSubPledge creates a sub-pledge.
// Returns ErrClanLevelTooLow if the clan level is insufficient.
func (c *Clan) AddSubPledge(sp *SubPledge) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	reqLevel := SubPledgeRequiredLevel(sp.ID)
	if c.level < reqLevel {
		return ErrClanLevelTooLow
	}
	c.subPledges[sp.ID] = sp
	return nil
}

// RemoveSubPledge removes a sub-pledge.
func (c *Clan) RemoveSubPledge(pledgeType int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subPledges, pledgeType)
}

// SubPledges returns a snapshot of all sub-pledges.
func (c *Clan) SubPledges() []*SubPledge {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*SubPledge, 0, len(c.subPledges))
	for _, sp := range c.subPledges {
		result = append(result, sp)
	}
	return result
}

// SubPledgeMemberCount returns the number of members in a sub-pledge.
func (c *Clan) SubPledgeMemberCount(pledgeType int32) int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var count int32
	for _, m := range c.members {
		if m.PledgeType() == pledgeType {
			count++
		}
	}
	return count
}

// --- Clan Wars ---

// DeclareWar adds a clan to the war list.
func (c *Clan) DeclareWar(enemyClanID int32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.atWarWith[enemyClanID]; ok {
		return ErrAlreadyAtWar
	}
	c.atWarWith[enemyClanID] = struct{}{}
	return nil
}

// AcceptWar marks that an enemy has declared war on us.
func (c *Clan) AcceptWar(attackerClanID int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.atWarAttackers[attackerClanID] = struct{}{}
}

// EndWar removes a clan from war lists.
func (c *Clan) EndWar(enemyClanID int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.atWarWith, enemyClanID)
	delete(c.atWarAttackers, enemyClanID)
}

// IsAtWarWith returns true if we declared war on the given clan.
func (c *Clan) IsAtWarWith(clanID int32) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.atWarWith[clanID]
	return ok
}

// IsUnderAttack returns true if the given clan declared war on us.
func (c *Clan) IsUnderAttack(clanID int32) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.atWarAttackers[clanID]
	return ok
}

// WarList returns a snapshot of clan IDs we are at war with.
func (c *Clan) WarList() []int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]int32, 0, len(c.atWarWith))
	for id := range c.atWarWith {
		result = append(result, id)
	}
	return result
}

// AttackerList returns a snapshot of clan IDs that attacked us.
func (c *Clan) AttackerList() []int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]int32, 0, len(c.atWarAttackers))
	for id := range c.atWarAttackers {
		result = append(result, id)
	}
	return result
}

// --- Clan Skills ---

// SkillLevel returns the level of a clan skill, 0 if not learned.
func (c *Clan) SkillLevel(skillID int32) int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.skills[skillID]
}

// SetSkill sets the clan skill level.
func (c *Clan) SetSkill(skillID, level int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.skills[skillID] = level
}

// Skills returns a snapshot copy of the skills map.
func (c *Clan) Skills() map[int32]int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[int32]int32, len(c.skills))
	maps.Copy(result, c.skills)
	return result
}

// --- Rank Privileges ---

// RankPrivileges returns the privilege mask for a power grade.
func (c *Clan) RankPrivileges(powerGrade int32) Privilege {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rankPrivileges[powerGrade]
}

// SetRankPrivileges sets the privilege mask for a power grade.
func (c *Clan) SetRankPrivileges(powerGrade int32, priv Privilege) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rankPrivileges[powerGrade] = priv
}

// AllRankPrivileges returns a snapshot of all rank privileges.
func (c *Clan) AllRankPrivileges() map[int32]Privilege {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[int32]Privilege, len(c.rankPrivileges))
	maps.Copy(result, c.rankPrivileges)
	return result
}

// --- Notice ---

// Notice returns the clan notice text.
func (c *Clan) Notice() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.notice
}

// SetNotice sets the clan notice.
func (c *Clan) SetNotice(text string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.notice = text
}

// NoticeEnabled returns whether the notice is displayed on login.
func (c *Clan) NoticeEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.noticeEnabled
}

// SetNoticeEnabled enables/disables the clan notice.
func (c *Clan) SetNoticeEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.noticeEnabled = enabled
}

// IntroductionMessage returns the clan introduction message.
func (c *Clan) IntroductionMessage() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.introductionMsg
}

// SetIntroductionMessage sets the clan introduction message.
func (c *Clan) SetIntroductionMessage(msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.introductionMsg = msg
}

// --- Dissolution ---

// DissolutionTime returns the scheduled dissolution time (Unix millis), or 0 if not dissolving.
func (c *Clan) DissolutionTime() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.dissolutionTime
}

// SetDissolutionTime sets the scheduled dissolution time.
func (c *Clan) SetDissolutionTime(t int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dissolutionTime = t
}

// IsDissolving returns true if the clan is scheduled for dissolution.
func (c *Clan) IsDissolving() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.dissolutionTime > 0
}

// MemberByName returns a member by character name (case-insensitive), or nil.
func (c *Clan) MemberByName(name string) *Member {
	c.mu.RLock()
	defer c.mu.RUnlock()

	lower := strings.ToLower(name)
	for _, m := range c.members {
		if strings.ToLower(m.Name()) == lower {
			return m
		}
	}
	return nil
}

// Leader returns the clan leader Member, or nil if not found.
func (c *Clan) Leader() *Member {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.members[c.leaderID]
}

// --- Alliance Penalty ---

// AllyPenaltyExpiryTime returns the alliance penalty expiry time (Unix millis).
func (c *Clan) AllyPenaltyExpiryTime() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.allyPenaltyExpiryTime
}

// AllyPenaltyType returns the alliance penalty type.
func (c *Clan) AllyPenaltyType() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.allyPenaltyType
}

// SetAllyPenalty sets the alliance penalty expiry time and type.
func (c *Clan) SetAllyPenalty(expiryTime int64, penaltyType int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.allyPenaltyExpiryTime = expiryTime
	c.allyPenaltyType = penaltyType
}

// HasAllyPenalty returns true if the alliance penalty is currently active.
func (c *Clan) HasAllyPenalty() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.allyPenaltyExpiryTime > 0
}

// ClearAlly clears all alliance fields.
func (c *Clan) ClearAlly() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.allyID = 0
	c.allyName = ""
	c.allyCrest = 0
}

// IsAllyLeader returns true if this clan is the leader of its alliance (allyID == clanID).
func (c *Clan) IsAllyLeader() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.allyID != 0 && c.allyID == c.id
}
