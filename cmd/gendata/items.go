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

// --- XML structures (items) ---

type xmlItemList struct {
	XMLName xml.Name  `xml:"list"`
	Items   []xmlItem `xml:"item"`
}

type xmlItem struct {
	ID         int32            `xml:"id,attr"`
	Type       string           `xml:"type,attr"`
	Name       string           `xml:"name,attr"`
	Sets       []xmlItemSet     `xml:"set"`
	Stats      *xmlItemStats    `xml:"stats"`
	Conditions *xmlItemConditions `xml:"conditions"`
}

type xmlItemSet struct {
	Name string `xml:"name,attr"`
	Val  string `xml:"val,attr"`
}

type xmlItemStats struct {
	Sets []xmlItemStatEntry `xml:"set"`
	Adds []xmlItemStatEntry `xml:"add"`
}

type xmlItemStatEntry struct {
	Stat string `xml:"stat,attr"`
	Val  string `xml:"val,attr"`
}

type xmlItemConditions struct {
	MsgID int32 `xml:"msgId,attr"`
}

// --- Parsed structure (items) ---

type parsedItem struct {
	id       int32
	name     string
	itemType string // "Weapon","Armor","EtcItem"

	// Common
	icon          string
	defaultAction string
	material      string
	weight        int32
	price         int64
	stackable     bool
	tradeable     bool // XML: is_tradable; default true
	dropable      bool // XML: is_dropable; default true
	sellable      bool // XML: is_sellable; default true
	depositable   bool // XML: is_depositable; default true
	questItem     bool
	tradeableSet  bool // true когда атрибут явно указан в XML
	dropableSet   bool
	sellableSet   bool
	depositableSet bool

	// Weapon
	weaponType   string
	bodyPart     string
	randomDamage int32
	attackRange  int32
	soulshots    int32
	spiritshots  int32
	magicWeapon  bool

	// Armor
	armorType string

	// EtcItem
	etcItemType string
	handler     string

	// Item skill
	itemSkillID    int32
	itemSkillLevel int32
	reuseDelay     int32
	olyRestricted  bool
	forNpc         bool

	// Stats
	pAtk    int32
	mAtk    int32
	pDef    int32
	mDef    int32
	pAtkSpd int32
	mAtkSpd int32
	critRate int32

	// Crystal Type / Grade
	crystalType string // "NONE","D","C","B","A","S"

	// Enchant
	enchantable bool

	// Conditions
	condMsgID int32
}

func generateItems(javaDir, outDir string) error {
	itemsDir := filepath.Join(javaDir, "stats", "items")
	items, err := parseAllItems(itemsDir)
	if err != nil {
		return fmt.Errorf("parse items: %w", err)
	}

	sort.Slice(items, func(i, j int) bool { return items[i].id < items[j].id })

	outPath := filepath.Join(outDir, "item_data_generated.go")
	if err := generateItemsGoFile(items, outPath); err != nil {
		return fmt.Errorf("generate items: %w", err)
	}

	fmt.Printf("  Generated %s: %d items\n", outPath, len(items))
	return nil
}

func parseAllItems(dir string) ([]parsedItem, error) {
	files, err := globXMLFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("glob items dir: %w", err)
	}

	var all []parsedItem
	for _, f := range files {
		items, err := parseItemFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", filepath.Base(f), err)
		}
		all = append(all, items...)
	}
	return all, nil
}

func parseItemFile(path string) ([]parsedItem, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var list xmlItemList
	if err := xml.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	result := make([]parsedItem, 0, len(list.Items))
	for _, xi := range list.Items {
		result = append(result, convertItem(xi))
	}
	return result, nil
}

func convertItem(xi xmlItem) parsedItem {
	pi := parsedItem{
		id:       xi.ID,
		name:     xi.Name,
		itemType: xi.Type,
		// Дефолты L2: tradeable/dropable/sellable/depositable = true
		tradeable:   true,
		dropable:    true,
		sellable:    true,
		depositable: true,
	}

	sets := make(map[string]string, len(xi.Sets))
	for _, s := range xi.Sets {
		sets[s.Name] = s.Val
	}

	// Common
	pi.icon = sets["icon"]
	pi.defaultAction = sets["default_action"]
	pi.material = sets["material"]
	pi.weight = parseInt32(sets["weight"])
	pi.price = parseInt64(sets["price"])
	pi.stackable = parseBool(sets["is_stackable"])
	pi.questItem = parseBool(sets["is_questitem"])

	if v, ok := sets["is_tradable"]; ok {
		pi.tradeable = parseBool(v)
		pi.tradeableSet = true
	}
	if v, ok := sets["is_dropable"]; ok {
		pi.dropable = parseBool(v)
		pi.dropableSet = true
	}
	if v, ok := sets["is_sellable"]; ok {
		pi.sellable = parseBool(v)
		pi.sellableSet = true
	}
	if v, ok := sets["is_depositable"]; ok {
		pi.depositable = parseBool(v)
		pi.depositableSet = true
	}

	// Weapon
	pi.weaponType = sets["weapon_type"]
	pi.bodyPart = sets["bodypart"]
	pi.randomDamage = parseInt32(sets["random_damage"])
	pi.attackRange = parseInt32(sets["attack_range"])
	pi.soulshots = parseInt32(sets["soulshots"])
	pi.spiritshots = parseInt32(sets["spiritshots"])
	pi.magicWeapon = parseBool(sets["is_magic_weapon"])

	// Armor
	pi.armorType = sets["armor_type"]

	// EtcItem
	pi.etcItemType = sets["etcitem_type"]
	pi.handler = sets["handler"]

	// Item skill: format "skillID-level" (e.g., "2031-1")
	if skillStr := sets["item_skill"]; skillStr != "" {
		if parts := strings.SplitN(skillStr, "-", 2); len(parts) == 2 {
			pi.itemSkillID = parseInt32(parts[0])
			pi.itemSkillLevel = parseInt32(parts[1])
		}
	}
	pi.reuseDelay = parseInt32(sets["reuse_delay"])
	pi.olyRestricted = parseBool(sets["is_oly_restricted"])
	pi.forNpc = parseBool(sets["for_npc"])

	// Crystal Type / Grade
	pi.crystalType = sets["crystal_type"]

	// Enchant
	pi.enchantable = parseBool(sets["enchant_enabled"])

	// Stats
	if xi.Stats != nil {
		for _, s := range xi.Stats.Sets {
			applyItemStat(&pi, s.Stat, s.Val)
		}
		// Для pDef/mDef также обрабатываем <add> (в Armor файлах pDef часто через <add>)
		for _, s := range xi.Stats.Adds {
			applyItemStat(&pi, s.Stat, s.Val)
		}
	}

	// Conditions
	if xi.Conditions != nil {
		pi.condMsgID = xi.Conditions.MsgID
	}

	return pi
}

func applyItemStat(pi *parsedItem, stat, val string) {
	switch stat {
	case "pAtk":
		pi.pAtk = parseInt32(val)
	case "mAtk":
		pi.mAtk = parseInt32(val)
	case "pDef":
		pi.pDef = parseInt32(val)
	case "mDef":
		pi.mDef = parseInt32(val)
	case "pAtkSpd":
		pi.pAtkSpd = parseInt32(val)
	case "mAtkSpd":
		pi.mAtkSpd = parseInt32(val)
	case "critRate":
		pi.critRate = parseInt32(val)
	}
}

func parseInt32(s string) int32 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 32)
	if err != nil {
		return 0
	}
	return int32(n)
}

func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func parseBool(s string) bool {
	return s == "true" || s == "1"
}

// --- Code generation (items) ---

func generateItemsGoFile(items []parsedItem, outPath string) error {
	var buf bytes.Buffer
	writeHeader(&buf, "items")
	buf.WriteString("var itemDefs = []itemDef{\n")

	for _, it := range items {
		writeItemDef(&buf, it)
	}

	buf.WriteString("}\n")
	return writeGoFile(outPath, buf.Bytes())
}

func writeItemDef(buf *bytes.Buffer, it parsedItem) {
	buf.WriteString("{\n")
	fmt.Fprintf(buf, "id: %d, name: %q, itemType: %q,\n", it.id, it.name, it.itemType)

	// Common
	writeItemStr(buf, "icon", it.icon)
	writeItemStr(buf, "defaultAction", it.defaultAction)
	writeItemStr(buf, "material", it.material)
	writeItemInt32(buf, "weight", it.weight)
	writeItemInt64(buf, "price", it.price)
	writeItemBool(buf, "stackable", it.stackable)

	// tradeable/dropable/sellable: дефолт true в L2, записываем значение
	// когда атрибут явно указан в XML (т.е. false = переопределение дефолта)
	if it.tradeableSet {
		fmt.Fprintf(buf, "tradeable: %t,\n", it.tradeable)
	}
	if it.dropableSet {
		fmt.Fprintf(buf, "droppable: %t,\n", it.dropable)
	}
	if it.sellableSet {
		fmt.Fprintf(buf, "sellable: %t,\n", it.sellable)
	}
	if it.depositableSet {
		fmt.Fprintf(buf, "depositable: %t,\n", it.depositable)
	}

	writeItemBool(buf, "questItem", it.questItem)

	// Weapon
	writeItemStr(buf, "weaponType", it.weaponType)
	writeItemStr(buf, "bodyPart", it.bodyPart)
	writeItemInt32(buf, "randomDamage", it.randomDamage)
	writeItemInt32(buf, "attackRange", it.attackRange)
	writeItemInt32(buf, "soulshots", it.soulshots)
	writeItemInt32(buf, "spiritshots", it.spiritshots)
	writeItemBool(buf, "magicWeapon", it.magicWeapon)

	// Armor
	writeItemStr(buf, "armorType", it.armorType)

	// EtcItem
	writeItemStr(buf, "etcItemType", it.etcItemType)
	writeItemStr(buf, "handler", it.handler)

	// Item skill
	writeItemInt32(buf, "itemSkillID", it.itemSkillID)
	writeItemInt32(buf, "itemSkillLevel", it.itemSkillLevel)
	writeItemInt32(buf, "reuseDelay", it.reuseDelay)
	writeItemBool(buf, "olyRestricted", it.olyRestricted)
	writeItemBool(buf, "forNpc", it.forNpc)

	// Stats
	writeItemInt32(buf, "pAtk", it.pAtk)
	writeItemInt32(buf, "mAtk", it.mAtk)
	writeItemInt32(buf, "pDef", it.pDef)
	writeItemInt32(buf, "mDef", it.mDef)
	writeItemInt32(buf, "pAtkSpd", it.pAtkSpd)
	writeItemInt32(buf, "mAtkSpd", it.mAtkSpd)
	writeItemInt32(buf, "critRate", it.critRate)

	// Crystal Type / Grade
	writeItemStr(buf, "crystalType", it.crystalType)

	// Enchant
	writeItemBool(buf, "enchantable", it.enchantable)

	// Conditions
	writeItemInt32(buf, "condMsgId", it.condMsgID)

	buf.WriteString("},\n")
}

// writeItemStr записывает строковое поле, пропуская пустые значения.
func writeItemStr(buf *bytes.Buffer, field, val string) {
	if val == "" {
		return
	}
	fmt.Fprintf(buf, "%s: %q,\n", field, val)
}

// writeItemInt32 записывает int32 поле, пропуская нулевые значения.
func writeItemInt32(buf *bytes.Buffer, field string, val int32) {
	if val == 0 {
		return
	}
	fmt.Fprintf(buf, "%s: %d,\n", field, val)
}

// writeItemInt64 записывает int64 поле, пропуская нулевые значения.
func writeItemInt64(buf *bytes.Buffer, field string, val int64) {
	if val == 0 {
		return
	}
	fmt.Fprintf(buf, "%s: %d,\n", field, val)
}

// writeItemBool записывает bool поле, пропуская false значения (дефолт).
func writeItemBool(buf *bytes.Buffer, field string, val bool) {
	if !val {
		return
	}
	fmt.Fprintf(buf, "%s: true,\n", field)
}
