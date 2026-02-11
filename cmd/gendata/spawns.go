package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// --- XML structures (spawns) ---

type xmlSpawnList struct {
	XMLName xml.Name   `xml:"list"`
	Enabled string     `xml:"enabled,attr"`
	Spawns  []xmlSpawn `xml:"spawn"`
}

type xmlSpawn struct {
	Territory *xmlTerritory `xml:"territory"`
	NPCs      []xmlSpawnNPC `xml:"npc"`
}

type xmlTerritory struct {
	MinZ  int32     `xml:"minZ,attr"`
	MaxZ  int32     `xml:"maxZ,attr"`
	Nodes []xmlNode `xml:"node"`
}

type xmlNode struct {
	X int32 `xml:"x,attr"`
	Y int32 `xml:"y,attr"`
}

type xmlSpawnNPC struct {
	ID           int32 `xml:"id,attr"`
	Count        int32 `xml:"count,attr"`
	RespawnDelay int32 `xml:"respawnDelay,attr"`
	RespawnRand  int32 `xml:"respawnRandom,attr"`
	ChaseRange   int32 `xml:"chaseRange,attr"`
	X            int32 `xml:"x,attr"`
	Y            int32 `xml:"y,attr"`
	Z            int32 `xml:"z,attr"`
	Heading      int32 `xml:"heading,attr"`
}

// --- Parsed structures (spawns) ---

type parsedSpawn struct {
	npcID        int32
	count        int32
	respawnDelay int32
	respawnRand  int32
	chaseRange   int32
	x, y, z      int32
	heading      int32
	territory    *parsedTerritory
}

type parsedTerritory struct {
	minZ, maxZ int32
	nodes      []parsedPoint
}

type parsedPoint struct {
	x, y int32
}

func generateSpawns(javaDir, outDir string) error {
	spawnsDir := filepath.Join(javaDir, "spawns")
	spawns, err := parseAllSpawns(spawnsDir)
	if err != nil {
		return fmt.Errorf("parse spawns: %w", err)
	}

	sort.Slice(spawns, func(i, j int) bool { return spawns[i].npcID < spawns[j].npcID })

	outPath := filepath.Join(outDir, "spawn_data_generated.go")
	if err := generateSpawnsGoFile(spawns, outPath); err != nil {
		return fmt.Errorf("generate spawns: %w", err)
	}

	fmt.Printf("  Generated %s: %d spawn entries\n", outPath, len(spawns))
	return nil
}

func parseAllSpawns(dir string) ([]parsedSpawn, error) {
	files, err := walkXMLFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("walk spawns dir: %w", err)
	}

	var all []parsedSpawn
	for _, f := range files {
		spawns, err := parseSpawnFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filepath.Base(f), err)
		}
		all = append(all, spawns...)
	}
	return all, nil
}

func parseSpawnFile(path string) ([]parsedSpawn, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var list xmlSpawnList
	if err := xml.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	// Пропускаем отключённые списки
	if list.Enabled == "false" {
		return nil, nil
	}

	var result []parsedSpawn
	for _, xs := range list.Spawns {
		result = append(result, convertSpawn(xs)...)
	}
	return result, nil
}

func convertSpawn(xs xmlSpawn) []parsedSpawn {
	var territory *parsedTerritory
	if xs.Territory != nil {
		nodes := make([]parsedPoint, 0, len(xs.Territory.Nodes))
		for _, n := range xs.Territory.Nodes {
			nodes = append(nodes, parsedPoint{x: n.X, y: n.Y})
		}
		territory = &parsedTerritory{
			minZ: xs.Territory.MinZ,
			maxZ: xs.Territory.MaxZ,
			nodes: nodes,
		}
	}

	result := make([]parsedSpawn, 0, len(xs.NPCs))
	for _, npc := range xs.NPCs {
		ps := parsedSpawn{
			npcID:        npc.ID,
			count:        npc.Count,
			respawnDelay: npc.RespawnDelay,
			respawnRand:  npc.RespawnRand,
			chaseRange:   npc.ChaseRange,
		}

		// Значение по умолчанию — 1 NPC
		if ps.count == 0 {
			ps.count = 1
		}

		if territory != nil {
			// Территориальный спавн — координаты из territory
			ps.territory = territory
		} else {
			// Фиксированный спавн — координаты из атрибутов <npc>
			ps.x = npc.X
			ps.y = npc.Y
			ps.z = npc.Z
			ps.heading = npc.Heading
		}

		result = append(result, ps)
	}
	return result
}

// --- Code generation (spawns) ---

func generateSpawnsGoFile(spawns []parsedSpawn, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "spawns")
	buf.WriteString("var spawnDefs = []spawnDef{\n")

	for i := range spawns {
		writeSpawnDef(&buf, &spawns[i])
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeSpawnDef(buf *bytes.Buffer, s *parsedSpawn) {
	buf.WriteString("{")
	fmt.Fprintf(buf, "npcID: %d", s.npcID)

	if s.count != 0 {
		fmt.Fprintf(buf, ", count: %d", s.count)
	}
	if s.respawnDelay != 0 {
		fmt.Fprintf(buf, ", respawnDelay: %d", s.respawnDelay)
	}
	if s.respawnRand != 0 {
		fmt.Fprintf(buf, ", respawnRand: %d", s.respawnRand)
	}
	if s.chaseRange != 0 {
		fmt.Fprintf(buf, ", chaseRange: %d", s.chaseRange)
	}

	if s.territory != nil {
		writeSpawnTerritory(buf, s.territory)
	} else {
		if s.x != 0 {
			fmt.Fprintf(buf, ", x: %d", s.x)
		}
		if s.y != 0 {
			fmt.Fprintf(buf, ", y: %d", s.y)
		}
		if s.z != 0 {
			fmt.Fprintf(buf, ", z: %d", s.z)
		}
		if s.heading != 0 {
			fmt.Fprintf(buf, ", heading: %d", s.heading)
		}
	}

	buf.WriteString("},\n")
}

func writeSpawnTerritory(buf *bytes.Buffer, t *parsedTerritory) {
	fmt.Fprintf(buf, ", territory: &territoryDef{minZ: %d, maxZ: %d, nodes: []pointDef{", t.minZ, t.maxZ)
	for i, n := range t.nodes {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(buf, "{%d, %d}", n.x, n.y)
	}
	buf.WriteString("}}")
}
