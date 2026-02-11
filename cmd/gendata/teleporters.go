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

// --- XML structures (teleporters) ---

type xmlTeleporterFile struct {
	XMLName xml.Name         `xml:"list"`
	NPCs    []xmlTeleporterNPC `xml:"npc"`
}

type xmlTeleporterNPC struct {
	ID        int32                `xml:"id,attr"`
	Teleports []xmlTeleportGroup   `xml:"teleport"`
}

type xmlTeleportGroup struct {
	Type      string               `xml:"type,attr"`
	Locations []xmlTeleportLocation `xml:"location"`
}

type xmlTeleportLocation struct {
	Name     string `xml:"name,attr"`
	X        int32  `xml:"x,attr"`
	Y        int32  `xml:"y,attr"`
	Z        int32  `xml:"z,attr"`
	FeeCount int32  `xml:"feeCount,attr"`
	FeeId    int32  `xml:"feeId,attr"`
	CastleId string `xml:"castleId,attr"` // may contain "4;5" — semicolon-separated
}

// --- Parsed structures ---

type parsedTeleporter struct {
	npcID     int32
	teleports []parsedTeleportGroup
}

type parsedTeleportGroup struct {
	teleType  string
	locations []parsedTeleportLoc
}

type parsedTeleportLoc struct {
	name     string
	x, y, z  int32
	feeCount int32
	feeId    int32
	castleId int32
}

func generateTeleporters(javaDir, outDir string) error {
	teleDir := filepath.Join(javaDir, "teleporters")
	teleporters, err := parseAllTeleporters(teleDir)
	if err != nil {
		return fmt.Errorf("parse teleporters: %w", err)
	}

	sort.Slice(teleporters, func(i, j int) bool {
		return teleporters[i].npcID < teleporters[j].npcID
	})

	outPath := filepath.Join(outDir, "teleporter_data_generated.go")
	if err := generateTeleportersGoFile(teleporters, outPath); err != nil {
		return fmt.Errorf("generate teleporters: %w", err)
	}

	fmt.Printf("  Generated %s: %d teleporter NPCs\n", outPath, len(teleporters))
	return nil
}

func parseAllTeleporters(dir string) ([]parsedTeleporter, error) {
	files, err := walkXMLFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("walk teleporters dir: %w", err)
	}

	var all []parsedTeleporter
	for _, f := range files {
		entries, err := parseTeleporterFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filepath.Base(f), err)
		}
		all = append(all, entries...)
	}
	return all, nil
}

func parseTeleporterFile(path string) ([]parsedTeleporter, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var file xmlTeleporterFile
	if err := xml.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	var result []parsedTeleporter
	for _, npc := range file.NPCs {
		pt := parsedTeleporter{npcID: npc.ID}
		for _, tg := range npc.Teleports {
			pg := parsedTeleportGroup{teleType: tg.Type}
			for _, loc := range tg.Locations {
				pg.locations = append(pg.locations, parsedTeleportLoc{
					name:     loc.Name,
					x:        loc.X,
					y:        loc.Y,
					z:        loc.Z,
					feeCount: loc.FeeCount,
					feeId:    loc.FeeId,
					castleId: parseFirstInt32(loc.CastleId),
				})
			}
			pt.teleports = append(pt.teleports, pg)
		}
		result = append(result, pt)
	}
	return result, nil
}

// parseFirstInt32 parses the first integer from a possibly semicolon-separated string.
// E.g. "4;5" → 4, "0" → 0, "" → 0.
func parseFirstInt32(s string) int32 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if idx := strings.Index(s, ";"); idx >= 0 {
		s = s[:idx]
	}
	v, _ := strconv.ParseInt(s, 10, 32)
	return int32(v)
}

// --- Code generation ---

func generateTeleportersGoFile(teleporters []parsedTeleporter, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "teleporters")
	buf.WriteString("var teleporterDefs = []teleporterDef{\n")

	for i := range teleporters {
		writeTeleporterDef(&buf, &teleporters[i])
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeTeleporterDef(buf *bytes.Buffer, tp *parsedTeleporter) {
	fmt.Fprintf(buf, "{npcID: %d, teleports: []teleportGroupDef{", tp.npcID)
	for i, tg := range tp.teleports {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(buf, "{teleType: %q, locations: []teleportLocDef{", tg.teleType)
		for j, loc := range tg.locations {
			if j > 0 {
				buf.WriteString(", ")
			}
			fmt.Fprintf(buf, "{name: %q, x: %d, y: %d, z: %d, feeCount: %d, feeId: %d, castleId: %d}",
				escapeString(loc.name), loc.x, loc.y, loc.z, loc.feeCount, loc.feeId, loc.castleId)
		}
		buf.WriteString("}}")
	}
	buf.WriteString("}},\n")
}
