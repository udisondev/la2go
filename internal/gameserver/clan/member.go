package clan

import "sync"

// Member represents a clan member.
// Thread-safe: fields are protected by mu.
// When the member is online, Name/Level/ClassID are read from the live Player.
// When offline, cached values are used.
type Member struct {
	mu sync.RWMutex

	playerID   int64 // Character (object) ID
	name       string
	level      int32
	classID    int32
	pledgeType int32 // SubPledge type: MainClan=0, Academy=-1, Royal=100/200, Knight=1001+
	powerGrade int32 // Rank within clan (1=leader, 2-9=members)
	title      string
	online     bool
	sponsorID  int64 // Apprentice sponsor (for Academy)
	apprentice int64 // Apprentice player ID

	// Privilege mask for this member (based on powerGrade).
	privileges Privilege
}

// NewMember creates a clan member with initial values.
func NewMember(playerID int64, name string, level, classID, pledgeType, powerGrade int32) *Member {
	return &Member{
		playerID:   playerID,
		name:       name,
		level:      level,
		classID:    classID,
		pledgeType: pledgeType,
		powerGrade: powerGrade,
		privileges: DefaultRankPrivileges(powerGrade),
	}
}

// PlayerID returns the member's character object ID.
func (m *Member) PlayerID() int64 {
	return m.playerID // immutable, no lock needed
}

// Name returns the member's character name.
func (m *Member) Name() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.name
}

// SetName updates the cached name (called on login/name change).
func (m *Member) SetName(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.name = name
}

// Level returns the member's level.
func (m *Member) Level() int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.level
}

// SetLevel updates the cached level.
func (m *Member) SetLevel(level int32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.level = level
}

// ClassID returns the member's class ID.
func (m *Member) ClassID() int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.classID
}

// SetClassID updates the cached class ID.
func (m *Member) SetClassID(classID int32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.classID = classID
}

// PledgeType returns the member's sub-pledge type.
func (m *Member) PledgeType() int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pledgeType
}

// SetPledgeType sets the member's sub-pledge type.
func (m *Member) SetPledgeType(pledgeType int32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pledgeType = pledgeType
}

// PowerGrade returns the member's rank (1=leader).
func (m *Member) PowerGrade() int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.powerGrade
}

// SetPowerGrade sets the member's rank and updates default privileges.
func (m *Member) SetPowerGrade(grade int32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.powerGrade = grade
	m.privileges = DefaultRankPrivileges(grade)
}

// Title returns the member's title.
func (m *Member) Title() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.title
}

// SetTitle sets the member's title.
func (m *Member) SetTitle(title string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.title = title
}

// Online returns whether the member is currently online.
func (m *Member) Online() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.online
}

// SetOnline marks the member as online or offline.
func (m *Member) SetOnline(online bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.online = online
}

// SponsorID returns the sponsor (mentor) player ID for academy members.
func (m *Member) SponsorID() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sponsorID
}

// SetSponsorID sets the sponsor player ID.
func (m *Member) SetSponsorID(id int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sponsorID = id
}

// Apprentice returns the apprentice player ID.
func (m *Member) Apprentice() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.apprentice
}

// SetApprentice sets the apprentice player ID.
func (m *Member) SetApprentice(id int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.apprentice = id
}

// Privileges returns the member's privilege mask.
func (m *Member) Privileges() Privilege {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.privileges
}

// SetPrivileges sets the member's privilege mask.
func (m *Member) SetPrivileges(priv Privilege) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.privileges = priv
}

// HasPrivilege checks if the member has a specific privilege.
func (m *Member) HasPrivilege(priv Privilege) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.privileges.Has(priv)
}
