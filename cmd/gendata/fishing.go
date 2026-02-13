package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// --- XML structures (fishing) ---

type xmlFishesFile struct {
	XMLName xml.Name  `xml:"list"`
	Fishes  []xmlFish `xml:"fish"`
}

type xmlFish struct {
	FishID         int32   `xml:"fishId,attr"`
	ItemID         int32   `xml:"itemId,attr"`
	FishGroup      string  `xml:"fishGroup,attr"`
	FishLevel      int32   `xml:"fishLevel,attr"`
	FishHP         int32   `xml:"fishHp,attr"`
	FishGrade      string  `xml:"fishGrade,attr"`
	HPRegen        float64 `xml:"hpRegen,attr"`
	CombatDuration int32   `xml:"combatDuration,attr"`
}

type xmlRodsFile struct {
	XMLName xml.Name     `xml:"list"`
	Rods    []xmlRodItem `xml:"fishingRod"`
}

type xmlRodItem struct {
	ID     int32   `xml:"fishingRodId,attr"`
	ItemID int32   `xml:"fishingRodItemId,attr"`
	Level  int32   `xml:"fishingRodLevel,attr"`
	Name   string  `xml:"fishingRodName,attr"`
	Damage float64 `xml:"fishingRodDamage,attr"`
}

type xmlMonstersFile struct {
	XMLName  xml.Name          `xml:"list"`
	Monsters []xmlMonsterEntry `xml:"fishingMonster"`
}

type xmlMonsterEntry struct {
	MinLevel  int32 `xml:"userMinLevel,attr"`
	MaxLevel  int32 `xml:"userMaxLevel,attr"`
	MonsterID int32 `xml:"fishingMonsterId,attr"`
	Chance    int32 `xml:"probability,attr"`
}

// --- Parsed structures ---

type parsedFish struct {
	id             int32
	itemID         int32
	fishType       string // fishGroup in XML
	group          int32  // derived from fishGrade
	level          int32
	hp             int32
	hpRegen        float64
	combatDuration int32
	fishGrade      int32 // 0=easy,1=normal,2=hard
}

func generateFishing(javaDir, outDir string) error {
	fishFile := filepath.Join(javaDir, "stats", "fishing", "fishes.xml")
	fishes, err := parseFishesFile(fishFile)
	if err != nil {
		return fmt.Errorf("parse fishes: %w", err)
	}

	sort.Slice(fishes, func(i, j int) bool {
		return fishes[i].id < fishes[j].id
	})

	rodsFile := filepath.Join(javaDir, "stats", "fishing", "fishingRods.xml")
	rods, err := parseRodsFile(rodsFile)
	if err != nil {
		return fmt.Errorf("parse rods: %w", err)
	}

	monstersFile := filepath.Join(javaDir, "stats", "fishing", "fishingMonsters.xml")
	monsters, err := parseMonstersFile(monstersFile)
	if err != nil {
		return fmt.Errorf("parse monsters: %w", err)
	}

	outPath := filepath.Join(outDir, "fishing_data_generated.go")
	if err := generateFishingGoFile(fishes, rods, monsters, outPath); err != nil {
		return fmt.Errorf("generate fishing: %w", err)
	}

	fmt.Printf("  Generated %s: %d fish, %d rods, %d monsters\n", outPath, len(fishes), len(rods), len(monsters))
	return nil
}

func parseFishesFile(path string) ([]parsedFish, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var file xmlFishesFile
	if err := xml.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	fishes := make([]parsedFish, 0, len(file.Fishes))
	for _, xf := range file.Fishes {
		f := parsedFish{
			id:             xf.FishID,
			itemID:         xf.ItemID,
			fishType:       xf.FishGroup,
			level:          xf.FishLevel,
			hp:             xf.FishHP,
			group:          fishGradeToGroup(xf.FishGrade),
			hpRegen:        xf.HPRegen,
			combatDuration: xf.CombatDuration,
			fishGrade:      fishGradeToNumeric(xf.FishGrade),
		}
		fishes = append(fishes, f)
	}
	return fishes, nil
}

func parseRodsFile(path string) ([]xmlRodItem, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var file xmlRodsFile
	if err := xml.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return file.Rods, nil
}

func parseMonstersFile(path string) ([]xmlMonsterEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var file xmlMonstersFile
	if err := xml.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return file.Monsters, nil
}

func fishGradeToGroup(grade string) int32 {
	switch grade {
	case "fish_normal":
		return 0
	case "fish_easy":
		return 1
	case "fish_hard":
		return 2
	default:
		return 0
	}
}

func fishGradeToNumeric(grade string) int32 {
	switch grade {
	case "fish_easy":
		return 0
	case "fish_normal":
		return 1
	case "fish_hard":
		return 2
	default:
		return 1
	}
}

// --- Code generation ---

func generateFishingGoFile(fishes []parsedFish, rods []xmlRodItem, monsters []xmlMonsterEntry, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "fishing")

	// Fish defs
	buf.WriteString("var fishDefs = []fishDef{\n")
	for i := range fishes {
		writeFishDef(&buf, &fishes[i])
	}
	buf.WriteString("}\n\n")

	// Rod defs
	buf.WriteString("var fishingRodDefs = []fishingRodDef{\n")
	for _, r := range rods {
		fmt.Fprintf(&buf, "\t{id: %d, itemID: %d, level: %d, name: %q, damage: %.1f},\n",
			r.ID, r.ItemID, r.Level, r.Name, r.Damage)
	}
	buf.WriteString("}\n\n")

	// Monster defs
	buf.WriteString("var fishingMonsterDefs = []fishingMonsterDef{\n")
	for _, m := range monsters {
		fmt.Fprintf(&buf, "\t{minLevel: %d, maxLevel: %d, monsterID: %d, chance: %d},\n",
			m.MinLevel, m.MaxLevel, m.MonsterID, m.Chance)
	}
	buf.WriteString("}\n")

	return writeGoFile(outPath, buf.Bytes())
}

func writeFishDef(buf *bytes.Buffer, f *parsedFish) {
	fmt.Fprintf(buf, "\t{id: %d, itemID: %d, fishType: %q, group: %d, level: %d, hp: %d, hpRegen: %.1f, combatDuration: %d, fishGrade: %d},\n",
		f.id, f.itemID, f.fishType, f.group, f.level, f.hp, f.hpRegen, f.combatDuration, f.fishGrade)
}
