package data

import (
	"log/slog"
	"slices"
)

// NpcTable — глобальный registry всех NPC templates.
// map[templateID]*npcDef
var NpcTable map[int32]*npcDef

// GetNpcDef возвращает npcDef по template ID.
// Returns nil если NPC не найден.
func GetNpcDef(templateID int32) *npcDef {
	if NpcTable == nil {
		return nil
	}
	return NpcTable[templateID]
}

// LoadNpcTemplates строит NpcTable из Go-литералов (npcDefs).
func LoadNpcTemplates() error {
	NpcTable = make(map[int32]*npcDef, len(npcDefs))

	for i := range npcDefs {
		NpcTable[npcDefs[i].id] = &npcDefs[i]
	}

	slog.Info("loaded NPC templates", "count", len(NpcTable))
	return nil
}

// NpcDef accessor methods — provide read access to npcDef fields.
// These are used by spawn/manager.go and other packages that need NPC data.

func (d *npcDef) ID() int32         { return d.id }
func (d *npcDef) Name() string      { return d.name }
func (d *npcDef) Title() string     { return d.title }
func (d *npcDef) NpcType() string   { return d.npcType }
func (d *npcDef) Level() int32      { return d.level }
func (d *npcDef) Race() string      { return d.race }
func (d *npcDef) Sex() string       { return d.sex }
func (d *npcDef) HP() float64       { return d.hp }
func (d *npcDef) MP() float64       { return d.mp }
func (d *npcDef) HPRegen() float64  { return d.hpRegen }
func (d *npcDef) MPRegen() float64  { return d.mpRegen }
func (d *npcDef) PAtk() float64     { return d.pAtk }
func (d *npcDef) MAtk() float64     { return d.mAtk }
func (d *npcDef) PDef() float64     { return d.pDef }
func (d *npcDef) MDef() float64     { return d.mDef }
func (d *npcDef) AggroRange() int32 { return d.aggroRange }
func (d *npcDef) RunSpeed() int32   { return d.runSpeed }
func (d *npcDef) AtkSpeed() int32   { return d.atkSpeed }
func (d *npcDef) BaseExp() int64    { return d.baseExp }
func (d *npcDef) BaseSP() int32     { return d.baseSP }

func (d *npcDef) CollisionRadius() float64 { return d.collisionRadius }
func (d *npcDef) CollisionHeight() float64 { return d.collisionHeight }

func (d *npcDef) IsAggressive() bool { return d.isAggressive || d.aggroRange > 0 }
func (d *npcDef) Rhand() int32       { return d.rhand }
func (d *npcDef) Lhand() int32       { return d.lhand }
func (d *npcDef) Chest() int32       { return d.chest }

func (d *npcDef) Drops() []dropGroupDef { return d.drops }
func (d *npcDef) Spoils() []dropItemDef { return d.spoils }
func (d *npcDef) Skills() []npcSkillDef { return d.skills }
func (d *npcDef) Minions() []minionDef  { return d.minions }

func (d *npcDef) Clans() []string        { return d.clans }
func (d *npcDef) ClanHelpRange() int32   { return d.clanHelpRange }
func (d *npcDef) IgnoreNpcIDs() []int32  { return d.ignoreNpcIds }
func (d *npcDef) AtkRange() int32        { return d.atkRange }
func (d *npcDef) WalkSpeed() int32       { return d.walkSpeed }

// NpcSkillDef accessor methods
func (s *npcSkillDef) SkillID() int32 { return s.skillID }
func (s *npcSkillDef) SkillLevel() int32 { return s.level }

// IsClan checks if this NPC shares a clan with the given clan list.
func (d *npcDef) IsClan(otherClans []string) bool {
	if len(d.clans) == 0 || len(otherClans) == 0 {
		return false
	}
	for _, c := range d.clans {
		if slices.Contains(otherClans, c) {
			return true
		}
	}
	return false
}

// IgnoresNpcID checks if this NPC ignores faction calls from the given NPC ID.
func (d *npcDef) IgnoresNpcID(npcID int32) bool {
	return slices.Contains(d.ignoreNpcIds, npcID)
}

// DropGroupDef accessor methods
func (g *dropGroupDef) Chance() float64      { return g.chance }
func (g *dropGroupDef) Items() []dropItemDef { return g.items }

// DropItemDef accessor methods
func (d *dropItemDef) ItemID() int32   { return d.itemID }
func (d *dropItemDef) Min() int32      { return d.min }
func (d *dropItemDef) Max() int32      { return d.max }
func (d *dropItemDef) Chance() float64 { return d.chance }
