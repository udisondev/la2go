package itemhandler

import (
	"log/slog"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

// itemSkillsHandler processes items that cast a skill on use.
// Covers ~696 items: potions, buff scrolls, consumable skill items.
//
// Java reference: ItemSkillsTemplate.java, ItemSkills.java
type itemSkillsHandler struct{}

func (h *itemSkillsHandler) UseItem(player *model.Player, item *model.Item, skillID, skillLevel int32) *UseResult {
	if skillID == 0 {
		slog.Debug("ItemSkills: no skill for item",
			"itemID", item.ItemID(),
			"player", player.Name())
		return nil
	}

	// Verify skill template exists
	tmpl := data.GetSkillTemplate(skillID, skillLevel)
	if tmpl == nil {
		slog.Warn("ItemSkills: skill template not found",
			"skillID", skillID,
			"level", skillLevel,
			"itemID", item.ItemID())
		return nil
	}

	def := data.GetItemDef(item.ItemID())
	reuseDelay := int32(0)
	if def != nil {
		reuseDelay = def.ReuseDelay()
	}

	return &UseResult{
		ConsumeCount: 1,
		SkillID:      skillID,
		SkillLevel:   skillLevel,
		ReuseDelay:   reuseDelay,
		Broadcast:    true,
	}
}

// elixirHandler is identical to itemSkillsHandler but restricted to player chars only (no pets).
// In our implementation pets aren't supported yet, so it's functionally identical.
//
// Java reference: Elixir.java extends ItemSkills
type elixirHandler struct {
	itemSkillsHandler
}
