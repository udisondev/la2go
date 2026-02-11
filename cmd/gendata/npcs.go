package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// --- XML structures (npcs) ---

type xmlNpcList struct {
	XMLName xml.Name `xml:"list"`
	Npcs    []xmlNpc `xml:"npc"`
}

type xmlNpc struct {
	ID        int32  `xml:"id,attr"`
	Level     int32  `xml:"level,attr"`
	Type      string `xml:"type,attr"`
	Name      string `xml:"name,attr"`
	Title     string `xml:"title,attr"`
	DisplayID int32  `xml:"displayId,attr"`

	Parameters *xmlNpcParameters `xml:"parameters"`
	Race       string            `xml:"race"`
	Sex        string            `xml:"sex"`
	Acquire    *xmlNpcAcquire    `xml:"acquire"`
	Stats      *xmlNpcStats      `xml:"stats"`
	Status     *xmlNpcStatus     `xml:"status"`
	SkillList  *xmlNpcSkillList  `xml:"skillList"`
	CorpseTime int32             `xml:"corpseTime"`
	AI         *xmlNpcAI         `xml:"ai"`
	DropLists  *xmlNpcDropLists  `xml:"dropLists"`
	Collision  *xmlNpcCollision  `xml:"collision"`
	Equipment  *xmlNpcEquipment  `xml:"equipment"`
}

type xmlNpcParameters struct {
	Minions *xmlNpcMinions `xml:"minions"`
}

type xmlNpcMinions struct {
	Npcs []xmlNpcMinion `xml:"npc"`
}

type xmlNpcMinion struct {
	ID    int32 `xml:"id,attr"`
	Count int32 `xml:"count,attr"`
}

type xmlNpcAcquire struct {
	Exp int64 `xml:"exp,attr"`
	SP  int32 `xml:"sp,attr"`
}

type xmlNpcStats struct {
	STR int32 `xml:"str,attr"`
	INT int32 `xml:"int,attr"`
	DEX int32 `xml:"dex,attr"`
	WIT int32 `xml:"wit,attr"`
	CON int32 `xml:"con,attr"`
	MEN int32 `xml:"men,attr"`

	Vitals  *xmlNpcVitals  `xml:"vitals"`
	Attack  *xmlNpcAttack  `xml:"attack"`
	Defence *xmlNpcDefence `xml:"defence"`
	Speed   *xmlNpcSpeed   `xml:"speed"`
	HitTime int32          `xml:"hitTime"`
}

type xmlNpcVitals struct {
	HP      float64 `xml:"hp,attr"`
	HPRegen float64 `xml:"hpRegen,attr"`
	MP      float64 `xml:"mp,attr"`
	MPRegen float64 `xml:"mpRegen,attr"`
}

type xmlNpcAttack struct {
	Physical    float64 `xml:"physical,attr"`
	Magical     float64 `xml:"magical,attr"`
	Random      int32   `xml:"random,attr"`
	Critical    int32   `xml:"critical,attr"`
	Accuracy    float64 `xml:"accuracy,attr"`
	AttackSpeed int32   `xml:"attackSpeed,attr"`
	Type        string  `xml:"type,attr"`
	Range       int32   `xml:"range,attr"`
	Distance    int32   `xml:"distance,attr"`
	Width       int32   `xml:"width,attr"`
}

type xmlNpcDefence struct {
	Physical float64 `xml:"physical,attr"`
	Magical  float64 `xml:"magical,attr"`
	Evasion  float64 `xml:"evasion,attr"`
}

type xmlNpcSpeed struct {
	Walk *xmlNpcSpeedGround `xml:"walk"`
	Run  *xmlNpcSpeedGround `xml:"run"`
}

type xmlNpcSpeedGround struct {
	Ground float64 `xml:"ground,attr"`
}

type xmlNpcStatus struct {
	Undying    string `xml:"undying,attr"`
	Attackable string `xml:"attackable,attr"`
	Talkable   string `xml:"talkable,attr"`
	CanBeSown  string `xml:"canBeSown,attr"`
}

type xmlNpcSkillList struct {
	Skills []xmlNpcSkill `xml:"skill"`
}

type xmlNpcSkill struct {
	ID    int32 `xml:"id,attr"`
	Level int32 `xml:"level,attr"`
}

type xmlNpcAI struct {
	AggroRange    int32  `xml:"aggroRange,attr"`
	ClanHelpRange int32  `xml:"clanHelpRange,attr"`
	IsAggressive  string `xml:"isAggressive,attr"`
	ClanList      *xmlNpcClanList `xml:"clanList"`
}

type xmlNpcClanList struct {
	Clans        []string `xml:"clan"`
	IgnoreNpcIds []int32  `xml:"ignoreNpcId"`
}

type xmlNpcDropLists struct {
	Drop  *xmlNpcDrop  `xml:"drop"`
	Spoil *xmlNpcSpoil `xml:"spoil"`
}

type xmlNpcDrop struct {
	Groups []xmlNpcDropGroup `xml:"group"`
}

type xmlNpcDropGroup struct {
	Chance float64           `xml:"chance,attr"`
	Items  []xmlNpcDropItem  `xml:"item"`
}

type xmlNpcDropItem struct {
	ID     int32   `xml:"id,attr"`
	Min    int32   `xml:"min,attr"`
	Max    int32   `xml:"max,attr"`
	Chance float64 `xml:"chance,attr"`
}

type xmlNpcSpoil struct {
	Items []xmlNpcDropItem `xml:"item"`
}

type xmlNpcCollision struct {
	Radius *xmlNpcCollisionValue `xml:"radius"`
	Height *xmlNpcCollisionValue `xml:"height"`
}

type xmlNpcCollisionValue struct {
	Normal float64 `xml:"normal,attr"`
}

type xmlNpcEquipment struct {
	RHand int32 `xml:"rhand,attr"`
	LHand int32 `xml:"lhand,attr"`
	Chest int32 `xml:"chest,attr"`
}

// --- Parsed structure (npc) ---

type parsedNpc struct {
	id      int32
	name    string
	title   string
	npcType string
	level   int32
	race    string
	sex     string

	str, intel, dex, wit, con, men int32
	hp, mp                         float64
	hpRegen, mpRegen               float64
	pAtk, mAtk                     float64
	pDef, mDef                     float64
	accuracy                       float64
	critical                       int32
	randomDamage                   int32
	evasion                        float64
	atkSpeed                       int32
	atkType                        string
	atkRange                       int32
	atkDistance                     int32
	atkWidth                       int32
	walkSpeed                      int32
	runSpeed                       int32
	hitTime                        int32

	aggroRange    int32
	clanHelpRange int32
	isAggressive  bool
	clans         []string
	ignoreNpcIds  []int32

	baseExp int64
	baseSP  int32

	drops  []parsedDropGroup
	spoils []parsedDropItem

	skills []parsedNpcSkill

	collisionRadius float64
	collisionHeight float64

	rhand int32
	lhand int32
	chest int32

	undying    bool
	attackable bool
	talkable   bool
	canBeSown  bool

	corpseTime int32
	displayId  int32

	minions []parsedMinion
}

type parsedDropGroup struct {
	chance float64
	items  []parsedDropItem
}

type parsedDropItem struct {
	itemID int32
	min    int32
	max    int32
	chance float64
}

type parsedNpcSkill struct {
	skillID int32
	level   int32
}

type parsedMinion struct {
	npcID int32
	count int32
}

// generateNpcs parses all NPC XML files and generates Go source with npcDef literals.
func generateNpcs(javaDir, outDir string) error {
	npcsDir := filepath.Join(javaDir, "stats", "npcs")
	npcs, err := parseAllNpcs(npcsDir)
	if err != nil {
		return fmt.Errorf("parse npcs: %w", err)
	}

	sort.Slice(npcs, func(i, j int) bool { return npcs[i].id < npcs[j].id })

	outPath := filepath.Join(outDir, "npc_data_generated.go")
	if err := generateNpcsGoFile(npcs, outPath); err != nil {
		return fmt.Errorf("generate npcs: %w", err)
	}

	fmt.Printf("  Generated %s: %d npcs\n", outPath, len(npcs))
	return nil
}

func parseAllNpcs(dir string) ([]parsedNpc, error) {
	files, err := globXMLFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("glob npcs dir: %w", err)
	}

	var all []parsedNpc
	for _, f := range files {
		npcs, err := parseNpcFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filepath.Base(f), err)
		}
		all = append(all, npcs...)
	}
	return all, nil
}

func parseNpcFile(path string) ([]parsedNpc, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var list xmlNpcList
	if err := xml.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	result := make([]parsedNpc, 0, len(list.Npcs))
	for _, xn := range list.Npcs {
		result = append(result, convertNpc(xn))
	}
	return result, nil
}

func convertNpc(xn xmlNpc) parsedNpc {
	pn := parsedNpc{
		id:         xn.ID,
		name:       xn.Name,
		title:      xn.Title,
		npcType:    xn.Type,
		level:      xn.Level,
		race:       strings.TrimSpace(xn.Race),
		sex:        strings.TrimSpace(xn.Sex),
		displayId:  xn.DisplayID,
		corpseTime: xn.CorpseTime,
	}

	// Acquire
	if xn.Acquire != nil {
		pn.baseExp = xn.Acquire.Exp
		pn.baseSP = xn.Acquire.SP
	}

	// Stats
	if xn.Stats != nil {
		pn.str = xn.Stats.STR
		pn.intel = xn.Stats.INT
		pn.dex = xn.Stats.DEX
		pn.wit = xn.Stats.WIT
		pn.con = xn.Stats.CON
		pn.men = xn.Stats.MEN
		pn.hitTime = xn.Stats.HitTime

		if xn.Stats.Vitals != nil {
			pn.hp = xn.Stats.Vitals.HP
			pn.mp = xn.Stats.Vitals.MP
			pn.hpRegen = xn.Stats.Vitals.HPRegen
			pn.mpRegen = xn.Stats.Vitals.MPRegen
		}

		if xn.Stats.Attack != nil {
			pn.pAtk = xn.Stats.Attack.Physical
			pn.mAtk = xn.Stats.Attack.Magical
			pn.randomDamage = xn.Stats.Attack.Random
			pn.critical = xn.Stats.Attack.Critical
			pn.accuracy = xn.Stats.Attack.Accuracy
			pn.atkSpeed = xn.Stats.Attack.AttackSpeed
			pn.atkType = xn.Stats.Attack.Type
			pn.atkRange = xn.Stats.Attack.Range
			pn.atkDistance = xn.Stats.Attack.Distance
			pn.atkWidth = xn.Stats.Attack.Width
		}

		if xn.Stats.Defence != nil {
			pn.pDef = xn.Stats.Defence.Physical
			pn.mDef = xn.Stats.Defence.Magical
			pn.evasion = xn.Stats.Defence.Evasion
		}

		if xn.Stats.Speed != nil {
			if xn.Stats.Speed.Walk != nil {
				pn.walkSpeed = int32(xn.Stats.Speed.Walk.Ground)
			}
			if xn.Stats.Speed.Run != nil {
				pn.runSpeed = int32(xn.Stats.Speed.Run.Ground)
			}
		}
	}

	// Status
	if xn.Status != nil {
		pn.undying = parseBoolAttr(xn.Status.Undying)
		pn.attackable = parseBoolAttr(xn.Status.Attackable)
		pn.talkable = parseBoolAttr(xn.Status.Talkable)
		pn.canBeSown = parseBoolAttr(xn.Status.CanBeSown)
	}

	// Skills
	if xn.SkillList != nil {
		pn.skills = make([]parsedNpcSkill, 0, len(xn.SkillList.Skills))
		for _, xs := range xn.SkillList.Skills {
			pn.skills = append(pn.skills, parsedNpcSkill{
				skillID: xs.ID,
				level:   xs.Level,
			})
		}
	}

	// AI
	if xn.AI != nil {
		pn.aggroRange = xn.AI.AggroRange
		pn.clanHelpRange = xn.AI.ClanHelpRange
		pn.isAggressive = parseBoolAttr(xn.AI.IsAggressive)

		if xn.AI.ClanList != nil {
			pn.clans = xn.AI.ClanList.Clans
			pn.ignoreNpcIds = xn.AI.ClanList.IgnoreNpcIds
		}
	}

	// DropLists
	if xn.DropLists != nil {
		if xn.DropLists.Drop != nil {
			pn.drops = make([]parsedDropGroup, 0, len(xn.DropLists.Drop.Groups))
			for _, xg := range xn.DropLists.Drop.Groups {
				g := parsedDropGroup{chance: xg.Chance}
				g.items = make([]parsedDropItem, 0, len(xg.Items))
				for _, xi := range xg.Items {
					g.items = append(g.items, parsedDropItem{
						itemID: xi.ID,
						min:    xi.Min,
						max:    xi.Max,
						chance: xi.Chance,
					})
				}
				pn.drops = append(pn.drops, g)
			}
		}
		if xn.DropLists.Spoil != nil {
			pn.spoils = make([]parsedDropItem, 0, len(xn.DropLists.Spoil.Items))
			for _, xi := range xn.DropLists.Spoil.Items {
				pn.spoils = append(pn.spoils, parsedDropItem{
					itemID: xi.ID,
					min:    xi.Min,
					max:    xi.Max,
					chance: xi.Chance,
				})
			}
		}
	}

	// Collision
	if xn.Collision != nil {
		if xn.Collision.Radius != nil {
			pn.collisionRadius = xn.Collision.Radius.Normal
		}
		if xn.Collision.Height != nil {
			pn.collisionHeight = xn.Collision.Height.Normal
		}
	}

	// Equipment
	if xn.Equipment != nil {
		pn.rhand = xn.Equipment.RHand
		pn.lhand = xn.Equipment.LHand
		pn.chest = xn.Equipment.Chest
	}

	// Minions (inside <parameters><minions>)
	if xn.Parameters != nil && xn.Parameters.Minions != nil {
		pn.minions = make([]parsedMinion, 0, len(xn.Parameters.Minions.Npcs))
		for _, xm := range xn.Parameters.Minions.Npcs {
			pn.minions = append(pn.minions, parsedMinion{
				npcID: xm.ID,
				count: xm.Count,
			})
		}
	}

	return pn
}

func parseBoolAttr(s string) bool {
	return s == "true" || s == "1"
}

// --- Code generation (npcs) ---

func generateNpcsGoFile(npcs []parsedNpc, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "npcs")
	buf.WriteString("var npcDefs = []npcDef{\n")

	for _, n := range npcs {
		writeNpcDef(&buf, n)
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeNpcDef(buf *bytes.Buffer, n parsedNpc) {
	buf.WriteString("{\n")

	// Identity
	fmt.Fprintf(buf, "id: %d,\n", n.id)
	if n.name != "" {
		fmt.Fprintf(buf, "name: %q,\n", n.name)
	}
	if n.title != "" {
		fmt.Fprintf(buf, "title: %q,\n", n.title)
	}
	if n.npcType != "" {
		fmt.Fprintf(buf, "npcType: %q,\n", n.npcType)
	}
	if n.level != 0 {
		fmt.Fprintf(buf, "level: %d,\n", n.level)
	}
	if n.race != "" {
		fmt.Fprintf(buf, "race: %q,\n", n.race)
	}
	if n.sex != "" {
		fmt.Fprintf(buf, "sex: %q,\n", n.sex)
	}
	if n.displayId != 0 {
		fmt.Fprintf(buf, "displayId: %d,\n", n.displayId)
	}

	// Base stats
	writeNpcInt32(buf, "str", n.str)
	writeNpcInt32(buf, "intel", n.intel)
	writeNpcInt32(buf, "dex", n.dex)
	writeNpcInt32(buf, "wit", n.wit)
	writeNpcInt32(buf, "con", n.con)
	writeNpcInt32(buf, "men", n.men)

	// Vitals
	writeNpcFloat64(buf, "hp", n.hp)
	writeNpcFloat64(buf, "mp", n.mp)
	writeNpcFloat64(buf, "hpRegen", n.hpRegen)
	writeNpcFloat64(buf, "mpRegen", n.mpRegen)

	// Attack
	writeNpcFloat64(buf, "pAtk", n.pAtk)
	writeNpcFloat64(buf, "mAtk", n.mAtk)
	writeNpcFloat64(buf, "pDef", n.pDef)
	writeNpcFloat64(buf, "mDef", n.mDef)
	writeNpcFloat64(buf, "accuracy", n.accuracy)
	writeNpcInt32(buf, "critical", n.critical)
	writeNpcInt32(buf, "randomDamage", n.randomDamage)
	writeNpcFloat64(buf, "evasion", n.evasion)
	writeNpcInt32(buf, "atkSpeed", n.atkSpeed)
	if n.atkType != "" {
		fmt.Fprintf(buf, "atkType: %q,\n", n.atkType)
	}
	writeNpcInt32(buf, "atkRange", n.atkRange)
	writeNpcInt32(buf, "atkDistance", n.atkDistance)
	writeNpcInt32(buf, "atkWidth", n.atkWidth)

	// Speed
	writeNpcInt32(buf, "walkSpeed", n.walkSpeed)
	writeNpcInt32(buf, "runSpeed", n.runSpeed)
	writeNpcInt32(buf, "hitTime", n.hitTime)

	// AI
	writeNpcInt32(buf, "aggroRange", n.aggroRange)
	writeNpcInt32(buf, "clanHelpRange", n.clanHelpRange)
	if n.isAggressive {
		buf.WriteString("isAggressive: true,\n")
	}
	if len(n.clans) > 0 {
		buf.WriteString("clans: []string{")
		for i, c := range n.clans {
			if i > 0 {
				buf.WriteString(", ")
			}
			fmt.Fprintf(buf, "%q", c)
		}
		buf.WriteString("},\n")
	}
	if len(n.ignoreNpcIds) > 0 {
		buf.WriteString("ignoreNpcIds: []int32{")
		for i, id := range n.ignoreNpcIds {
			if i > 0 {
				buf.WriteString(", ")
			}
			fmt.Fprintf(buf, "%d", id)
		}
		buf.WriteString("},\n")
	}

	// Rewards
	writeNpcInt64(buf, "baseExp", n.baseExp)
	writeNpcInt32(buf, "baseSP", n.baseSP)

	// Drops
	if len(n.drops) > 0 {
		buf.WriteString("drops: []dropGroupDef{\n")
		for _, g := range n.drops {
			fmt.Fprintf(buf, "{chance: %g, items: []dropItemDef{", g.chance)
			for i, item := range g.items {
				if i > 0 {
					buf.WriteString(", ")
				}
				fmt.Fprintf(buf, "{itemID: %d, min: %d, max: %d, chance: %g}",
					item.itemID, item.min, item.max, item.chance)
			}
			buf.WriteString("}},\n")
		}
		buf.WriteString("},\n")
	}

	// Spoils
	if len(n.spoils) > 0 {
		buf.WriteString("spoils: []dropItemDef{")
		for i, item := range n.spoils {
			if i > 0 {
				buf.WriteString(", ")
			}
			fmt.Fprintf(buf, "{itemID: %d, min: %d, max: %d, chance: %g}",
				item.itemID, item.min, item.max, item.chance)
		}
		buf.WriteString("},\n")
	}

	// Skills
	if len(n.skills) > 0 {
		buf.WriteString("skills: []npcSkillDef{")
		for i, s := range n.skills {
			if i > 0 {
				buf.WriteString(", ")
			}
			fmt.Fprintf(buf, "{skillID: %d, level: %d}", s.skillID, s.level)
		}
		buf.WriteString("},\n")
	}

	// Collision
	writeNpcFloat64(buf, "collisionRadius", n.collisionRadius)
	writeNpcFloat64(buf, "collisionHeight", n.collisionHeight)

	// Equipment
	writeNpcInt32(buf, "rhand", n.rhand)
	writeNpcInt32(buf, "lhand", n.lhand)
	writeNpcInt32(buf, "chest", n.chest)

	// Status flags
	if n.undying {
		buf.WriteString("undying: true,\n")
	}
	if n.attackable {
		buf.WriteString("attackable: true,\n")
	}
	if n.talkable {
		buf.WriteString("talkable: true,\n")
	}
	if n.canBeSown {
		buf.WriteString("canBeSown: true,\n")
	}

	writeNpcInt32(buf, "corpseTime", n.corpseTime)

	// Minions
	if len(n.minions) > 0 {
		buf.WriteString("minions: []minionDef{")
		for i, m := range n.minions {
			if i > 0 {
				buf.WriteString(", ")
			}
			fmt.Fprintf(buf, "{npcID: %d, count: %d}", m.npcID, m.count)
		}
		buf.WriteString("},\n")
	}

	buf.WriteString("},\n")
}

func writeNpcInt32(buf *bytes.Buffer, field string, val int32) {
	if val != 0 {
		fmt.Fprintf(buf, "%s: %d,\n", field, val)
	}
}

func writeNpcInt64(buf *bytes.Buffer, field string, val int64) {
	if val != 0 {
		fmt.Fprintf(buf, "%s: %d,\n", field, val)
	}
}

func writeNpcFloat64(buf *bytes.Buffer, field string, val float64) {
	if val != 0 {
		fmt.Fprintf(buf, "%s: %g,\n", field, val)
	}
}

