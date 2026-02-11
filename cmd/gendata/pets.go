package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// --- XML structures (pets) ---

type xmlPetsFile struct {
	XMLName xml.Name `xml:"pets"`
	Pets    []xmlPet `xml:"pet"`
}

type xmlPet struct {
	ID     int32      `xml:"id,attr"`
	ItemID int32      `xml:"itemId,attr"`
	Sets   []xmlPetSet `xml:"set"`
	Stats  xmlPetStats `xml:"stats"`
}

type xmlPetSet struct {
	Name string `xml:"name,attr"`
	Val  string `xml:"val,attr"`
}

type xmlPetStats struct {
	Stats []xmlPetStat `xml:"stat"`
}

type xmlPetStat struct {
	Level int32      `xml:"level,attr"`
	Sets  []xmlPetSet `xml:"set"`
}

// --- Parsed structures ---

type parsedPet struct {
	npcID  int32
	itemID int32
	food   int32
	levels []parsedPetLevel
}

type parsedPetLevel struct {
	level  int32
	exp    int64
	hp     float64
	mp     float64
	pAtk   float64
	pDef   float64
	mAtk   float64
	mDef   float64
	maxFeed int32
	feedBattle float64
	feedNormal float64
}

func generatePets(javaDir, outDir string) error {
	petsDir := filepath.Join(javaDir, "stats", "pets")
	pets, err := parseAllPets(petsDir)
	if err != nil {
		return fmt.Errorf("parse pets: %w", err)
	}

	sort.Slice(pets, func(i, j int) bool {
		return pets[i].npcID < pets[j].npcID
	})

	outPath := filepath.Join(outDir, "pet_data_generated.go")
	if err := generatePetsGoFile(pets, outPath); err != nil {
		return fmt.Errorf("generate pets: %w", err)
	}

	fmt.Printf("  Generated %s: %d pets\n", outPath, len(pets))
	return nil
}

func parseAllPets(dir string) ([]parsedPet, error) {
	files, err := globXMLFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("glob pets dir: %w", err)
	}

	var all []parsedPet
	for _, f := range files {
		pet, err := parsePetFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filepath.Base(f), err)
		}
		all = append(all, *pet)
	}
	return all, nil
}

func parsePetFile(path string) (*parsedPet, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var file xmlPetsFile
	if err := xml.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	if len(file.Pets) == 0 {
		return nil, fmt.Errorf("no pets in file")
	}

	xp := file.Pets[0]
	pet := &parsedPet{
		npcID:  xp.ID,
		itemID: xp.ItemID,
	}

	// Parse top-level sets (food, hungry_limit, etc.)
	for _, s := range xp.Sets {
		switch s.Name {
		case "food":
			v, _ := strconv.ParseInt(s.Val, 10, 32)
			pet.food = int32(v)
		}
	}

	// Parse per-level stats
	for _, stat := range xp.Stats.Stats {
		lvl := parsedPetLevel{level: stat.Level}
		for _, s := range stat.Sets {
			switch s.Name {
			case "exp":
				lvl.exp, _ = strconv.ParseInt(s.Val, 10, 64)
			case "org_hp":
				lvl.hp, _ = strconv.ParseFloat(s.Val, 64)
			case "org_mp":
				lvl.mp, _ = strconv.ParseFloat(s.Val, 64)
			case "org_pattack":
				lvl.pAtk, _ = strconv.ParseFloat(s.Val, 64)
			case "org_pdefend":
				lvl.pDef, _ = strconv.ParseFloat(s.Val, 64)
			case "org_mattack":
				lvl.mAtk, _ = strconv.ParseFloat(s.Val, 64)
			case "org_mdefend":
				lvl.mDef, _ = strconv.ParseFloat(s.Val, 64)
			case "max_meal":
				v, _ := strconv.ParseInt(s.Val, 10, 32)
				lvl.maxFeed = int32(v)
			case "consume_meal_in_battle":
				lvl.feedBattle, _ = strconv.ParseFloat(s.Val, 64)
			case "consume_meal_in_normal":
				lvl.feedNormal, _ = strconv.ParseFloat(s.Val, 64)
			}
		}
		pet.levels = append(pet.levels, lvl)
	}

	return pet, nil
}

// --- Code generation ---

func generatePetsGoFile(pets []parsedPet, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "pets")
	buf.WriteString("var petDefs = []petDef{\n")

	for i := range pets {
		writePetDef(&buf, &pets[i])
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writePetDef(buf *bytes.Buffer, p *parsedPet) {
	fmt.Fprintf(buf, "{npcID: %d, itemID: %d, levels: []petLevelDef{", p.npcID, p.itemID)
	for i, lvl := range p.levels {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(buf, "{level: %d, exp: %d, hp: %g, mp: %g, pAtk: %g, pDef: %g, mAtk: %g, mDef: %g, maxFeed: %d, feedRate: %g}",
			lvl.level, lvl.exp, lvl.hp, lvl.mp, lvl.pAtk, lvl.pDef, lvl.mAtk, lvl.mDef, lvl.maxFeed, lvl.feedBattle)
	}
	buf.WriteString("}},\n")
}
