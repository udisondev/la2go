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

// --- XML structures (category data) ---

type xmlCategoryList struct {
	XMLName    xml.Name          `xml:"list"`
	Categories []xmlCategoryItem `xml:"category"`
}

type xmlCategoryItem struct {
	Name string   `xml:"name,attr"`
	IDs  []string `xml:"id"`
}

// --- Parsed structures (category data) ---

type parsedCategory struct {
	name     string
	classIDs []int32
}

func generateCategoryData(javaDir, outDir string) error {
	xmlPath := filepath.Join(javaDir, "CategoryData.xml")
	categories, err := parseCategoryData(xmlPath)
	if err != nil {
		return fmt.Errorf("parse category data: %w", err)
	}

	// Сохраняем порядок из XML (не сортируем — порядок может быть важен)
	outPath := filepath.Join(outDir, "category_data_generated.go")
	if err := generateCategoryGoFile(categories, outPath); err != nil {
		return fmt.Errorf("generate category data: %w", err)
	}

	fmt.Printf("  Generated %s: %d categories\n", outPath, len(categories))
	return nil
}

func parseCategoryData(path string) ([]parsedCategory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var list xmlCategoryList
	if err := xml.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	categories := make([]parsedCategory, 0, len(list.Categories))
	for _, xc := range list.Categories {
		pc, err := convertCategory(xc)
		if err != nil {
			return nil, fmt.Errorf("category %q: %w", xc.Name, err)
		}
		categories = append(categories, pc)
	}
	return categories, nil
}

func convertCategory(xc xmlCategoryItem) (parsedCategory, error) {
	classIDs := make([]int32, 0, len(xc.IDs))
	for _, raw := range xc.IDs {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		id, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return parsedCategory{}, fmt.Errorf("parse id %q: %w", s, err)
		}
		classIDs = append(classIDs, int32(id))
	}

	// Сортируем ID внутри категории для детерминированного вывода
	sort.Slice(classIDs, func(i, j int) bool { return classIDs[i] < classIDs[j] })

	return parsedCategory{
		name:     xc.Name,
		classIDs: classIDs,
	}, nil
}

// --- Code generation (category data) ---

func generateCategoryGoFile(categories []parsedCategory, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "categorydata")
	buf.WriteString("var categoryDefs = []categoryDef{\n")

	for i := range categories {
		writeCategoryDef(&buf, &categories[i])
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeCategoryDef(buf *bytes.Buffer, c *parsedCategory) {
	fmt.Fprintf(buf, "{name: %q, classIDs: []int32{", c.name)
	for i, id := range c.classIDs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(strconv.FormatInt(int64(id), 10))
	}
	buf.WriteString("}},\n")
}
