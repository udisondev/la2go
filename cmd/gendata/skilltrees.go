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

// --- XML structures (skill trees) ---

type xmlTreeList struct {
	XMLName xml.Name       `xml:"list"`
	Trees   []xmlSkillTree `xml:"skillTree"`
}

type xmlSkillTree struct {
	Type          string         `xml:"type,attr"`
	ClassID       int32          `xml:"classId,attr"`
	ParentClassID int32          `xml:"parentClassId,attr"`
	Skills        []xmlTreeSkill `xml:"skill"`
}

type xmlTreeSkill struct {
	SkillName    string        `xml:"skillName,attr"`
	SkillID      int32         `xml:"skillId,attr"`
	SkillLevel   int32         `xml:"skillLevel,attr"`
	GetLevel     int32         `xml:"getLevel,attr"`
	LevelUpSp    int64         `xml:"levelUpSp,attr"`
	AutoGet      xmlTreeBool   `xml:"autoGet,attr"`
	LearnedByNpc xmlTreeBool   `xml:"learnedByNpc,attr"`
	Items        []xmlTreeItem `xml:"item"`
	SocialClass  string        `xml:"socialClass"`
	Race         string        `xml:"race"`
}

type xmlTreeItem struct {
	ID    int32 `xml:"id,attr"`
	Count int32 `xml:"count,attr"`
}

type xmlTreeBool bool

func (b *xmlTreeBool) UnmarshalXMLAttr(attr xml.Attr) error {
	*b = xmlTreeBool(attr.Value == "true")
	return nil
}

// --- Parsed structures (skill trees) ---

type parsedTree struct {
	treeType      string
	classID       int32
	parentClassID int32
	skills        []parsedTreeEntry
}

type parsedTreeEntry struct {
	skillID      int32
	skillLevel   int32
	minLevel     int32
	spCost       int64
	autoGet      bool
	learnedByNpc bool
	items        []parsedTreeItem
	socialClass  string
	race         string
}

type parsedTreeItem struct {
	id    int32
	count int32
}

func generateSkillTrees(javaDir, outDir string) error {
	treesDir := filepath.Join(javaDir, "stats", "players", "skillTrees")
	trees, err := parseAllTrees(treesDir)
	if err != nil {
		return fmt.Errorf("parse skill trees: %w", err)
	}

	sort.Slice(trees, func(i, j int) bool {
		ti, tj := trees[i], trees[j]
		if ti.treeType != tj.treeType {
			if ti.treeType == "classSkillTree" {
				return true
			}
			if tj.treeType == "classSkillTree" {
				return false
			}
			return ti.treeType < tj.treeType
		}
		return ti.classID < tj.classID
	})

	outPath := filepath.Join(outDir, "skill_tree_data_generated.go")
	if err := generateTreesGoFile(trees, outPath); err != nil {
		return fmt.Errorf("generate skill trees: %w", err)
	}

	var totalEntries int
	for _, t := range trees {
		totalEntries += len(t.skills)
	}
	fmt.Printf("  Generated %s: %d trees, %d entries\n", outPath, len(trees), totalEntries)
	return nil
}

func parseAllTrees(dir string) ([]parsedTree, error) {
	var all []parsedTree

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".xml") {
			return nil
		}
		trees, err := parseTreeFile(path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", filepath.Base(path), err)
		}
		all = append(all, trees...)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk trees dir: %w", err)
	}

	return all, nil
}

func parseTreeFile(path string) ([]parsedTree, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var list xmlTreeList
	if err := xml.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	result := make([]parsedTree, 0, len(list.Trees))
	for _, xt := range list.Trees {
		pt := parsedTree{
			treeType:      xt.Type,
			classID:       xt.ClassID,
			parentClassID: xt.ParentClassID,
			skills:        make([]parsedTreeEntry, 0, len(xt.Skills)),
		}

		for _, xs := range xt.Skills {
			entry := parsedTreeEntry{
				skillID:      xs.SkillID,
				skillLevel:   xs.SkillLevel,
				minLevel:     xs.GetLevel,
				spCost:       xs.LevelUpSp,
				autoGet:      bool(xs.AutoGet),
				learnedByNpc: bool(xs.LearnedByNpc),
				socialClass:  strings.TrimSpace(xs.SocialClass),
				race:         strings.TrimSpace(xs.Race),
			}

			for _, xi := range xs.Items {
				entry.items = append(entry.items, parsedTreeItem{id: xi.ID, count: xi.Count})
			}

			pt.skills = append(pt.skills, entry)
		}

		result = append(result, pt)
	}
	return result, nil
}

func generateTreesGoFile(trees []parsedTree, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "skilltrees")
	buf.WriteString("var skillTreeDefs = []skillTreeDef{\n")

	for _, t := range trees {
		writeTreeDef(&buf, t)
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeTreeDef(buf *bytes.Buffer, t parsedTree) {
	fmt.Fprintf(buf, "{\n")
	fmt.Fprintf(buf, "treeType: %q, classID: %d", t.treeType, t.classID)
	if t.parentClassID != 0 {
		fmt.Fprintf(buf, ", parentClassID: %d", t.parentClassID)
	}
	buf.WriteString(",\n")
	buf.WriteString("skills: []skillTreeEntry{\n")

	for _, s := range t.skills {
		writeTreeEntryDef(buf, s)
	}

	buf.WriteString("},\n")
	buf.WriteString("},\n")
}

func writeTreeEntryDef(buf *bytes.Buffer, s parsedTreeEntry) {
	fmt.Fprintf(buf, "{skillID: %d, skillLevel: %d, minLevel: %d", s.skillID, s.skillLevel, s.minLevel)

	if s.spCost != 0 {
		fmt.Fprintf(buf, ", spCost: %d", s.spCost)
	}
	if s.autoGet {
		buf.WriteString(", autoGet: true")
	}
	if s.learnedByNpc {
		buf.WriteString(", learnedByNpc: true")
	}
	if s.socialClass != "" {
		fmt.Fprintf(buf, ", socialClass: %q", s.socialClass)
	}
	if s.race != "" {
		fmt.Fprintf(buf, ", race: %q", s.race)
	}

	if len(s.items) > 0 {
		buf.WriteString(", items: []itemReq{")
		for i, item := range s.items {
			if i > 0 {
				buf.WriteString(", ")
			}
			fmt.Fprintf(buf, "{itemID: %d, count: %d}", item.id, item.count)
		}
		buf.WriteString("}")
	}

	buf.WriteString("},\n")
}
