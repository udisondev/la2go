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

// --- XML structures (doors) ---

type xmlDoorsFile struct {
	XMLName xml.Name  `xml:"list"`
	Doors   []xmlDoor `xml:"door"`
}

type xmlDoor struct {
	ID            int32  `xml:"id,attr"`
	Name          string `xml:"name,attr"`
	OpenMethod    int32  `xml:"open_method,attr"`
	Level         int32  `xml:"level,attr"`
	CloseTime     int32  `xml:"close_time,attr"`
	Height        int32  `xml:"height,attr"`
	BaseHpMax     int32  `xml:"baseHpMax,attr"`
	HpShowable    string `xml:"hp_showable,attr"`
	BasePDef      int32  `xml:"basePDef,attr"`
	BaseMDef      int32  `xml:"baseMDef,attr"`
	Pos           string `xml:"pos,attr"`
	Node1         string `xml:"node1,attr"`
	Node2         string `xml:"node2,attr"`
	Node3         string `xml:"node3,attr"`
	Node4         string `xml:"node4,attr"`
	NodeZ         int32  `xml:"nodeZ,attr"`
	ClanhallID    int32  `xml:"clanhall_id,attr"`
	CastleID      int32  `xml:"castle_id,attr"`
	FortID        int32  `xml:"fort_id,attr"`
	IsWall        string `xml:"is_wall,attr"`
	Group         string `xml:"group,attr"`
	DefaultStatus string `xml:"default_status,attr"`
}

// --- Parsed structure ---

type parsedDoor struct {
	id            int32
	name          string
	openMethod    int32
	height        int32
	hp            int32
	pDef          int32
	mDef          int32
	posX, posY, posZ int32
	nodes         [4][2]int32 // 4 nodes: [x,y]
	nodeZ         int32
	clanhallID    int32
	castleID      int32
	fortID        int32
	showHP        bool
}

func generateDoors(javaDir, outDir string) error {
	doorsPath := filepath.Join(javaDir, "Doors.xml")
	doors, err := parseDoorsFile(doorsPath)
	if err != nil {
		return fmt.Errorf("parse doors: %w", err)
	}

	sort.Slice(doors, func(i, j int) bool {
		return doors[i].id < doors[j].id
	})

	outPath := filepath.Join(outDir, "door_data_generated.go")
	if err := generateDoorsGoFile(doors, outPath); err != nil {
		return fmt.Errorf("generate doors: %w", err)
	}

	fmt.Printf("  Generated %s: %d doors\n", outPath, len(doors))
	return nil
}

func parseDoorsFile(path string) ([]parsedDoor, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var file xmlDoorsFile
	if err := xml.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	doors := make([]parsedDoor, 0, len(file.Doors))
	for _, xd := range file.Doors {
		d := parsedDoor{
			id:         xd.ID,
			name:       xd.Name,
			openMethod: xd.OpenMethod,
			height:     xd.Height,
			hp:         xd.BaseHpMax,
			pDef:       xd.BasePDef,
			mDef:       xd.BaseMDef,
			nodeZ:      xd.NodeZ,
			clanhallID: xd.ClanhallID,
			castleID:   xd.CastleID,
			fortID:     xd.FortID,
			showHP:     xd.HpShowable != "false",
		}

		// Parse pos="x,y,z"
		if xd.Pos != "" {
			px, py, pz := parseCoordTriple(xd.Pos)
			d.posX, d.posY, d.posZ = px, py, pz
		}

		// Parse node1..node4 "x,y"
		d.nodes[0] = parseCoordPair(xd.Node1)
		d.nodes[1] = parseCoordPair(xd.Node2)
		d.nodes[2] = parseCoordPair(xd.Node3)
		d.nodes[3] = parseCoordPair(xd.Node4)

		doors = append(doors, d)
	}
	return doors, nil
}

func parseCoordPair(s string) [2]int32 {
	parts := strings.Split(s, ",")
	if len(parts) < 2 {
		return [2]int32{}
	}
	x, _ := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 32)
	y, _ := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32)
	return [2]int32{int32(x), int32(y)}
}

func parseCoordTriple(s string) (int32, int32, int32) {
	parts := strings.Split(s, ",")
	if len(parts) < 3 {
		return 0, 0, 0
	}
	x, _ := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 32)
	y, _ := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32)
	z, _ := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 32)
	return int32(x), int32(y), int32(z)
}

// --- Code generation ---

func generateDoorsGoFile(doors []parsedDoor, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "doors")
	buf.WriteString("var doorDefs = []doorDef{\n")

	for i := range doors {
		writeDoorDef(&buf, &doors[i])
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeDoorDef(buf *bytes.Buffer, d *parsedDoor) {
	fmt.Fprintf(buf, "{id: %d, name: %q, openMethod: %d, height: %d, hp: %d, pDef: %d, mDef: %d, posX: %d, posY: %d, posZ: %d, nodes: [4]pointDef{{x: %d, y: %d}, {x: %d, y: %d}, {x: %d, y: %d}, {x: %d, y: %d}}, nodeZ: %d, clanhallID: %d, castleID: %d, fortID: %d, showHP: %t},\n",
		d.id, escapeString(d.name), d.openMethod, d.height, d.hp, d.pDef, d.mDef,
		d.posX, d.posY, d.posZ,
		d.nodes[0][0], d.nodes[0][1],
		d.nodes[1][0], d.nodes[1][1],
		d.nodes[2][0], d.nodes[2][1],
		d.nodes[3][0], d.nodes[3][1],
		d.nodeZ, d.clanhallID, d.castleID, d.fortID, d.showHP)
}
