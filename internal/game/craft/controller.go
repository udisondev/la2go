// Package craft implements the recipe crafting system.
// Phase 15: Recipe/Craft System.
//
// Java reference: RecipeManager.java, RecipeItemMaker (inner class).
package craft

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

// ItemCreator creates new items in the world (injected dependency).
type ItemCreator interface {
	CreateItem(itemID int32, count int32, ownerID int64) (*model.Item, error)
}

// Controller manages recipe crafting operations.
type Controller struct {
	mu           sync.Mutex
	activeCrafts map[uint32]struct{} // objectID → in progress
	itemCreator  ItemCreator
}

// NewController creates a new craft controller.
func NewController(creator ItemCreator) *Controller {
	return &Controller{
		activeCrafts: make(map[uint32]struct{}),
		itemCreator:  creator,
	}
}

// CraftResult represents the outcome of a craft attempt.
type CraftResult struct {
	Success bool
	Item    *model.Item // nil on failure
	Recipe  *data.RecipeTemplate
}

// Craft attempts to craft an item from a recipe.
//
// Business rules (Java reference: RecipeItemMaker.finishCrafting):
//  1. Player must have learned the recipe
//  2. Player must have all required materials
//  3. Player must have enough MP
//  4. Materials are consumed BEFORE success check (failure = lost materials)
//  5. Success determined by recipe.SuccessRate (0-100)
//  6. Player must not be in combat, dead, trading, or already crafting
func (c *Controller) Craft(player *model.Player, recipeListID int32) (*CraftResult, error) {
	if player == nil {
		return nil, fmt.Errorf("player is nil")
	}

	objID := player.ObjectID()

	// Check not already crafting (atomic check+insert)
	c.mu.Lock()
	if _, busy := c.activeCrafts[objID]; busy {
		c.mu.Unlock()
		return nil, fmt.Errorf("already crafting")
	}
	c.activeCrafts[objID] = struct{}{}
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.activeCrafts, objID)
		c.mu.Unlock()
	}()

	// Validate state
	if err := c.validateState(player); err != nil {
		return nil, err
	}

	// Get recipe
	recipe := data.GetRecipeTemplate(recipeListID)
	if recipe == nil {
		return nil, fmt.Errorf("recipe %d not found", recipeListID)
	}

	// Check recipe is learned
	if !player.HasRecipe(recipeListID) {
		return nil, fmt.Errorf("recipe %d not learned", recipeListID)
	}

	// Check materials
	if err := c.checkMaterials(player, recipe); err != nil {
		return nil, err
	}

	// Check MP
	if recipe.MPCost > 0 && player.CurrentMP() < recipe.MPCost {
		return nil, fmt.Errorf("not enough MP: have %d, need %d", player.CurrentMP(), recipe.MPCost)
	}

	// Consume MP
	if recipe.MPCost > 0 {
		player.SetCurrentMP(player.CurrentMP() - recipe.MPCost)
	}

	// Consume materials (BEFORE success check — Java behavior)
	c.consumeMaterials(player, recipe)

	// Success check
	result := &CraftResult{Recipe: recipe}
	if recipe.SuccessRate >= 100 || rand.IntN(100) < int(recipe.SuccessRate) {
		// Create product
		if len(recipe.Productions) > 0 {
			prod := recipe.Productions[0]
			item, err := c.itemCreator.CreateItem(prod.ItemID, prod.Count, player.CharacterID())
			if err != nil {
				slog.Error("craft product creation failed",
					"recipeID", recipeListID,
					"productID", prod.ItemID,
					"error", err)
				return nil, fmt.Errorf("create craft product %d: %w", prod.ItemID, err)
			}

			if err := player.Inventory().AddItem(item); err != nil {
				slog.Error("adding craft product to inventory",
					"recipeID", recipeListID,
					"productID", prod.ItemID,
					"error", err)
				return nil, fmt.Errorf("add craft product to inventory: %w", err)
			}

			result.Success = true
			result.Item = item

			slog.Info("craft success",
				"player", player.Name(),
				"recipe", recipe.Name,
				"product", prod.ItemID,
				"count", prod.Count)
		}
	} else {
		slog.Info("craft failed",
			"player", player.Name(),
			"recipe", recipe.Name,
			"successRate", recipe.SuccessRate)
	}

	return result, nil
}

// IsCrafting returns true if the player is currently crafting.
func (c *Controller) IsCrafting(objectID uint32) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, busy := c.activeCrafts[objectID]
	return busy
}

func (c *Controller) validateState(player *model.Player) error {
	if player.IsDead() {
		return fmt.Errorf("cannot craft while dead")
	}
	if player.HasAttackStance() {
		return fmt.Errorf("cannot craft in combat")
	}
	if player.IsInStoreMode() {
		return fmt.Errorf("cannot craft while in store mode")
	}
	return nil
}

func (c *Controller) checkMaterials(player *model.Player, recipe *data.RecipeTemplate) error {
	inv := player.Inventory()
	for _, ing := range recipe.Ingredients {
		have := inv.CountItemsByID(ing.ItemID)
		if have < int64(ing.Count) {
			return fmt.Errorf("missing material itemID=%d: have %d, need %d",
				ing.ItemID, have, ing.Count)
		}
	}
	return nil
}

func (c *Controller) consumeMaterials(player *model.Player, recipe *data.RecipeTemplate) {
	inv := player.Inventory()
	for _, ing := range recipe.Ingredients {
		removed := inv.RemoveItemsByID(ing.ItemID, int64(ing.Count))
		if removed < int64(ing.Count) {
			slog.Warn("partial material consumption",
				"player", player.Name(),
				"itemID", ing.ItemID,
				"expected", ing.Count,
				"removed", removed)
		}
	}
}
