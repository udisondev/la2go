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
	FishID    int32   `xml:"fishId,attr"`
	ItemID    int32   `xml:"itemId,attr"`
	FishGroup string  `xml:"fishGroup,attr"`
	FishLevel int32   `xml:"fishLevel,attr"`
	FishHP    int32   `xml:"fishHp,attr"`
	FishGrade string  `xml:"fishGrade,attr"`
	HPRegen   float64 `xml:"hpRegen,attr"`
}

// --- Parsed structure ---

type parsedFish struct {
	id       int32
	itemID   int32
	fishType string // fishGroup in XML
	group    int32  // derived from fishGrade
	level    int32
	hp       int32
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

	outPath := filepath.Join(outDir, "fishing_data_generated.go")
	if err := generateFishingGoFile(fishes, outPath); err != nil {
		return fmt.Errorf("generate fishing: %w", err)
	}

	fmt.Printf("  Generated %s: %d fish entries\n", outPath, len(fishes))
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
			id:       xf.FishID,
			itemID:   xf.ItemID,
			fishType: xf.FishGroup,
			level:    xf.FishLevel,
			hp:       xf.FishHP,
			group:    fishGradeToGroup(xf.FishGrade),
		}
		fishes = append(fishes, f)
	}
	return fishes, nil
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

// --- Code generation ---

func generateFishingGoFile(fishes []parsedFish, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "fishing")
	buf.WriteString("var fishDefs = []fishDef{\n")

	for i := range fishes {
		writeFishDef(&buf, &fishes[i])
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeFishDef(buf *bytes.Buffer, f *parsedFish) {
	fmt.Fprintf(buf, "{id: %d, itemID: %d, fishType: %q, group: %d, level: %d, hp: %d},\n",
		f.id, f.itemID, f.fishType, f.group, f.level, f.hp)
}
