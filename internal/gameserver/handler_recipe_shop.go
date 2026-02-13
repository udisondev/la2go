package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	skilldata "github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// ─── Recipe Shop (Manufacture) handlers ─────────────────────────────────────
// Phase 54: Recipe Shop (Manufacture) System.
// Java reference: RequestRecipeShopMessageSet, RequestRecipeShopListSet,
// RequestRecipeShopManageQuit, RequestRecipeShopMakeInfo,
// RequestRecipeShopMakeItem, RequestRecipeShopManagePrev.

// handleRequestRecipeShopMessageSet processes 0xB1 — set manufacture shop message.
// Client sends the store title text. Max 29 characters (enforced by SetStoreMessage).
func (h *Handler) handleRequestRecipeShopMessageSet(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	r := packet.NewReader(data)

	msg, err := r.ReadString()
	if err != nil {
		return 0, false, fmt.Errorf("read recipe shop message: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	player.SetStoreMessage(msg)

	// Broadcast manufacture message to nearby players
	msgPkt := &serverpackets.RecipeShopMsg{
		ObjectID: player.ObjectID(),
		Message:  player.StoreMessage(),
	}
	msgData, err := msgPkt.Write()
	if err != nil {
		slog.Error("serialize RecipeShopMsg", "error", err)
		return 0, true, nil
	}

	h.clientManager.BroadcastToVisibleNear(player, msgData, len(msgData))

	return 0, true, nil
}

// handleRequestRecipeShopListSet processes 0xB2 — open manufacture store.
// Client sends the list of recipes to offer with prices.
func (h *Handler) handleRequestRecipeShopListSet(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Не открываем магазин в бою
	if player.HasAttackStance() {
		return h.sendActionFailed(buf)
	}

	// Не открываем если уже в торговом режиме (кроме manage)
	storeType := player.PrivateStoreType()
	if storeType != model.StoreNone && storeType != model.StoreManufactureManage {
		return h.sendActionFailed(buf)
	}

	r := packet.NewReader(data)

	count, err := r.ReadInt()
	if err != nil {
		return 0, false, fmt.Errorf("read recipe shop list count: %w", err)
	}

	// Защита от переполнения: максимум 100 рецептов
	if count <= 0 || count > 100 {
		player.SetPrivateStoreType(model.StoreNone)
		return h.sendActionFailed(buf)
	}

	items := make([]*model.ManufactureItem, 0, count)

	for range count {
		recipeID, rErr := r.ReadInt()
		if rErr != nil {
			return 0, false, fmt.Errorf("read recipe shop recipeID: %w", rErr)
		}

		cost, cErr := r.ReadLong()
		if cErr != nil {
			return 0, false, fmt.Errorf("read recipe shop cost: %w", cErr)
		}

		// Цена должна быть положительной
		if cost <= 0 {
			slog.Debug("recipe shop: invalid cost",
				"recipeID", recipeID,
				"cost", cost,
				"player", player.Name())
			continue
		}

		// Overflow protection
		if cost > model.MaxAdena {
			slog.Debug("recipe shop: cost exceeds max adena",
				"recipeID", recipeID,
				"cost", cost,
				"player", player.Name())
			continue
		}

		// Проверяем что игрок знает этот рецепт
		if !player.HasRecipe(recipeID) {
			slog.Warn("recipe shop: player does not know recipe",
				"recipeID", recipeID,
				"player", player.Name())
			continue
		}

		// Определяем тип рецепта
		recipe := skilldata.GetRecipeTemplate(recipeID)
		isDwarven := false
		if recipe != nil {
			isDwarven = recipe.IsDwarven
		}

		items = append(items, &model.ManufactureItem{
			RecipeID:  recipeID,
			Cost:      cost,
			IsDwarven: isDwarven,
		})
	}

	if len(items) == 0 {
		player.SetPrivateStoreType(model.StoreNone)
		player.ClearManufactureItems()
		return h.sendActionFailed(buf)
	}

	// Сохраняем список и активируем магазин
	player.SetManufactureItems(items)
	player.SetPrivateStoreType(model.StoreManufacture)

	// Крафтер садится (как при обычном private store)
	player.SetSitting(true)
	sitPkt := serverpackets.NewChangeWaitType(player, serverpackets.WaitTypeSitting)
	sitData, _ := sitPkt.Write()
	h.clientManager.BroadcastToVisibleNear(player, sitData, len(sitData))

	// Broadcast UserInfo чтобы рядом стоящие игроки увидели иконку магазина
	userInfo := serverpackets.NewUserInfo(player)
	uiData, _ := userInfo.Write()
	h.clientManager.BroadcastToVisibleNear(player, uiData, len(uiData))

	// Broadcast store message
	storeMsg := player.StoreMessage()
	if storeMsg != "" {
		msgPkt := &serverpackets.RecipeShopMsg{
			ObjectID: player.ObjectID(),
			Message:  storeMsg,
		}
		msgData, mErr := msgPkt.Write()
		if mErr == nil {
			h.clientManager.BroadcastToVisibleNear(player, msgData, len(msgData))
		}
	}

	slog.Info("manufacture store opened",
		"player", player.Name(),
		"recipes", len(items))

	return 0, true, nil
}

// handleRequestRecipeShopManageQuit processes 0xB3 — close manufacture management window.
// This is a client UI event, no server state change needed
// (the player may still have their store open).
func (h *Handler) handleRequestRecipeShopManageQuit(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Если игрок в режиме manage — сбрасываем в StoreNone
	if player.PrivateStoreType() == model.StoreManufactureManage {
		player.SetPrivateStoreType(model.StoreNone)
	}

	return h.sendActionFailed(buf)
}

// handleRequestRecipeShopMakeInfo processes 0xB5 — request recipe shop item info.
// The buyer requests crafting info about a specific recipe in the crafter's shop.
func (h *Handler) handleRequestRecipeShopMakeInfo(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	r := packet.NewReader(data)

	shopObjectID, err := r.ReadInt()
	if err != nil {
		return 0, false, fmt.Errorf("read shop objectID: %w", err)
	}

	recipeID, err := r.ReadInt()
	if err != nil {
		return 0, false, fmt.Errorf("read recipeID: %w", err)
	}

	// Находим владельца магазина
	shopClient := h.clientManager.GetClientByObjectID(uint32(shopObjectID))
	if shopClient == nil {
		return h.sendActionFailed(buf)
	}

	shopOwner := shopClient.ActivePlayer()
	if shopOwner == nil {
		return h.sendActionFailed(buf)
	}

	// Проверяем что у владельца открыт магазин рецептов
	if shopOwner.PrivateStoreType() != model.StoreManufacture {
		return h.sendActionFailed(buf)
	}

	// Отправляем RecipeShopItemInfo с данными крафтера
	pkt := &serverpackets.RecipeShopItemInfo{
		CrafterObjectID: shopOwner.ObjectID(),
		RecipeID:        recipeID,
		CurrentMP:       shopOwner.CurrentMP(),
		MaxMP:           shopOwner.MaxMP(),
	}
	pktData, err := pkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("serialize RecipeShopItemInfo: %w", err)
	}

	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestRecipeShopMakeItem processes 0xB6 — craft item from recipe shop.
// The buyer initiates crafting from another player's manufacture store.
// Materials come from the buyer's inventory, crafting is done by the shop owner.
func (h *Handler) handleRequestRecipeShopMakeItem(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	r := packet.NewReader(data)

	shopObjectID, err := r.ReadInt()
	if err != nil {
		return 0, false, fmt.Errorf("read shop objectID: %w", err)
	}

	recipeID, err := r.ReadInt()
	if err != nil {
		return 0, false, fmt.Errorf("read recipeID: %w", err)
	}

	// Пропускаем unknown (int64) и cost (int64) — клиент шлёт, но мы берём цену из списка магазина
	if _, skipErr := r.ReadLong(); skipErr != nil {
		// Допускаем отсутствие необязательных полей
		slog.Debug("recipe shop make item: short packet (no unknown field)")
	}

	// Находим владельца магазина
	shopClient := h.clientManager.GetClientByObjectID(uint32(shopObjectID))
	if shopClient == nil {
		slog.Warn("recipe shop: shop owner not found",
			"shopObjectID", shopObjectID,
			"buyer", player.Name())
		return h.sendActionFailed(buf)
	}

	shopOwner := shopClient.ActivePlayer()
	if shopOwner == nil {
		return h.sendActionFailed(buf)
	}

	// Проверяем что у владельца открыт магазин
	if shopOwner.PrivateStoreType() != model.StoreManufacture {
		slog.Warn("recipe shop: owner not in manufacture mode",
			"shopOwner", shopOwner.Name(),
			"buyer", player.Name())
		return h.sendActionFailed(buf)
	}

	// Ищем рецепт в списке магазина
	manufactureItem := shopOwner.FindManufactureItem(recipeID)
	if manufactureItem == nil {
		slog.Warn("recipe shop: recipe not in manufacture list",
			"recipeID", recipeID,
			"shopOwner", shopOwner.Name(),
			"buyer", player.Name())
		return h.sendActionFailed(buf)
	}

	cost := manufactureItem.Cost

	// Проверяем что у покупателя достаточно адены
	buyerAdena := player.Inventory().GetAdena()
	if buyerAdena < cost {
		slog.Debug("recipe shop: buyer has insufficient adena",
			"have", buyerAdena,
			"need", cost,
			"buyer", player.Name())
		return h.sendActionFailed(buf)
	}

	// Выполняем крафт через craft.Controller (материалы берём из инвентаря покупателя)
	if h.craftController == nil {
		slog.Warn("recipe shop: craft controller not initialized")
		return h.sendActionFailed(buf)
	}

	// Крафт (материалы из инвентаря покупателя)
	result, craftErr := h.craftController.Craft(player, recipeID)

	// Списываем адену с покупателя и начисляем крафтеру НЕЗАВИСИМО от результата крафта.
	// В Java-реализации оплата происходит даже при провале крафта (failure = materials lost + adena paid).
	if err := player.Inventory().RemoveAdena(int32(cost)); err != nil {
		slog.Error("recipe shop: remove adena from buyer",
			"cost", cost,
			"buyer", player.Name(),
			"error", err)
		return h.sendActionFailed(buf)
	}
	if err := shopOwner.Inventory().AddAdena(int32(cost)); err != nil {
		slog.Error("recipe shop: add adena to crafter",
			"cost", cost,
			"shopOwner", shopOwner.Name(),
			"error", err)
		// Адена уже списана с покупателя, но не пришла крафтеру — логируем ошибку
	}

	// Отправляем результат покупателю
	recipe := skilldata.GetRecipeTemplate(recipeID)
	isDwarven := false
	if recipe != nil {
		isDwarven = recipe.IsDwarven
	}

	craftSuccess := craftErr == nil && result != nil && result.Success

	buyerResp := &serverpackets.RecipeShopItemInfo{
		CrafterObjectID: shopOwner.ObjectID(),
		RecipeID:        recipeID,
		CurrentMP:       shopOwner.CurrentMP(),
		MaxMP:           shopOwner.MaxMP(),
	}
	buyerRespData, _ := buyerResp.Write()
	n := copy(buf, buyerRespData)

	// Отправляем результат крафтеру через SendToPlayer
	ownerResp := &serverpackets.RecipeItemMakeInfo{
		RecipeListID:   recipeID,
		IsDwarvenCraft: isDwarven,
		CurrentMP:      shopOwner.CurrentMP(),
		MaxMP:          shopOwner.MaxMP(),
		Success:        craftSuccess,
	}
	ownerRespData, _ := ownerResp.Write()
	if sendErr := h.clientManager.SendToPlayer(shopOwner.ObjectID(), ownerRespData, len(ownerRespData)); sendErr != nil {
		slog.Error("recipe shop: send result to crafter",
			"shopOwner", shopOwner.Name(),
			"error", sendErr)
	}

	if craftErr != nil {
		slog.Warn("recipe shop: craft failed",
			"recipeID", recipeID,
			"shopOwner", shopOwner.Name(),
			"buyer", player.Name(),
			"error", craftErr)
	} else if craftSuccess {
		slog.Info("recipe shop: craft success",
			"recipeID", recipeID,
			"shopOwner", shopOwner.Name(),
			"buyer", player.Name(),
			"cost", cost)
	} else {
		slog.Info("recipe shop: craft failed (random)",
			"recipeID", recipeID,
			"shopOwner", shopOwner.Name(),
			"buyer", player.Name(),
			"cost", cost)
	}

	return n, true, nil
}

// handleRequestRecipeShopManagePrev processes 0xB7 — return to recipe management window.
// Re-opens the manufacture management window showing the player's known recipes.
func (h *Handler) handleRequestRecipeShopManagePrev(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	player.SetPrivateStoreType(model.StoreManufactureManage)

	// Собираем все рецепты (dwarven + common)
	dwarvenIDs := player.GetRecipeBook(true)
	commonIDs := player.GetRecipeBook(false)

	allRecipes := make([]int32, 0, len(dwarvenIDs)+len(commonIDs))
	allRecipes = append(allRecipes, dwarvenIDs...)
	allRecipes = append(allRecipes, commonIDs...)

	pkt := &serverpackets.RecipeShopManageList{
		PlayerObjectID: player.ObjectID(),
		CurrentMP:      player.CurrentMP(),
		MaxMP:          player.MaxMP(),
		RecipeIDs:      allRecipes,
	}
	pktData, err := pkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("serialize RecipeShopManageList: %w", err)
	}

	n := copy(buf, pktData)
	return n, true, nil
}

