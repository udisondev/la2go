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

// --- XML structures (augmentation_skillmap) ---

type xmlAugmentationList struct {
	XMLName       xml.Name           `xml:"list"`
	Augmentations []xmlAugmentation  `xml:"augmentation"`
}

type xmlAugmentation struct {
	ID         int32           `xml:"id,attr"`
	SkillID    xmlValInt32     `xml:"skillId"`
	SkillLevel xmlValInt32     `xml:"skillLevel"`
	Type       xmlValString    `xml:"type"`
}

// xmlValInt32 parses child elements like <skillId val="3203" />.
type xmlValInt32 struct {
	Val string `xml:"val,attr"`
}

func (v xmlValInt32) Int32() (int32, error) {
	n, err := strconv.ParseInt(v.Val, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse int32 %q: %w", v.Val, err)
	}
	return int32(n), nil
}

// xmlValString parses child elements like <type val="blue" />.
type xmlValString struct {
	Val string `xml:"val,attr"`
}

// --- Parsed structures (augmentation) ---

type parsedAugmentation struct {
	id         int32
	skillID    int32
	skillLevel int32
	augType    string
}

func generateAugmentation(javaDir, outDir string) error {
	augDir := filepath.Join(javaDir, "stats", "augmentation")
	skillmapPath := filepath.Join(augDir, "augmentation_skillmap.xml")

	augs, err := parseAugmentationSkillmap(skillmapPath)
	if err != nil {
		return fmt.Errorf("parse augmentation skillmap: %w", err)
	}

	sort.Slice(augs, func(i, j int) bool { return augs[i].id < augs[j].id })

	outPath := filepath.Join(outDir, "augmentation_data_generated.go")
	if err := generateAugmentationGoFile(augs, outPath); err != nil {
		return fmt.Errorf("generate augmentation: %w", err)
	}

	fmt.Printf("  Generated %s: %d augmentations\n", outPath, len(augs))
	return nil
}

func parseAugmentationSkillmap(path string) ([]parsedAugmentation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", filepath.Base(path), err)
	}

	var list xmlAugmentationList
	if err := xml.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", filepath.Base(path), err)
	}

	augs := make([]parsedAugmentation, 0, len(list.Augmentations))
	for _, xa := range list.Augmentations {
		skillID, err := xa.SkillID.Int32()
		if err != nil {
			return nil, fmt.Errorf("augmentation %d skillId: %w", xa.ID, err)
		}

		skillLevel, err := xa.SkillLevel.Int32()
		if err != nil {
			return nil, fmt.Errorf("augmentation %d skillLevel: %w", xa.ID, err)
		}

		augType := xa.Type.Val
		if augType == "" {
			return nil, fmt.Errorf("augmentation %d: empty type", xa.ID)
		}

		augs = append(augs, parsedAugmentation{
			id:         xa.ID,
			skillID:    skillID,
			skillLevel: skillLevel,
			augType:    augType,
		})
	}

	return augs, nil
}

func generateAugmentationGoFile(augs []parsedAugmentation, outPath string) error {
	var buf bytes.Buffer

	writeHeader(&buf, "augmentation.go")

	fmt.Fprintf(&buf, "var augmentationDefs = []augmentationDef{\n")
	for _, a := range augs {
		fmt.Fprintf(&buf, "\t{id: %d, skillID: %d, skillLevel: %d, augType: %q},\n",
			a.id, a.skillID, a.skillLevel, a.augType)
	}
	fmt.Fprintf(&buf, "}\n")

	return writeGoFile(outPath, buf.Bytes())
}
