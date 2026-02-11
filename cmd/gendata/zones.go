package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// --- XML structures (zones) ---

type xmlZoneList struct {
	XMLName xml.Name  `xml:"list"`
	Enabled string    `xml:"enabled,attr"`
	Zones   []xmlZone `xml:"zone"`
}

type xmlZone struct {
	Name  string         `xml:"name,attr"`
	ID    int32          `xml:"id,attr"`
	Type  string         `xml:"type,attr"`
	Shape string         `xml:"shape,attr"`
	MinZ  int32          `xml:"minZ,attr"`
	MaxZ  int32          `xml:"maxZ,attr"`
	Nodes []xmlZoneNode  `xml:"node"`
	Stats []xmlZoneStat  `xml:"stat"`
	Spawns []xmlZoneSpawn `xml:"spawn"`
}

type xmlZoneNode struct {
	X int32 `xml:"X,attr"`
	Y int32 `xml:"Y,attr"`
}

type xmlZoneStat struct {
	Name string `xml:"name,attr"`
	Val  string `xml:"val,attr"`
}

type xmlZoneSpawn struct {
	X    int32  `xml:"X,attr"`
	Y    int32  `xml:"Y,attr"`
	Z    int32  `xml:"Z,attr"`
	Type string `xml:"type,attr"`
}

// --- Parsed structures (zones) ---

type parsedZone struct {
	name     string
	id       int32
	zoneType string
	shape    string
	minZ     int32
	maxZ     int32
	nodes    []parsedPoint
	params   map[string]string
	spawns   []parsedZoneSpawn
}

type parsedZoneSpawn struct {
	x, y, z   int32
	spawnType string
}

func generateZones(javaDir, outDir string) error {
	zonesDir := filepath.Join(javaDir, "zones")
	zones, err := parseAllZones(zonesDir)
	if err != nil {
		return fmt.Errorf("parse zones: %w", err)
	}

	sort.Slice(zones, func(i, j int) bool {
		if zones[i].id != zones[j].id {
			return zones[i].id < zones[j].id
		}
		return zones[i].name < zones[j].name
	})

	outPath := filepath.Join(outDir, "zone_data_generated.go")
	if err := generateZonesGoFile(zones, outPath); err != nil {
		return fmt.Errorf("generate zones: %w", err)
	}

	fmt.Printf("  Generated %s: %d zone entries\n", outPath, len(zones))
	return nil
}

func parseAllZones(dir string) ([]parsedZone, error) {
	files, err := globXMLFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("glob zones dir: %w", err)
	}

	var all []parsedZone
	for _, f := range files {
		zones, err := parseZoneFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filepath.Base(f), err)
		}
		all = append(all, zones...)
	}
	return all, nil
}

func parseZoneFile(path string) ([]parsedZone, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var list xmlZoneList
	if err := xml.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	// Пропускаем отключённые списки
	if list.Enabled == "false" {
		return nil, nil
	}

	result := make([]parsedZone, 0, len(list.Zones))
	for _, xz := range list.Zones {
		result = append(result, convertZone(xz))
	}
	return result, nil
}

func convertZone(xz xmlZone) parsedZone {
	nodes := make([]parsedPoint, 0, len(xz.Nodes))
	for _, n := range xz.Nodes {
		nodes = append(nodes, parsedPoint{x: n.X, y: n.Y})
	}

	var params map[string]string
	if len(xz.Stats) > 0 {
		params = make(map[string]string, len(xz.Stats))
		for _, s := range xz.Stats {
			params[s.Name] = s.Val
		}
	}

	spawns := make([]parsedZoneSpawn, 0, len(xz.Spawns))
	for _, sp := range xz.Spawns {
		spawns = append(spawns, parsedZoneSpawn{
			x:         sp.X,
			y:         sp.Y,
			z:         sp.Z,
			spawnType: sp.Type,
		})
	}

	return parsedZone{
		name:     xz.Name,
		id:       xz.ID,
		zoneType: xz.Type,
		shape:    xz.Shape,
		minZ:     xz.MinZ,
		maxZ:     xz.MaxZ,
		nodes:    nodes,
		params:   params,
		spawns:   spawns,
	}
}

// --- Code generation (zones) ---

func generateZonesGoFile(zones []parsedZone, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "zones")
	buf.WriteString("var zoneDefs = []zoneDef{\n")

	for i := range zones {
		writeZoneDef(&buf, &zones[i])
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeZoneDef(buf *bytes.Buffer, z *parsedZone) {
	buf.WriteString("{")
	fmt.Fprintf(buf, "name: %q, zoneType: %q, shape: %q, minZ: %d, maxZ: %d",
		z.name, z.zoneType, z.shape, z.minZ, z.maxZ)

	if z.id != 0 {
		fmt.Fprintf(buf, ", id: %d", z.id)
	}

	// Nodes
	if len(z.nodes) > 0 {
		buf.WriteString(", nodes: []pointDef{")
		for i, n := range z.nodes {
			if i > 0 {
				buf.WriteString(", ")
			}
			fmt.Fprintf(buf, "{%d, %d}", n.x, n.y)
		}
		buf.WriteString("}")
	}

	// Params
	if len(z.params) > 0 {
		buf.WriteString(", params: map[string]string{")
		keys := sortedKeys(z.params)
		for i, k := range keys {
			if i > 0 {
				buf.WriteString(", ")
			}
			fmt.Fprintf(buf, "%q: %q", k, z.params[k])
		}
		buf.WriteString("}")
	}

	// Spawns
	if len(z.spawns) > 0 {
		buf.WriteString(", spawns: []zoneSpawnDef{")
		for i, sp := range z.spawns {
			if i > 0 {
				buf.WriteString(", ")
			}
			fmt.Fprintf(buf, "{x: %d, y: %d, z: %d", sp.x, sp.y, sp.z)
			if sp.spawnType != "" {
				fmt.Fprintf(buf, ", spawnType: %q", sp.spawnType)
			}
			buf.WriteString("}")
		}
		buf.WriteString("}")
	}

	buf.WriteString("},\n")
}
