package model

// NpcTemplate represents NPC stats and AI parameters from npc_templates table
type NpcTemplate struct {
	templateID  int32
	name        string
	title       string
	level       int32
	maxHP       int32
	maxMP       int32
	pAtk        int32
	pDef        int32
	mAtk        int32
	mDef        int32
	aggroRange  int32
	moveSpeed   int32
	atkSpeed    int32
	respawnMin  int32 // seconds
	respawnMax  int32 // seconds
	baseExp     int64
	baseSP      int32
}

// NewNpcTemplate creates a new NPC template
func NewNpcTemplate(
	templateID int32,
	name, title string,
	level, maxHP, maxMP int32,
	pAtk, pDef, mAtk, mDef int32,
	aggroRange, moveSpeed, atkSpeed int32,
	respawnMin, respawnMax int32,
	baseExp int64, baseSP int32,
) *NpcTemplate {
	return &NpcTemplate{
		templateID: templateID,
		name:       name,
		title:      title,
		level:      level,
		maxHP:      maxHP,
		maxMP:      maxMP,
		pAtk:       pAtk,
		pDef:       pDef,
		mAtk:       mAtk,
		mDef:       mDef,
		aggroRange: aggroRange,
		moveSpeed:  moveSpeed,
		atkSpeed:   atkSpeed,
		respawnMin: respawnMin,
		respawnMax: respawnMax,
		baseExp:    baseExp,
		baseSP:     baseSP,
	}
}

// TemplateID returns template ID
func (t *NpcTemplate) TemplateID() int32 {
	return t.templateID
}

// Name returns NPC name
func (t *NpcTemplate) Name() string {
	return t.name
}

// Title returns NPC title
func (t *NpcTemplate) Title() string {
	return t.title
}

// Level returns NPC level
func (t *NpcTemplate) Level() int32 {
	return t.level
}

// MaxHP returns max HP
func (t *NpcTemplate) MaxHP() int32 {
	return t.maxHP
}

// MaxMP returns max MP
func (t *NpcTemplate) MaxMP() int32 {
	return t.maxMP
}

// PAtk returns physical attack
func (t *NpcTemplate) PAtk() int32 {
	return t.pAtk
}

// PDef returns physical defense
func (t *NpcTemplate) PDef() int32 {
	return t.pDef
}

// MAtk returns magical attack
func (t *NpcTemplate) MAtk() int32 {
	return t.mAtk
}

// MDef returns magical defense
func (t *NpcTemplate) MDef() int32 {
	return t.mDef
}

// AggroRange returns aggro range (0 for non-aggressive NPCs)
func (t *NpcTemplate) AggroRange() int32 {
	return t.aggroRange
}

// MoveSpeed returns movement speed
func (t *NpcTemplate) MoveSpeed() int32 {
	return t.moveSpeed
}

// AtkSpeed returns attack speed
func (t *NpcTemplate) AtkSpeed() int32 {
	return t.atkSpeed
}

// RespawnMin returns minimum respawn time in seconds
func (t *NpcTemplate) RespawnMin() int32 {
	return t.respawnMin
}

// RespawnMax returns maximum respawn time in seconds
func (t *NpcTemplate) RespawnMax() int32 {
	return t.respawnMax
}

// BaseExp returns base experience reward
func (t *NpcTemplate) BaseExp() int64 {
	return t.baseExp
}

// BaseSP returns base skill point reward
func (t *NpcTemplate) BaseSP() int32 {
	return t.baseSP
}
