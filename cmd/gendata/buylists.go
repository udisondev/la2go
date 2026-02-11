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

// --- XML structures (buylists) ---

type xmlBuylistFile struct {
	XMLName xml.Name       `xml:"list"`
	NPCs    *xmlBuylistNPCs `xml:"npcs"`
	Items   []xmlBuylistItem `xml:"item"`
}

type xmlBuylistNPCs struct {
	NPCs []xmlBuylistNPC `xml:"npc"`
}

type xmlBuylistNPC struct {
	ID string `xml:",chardata"`
}

type xmlBuylistItem struct {
	ID           int32 `xml:"id,attr"`
	Price        int64 `xml:"price,attr"`
	Count        int32 `xml:"count,attr"`
	RestockDelay int32 `xml:"restock_delay,attr"`
}

// --- Parsed structures (buylists) ---

type parsedBuylist struct {
	listID int32
	npcID  int32
	items  []parsedBuylistItem
}

type parsedBuylistItem struct {
	itemID       int32
	count        int32
	price        int64
	restockDelay int32
}

func generateBuylists(javaDir, outDir string) error {
	buylistsDir := filepath.Join(javaDir, "buylists")
	buylists, err := parseAllBuylists(buylistsDir)
	if err != nil {
		return fmt.Errorf("parse buylists: %w", err)
	}

	sort.Slice(buylists, func(i, j int) bool {
		if buylists[i].listID != buylists[j].listID {
			return buylists[i].listID < buylists[j].listID
		}
		return buylists[i].npcID < buylists[j].npcID
	})

	outPath := filepath.Join(outDir, "buylist_data_generated.go")
	if err := generateBuylistsGoFile(buylists, outPath); err != nil {
		return fmt.Errorf("generate buylists: %w", err)
	}

	fmt.Printf("  Generated %s: %d buylist entries\n", outPath, len(buylists))
	return nil
}

func parseAllBuylists(dir string) ([]parsedBuylist, error) {
	files, err := globXMLFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("glob buylists dir: %w", err)
	}

	var all []parsedBuylist
	for _, f := range files {
		entries, err := parseBuylistFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filepath.Base(f), err)
		}
		all = append(all, entries...)
	}
	return all, nil
}

func parseBuylistFile(path string) ([]parsedBuylist, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var list xmlBuylistFile
	if err := xml.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	// ID buylist берётся из имени файла (без ведущих нулей и расширения).
	base := strings.TrimSuffix(filepath.Base(path), ".xml")
	listID, err := strconv.ParseInt(base, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse list id from filename %q: %w", base, err)
	}

	// Парсим предметы. count=0 в XML означает unlimited — записываем как -1.
	items := make([]parsedBuylistItem, 0, len(list.Items))
	for _, xi := range list.Items {
		count := xi.Count
		if count == 0 {
			count = -1
		}
		items = append(items, parsedBuylistItem{
			itemID:       xi.ID,
			count:        count,
			price:        xi.Price,
			restockDelay: xi.RestockDelay,
		})
	}

	// Собираем NPC ID. Каждый NPC получает свою запись buylistDef с теми же items.
	var npcIDs []int32
	if list.NPCs != nil {
		for _, n := range list.NPCs.NPCs {
			text := strings.TrimSpace(n.ID)
			id, err := strconv.ParseInt(text, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("parse npc id %q: %w", text, err)
			}
			npcIDs = append(npcIDs, int32(id))
		}
	}

	// Если NPC не указаны — одна запись с npcID=0.
	if len(npcIDs) == 0 {
		npcIDs = []int32{0}
	}

	result := make([]parsedBuylist, 0, len(npcIDs))
	for _, npcID := range npcIDs {
		result = append(result, parsedBuylist{
			listID: int32(listID),
			npcID:  npcID,
			items:  items,
		})
	}
	return result, nil
}

// --- Code generation (buylists) ---

func generateBuylistsGoFile(buylists []parsedBuylist, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "buylists")
	buf.WriteString("var buylistDefs = []buylistDef{\n")

	for i := range buylists {
		writeBuylistDef(&buf, &buylists[i])
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeBuylistDef(buf *bytes.Buffer, bl *parsedBuylist) {
	fmt.Fprintf(buf, "{listID: %d, npcID: %d, items: []buylistItemDef{", bl.listID, bl.npcID)
	for i, item := range bl.items {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(buf, "{itemID: %d, count: %d, price: %d, restockDelay: %d}",
			item.itemID, item.count, item.price, item.restockDelay)
	}
	buf.WriteString("}},\n")
}
