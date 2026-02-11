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

// --- XML structures (multisell) ---

type xmlMultisellList struct {
	XMLName xml.Name           `xml:"list"`
	Items   []xmlMultisellItem `xml:"item"`
}

type xmlMultisellItem struct {
	Ingredients []xmlMultisellIng `xml:"ingredient"`
	Productions []xmlMultisellIng `xml:"production"`
}

type xmlMultisellIng struct {
	ID    int32 `xml:"id,attr"`
	Count int64 `xml:"count,attr"`
}

// --- Parsed structures (multisell) ---

type parsedMultisell struct {
	listID int32
	items  []parsedMultisellEntry
}

type parsedMultisellEntry struct {
	ingredients []parsedMultisellIng
	productions []parsedMultisellIng
}

type parsedMultisellIng struct {
	itemID int32
	count  int64
}

func generateMultisell(javaDir, outDir string) error {
	msDir := filepath.Join(javaDir, "multisell")
	lists, err := parseAllMultisell(msDir)
	if err != nil {
		return fmt.Errorf("parse multisell: %w", err)
	}

	sort.Slice(lists, func(i, j int) bool { return lists[i].listID < lists[j].listID })

	outPath := filepath.Join(outDir, "multisell_data_generated.go")
	if err := generateMultisellGoFile(lists, outPath); err != nil {
		return fmt.Errorf("generate multisell: %w", err)
	}

	totalItems := 0
	for i := range lists {
		totalItems += len(lists[i].items)
	}
	fmt.Printf("  Generated %s: %d lists, %d entries\n", outPath, len(lists), totalItems)
	return nil
}

func parseAllMultisell(dir string) ([]parsedMultisell, error) {
	files, err := globXMLFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("glob multisell dir: %w", err)
	}

	lists := make([]parsedMultisell, 0, len(files))
	for _, f := range files {
		ms, err := parseMultisellFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filepath.Base(f), err)
		}
		lists = append(lists, ms)
	}
	return lists, nil
}

func parseMultisellFile(path string) (parsedMultisell, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return parsedMultisell{}, fmt.Errorf("read file: %w", err)
	}

	// listID из имени файла: "001.xml" -> 1, "350980009.xml" -> 350980009
	base := strings.TrimSuffix(filepath.Base(path), ".xml")
	listID, err := strconv.ParseInt(base, 10, 32)
	if err != nil {
		return parsedMultisell{}, fmt.Errorf("parse listID from filename %q: %w", base, err)
	}

	var list xmlMultisellList
	if err := xml.Unmarshal(data, &list); err != nil {
		return parsedMultisell{}, fmt.Errorf("unmarshal: %w", err)
	}

	entries := make([]parsedMultisellEntry, 0, len(list.Items))
	for _, xi := range list.Items {
		entries = append(entries, convertMultisellEntry(xi))
	}

	return parsedMultisell{
		listID: int32(listID),
		items:  entries,
	}, nil
}

func convertMultisellEntry(xi xmlMultisellItem) parsedMultisellEntry {
	ingredients := make([]parsedMultisellIng, 0, len(xi.Ingredients))
	for _, ing := range xi.Ingredients {
		ingredients = append(ingredients, parsedMultisellIng{
			itemID: ing.ID,
			count:  ing.Count,
		})
	}

	productions := make([]parsedMultisellIng, 0, len(xi.Productions))
	for _, prod := range xi.Productions {
		productions = append(productions, parsedMultisellIng{
			itemID: prod.ID,
			count:  prod.Count,
		})
	}

	return parsedMultisellEntry{
		ingredients: ingredients,
		productions: productions,
	}
}

// --- Code generation (multisell) ---

func generateMultisellGoFile(lists []parsedMultisell, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "multisell")
	buf.WriteString("var multisellDefs = []multisellDef{\n")

	for i := range lists {
		writeMultisellDef(&buf, &lists[i])
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeMultisellDef(buf *bytes.Buffer, ms *parsedMultisell) {
	fmt.Fprintf(buf, "{listID: %d, items: []multisellEntryDef{\n", ms.listID)
	for i := range ms.items {
		writeMultisellEntry(buf, &ms.items[i])
	}
	buf.WriteString("}},\n")
}

func writeMultisellEntry(buf *bytes.Buffer, entry *parsedMultisellEntry) {
	buf.WriteString("{ingredients: []multisellIngDef{")
	for i, ing := range entry.ingredients {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(buf, "{itemID: %d, count: %d}", ing.itemID, ing.count)
	}
	buf.WriteString("}, productions: []multisellIngDef{")
	for i, prod := range entry.productions {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(buf, "{itemID: %d, count: %d}", prod.itemID, prod.count)
	}
	buf.WriteString("}},\n")
}
