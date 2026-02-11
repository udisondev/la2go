package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// --- XML structures (seeds) ---

type xmlSeedsFile struct {
	XMLName xml.Name        `xml:"list"`
	Castles []xmlSeedCastle `xml:"castle"`
}

type xmlSeedCastle struct {
	ID    int32         `xml:"id,attr"`
	Crops []xmlSeedCrop `xml:"crop"`
}

type xmlSeedCrop struct {
	CropID   int32 `xml:"id,attr"`
	SeedID   int32 `xml:"seedId,attr"`
	MatureID int32 `xml:"mature_Id,attr"`
	Reward1  int32 `xml:"reward1,attr"`
	Reward2  int32 `xml:"reward2,attr"`
	Level    int32 `xml:"level,attr"`
}

// --- Parsed structure ---

type parsedSeed struct {
	castleID int32
	cropID   int32
	seedID   int32
	matureID int32
	reward1  int32
	reward2  int32
	level    int32
}

func generateSeeds(javaDir, outDir string) error {
	seedsPath := filepath.Join(javaDir, "Seeds.xml")
	seeds, err := parseSeedsFile(seedsPath)
	if err != nil {
		return fmt.Errorf("parse seeds: %w", err)
	}

	sort.Slice(seeds, func(i, j int) bool {
		if seeds[i].castleID != seeds[j].castleID {
			return seeds[i].castleID < seeds[j].castleID
		}
		return seeds[i].seedID < seeds[j].seedID
	})

	outPath := filepath.Join(outDir, "seed_data_generated.go")
	if err := generateSeedsGoFile(seeds, outPath); err != nil {
		return fmt.Errorf("generate seeds: %w", err)
	}

	fmt.Printf("  Generated %s: %d seed entries\n", outPath, len(seeds))
	return nil
}

func parseSeedsFile(path string) ([]parsedSeed, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var file xmlSeedsFile
	if err := xml.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	var seeds []parsedSeed
	for _, castle := range file.Castles {
		for _, crop := range castle.Crops {
			seeds = append(seeds, parsedSeed{
				castleID: castle.ID,
				cropID:   crop.CropID,
				seedID:   crop.SeedID,
				matureID: crop.MatureID,
				reward1:  crop.Reward1,
				reward2:  crop.Reward2,
				level:    crop.Level,
			})
		}
	}
	return seeds, nil
}

// --- Code generation ---

func generateSeedsGoFile(seeds []parsedSeed, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "seeds")
	buf.WriteString("var seedDefs = []seedDef{\n")

	for i := range seeds {
		writeSeedDef(&buf, &seeds[i])
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeSeedDef(buf *bytes.Buffer, s *parsedSeed) {
	fmt.Fprintf(buf, "{castleID: %d, cropID: %d, seedID: %d, matureID: %d, reward1: %d, reward2: %d, level: %d},\n",
		s.castleID, s.cropID, s.seedID, s.matureID, s.reward1, s.reward2, s.level)
}
