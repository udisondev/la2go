package data

import (
	"embed"
	"encoding/xml"
	"fmt"
	"log/slog"
	"path/filepath"
)

//go:embed xml
var playersFS embed.FS

// XML structures (для парсинга L2J Mobius templates)

type xmlPlayerTemplate struct {
	XMLName    xml.Name       `xml:"list"`
	ClassID    uint8          `xml:"classId"`
	StaticData xmlStaticData  `xml:"staticData"`
	LvlUpgain  []xmlLevelData `xml:"lvlUpgainData>level"`
}

type xmlStaticData struct {
	BaseSTR      uint8        `xml:"baseSTR"`
	BaseCON      uint8        `xml:"baseCON"`
	BaseDEX      uint8        `xml:"baseDEX"`
	BaseINT      uint8        `xml:"baseINT"`
	BaseWIT      uint8        `xml:"baseWIT"`
	BaseMEN      uint8        `xml:"baseMEN"`
	BasePAtk     int32        `xml:"basePAtk"`
	BaseMAtk     int32        `xml:"baseMAtk"`
	BasePDef     xmlPDefSlots `xml:"basePDef"`
	BaseMDef     xmlMDefSlots `xml:"baseMDef"`
	BasePAtkSpd  int32        `xml:"basePAtkSpd"`
	BaseMAtkSpd  int32        `xml:"baseMAtkSpd"`
	BaseCritRate int32        `xml:"baseCritRate"`
	BaseAtkRange int32        `xml:"baseAtkRange"`
	BaseRndDam   int32        `xml:"baseRndDam"`

	CollisionMale   xmlCollision  `xml:"collisionMale"`
	CollisionFemale xmlCollision  `xml:"collisionFemale"`
	CreationPoints  []xmlLocation `xml:"creationPoints>node"`
}

type xmlPDefSlots struct {
	Chest     int32 `xml:"chest"`
	Legs      int32 `xml:"legs"`
	Head      int32 `xml:"head"`
	Feet      int32 `xml:"feet"`
	Gloves    int32 `xml:"gloves"`
	Underwear int32 `xml:"underwear"`
	Cloak     int32 `xml:"cloak"`
}

type xmlMDefSlots struct {
	Rear    int32 `xml:"rear"`
	Lear    int32 `xml:"lear"`
	RFinger int32 `xml:"rfinger"`
	LFinger int32 `xml:"lfinger"`
	Neck    int32 `xml:"neck"`
}

type xmlLevelData struct {
	Val   uint8   `xml:"val,attr"`
	HP    float32 `xml:"hp"`
	MP    float32 `xml:"mp"`
	CP    float32 `xml:"cp"`
	HPReg float64 `xml:"hpRegen"`
	MPReg float64 `xml:"mpRegen"`
	CPReg float64 `xml:"cpRegen"`
}

type xmlCollision struct {
	Radius float32 `xml:"radius"`
	Height float32 `xml:"height"`
}

type xmlLocation struct {
	X int32 `xml:"x,attr"`
	Y int32 `xml:"y,attr"`
	Z int32 `xml:"z,attr"`
}

// LoadPlayerTemplates загружает все player templates из embedded XML файлов.
// Вызывается при старте сервера (cmd/gameserver/main.go).
//
// Phase 5.4: Character Templates & Stats System.
// Java reference: PlayerTemplateData.java (line 213)
func LoadPlayerTemplates() error {
	PlayerTemplates = make(map[uint8]*PlayerTemplate)

	// Walk через embedded FS (StartingClass, 1stClass, 2ndClass)
	// Note: 3rdClass excluded (Chronicle 5+, не в Interlude)
	dirs := []string{
		"xml/players/templates/StartingClass",
		"xml/players/templates/1stClass",
		"xml/players/templates/2ndClass",
	}

	for _, dir := range dirs {
		if err := loadTemplatesFromDir(dir); err != nil {
			return fmt.Errorf("loading from %s: %w", dir, err)
		}
	}

	slog.Info("loaded player templates", "count", len(PlayerTemplates))
	return nil
}

// loadTemplatesFromDir загружает все XML templates из указанной директории.
func loadTemplatesFromDir(dir string) error {
	entries, err := playersFS.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading dir %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Parse только .xml файлы
		if filepath.Ext(entry.Name()) != ".xml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := playersFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		var xmlT xmlPlayerTemplate
		if err := xml.Unmarshal(data, &xmlT); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		// Convert XML → PlayerTemplate
		template := convertXMLTemplate(&xmlT)
		PlayerTemplates[template.ClassID] = template

		slog.Debug("loaded template",
			"classID", template.ClassID,
			"className", entry.Name(),
			"baseSTR", template.BaseSTR,
			"basePAtk", template.BasePAtk)
	}

	return nil
}

// convertXMLTemplate конвертирует XML structure → PlayerTemplate.
func convertXMLTemplate(xmlT *xmlPlayerTemplate) *PlayerTemplate {
	t := &PlayerTemplate{
		ClassID: xmlT.ClassID,
		BaseSTR: xmlT.StaticData.BaseSTR,
		BaseCON: xmlT.StaticData.BaseCON,
		BaseDEX: xmlT.StaticData.BaseDEX,
		BaseINT: xmlT.StaticData.BaseINT,
		BaseWIT: xmlT.StaticData.BaseWIT,
		BaseMEN: xmlT.StaticData.BaseMEN,

		BasePAtk:     xmlT.StaticData.BasePAtk,
		BaseMAtk:     xmlT.StaticData.BaseMAtk,
		BasePAtkSpd:  xmlT.StaticData.BasePAtkSpd,
		BaseMAtkSpd:  xmlT.StaticData.BaseMAtkSpd,
		BaseCritRate: xmlT.StaticData.BaseCritRate,
		BaseAtkRange: xmlT.StaticData.BaseAtkRange,
		RandomDamage: xmlT.StaticData.BaseRndDam,

		CollisionRadiusMale:   xmlT.StaticData.CollisionMale.Radius,
		CollisionHeightMale:   xmlT.StaticData.CollisionMale.Height,
		CollisionRadiusFemale: xmlT.StaticData.CollisionFemale.Radius,
		CollisionHeightFemale: xmlT.StaticData.CollisionFemale.Height,

		HPByLevel:      make([]float32, 80),
		MPByLevel:      make([]float32, 80),
		CPByLevel:      make([]float32, 80),
		HPRegenByLevel: make([]float64, 80),
		MPRegenByLevel: make([]float64, 80),
		CPRegenByLevel: make([]float64, 80),
		SlotDef:        make(map[uint8]int32),
		CreationPoints: make([]Location, 0, len(xmlT.StaticData.CreationPoints)),
	}

	// Fill HP/MP/CP arrays from lvlUpgainData (levels 1-80)
	for _, lvl := range xmlT.LvlUpgain {
		idx := lvl.Val - 1 // Level 1 → index 0
		if idx >= 80 {
			continue
		}
		t.HPByLevel[idx] = lvl.HP
		t.MPByLevel[idx] = lvl.MP
		t.CPByLevel[idx] = lvl.CP
		t.HPRegenByLevel[idx] = lvl.HPReg
		t.MPRegenByLevel[idx] = lvl.MPReg
		t.CPRegenByLevel[idx] = lvl.CPReg
	}

	// Fill SlotDef map (Physical Defense по слотам)
	pdef := xmlT.StaticData.BasePDef
	t.SlotDef[SlotChest] = pdef.Chest
	t.SlotDef[SlotLegs] = pdef.Legs
	t.SlotDef[SlotHead] = pdef.Head
	t.SlotDef[SlotFeet] = pdef.Feet
	t.SlotDef[SlotGloves] = pdef.Gloves
	t.SlotDef[SlotUnderwear] = pdef.Underwear
	t.SlotDef[SlotCloak] = pdef.Cloak

	// BasePDef — sum of all PDef slots (nude defense)
	t.BasePDef = pdef.Chest + pdef.Legs + pdef.Head + pdef.Feet + pdef.Gloves + pdef.Underwear + pdef.Cloak

	// Fill MDef slots (jewelry)
	mdef := xmlT.StaticData.BaseMDef
	t.SlotDef[SlotRightEar] = mdef.Rear
	t.SlotDef[SlotLeftEar] = mdef.Lear
	t.SlotDef[SlotRightFinger] = mdef.RFinger
	t.SlotDef[SlotLeftFinger] = mdef.LFinger
	t.SlotDef[SlotNeck] = mdef.Neck

	// BaseMDef — sum of all MDef slots
	t.BaseMDef = mdef.Rear + mdef.Lear + mdef.RFinger + mdef.LFinger + mdef.Neck

	// Creation points
	for _, loc := range xmlT.StaticData.CreationPoints {
		t.CreationPoints = append(t.CreationPoints, Location{
			X: loc.X,
			Y: loc.Y,
			Z: loc.Z,
		})
	}

	return t
}
