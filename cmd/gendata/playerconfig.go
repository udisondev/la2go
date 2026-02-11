package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// --- XML structures (initialEquipment) ---

type xmlEquipmentList struct {
	XMLName   xml.Name         `xml:"list"`
	Equipment []xmlEquipment   `xml:"equipment"`
}

type xmlEquipment struct {
	ClassID int32          `xml:"classId,attr"`
	Items   []xmlEquipItem `xml:"item"`
}

type xmlEquipItem struct {
	ID       int32  `xml:"id,attr"`
	Count    int32  `xml:"count,attr"`
	Equipped string `xml:"equipped,attr"`
}

// --- XML structures (initialShortcuts) ---

type xmlShortcutList struct {
	XMLName   xml.Name        `xml:"list"`
	Shortcuts []xmlShortcuts  `xml:"shortcuts"`
}

type xmlShortcuts struct {
	ClassID int32          `xml:"classId,attr"` // 0 если не указан (глобальные)
	Pages   []xmlShortPage `xml:"page"`
	// hasClassID заполняется вручную после парсинга
}

type xmlShortPage struct {
	PageID int32         `xml:"pageId,attr"`
	Slots  []xmlShortSlot `xml:"slot"`
}

type xmlShortSlot struct {
	SlotID        int32  `xml:"slotId,attr"`
	ShortcutType  string `xml:"shortcutType,attr"`
	ShortcutID    int32  `xml:"shortcutId,attr"`
	ShortcutLevel int32  `xml:"shortcutLevel,attr"`
}

// --- Parsed structures ---

type parsedEquipment struct {
	classID int32
	items   []parsedEquipItem
}

type parsedEquipItem struct {
	itemID   int32
	count    int32
	equipped bool
}

type parsedShortcuts struct {
	classID   int32 // -1 для глобальных (без classId)
	shortcuts []parsedShortcut
}

type parsedShortcut struct {
	page          int32
	slot          int32
	shortcutType  string
	shortcutID    int32
	shortcutLevel int32
}

func generatePlayerConfig(javaDir, outDir string) error {
	playersDir := filepath.Join(javaDir, "stats", "players")

	equips, err := parseInitialEquipment(filepath.Join(playersDir, "initialEquipment.xml"))
	if err != nil {
		return fmt.Errorf("parse initial equipment: %w", err)
	}

	shortcuts, err := parseInitialShortcuts(filepath.Join(playersDir, "initialShortcuts.xml"))
	if err != nil {
		return fmt.Errorf("parse initial shortcuts: %w", err)
	}

	sort.Slice(equips, func(i, j int) bool { return equips[i].classID < equips[j].classID })
	sort.Slice(shortcuts, func(i, j int) bool { return shortcuts[i].classID < shortcuts[j].classID })

	outPath := filepath.Join(outDir, "player_config_generated.go")
	if err := generatePlayerConfigGoFile(equips, shortcuts, outPath); err != nil {
		return fmt.Errorf("generate player config: %w", err)
	}

	fmt.Printf("  Generated %s: %d equip classes, %d shortcut classes\n", outPath, len(equips), len(shortcuts))
	return nil
}

// --- Parsing ---

func parseInitialEquipment(path string) ([]parsedEquipment, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var list xmlEquipmentList
	if err := xml.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	result := make([]parsedEquipment, 0, len(list.Equipment))
	for _, xe := range list.Equipment {
		items := make([]parsedEquipItem, 0, len(xe.Items))
		for _, xi := range xe.Items {
			items = append(items, parsedEquipItem{
				itemID:   xi.ID,
				count:    xi.Count,
				equipped: parseBool(xi.Equipped),
			})
		}
		result = append(result, parsedEquipment{
			classID: xe.ClassID,
			items:   items,
		})
	}
	return result, nil
}

func parseInitialShortcuts(path string) ([]parsedShortcuts, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	// encoding/xml не различает "classId не указан" и "classId=0".
	// В XML глобальный блок <shortcuts> не имеет classId, а classId="0" (Human Fighter)
	// вообще не встречается в shortcuts (только в equipment).
	// Поэтому используем ручной подход: декодируем через xml.Decoder
	// и проверяем наличие атрибута classId.
	var list xmlShortcutList
	if err := xml.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	// Определяем, какие блоки не имеют classId (глобальные).
	// Для этого ещё раз проходим через decoder и собираем set'ы classId.
	classIDPresent := detectShortcutClassIDs(raw)

	result := make([]parsedShortcuts, 0, len(list.Shortcuts))
	for idx, xs := range list.Shortcuts {
		classID := xs.ClassID
		if !classIDPresent[idx] {
			classID = -1 // глобальные shortcuts
		}

		var shortcuts []parsedShortcut
		for _, page := range xs.Pages {
			for _, slot := range page.Slots {
				shortcuts = append(shortcuts, parsedShortcut{
					page:          page.PageID,
					slot:          slot.SlotID,
					shortcutType:  slot.ShortcutType,
					shortcutID:    slot.ShortcutID,
					shortcutLevel: slot.ShortcutLevel,
				})
			}
		}
		result = append(result, parsedShortcuts{
			classID:   classID,
			shortcuts: shortcuts,
		})
	}
	return result, nil
}

// detectShortcutClassIDs проходит по XML и для каждого <shortcuts> элемента
// определяет, присутствует ли атрибут classId. Возвращает map[index]bool.
func detectShortcutClassIDs(raw []byte) map[int]bool {
	result := make(map[int]bool)
	decoder := xml.NewDecoder(bytes.NewReader(raw))
	idx := 0
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "shortcuts" {
			continue
		}
		for _, attr := range se.Attr {
			if attr.Name.Local == "classId" {
				result[idx] = true
				break
			}
		}
		idx++
	}
	return result
}

// --- Code generation ---

func generatePlayerConfigGoFile(equips []parsedEquipment, shortcuts []parsedShortcuts, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "playerconfig")

	// initialEquipDefs
	buf.WriteString("var initialEquipDefs = []initialEquipDef{\n")
	for i := range equips {
		writeEquipDef(&buf, &equips[i])
	}
	buf.WriteString("}\n\n")

	// initialShortcutsDefs
	buf.WriteString("var initialShortcutsDefs = []initialShortcutsDef{\n")
	for i := range shortcuts {
		writeShortcutsDef(&buf, &shortcuts[i])
	}
	buf.WriteString("}\n")

	return writeGoFile(outPath, buf.Bytes())
}

func writeEquipDef(buf *bytes.Buffer, eq *parsedEquipment) {
	fmt.Fprintf(buf, "{classID: %d, items: []initialItemDef{", eq.classID)
	for i, it := range eq.items {
		if i > 0 {
			buf.WriteString(", ")
		}
		if it.equipped {
			fmt.Fprintf(buf, "{itemID: %d, count: %d, equipped: true}", it.itemID, it.count)
		} else {
			fmt.Fprintf(buf, "{itemID: %d, count: %d}", it.itemID, it.count)
		}
	}
	buf.WriteString("}},\n")
}

func writeShortcutsDef(buf *bytes.Buffer, sc *parsedShortcuts) {
	fmt.Fprintf(buf, "{classID: %d, shortcuts: []shortcutDef{\n", sc.classID)
	for _, s := range sc.shortcuts {
		if s.shortcutLevel != 0 {
			fmt.Fprintf(buf, "{page: %d, slot: %d, shortcutType: %q, shortcutID: %d, shortcutLevel: %d},\n",
				s.page, s.slot, s.shortcutType, s.shortcutID, s.shortcutLevel)
		} else {
			fmt.Fprintf(buf, "{page: %d, slot: %d, shortcutType: %q, shortcutID: %d},\n",
				s.page, s.slot, s.shortcutType, s.shortcutID)
		}
	}
	buf.WriteString("}},\n")
}
