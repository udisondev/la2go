package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// --- XML structures (armorsets) ---

type xmlArmorsetList struct {
	XMLName xml.Name       `xml:"list"`
	Sets    []xmlArmorset  `xml:"set"`
}

type xmlArmorset struct {
	ID            int32              `xml:"id,attr"`
	Chests        []xmlArmorsetPart  `xml:"chest"`
	Legs          []xmlArmorsetPart  `xml:"legs"`
	Heads         []xmlArmorsetPart  `xml:"head"`
	Gloves        []xmlArmorsetPart  `xml:"gloves"`
	Feet          []xmlArmorsetPart  `xml:"feet"`
	Shields       []xmlArmorsetPart  `xml:"shield"`
	Skills        []xmlArmorsetSkill `xml:"skill"`
	ShieldSkills  []xmlArmorsetSkill `xml:"shield_skill"`
	Enchant6Skills []xmlArmorsetSkill `xml:"enchant6skill"`
	Str           []xmlArmorsetStat  `xml:"str"`
	Con           []xmlArmorsetStat  `xml:"con"`
	Dex           []xmlArmorsetStat  `xml:"dex"`
	Int           []xmlArmorsetStat  `xml:"int"`
	Men           []xmlArmorsetStat  `xml:"men"`
	Wit           []xmlArmorsetStat  `xml:"wit"`
}

type xmlArmorsetPart struct {
	ID int32 `xml:"id,attr"`
}

type xmlArmorsetSkill struct {
	ID    int32 `xml:"id,attr"`
	Level int32 `xml:"level,attr"`
}

type xmlArmorsetStat struct {
	Val int32 `xml:"val,attr"`
}

// --- Parsed structure (armorsets) ---

type parsedArmorset struct {
	setID                                   int32
	chest, legs, head, gloves, feet, shield int32
	skillID, skillLevel                     int32
	shieldSkillID, shieldSkillLevel         int32
	enchant6SkillID, enchant6SkillLevel     int32
	strMod, conMod, dexMod, intMod          int32
}

func generateArmorsets(javaDir, outDir string) error {
	armorsetsDir := filepath.Join(javaDir, "stats", "armorsets")
	sets, err := parseAllArmorsets(armorsetsDir)
	if err != nil {
		return fmt.Errorf("parse armorsets: %w", err)
	}

	sort.Slice(sets, func(i, j int) bool { return sets[i].setID < sets[j].setID })

	outPath := filepath.Join(outDir, "armorset_data_generated.go")
	if err := generateArmorsetsGoFile(sets, outPath); err != nil {
		return fmt.Errorf("generate armorsets: %w", err)
	}

	fmt.Printf("  Generated %s: %d armor sets\n", outPath, len(sets))
	return nil
}

func parseAllArmorsets(dir string) ([]parsedArmorset, error) {
	files, err := globXMLFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("glob armorsets dir: %w", err)
	}

	var all []parsedArmorset
	for _, f := range files {
		sets, err := parseArmorsetFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filepath.Base(f), err)
		}
		all = append(all, sets...)
	}
	return all, nil
}

func parseArmorsetFile(path string) ([]parsedArmorset, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var list xmlArmorsetList
	if err := xml.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	result := make([]parsedArmorset, 0, len(list.Sets))
	for _, xs := range list.Sets {
		result = append(result, convertArmorset(xs))
	}
	return result, nil
}

func convertArmorset(xs xmlArmorset) parsedArmorset {
	ps := parsedArmorset{
		setID: xs.ID,
	}

	// Части сета (берём первый элемент, если есть)
	if len(xs.Chests) > 0 {
		ps.chest = xs.Chests[0].ID
	}
	if len(xs.Legs) > 0 {
		ps.legs = xs.Legs[0].ID
	}
	if len(xs.Heads) > 0 {
		ps.head = xs.Heads[0].ID
	}
	if len(xs.Gloves) > 0 {
		ps.gloves = xs.Gloves[0].ID
	}
	if len(xs.Feet) > 0 {
		ps.feet = xs.Feet[0].ID
	}
	if len(xs.Shields) > 0 {
		ps.shield = xs.Shields[0].ID
	}

	// Скиллы: в XML обычно 2 <skill> — первый id=3006 (общий "Equip Set Items"),
	// второй — уникальный бонус сета. Берём последний не-3006 скилл.
	for i := len(xs.Skills) - 1; i >= 0; i-- {
		if xs.Skills[i].ID != 3006 {
			ps.skillID = xs.Skills[i].ID
			ps.skillLevel = xs.Skills[i].Level
			break
		}
	}

	if len(xs.ShieldSkills) > 0 {
		ps.shieldSkillID = xs.ShieldSkills[0].ID
		ps.shieldSkillLevel = xs.ShieldSkills[0].Level
	}

	if len(xs.Enchant6Skills) > 0 {
		ps.enchant6SkillID = xs.Enchant6Skills[0].ID
		ps.enchant6SkillLevel = xs.Enchant6Skills[0].Level
	}

	// Стат-модификаторы
	if len(xs.Str) > 0 {
		ps.strMod = xs.Str[0].Val
	}
	if len(xs.Con) > 0 {
		ps.conMod = xs.Con[0].Val
	}
	if len(xs.Dex) > 0 {
		ps.dexMod = xs.Dex[0].Val
	}
	if len(xs.Int) > 0 {
		ps.intMod = xs.Int[0].Val
	}
	// men и wit в XML присутствуют, но в armorsetDef для них нет полей — игнорируем

	return ps
}

// --- Code generation (armorsets) ---

func generateArmorsetsGoFile(sets []parsedArmorset, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "armorsets")
	buf.WriteString("var armorsetDefs = []armorsetDef{\n")

	for _, s := range sets {
		writeArmorsetDef(&buf, s)
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeArmorsetDef(buf *bytes.Buffer, s parsedArmorset) {
	buf.WriteString("{")
	fmt.Fprintf(buf, "setID: %d", s.setID)
	fmt.Fprintf(buf, ", chest: %d", s.chest)

	writeArmorsetInt32(buf, "legs", s.legs)
	writeArmorsetInt32(buf, "head", s.head)
	writeArmorsetInt32(buf, "gloves", s.gloves)
	writeArmorsetInt32(buf, "feet", s.feet)
	writeArmorsetInt32(buf, "shield", s.shield)
	writeArmorsetInt32(buf, "skillID", s.skillID)
	writeArmorsetInt32(buf, "skillLevel", s.skillLevel)
	writeArmorsetInt32(buf, "shieldSkillID", s.shieldSkillID)
	writeArmorsetInt32(buf, "shieldSkillLevel", s.shieldSkillLevel)
	writeArmorsetInt32(buf, "enchant6SkillID", s.enchant6SkillID)
	writeArmorsetInt32(buf, "enchant6SkillLevel", s.enchant6SkillLevel)
	writeArmorsetNonZeroInt32(buf, "strMod", s.strMod)
	writeArmorsetNonZeroInt32(buf, "conMod", s.conMod)
	writeArmorsetNonZeroInt32(buf, "dexMod", s.dexMod)
	writeArmorsetNonZeroInt32(buf, "intMod", s.intMod)

	buf.WriteString("},\n")
}

// writeArmorsetInt32 записывает int32 поле (включая нулевые значения).
func writeArmorsetInt32(buf *bytes.Buffer, field string, val int32) {
	fmt.Fprintf(buf, ", %s: %d", field, val)
}

// writeArmorsetNonZeroInt32 записывает int32 поле, пропуская нулевые значения.
func writeArmorsetNonZeroInt32(buf *bytes.Buffer, field string, val int32) {
	if val == 0 {
		return
	}
	fmt.Fprintf(buf, ", %s: %d", field, val)
}
