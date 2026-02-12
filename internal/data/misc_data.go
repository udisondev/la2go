package data

import "log/slog"

// --- Buylists ---

type buylistDef struct {
	listID int32
	npcID  int32
	items  []buylistItemDef
}

type buylistItemDef struct {
	itemID       int32
	count        int32
	price        int64
	restockDelay int32 // seconds, 0 = unlimited
}

var BuylistTable map[int32]*buylistDef

// NpcBuylistIndex maps npcID → list of buylist IDs for that NPC.
// Phase 8.3: NPC Shops.
var NpcBuylistIndex map[int32][]int32

func LoadBuylists() error {
	BuylistTable = make(map[int32]*buylistDef, len(buylistDefs))
	NpcBuylistIndex = make(map[int32][]int32)
	for i := range buylistDefs {
		def := &buylistDefs[i]
		BuylistTable[def.listID] = def
		if def.npcID != 0 {
			NpcBuylistIndex[def.npcID] = append(NpcBuylistIndex[def.npcID], def.listID)
		}
	}
	slog.Info("loaded buylists", "count", len(BuylistTable), "npc_index", len(NpcBuylistIndex))
	return nil
}

func GetBuylist(listID int32) *buylistDef { return BuylistTable[listID] }

// GetBuylistsByNpc returns all buylist IDs available for given NPC.
// Phase 8.3: NPC Shops.
func GetBuylistsByNpc(npcID int32) []int32 { return NpcBuylistIndex[npcID] }

// BuylistProduct — exported view of a buylist item for use outside the data package.
// Phase 8.3: NPC Shops.
type BuylistProduct struct {
	ItemID       int32
	Count        int32 // -1 = unlimited
	Price        int64
	RestockDelay int32
}

// GetBuylistProducts returns all products in a buylist as exported structs.
// Returns nil if buylist not found.
// Phase 8.3: NPC Shops.
func GetBuylistProducts(listID int32) []BuylistProduct {
	bl := BuylistTable[listID]
	if bl == nil {
		return nil
	}

	products := make([]BuylistProduct, len(bl.items))
	for i, item := range bl.items {
		products[i] = BuylistProduct{
			ItemID:       item.itemID,
			Count:        item.count,
			Price:        item.price,
			RestockDelay: item.restockDelay,
		}
	}
	return products
}

// FindProductInBuylist searches for an item in a buylist by template ID.
// Returns nil if not found.
// Phase 8.3: NPC Shops.
func FindProductInBuylist(listID int32, itemID int32) *BuylistProduct {
	bl := BuylistTable[listID]
	if bl == nil {
		return nil
	}
	for _, item := range bl.items {
		if item.itemID == itemID {
			return &BuylistProduct{
				ItemID:       item.itemID,
				Count:        item.count,
				Price:        item.price,
				RestockDelay: item.restockDelay,
			}
		}
	}
	return nil
}

// --- Teleporters ---

type teleporterDef struct {
	npcID     int32
	teleports []teleportGroupDef
}

type teleportGroupDef struct {
	teleType  string // "NORMAL","NOBLES_TOKEN","NOBLES_ADENA"
	locations []teleportLocDef
}

type teleportLocDef struct {
	name     string
	x, y, z  int32
	feeCount int32
	feeId    int32
	castleId int32
}

var TeleporterTable map[int32]*teleporterDef

func LoadTeleporters() error {
	TeleporterTable = make(map[int32]*teleporterDef, len(teleporterDefs))
	for i := range teleporterDefs {
		TeleporterTable[teleporterDefs[i].npcID] = &teleporterDefs[i]
	}
	slog.Info("loaded teleporters", "count", len(TeleporterTable))
	return nil
}

func GetTeleporter(npcID int32) *teleporterDef { return TeleporterTable[npcID] }

// --- Multisell ---

type multisellDef struct {
	listID int32
	items  []multisellEntryDef
}

type multisellEntryDef struct {
	ingredients []multisellIngDef
	productions []multisellIngDef
}

type multisellIngDef struct {
	itemID int32
	count  int64
}

var MultisellTable map[int32]*multisellDef

func LoadMultisell() error {
	MultisellTable = make(map[int32]*multisellDef, len(multisellDefs))
	for i := range multisellDefs {
		MultisellTable[multisellDefs[i].listID] = &multisellDefs[i]
	}
	slog.Info("loaded multisell lists", "count", len(MultisellTable))
	return nil
}

func GetMultisell(listID int32) *multisellDef { return MultisellTable[listID] }

// --- Doors ---

type doorDef struct {
	id          int32
	name        string
	openMethod  int32
	height      int32
	hp          int32
	pDef        int32
	mDef        int32
	posX, posY, posZ int32
	nodes       [4]pointDef
	nodeZ       int32
	clanhallID  int32
	castleID    int32
	fortID      int32
	showHP      bool
}

var DoorTable map[int32]*doorDef

func LoadDoors() error {
	DoorTable = make(map[int32]*doorDef, len(doorDefs))
	for i := range doorDefs {
		DoorTable[doorDefs[i].id] = &doorDefs[i]
	}
	slog.Info("loaded doors", "count", len(DoorTable))
	return nil
}

func GetDoor(id int32) *doorDef { return DoorTable[id] }

// --- Armorsets ---

type armorsetDef struct {
	setID                                   int32
	chest, legs, head, gloves, feet, shield int32
	skillID, skillLevel                     int32
	shieldSkillID, shieldSkillLevel         int32
	enchant6SkillID, enchant6SkillLevel     int32
	strMod, conMod, dexMod, intMod          int32
}

var ArmorsetTable map[int32]*armorsetDef

func LoadArmorsets() error {
	ArmorsetTable = make(map[int32]*armorsetDef, len(armorsetDefs))
	for i := range armorsetDefs {
		ArmorsetTable[armorsetDefs[i].setID] = &armorsetDefs[i]
	}
	slog.Info("loaded armorsets", "count", len(ArmorsetTable))
	return nil
}

func GetArmorset(setID int32) *armorsetDef { return ArmorsetTable[setID] }

// --- Recipes ---

type recipeDef struct {
	id          int32
	recipeID    int32
	name        string
	craftLevel  int32
	recipeType  string // "dwarven","common"
	successRate int32
	mpCost      int32
	ingredients []recipeIngDef
	productions []recipeIngDef
}

type recipeIngDef struct {
	itemID int32
	count  int32
}

var RecipeTable map[int32]*recipeDef

func LoadRecipes() error {
	RecipeTable = make(map[int32]*recipeDef, len(recipeDefs))
	for i := range recipeDefs {
		RecipeTable[recipeDefs[i].id] = &recipeDefs[i]
	}
	slog.Info("loaded recipes", "count", len(RecipeTable))
	return nil
}

func GetRecipe(id int32) *recipeDef { return RecipeTable[id] }

// --- Augmentation ---

type augmentationDef struct {
	id         int32
	skillID    int32
	skillLevel int32
	augType    string // "blue","red","yellow","purple"
}

var AugmentationTable map[int32]*augmentationDef

func LoadAugmentations() error {
	AugmentationTable = make(map[int32]*augmentationDef, len(augmentationDefs))
	for i := range augmentationDefs {
		AugmentationTable[augmentationDefs[i].id] = &augmentationDefs[i]
	}
	slog.Info("loaded augmentations", "count", len(AugmentationTable))
	return nil
}

// --- PlayerConfig ---

type initialEquipDef struct {
	classID int32
	items   []initialItemDef
}

type initialItemDef struct {
	itemID   int32
	count    int32
	equipped bool
}

type initialShortcutsDef struct {
	classID   int32
	shortcuts []shortcutDef
}

type shortcutDef struct {
	page, slot    int32
	shortcutType  string // "ACTION","SKILL","ITEM"
	shortcutID    int32
	shortcutLevel int32
}

var InitialEquipTable map[int32]*initialEquipDef
var InitialShortcutsTable map[int32]*initialShortcutsDef

func LoadPlayerConfig() error {
	InitialEquipTable = make(map[int32]*initialEquipDef, len(initialEquipDefs))
	for i := range initialEquipDefs {
		InitialEquipTable[initialEquipDefs[i].classID] = &initialEquipDefs[i]
	}
	InitialShortcutsTable = make(map[int32]*initialShortcutsDef, len(initialShortcutsDefs))
	for i := range initialShortcutsDefs {
		InitialShortcutsTable[initialShortcutsDefs[i].classID] = &initialShortcutsDefs[i]
	}
	slog.Info("loaded player config", "equip_classes", len(InitialEquipTable), "shortcut_classes", len(InitialShortcutsTable))
	return nil
}

// --- CategoryData ---

type categoryDef struct {
	name     string
	classIDs []int32
}

var CategoryTable map[string]*categoryDef

func LoadCategoryData() error {
	CategoryTable = make(map[string]*categoryDef, len(categoryDefs))
	for i := range categoryDefs {
		CategoryTable[categoryDefs[i].name] = &categoryDefs[i]
	}
	slog.Info("loaded category data", "count", len(CategoryTable))
	return nil
}

// --- Pets ---

type petDef struct {
	npcID  int32
	itemID int32 // control item
	levels []petLevelDef
	skills []npcSkillDef
}

type petLevelDef struct {
	level       int32
	exp         int64
	hp, mp      float64
	pAtk, pDef  float64
	mAtk, mDef  float64
	maxFeed     int32
	feedRate    float64
}

var PetTable map[int32]*petDef

func LoadPetData() error {
	PetTable = make(map[int32]*petDef, len(petDefs))
	for i := range petDefs {
		PetTable[petDefs[i].npcID] = &petDefs[i]
	}
	slog.Info("loaded pet data", "count", len(PetTable))
	return nil
}

// --- Fishing ---

type fishDef struct {
	id       int32
	itemID   int32
	fishType string
	group    int32
	level    int32
	hp       int32
}

var FishTable map[int32]*fishDef

func LoadFishingData() error {
	FishTable = make(map[int32]*fishDef, len(fishDefs))
	for i := range fishDefs {
		FishTable[fishDefs[i].id] = &fishDefs[i]
	}
	slog.Info("loaded fishing data", "count", len(FishTable))
	return nil
}

// --- Seeds ---

type seedDef struct {
	castleID int32
	cropID   int32
	seedID   int32
	matureID int32
	reward1  int32
	reward2  int32
	level    int32
}

var SeedTable map[int32]*seedDef

func LoadSeeds() error {
	SeedTable = make(map[int32]*seedDef, len(seedDefs))
	for i := range seedDefs {
		SeedTable[seedDefs[i].seedID] = &seedDefs[i]
	}
	slog.Info("loaded seeds", "count", len(SeedTable))
	return nil
}
