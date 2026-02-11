package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// --- XML structures (skills) ---

type xmlSkillList struct {
	XMLName xml.Name   `xml:"list"`
	Skills  []xmlSkill `xml:"skill"`
}

type xmlSkill struct {
	ID              int32             `xml:"id,attr"`
	Levels          int32             `xml:"levels,attr"`
	Name            string            `xml:"name,attr"`
	EnchantGroup1   int32             `xml:"enchantGroup1,attr"`
	EnchantGroup2   int32             `xml:"enchantGroup2,attr"`
	Tables          []xmlSkillTable   `xml:"table"`
	Sets            []xmlSkillSet     `xml:"set"`
	Enchant1Sets    []xmlSkillSet     `xml:"enchant1"`
	Enchant2Sets    []xmlSkillSet     `xml:"enchant2"`
	Effects         *xmlSkillEffects  `xml:"effects"`
	Enchant1Effects *xmlSkillEffects  `xml:"enchant1effects"`
	Enchant2Effects *xmlSkillEffects  `xml:"enchant2effects"`
	Conditions      *xmlSkillConds    `xml:"conditions"`
}

type xmlSkillTable struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type xmlSkillSet struct {
	Name string `xml:"name,attr"`
	Val  string `xml:"val,attr"`
}

type xmlSkillEffects struct {
	Effects []xmlSkillEffect `xml:"effect"`
}

type xmlSkillEffect struct {
	Name   string              `xml:"name,attr"`
	Params []xmlSkillAnyAttrs  `xml:"param"`
	Adds   []xmlSkillStatMod   `xml:"add"`
	Muls   []xmlSkillStatMod   `xml:"mul"`
	Subs   []xmlSkillStatMod   `xml:"sub"`
	Sets   []xmlSkillStatMod   `xml:"set"`
}

type xmlSkillStatMod struct {
	Stat string `xml:"stat,attr"`
	Val  string `xml:"val,attr"`
}

type xmlSkillAnyAttrs struct {
	Attrs map[string]string
}

func (x *xmlSkillAnyAttrs) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	x.Attrs = make(map[string]string, len(start.Attr))
	for _, a := range start.Attr {
		x.Attrs[a.Name.Local] = a.Value
	}
	return d.Skip()
}

type xmlSkillConds struct {
	MsgId   int32              `xml:"msgId,attr"`
	AddName int32              `xml:"addName,attr"`
	Using   []xmlSkillAnyAttrs `xml:"using"`
	Player  []xmlSkillAnyAttrs `xml:"player"`
	Target  []xmlSkillAnyAttrs `xml:"target"`
	And     *xmlSkillCondGroup `xml:"and"`
	Or      *xmlSkillCondGroup `xml:"or"`
}

type xmlSkillCondGroup struct {
	Using  []xmlSkillAnyAttrs `xml:"using"`
	Player []xmlSkillAnyAttrs `xml:"player"`
	Target []xmlSkillAnyAttrs `xml:"target"`
	And    *xmlSkillCondGroup `xml:"and"`
	Or     *xmlSkillCondGroup `xml:"or"`
}

// --- Parsed structures (skills) ---

type parsedSkill struct {
	id, levels                   int32
	name                         string
	enchantGroup1, enchantGroup2 int32
	tables                       map[string][]string
	sets                         map[string]string
	effects                      []parsedSkillEffect
	enchant1Effects              []parsedSkillEffect
	enchant2Effects              []parsedSkillEffect
	enchant1Sets                 []parsedSkillEnchant
	enchant2Sets                 []parsedSkillEnchant
	conditions                   []parsedSkillCondition
	condMsgId, condAddName       int32
}

type parsedSkillEffect struct {
	name     string
	params   map[string]string
	perLvl   map[string][]string
	statMods []parsedSkillStatMod
}

type parsedSkillStatMod struct {
	op, stat string
	val      string
}

type parsedSkillEnchant struct {
	attr   string
	values []string
}

type parsedSkillCondition struct {
	typ      string
	params   map[string]string
	children []parsedSkillCondition
}

func generateSkills(javaDir, outDir string) error {
	skillsDir := filepath.Join(javaDir, "stats", "skills")
	skills, err := parseAllSkills(skillsDir)
	if err != nil {
		return fmt.Errorf("parse skills: %w", err)
	}

	sort.Slice(skills, func(i, j int) bool { return skills[i].id < skills[j].id })

	outPath := filepath.Join(outDir, "skill_data_generated.go")
	if err := generateSkillsGoFile(skills, outPath); err != nil {
		return fmt.Errorf("generate skills: %w", err)
	}

	fmt.Printf("  Generated %s: %d skills\n", outPath, len(skills))
	return nil
}

func parseAllSkills(dir string) ([]parsedSkill, error) {
	files, err := globXMLFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("glob skills dir: %w", err)
	}

	var all []parsedSkill
	for _, f := range files {
		skills, err := parseSkillFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filepath.Base(f), err)
		}
		all = append(all, skills...)
	}
	return all, nil
}

func parseSkillFile(path string) ([]parsedSkill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var list xmlSkillList
	if err := xml.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	result := make([]parsedSkill, 0, len(list.Skills))
	for _, xs := range list.Skills {
		ps := convertSkill(xs)
		result = append(result, ps)
	}
	return result, nil
}

func convertSkill(xs xmlSkill) parsedSkill {
	ps := parsedSkill{
		id:            xs.ID,
		levels:        xs.Levels,
		name:          xs.Name,
		enchantGroup1: xs.EnchantGroup1,
		enchantGroup2: xs.EnchantGroup2,
		tables:        make(map[string][]string),
		sets:          make(map[string]string),
	}

	for _, t := range xs.Tables {
		vals := strings.Fields(strings.TrimSpace(t.Value))
		ps.tables[t.Name] = vals
	}

	for _, s := range xs.Sets {
		ps.sets[s.Name] = s.Val
	}

	ps.enchant1Sets = convertSkillEnchantSets(xs.Enchant1Sets, ps.tables)
	ps.enchant2Sets = convertSkillEnchantSets(xs.Enchant2Sets, ps.tables)

	if xs.Effects != nil {
		ps.effects = convertSkillEffects(xs.Effects.Effects, ps.tables)
	}
	if xs.Enchant1Effects != nil {
		ps.enchant1Effects = convertSkillEffects(xs.Enchant1Effects.Effects, ps.tables)
	}
	if xs.Enchant2Effects != nil {
		ps.enchant2Effects = convertSkillEffects(xs.Enchant2Effects.Effects, ps.tables)
	}

	if xs.Conditions != nil {
		ps.condMsgId = xs.Conditions.MsgId
		ps.condAddName = xs.Conditions.AddName
		ps.conditions = convertSkillConditions(xs.Conditions)
	}

	return ps
}

func convertSkillEnchantSets(sets []xmlSkillSet, tables map[string][]string) []parsedSkillEnchant {
	if len(sets) == 0 {
		return nil
	}
	result := make([]parsedSkillEnchant, 0, len(sets))
	for _, s := range sets {
		pe := parsedSkillEnchant{attr: s.Name}
		if strings.HasPrefix(s.Val, "#") {
			if vals, ok := tables[s.Val]; ok {
				pe.values = vals
			} else {
				pe.values = []string{s.Val}
			}
		} else {
			pe.values = []string{s.Val}
		}
		result = append(result, pe)
	}
	return result
}

func convertSkillEffects(effects []xmlSkillEffect, tables map[string][]string) []parsedSkillEffect {
	result := make([]parsedSkillEffect, 0, len(effects))
	for _, xe := range effects {
		pe := parsedSkillEffect{
			name:   xe.Name,
			params: make(map[string]string),
			perLvl: make(map[string][]string),
		}

		for _, p := range xe.Params {
			for k, v := range p.Attrs {
				if strings.HasPrefix(v, "#") {
					if vals, ok := tables[v]; ok {
						pe.perLvl[k] = vals
					} else {
						pe.params[k] = v
					}
				} else {
					pe.params[k] = v
				}
			}
		}

		pe.statMods = append(pe.statMods, convertSkillStatMods("add", xe.Adds, tables)...)
		pe.statMods = append(pe.statMods, convertSkillStatMods("mul", xe.Muls, tables)...)
		pe.statMods = append(pe.statMods, convertSkillStatMods("sub", xe.Subs, tables)...)
		pe.statMods = append(pe.statMods, convertSkillStatMods("set", xe.Sets, tables)...)

		result = append(result, pe)
	}
	return result
}

func convertSkillStatMods(op string, mods []xmlSkillStatMod, tables map[string][]string) []parsedSkillStatMod {
	result := make([]parsedSkillStatMod, 0, len(mods))
	for _, m := range mods {
		psm := parsedSkillStatMod{op: op, stat: m.Stat}
		if strings.HasPrefix(m.Val, "#") {
			if vals, ok := tables[m.Val]; ok {
				psm.val = strings.Join(vals, " ")
			} else {
				psm.val = m.Val
			}
		} else {
			psm.val = m.Val
		}
		result = append(result, psm)
	}
	return result
}

func convertSkillConditions(xc *xmlSkillConds) []parsedSkillCondition {
	var result []parsedSkillCondition
	for _, u := range xc.Using {
		result = append(result, parsedSkillCondition{typ: "using", params: u.Attrs})
	}
	for _, p := range xc.Player {
		result = append(result, parsedSkillCondition{typ: "player", params: p.Attrs})
	}
	for _, t := range xc.Target {
		result = append(result, parsedSkillCondition{typ: "target", params: t.Attrs})
	}
	if xc.And != nil {
		result = append(result, parsedSkillCondition{typ: "and", children: convertSkillCondGroup(xc.And)})
	}
	if xc.Or != nil {
		result = append(result, parsedSkillCondition{typ: "or", children: convertSkillCondGroup(xc.Or)})
	}
	return result
}

func convertSkillCondGroup(g *xmlSkillCondGroup) []parsedSkillCondition {
	var result []parsedSkillCondition
	for _, u := range g.Using {
		result = append(result, parsedSkillCondition{typ: "using", params: u.Attrs})
	}
	for _, p := range g.Player {
		result = append(result, parsedSkillCondition{typ: "player", params: p.Attrs})
	}
	for _, t := range g.Target {
		result = append(result, parsedSkillCondition{typ: "target", params: t.Attrs})
	}
	if g.And != nil {
		result = append(result, parsedSkillCondition{typ: "and", children: convertSkillCondGroup(g.And)})
	}
	if g.Or != nil {
		result = append(result, parsedSkillCondition{typ: "or", children: convertSkillCondGroup(g.Or)})
	}
	return result
}

// --- Code generation (skills) ---

func generateSkillsGoFile(skills []parsedSkill, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "skills")
	buf.WriteString("var skillDefs = []skillDef{\n")

	for _, s := range skills {
		writeSkillDef(&buf, s)
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeSkillDef(buf *bytes.Buffer, s parsedSkill) {
	tables := s.tables

	fmt.Fprintf(buf, "{\n")
	fmt.Fprintf(buf, "id: %d, name: %q, levels: %d,\n", s.id, s.name, s.levels)

	if v, ok := s.sets["operateType"]; ok {
		fmt.Fprintf(buf, "operateType: %q,\n", v)
	}
	if v, ok := s.sets["targetType"]; ok {
		fmt.Fprintf(buf, "targetType: %q,\n", v)
	}

	writeSkillBoolSet(buf, s.sets, "isMagic", "isMagic")
	writeSkillBoolSet(buf, s.sets, "isDebuff", "isDebuff")
	writeSkillBoolSet(buf, s.sets, "overHit", "overHit")
	writeSkillBoolSet(buf, s.sets, "ignoreShld", "ignoreShld")
	writeSkillBoolSet(buf, s.sets, "nextActionAttack", "nextActionAttack")
	writeSkillBoolSet(buf, s.sets, "isSuicideAttack", "isSuicideAttack")
	writeSkillBoolSet(buf, s.sets, "stayAfterDeath", "stayAfterDeath")
	writeSkillBoolSet(buf, s.sets, "staticReuse", "staticReuse")
	writeSkillBoolSet(buf, s.sets, "removedOnDamage", "removedOnDamage")

	writeSkillInt32Set(buf, s.sets, "hitTime", "hitTime")
	writeSkillInt32Set(buf, s.sets, "coolTime", "coolTime")
	writeSkillInt32Set(buf, s.sets, "reuseDelay", "reuseDelay")
	writeSkillInt32Set(buf, s.sets, "castRange", "castRange")
	writeSkillInt32Set(buf, s.sets, "effectRange", "effectRange")
	writeSkillInt32Set(buf, s.sets, "element", "element")
	writeSkillInt32Set(buf, s.sets, "elementPower", "elementPower")
	writeSkillInt32Set(buf, s.sets, "affectRange", "affectRange")
	writeSkillInt32Set(buf, s.sets, "itemConsumeId", "itemConsumeId")
	writeSkillInt32Set(buf, s.sets, "itemConsumeCount", "itemConsumeCount")
	writeSkillInt32Set(buf, s.sets, "chargeConsume", "chargeConsume")
	writeSkillInt32Set(buf, s.sets, "blowChance", "blowChance")
	writeSkillInt32Set(buf, s.sets, "activateRate", "activateRate")
	writeSkillInt32Set(buf, s.sets, "lvlBonusRate", "lvlBonusRate")
	writeSkillInt32Set(buf, s.sets, "baseCritRate", "baseCritRate")

	writeSkillStringSet(buf, s.sets, "trait", "trait")
	writeSkillStringSet(buf, s.sets, "basicProperty", "basicProperty")
	writeSkillStringSet(buf, s.sets, "affectLimit", "affectLimit")
	writeSkillStringSet(buf, s.sets, "flyType", "flyType")
	writeSkillStringSet(buf, s.sets, "fanRange", "fanRange")
	writeSkillInt32Set(buf, s.sets, "flyRadius", "flyRadius")
	writeSkillStringSet(buf, s.sets, "abnormalType", "abnormalType")

	writeSkillScalarOrArray(buf, s.sets, tables, "magicLevel", "magicLevel", "magicLevelByLvl", "int32")
	writeSkillScalarOrArray(buf, s.sets, tables, "abnormalLevel", "abnormalLevel", "abnormalLevelTbl", "int32")

	writeSkillArrayField(buf, s.sets, tables, "power", "power", "float64")
	writeSkillArrayField(buf, s.sets, tables, "mpConsume", "mpConsume", "int32")
	writeSkillArrayField(buf, s.sets, tables, "hpConsume", "hpConsume", "int32")
	writeSkillArrayField(buf, s.sets, tables, "abnormalTime", "abnormalTime", "int32")
	writeSkillArrayField(buf, s.sets, tables, "effectPoint", "effectPoint", "int32")
	writeSkillArrayField(buf, s.sets, tables, "mpInitialConsume", "mpInitialConsume", "int32")

	if s.enchantGroup1 > 0 {
		fmt.Fprintf(buf, "enchantGroup1: %d,\n", s.enchantGroup1)
	}
	if s.enchantGroup2 > 0 {
		fmt.Fprintf(buf, "enchantGroup2: %d,\n", s.enchantGroup2)
	}

	writeSkillEnchantOverrides(buf, "enchant1", s.enchant1Sets)
	writeSkillEnchantOverrides(buf, "enchant2", s.enchant2Sets)

	if s.condMsgId != 0 {
		fmt.Fprintf(buf, "condMsgId: %d,\n", s.condMsgId)
	}
	if s.condAddName != 0 {
		fmt.Fprintf(buf, "condAddName: %d,\n", s.condAddName)
	}
	if len(s.conditions) > 0 {
		writeSkillConditionsDef(buf, "conditions", s.conditions)
	}

	writeSkillEffectsDef(buf, "effects", s.effects)
	writeSkillEffectsDef(buf, "enchant1Effects", s.enchant1Effects)
	writeSkillEffectsDef(buf, "enchant2Effects", s.enchant2Effects)

	fmt.Fprintf(buf, "},\n")
}

func writeSkillBoolSet(buf *bytes.Buffer, sets map[string]string, xmlName, goField string) {
	v, ok := sets[xmlName]
	if !ok {
		return
	}
	if v == "true" || v == "1" {
		fmt.Fprintf(buf, "%s: true,\n", goField)
	}
}

func writeSkillInt32Set(buf *bytes.Buffer, sets map[string]string, xmlName, goField string) {
	v, ok := sets[xmlName]
	if !ok || strings.HasPrefix(v, "#") {
		return
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil || n == 0 {
		return
	}
	fmt.Fprintf(buf, "%s: %d,\n", goField, n)
}

func writeSkillStringSet(buf *bytes.Buffer, sets map[string]string, xmlName, goField string) {
	v, ok := sets[xmlName]
	if !ok || v == "" || strings.HasPrefix(v, "#") {
		return
	}
	fmt.Fprintf(buf, "%s: %q,\n", goField, v)
}

func writeSkillScalarOrArray(buf *bytes.Buffer, sets map[string]string, tables map[string][]string, xmlName, scalarField, arrayField, elemType string) {
	v, ok := sets[xmlName]
	if !ok {
		return
	}
	if strings.HasPrefix(v, "#") {
		vals, ok := tables[v]
		if !ok || len(vals) == 0 {
			return
		}
		fmt.Fprintf(buf, "%s: []%s{%s},\n", arrayField, elemType, strings.Join(vals, ", "))
	} else {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil || n == 0 {
			return
		}
		fmt.Fprintf(buf, "%s: %d,\n", scalarField, n)
	}
}

func writeSkillArrayField(buf *bytes.Buffer, sets map[string]string, tables map[string][]string, xmlName, goField, elemType string) {
	v, ok := sets[xmlName]
	if !ok {
		return
	}
	var vals []string
	if strings.HasPrefix(v, "#") {
		var found bool
		vals, found = tables[v]
		if !found || len(vals) == 0 {
			return
		}
	} else {
		vals = []string{v}
	}
	fmt.Fprintf(buf, "%s: []%s{%s},\n", goField, elemType, strings.Join(vals, ", "))
}

func writeSkillEnchantOverrides(buf *bytes.Buffer, field string, enchants []parsedSkillEnchant) {
	if len(enchants) == 0 {
		return
	}
	fmt.Fprintf(buf, "%s: []enchantOverride{\n", field)
	for _, e := range enchants {
		fmt.Fprintf(buf, "{attr: %q, values: []string{", e.attr)
		for i, v := range e.values {
			if i > 0 {
				buf.WriteString(", ")
			}
			fmt.Fprintf(buf, "%q", v)
		}
		buf.WriteString("}},\n")
	}
	buf.WriteString("},\n")
}

func writeSkillConditionsDef(buf *bytes.Buffer, field string, conds []parsedSkillCondition) {
	fmt.Fprintf(buf, "%s: []conditionDef{\n", field)
	for _, c := range conds {
		writeSkillConditionDef(buf, c)
	}
	buf.WriteString("},\n")
}

func writeSkillConditionDef(buf *bytes.Buffer, c parsedSkillCondition) {
	fmt.Fprintf(buf, "{typ: %q", c.typ)
	if len(c.params) > 0 {
		buf.WriteString(", params: map[string]string{")
		keys := sortedKeys(c.params)
		for i, k := range keys {
			if i > 0 {
				buf.WriteString(", ")
			}
			fmt.Fprintf(buf, "%q: %q", k, c.params[k])
		}
		buf.WriteString("}")
	}
	if len(c.children) > 0 {
		buf.WriteString(", children: []conditionDef{\n")
		for _, ch := range c.children {
			writeSkillConditionDef(buf, ch)
		}
		buf.WriteString("}")
	}
	buf.WriteString("},\n")
}

func writeSkillEffectsDef(buf *bytes.Buffer, field string, effects []parsedSkillEffect) {
	if len(effects) == 0 {
		return
	}
	fmt.Fprintf(buf, "%s: []effectDef{\n", field)
	for _, e := range effects {
		fmt.Fprintf(buf, "{name: %q", e.name)

		if len(e.params) > 0 {
			buf.WriteString(", params: map[string]string{")
			keys := sortedKeys(e.params)
			for i, k := range keys {
				if i > 0 {
					buf.WriteString(", ")
				}
				fmt.Fprintf(buf, "%q: %q", k, e.params[k])
			}
			buf.WriteString("}")
		}

		if len(e.perLvl) > 0 {
			buf.WriteString(", perLvl: map[string][]string{")
			keys := sortedKeys(e.perLvl)
			for i, k := range keys {
				if i > 0 {
					buf.WriteString(", ")
				}
				fmt.Fprintf(buf, "%q: {", k)
				for j, v := range e.perLvl[k] {
					if j > 0 {
						buf.WriteString(", ")
					}
					fmt.Fprintf(buf, "%q", v)
				}
				buf.WriteString("}")
			}
			buf.WriteString("}")
		}

		if len(e.statMods) > 0 {
			buf.WriteString(", statMods: []statModDef{")
			for i, sm := range e.statMods {
				if i > 0 {
					buf.WriteString(", ")
				}
				fmt.Fprintf(buf, "{op: %q, stat: %q, val: %q}", sm.op, sm.stat, sm.val)
			}
			buf.WriteString("}")
		}

		buf.WriteString("},\n")
	}
	buf.WriteString("},\n")
}
