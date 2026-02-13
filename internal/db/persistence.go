package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

// PlayerPersistenceService атомарно сохраняет/загружает данные игрока.
// MVP: последовательное сохранение (character → items → skills → recipes → hennas), без единой транзакции.
type PlayerPersistenceService struct {
	pool         *pgxpool.Pool
	charRepo     *CharacterRepository
	itemRepo     *ItemRepository
	skillRepo    *SkillRepository
	recipeRepo   *RecipeRepository
	hennaRepo    *HennaRepository
	subclassRepo *SubClassRepository
	friendRepo   *FriendRepository
}

// NewPlayerPersistenceService создаёт новый сервис.
func NewPlayerPersistenceService(
	pool *pgxpool.Pool,
	charRepo *CharacterRepository,
	itemRepo *ItemRepository,
	skillRepo *SkillRepository,
	recipeRepo *RecipeRepository,
	hennaRepo *HennaRepository,
	subclassRepo *SubClassRepository,
	friendRepo *FriendRepository,
) *PlayerPersistenceService {
	return &PlayerPersistenceService{
		pool:         pool,
		charRepo:     charRepo,
		itemRepo:     itemRepo,
		skillRepo:    skillRepo,
		recipeRepo:   recipeRepo,
		hennaRepo:    hennaRepo,
		subclassRepo: subclassRepo,
		friendRepo:   friendRepo,
	}
}

// SavePlayer saves all player data (character, items, skills) in a single transaction.
// Ensures consistency: either all data is saved or none.
func (s *PlayerPersistenceService) SavePlayer(ctx context.Context, player *model.Player) error {
	charID := player.CharacterID()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction for character %d: %w", charID, err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
			slog.Error("rollback failed", "characterID", charID, "error", err)
		}
	}()

	// 1. Save character (location, stats, exp, sp, level)
	if err := s.charRepo.UpdateTx(ctx, tx, player); err != nil {
		return fmt.Errorf("saving character %d: %w", charID, err)
	}

	// 2. Save items
	items := player.Inventory().GetItems()
	itemRows := make([]ItemRow, 0, len(items))
	for _, item := range items {
		itemRows = append(itemRows, ItemRow{
			ItemTypeID: item.ItemID(),
			OwnerID:    charID,
			Count:      item.Count(),
			Enchant:    item.Enchant(),
			Location:   int32(item.Location()),
			SlotID:     item.Slot(),
		})
	}

	if err := s.itemRepo.SaveAllTx(ctx, tx, charID, itemRows); err != nil {
		return fmt.Errorf("saving items for character %d: %w", charID, err)
	}

	// 3. Save skills
	skills := player.Skills()
	if err := s.skillRepo.SaveTx(ctx, tx, charID, skills); err != nil {
		return fmt.Errorf("saving skills for character %d: %w", charID, err)
	}

	// 4. Save recipes (Phase 15)
	recipeRows := playerRecipesToRows(player)
	if s.recipeRepo != nil {
		if err := s.recipeRepo.SaveTx(ctx, tx, charID, recipeRows); err != nil {
			return fmt.Errorf("saving recipes for character %d: %w", charID, err)
		}
	}

	// 5. Save hennas (Phase 13)
	hennaRows := playerHennasToRows(player)
	if s.hennaRepo != nil {
		if err := s.hennaRepo.SaveAllTx(ctx, tx, charID, hennaRows); err != nil {
			return fmt.Errorf("saving hennas for character %d: %w", charID, err)
		}
	}

	// 6. Save subclasses (Phase 14)
	subclassRows := playerSubClassesToRows(player)
	if s.subclassRepo != nil {
		if err := s.subclassRepo.SaveAllTx(ctx, tx, charID, subclassRows); err != nil {
			return fmt.Errorf("saving subclasses for character %d: %w", charID, err)
		}
	}

	// 7. Save friends/blocks (Phase 35)
	friendRows := playerFriendsToRows(player)
	if s.friendRepo != nil {
		if err := s.friendRepo.SaveTx(ctx, tx, charID, friendRows); err != nil {
			return fmt.Errorf("saving friends for character %d: %w", charID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction for character %d: %w", charID, err)
	}

	slog.Info("player data saved",
		"characterID", charID,
		"character", player.Name(),
		"items", len(itemRows),
		"skills", len(skills),
		"recipes", len(recipeRows),
		"hennas", len(hennaRows),
		"subclasses", len(subclassRows),
		"friends", len(friendRows))

	return nil
}

// PlayerData holds all loaded data for a player.
type PlayerData struct {
	Items      []ItemRow
	Skills     []*model.SkillInfo
	Recipes    []RecipeRow
	Hennas     []HennaRow
	SubClasses []SubClassRow
	Friends    []FriendRow
}

// LoadPlayerData загружает items, skills, recipes и hennas для существующего player.
func (s *PlayerPersistenceService) LoadPlayerData(ctx context.Context, charID int64) (*PlayerData, error) {
	itemRows, err := s.itemRepo.LoadByOwner(ctx, charID)
	if err != nil {
		return nil, fmt.Errorf("loading items for character %d: %w", charID, err)
	}

	skills, err := s.skillRepo.LoadByCharacterID(ctx, charID)
	if err != nil {
		return nil, fmt.Errorf("loading skills for character %d: %w", charID, err)
	}

	var recipes []RecipeRow
	if s.recipeRepo != nil {
		recipes, err = s.recipeRepo.LoadByCharacterID(ctx, charID)
		if err != nil {
			return nil, fmt.Errorf("loading recipes for character %d: %w", charID, err)
		}
	}

	var hennas []HennaRow
	if s.hennaRepo != nil {
		hennas, err = s.hennaRepo.LoadByCharacterID(ctx, charID)
		if err != nil {
			return nil, fmt.Errorf("loading hennas for character %d: %w", charID, err)
		}
	}

	var subclasses []SubClassRow
	if s.subclassRepo != nil {
		subclasses, err = s.subclassRepo.LoadByCharacterID(ctx, charID)
		if err != nil {
			return nil, fmt.Errorf("loading subclasses for character %d: %w", charID, err)
		}
	}

	var friends []FriendRow
	if s.friendRepo != nil {
		friends, err = s.friendRepo.LoadByCharacterID(ctx, charID)
		if err != nil {
			return nil, fmt.Errorf("loading friends for character %d: %w", charID, err)
		}
	}

	return &PlayerData{
		Items:      itemRows,
		Skills:     skills,
		Recipes:    recipes,
		Hennas:     hennas,
		SubClasses: subclasses,
		Friends:    friends,
	}, nil
}

// playerHennasToRows converts player's equipped hennas to DB rows.
func playerHennasToRows(player *model.Player) []HennaRow {
	hennaList := player.GetHennaList()
	rows := make([]HennaRow, 0, model.MaxHennaSlots)
	for i, h := range hennaList {
		if h != nil {
			rows = append(rows, HennaRow{
				Slot:  int32(i + 1),
				DyeID: h.DyeID,
			})
		}
	}
	return rows
}

// playerRecipesToRows converts player's recipe book to DB rows.
func playerRecipesToRows(player *model.Player) []RecipeRow {
	dwarven := player.GetRecipeBook(true)
	common := player.GetRecipeBook(false)

	rows := make([]RecipeRow, 0, len(dwarven)+len(common))
	for _, id := range dwarven {
		rows = append(rows, RecipeRow{RecipeID: id, IsDwarven: true})
	}
	for _, id := range common {
		rows = append(rows, RecipeRow{RecipeID: id, IsDwarven: false})
	}
	return rows
}

// playerSubClassesToRows converts player's subclasses to DB rows.
func playerSubClassesToRows(player *model.Player) []SubClassRow {
	player.SaveActiveSubClassState()
	subs := player.SubClasses()
	rows := make([]SubClassRow, 0, len(subs))
	for _, sub := range subs {
		rows = append(rows, SubClassRow{
			ClassID:    sub.ClassID,
			ClassIndex: sub.ClassIndex,
			Exp:        sub.Exp,
			SP:         sub.SP,
			Level:      sub.Level,
		})
	}
	return rows
}

// playerFriendsToRows converts player's friend and block lists to DB rows.
func playerFriendsToRows(player *model.Player) []FriendRow {
	friends := player.FriendList()
	blocks := player.BlockList()

	rows := make([]FriendRow, 0, len(friends)+len(blocks))
	for _, id := range friends {
		rows = append(rows, FriendRow{FriendID: id, Relation: 0})
	}
	for _, id := range blocks {
		rows = append(rows, FriendRow{FriendID: id, Relation: 1})
	}
	return rows
}

// ItemDefToTemplate ищет item definition по ID и конвертирует в *model.ItemTemplate.
// Возвращает nil если item definition не найден.
func ItemDefToTemplate(itemID int32) *model.ItemTemplate {
	def := data.GetItemDef(itemID)
	if def == nil {
		return nil
	}
	itemType := itemTypeFromString(def.Type())
	// Quest items identified by questItem flag in XML data,
	// regardless of item type string.
	// Java reference: EtcItem constructor sets TYPE2_QUEST for quest items.
	if def.IsQuestItem() {
		itemType = model.ItemTypeQuestItem
	}

	bodyPartStr := def.BodyPart()
	bodyPartMask := model.BodyPartMaskFromString(bodyPartStr)
	type1, type2 := itemClientTypes(itemType, bodyPartMask)

	return &model.ItemTemplate{
		ItemID:       def.ID(),
		Name:         def.Name(),
		Type:         itemType,
		Type1:        type1,
		Type2:        type2,
		BodyPartMask: bodyPartMask,
		PAtk:         def.PAtk(),
		AttackRange:  def.AttackRange(),
		CritRate:     def.CritRate(),
		RandomDamage: def.RandomDamage(),
		PDef:         def.PDef(),
		Weight:       def.Weight(),
		Stackable:    def.IsStackable(),
		Tradeable:    def.IsTradeable(),
		CrystalType:  model.CrystalTypeFromString(def.CrystalType()),
		BodyPartStr:  bodyPartStr,
	}
}

// itemClientTypes computes type1/type2 for client packets based on item type and body part.
// Java reference: ItemTemplate.java TYPE1_*/TYPE2_*, Weapon.java, Armor.java, EtcItem.java constructors.
func itemClientTypes(itemType model.ItemType, bodyPartMask int32) (int16, int16) {
	switch itemType {
	case model.ItemTypeWeapon:
		return model.Type1WeaponRingEarringNecklace, model.Type2Weapon
	case model.ItemTypeArmor:
		// Accessories (neck, earring, ring) have Type1=0 (same as weapons), Type2=2
		isAccessory := bodyPartMask == model.BodyPartNeck ||
			bodyPartMask == model.BodyPartREar || bodyPartMask == model.BodyPartLEar ||
			bodyPartMask == model.BodyPartRFinger || bodyPartMask == model.BodyPartLFinger
		if isAccessory {
			return model.Type1WeaponRingEarringNecklace, model.Type2Accessory
		}
		return model.Type1ShieldArmor, model.Type2ShieldArmor
	case model.ItemTypeQuestItem:
		return model.Type1ItemQuestItemAdena, model.Type2Quest
	default:
		return model.Type1ItemQuestItemAdena, model.Type2Other
	}
}

// itemTypeFromString конвертирует строку типа в model.ItemType.
func itemTypeFromString(s string) model.ItemType {
	switch s {
	case "Weapon":
		return model.ItemTypeWeapon
	case "Armor":
		return model.ItemTypeArmor
	default:
		return model.ItemTypeEtcItem
	}
}
