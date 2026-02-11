package data

// OperateType определяет тип активации скилла.
type OperateType int8

const (
	OperateTypeA1  OperateType = iota // Active instant (damage, heal)
	OperateTypeA2                     // Active continuous (buff/debuff)
	OperateTypeA3                     // Active continuous channeling
	OperateTypeCA1                    // Charge skill
	OperateTypeCA5                    // Charge skill type 5
	OperateTypeP                      // Passive
	OperateTypeT                      // Toggle
)

// TargetType определяет тип цели скилла.
type TargetType int8

const (
	TargetSelf           TargetType = iota // Self-cast
	TargetOne                              // Single target
	TargetArea                             // Area of effect
	TargetAura                             // Aura around caster
	TargetParty                            // Party members
	TargetClan                             // Clan members
	TargetAreaCorpseMob                    // Area corpse mob
	TargetAreaSummon                       // Area summon
	TargetAuraCorpseMob                    // Aura corpse mob
	TargetBehindAura                       // Behind aura
	TargetClanMember                       // Clan member
	TargetCorpse                           // Corpse
	TargetCorpseClan                       // Corpse clan
	TargetCorpseMob                        // Corpse mob
	TargetEnemySummon                      // Enemy summon
	TargetFrontArea                        // Front area
	TargetFrontAura                        // Front aura
	TargetGround                           // Ground target
	TargetHoly                             // Holy artifact
	TargetNone                             // No target
	TargetOwnerPet                         // Owner's pet
	TargetPartyClan                        // Party + clan
	TargetPartyMember                      // Party member
	TargetPartyNotMe                       // Party except self
	TargetPcBody                           // Player body
	TargetPet                              // Pet
	TargetServitor                         // Servitor
	TargetUnlockable                       // Unlockable (door/chest)
	TargetCorpsePlayer                     // Corpse player
	TargetCorpseAlly                       // Corpse ally
	TargetBehindArea                       // Behind area
	TargetUndead                           // Undead target
	TargetAlly                             // Ally members
	TargetEnemy                            // Enemy target
	TargetEnemyNot                         // Enemy not
	TargetSummon                           // Summon
)

// StatMod — runtime stat modifier inside an effect.
type StatMod struct {
	Op   string  // "add", "mul", "sub", "set"
	Stat string  // "pDef", "runSpd", "critRate"...
	Val  float64 // resolved numeric value
}

// EffectDef описывает один эффект скилла (из XML <effect>).
type EffectDef struct {
	Name     string            // "Buff", "PhysicalDamage", "Heal", etc.
	Params   map[string]string // effect-specific parameters
	StatMods []StatMod         // resolved stat modifiers
}

// SkillTemplate — immutable шаблон скилла, загруженный из XML.
// Один экземпляр на каждую пару (skillID, level).
// Shared across all players — НЕ модифицировать после загрузки.
type SkillTemplate struct {
	ID            int32
	Level         int32
	Name          string
	OperateType   OperateType
	MagicLevel    int32
	HitTime       int32   // ms — cast time animation
	CoolTime      int32   // ms — delay after cast before next action
	ReuseDelay    int32   // ms — cooldown before skill can be used again
	CastRange     int32   // max distance to target
	EffectRange   int32   // max distance for effect application
	MpConsume     int32   // MP consumed on cast
	HpConsume     int32   // HP consumed on cast
	Power         float64 // skill power (for damage/heal calculations)
	TargetType    TargetType
	IsMagic       bool   // true = magic, false = physical
	AbnormalType  string // "STUN", "ROOT", "SPEED_UP", etc.
	AbnormalLevel int32
	AbnormalTime  int32 // seconds — duration of abnormal effect
	Effects       []EffectDef

	// Extended attributes
	IsDebuff         bool
	OverHit          bool
	IgnoreShield     bool
	NextActionAttack bool
	Trait             string
	BasicProperty     string
	Element           int32
	ElementPower      int32
	AffectRange       int32
	ActivateRate      int32  // debuff land rate chance
	LvlBonusRate      int32
	BlowChance        int32
	ChargeConsume     int32
	MpInitialConsume  int32
	ItemConsumeId     int32
	ItemConsumeCount  int32
	EffectPoint       int32 // hate/aggro points
	BaseCritRate      int32
	EnchantGroup1     int32
	EnchantGroup2     int32
}

// IsPassive returns true if this skill is a passive skill.
func (s *SkillTemplate) IsPassive() bool {
	return s.OperateType == OperateTypeP
}

// IsToggle returns true if this skill is a toggle skill.
func (s *SkillTemplate) IsToggle() bool {
	return s.OperateType == OperateTypeT
}

// IsActive returns true if this skill is an active skill.
func (s *SkillTemplate) IsActive() bool {
	switch s.OperateType {
	case OperateTypeA1, OperateTypeA2, OperateTypeA3, OperateTypeCA1, OperateTypeCA5:
		return true
	default:
		return false
	}
}

// IsInstant returns true if skill is instant cast (HitTime == 0).
func (s *SkillTemplate) IsInstant() bool {
	return s.HitTime == 0
}

// IsContinuous returns true if skill applies a lasting effect (A2 type).
func (s *SkillTemplate) IsContinuous() bool {
	return s.OperateType == OperateTypeA2 || s.OperateType == OperateTypeA3
}

// ParseOperateType converts string to OperateType.
func ParseOperateType(s string) OperateType {
	switch s {
	case "A1":
		return OperateTypeA1
	case "A2":
		return OperateTypeA2
	case "A3":
		return OperateTypeA3
	case "CA1":
		return OperateTypeCA1
	case "CA5":
		return OperateTypeCA5
	case "P":
		return OperateTypeP
	case "T":
		return OperateTypeT
	default:
		return OperateTypeA1
	}
}

// ParseTargetType converts string to TargetType.
func ParseTargetType(s string) TargetType {
	switch s {
	case "SELF":
		return TargetSelf
	case "ONE":
		return TargetOne
	case "AREA":
		return TargetArea
	case "AURA":
		return TargetAura
	case "PARTY":
		return TargetParty
	case "CLAN":
		return TargetClan
	case "AREA_CORPSE_MOB":
		return TargetAreaCorpseMob
	case "AREA_SUMMON":
		return TargetAreaSummon
	case "AURA_CORPSE_MOB":
		return TargetAuraCorpseMob
	case "BEHIND_AURA":
		return TargetBehindAura
	case "CLAN_MEMBER":
		return TargetClanMember
	case "CORPSE":
		return TargetCorpse
	case "CORPSE_CLAN":
		return TargetCorpseClan
	case "CORPSE_MOB":
		return TargetCorpseMob
	case "ENEMY_SUMMON":
		return TargetEnemySummon
	case "FRONT_AREA":
		return TargetFrontArea
	case "FRONT_AURA":
		return TargetFrontAura
	case "GROUND":
		return TargetGround
	case "HOLY":
		return TargetHoly
	case "NONE":
		return TargetNone
	case "OWNER_PET":
		return TargetOwnerPet
	case "PARTY_CLAN":
		return TargetPartyClan
	case "PARTY_MEMBER":
		return TargetPartyMember
	case "PARTY_NOTME":
		return TargetPartyNotMe
	case "PC_BODY":
		return TargetPcBody
	case "PET":
		return TargetPet
	case "SERVITOR":
		return TargetServitor
	case "UNLOCKABLE":
		return TargetUnlockable
	case "CORPSE_PLAYER":
		return TargetCorpsePlayer
	case "CORPSE_ALLY":
		return TargetCorpseAlly
	case "BEHIND_AREA":
		return TargetBehindArea
	case "UNDEAD":
		return TargetUndead
	case "ALLY":
		return TargetAlly
	case "ENEMY":
		return TargetEnemy
	case "ENEMY_NOT":
		return TargetEnemyNot
	case "SUMMON":
		return TargetSummon
	default:
		return TargetOne
	}
}
