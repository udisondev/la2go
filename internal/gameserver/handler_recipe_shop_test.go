package gameserver

import (
	"context"
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// --- Helpers for recipe shop tests ---

func newRecipeShopHandler(t *testing.T) *Handler {
	t.Helper()
	cm := NewClientManager()
	return NewHandler(nil, cm, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
}

func newTestPlayerWithAdena(t *testing.T, objectID uint32, adena int32) *model.Player {
	t.Helper()
	player, err := model.NewPlayer(objectID, int64(objectID), int64(objectID), "TestPlayer", 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}

	if adena > 0 {
		adenaTmpl := &model.ItemTemplate{
			ItemID:    model.AdenaItemID,
			Name:      "Adena",
			Type:      model.ItemTypeEtcItem,
			Stackable: true,
			Tradeable: true,
		}
		adenaItem, iErr := model.NewItem(objectID+50000, model.AdenaItemID, 1, adena, adenaTmpl)
		if iErr != nil {
			t.Fatalf("NewItem (adena): %v", iErr)
		}
		if aErr := player.Inventory().AddItem(adenaItem); aErr != nil {
			t.Fatalf("AddItem (adena): %v", aErr)
		}
	}

	return player
}

func encodeUTF16LEString(s string) []byte {
	// UTF-16LE encode + null terminator
	w := packet.NewWriter(len(s)*2 + 2)
	w.WriteString(s)
	return w.Bytes()
}

// --- Tests ---

func TestHandleRequestRecipeShopMessageSet(t *testing.T) {
	h := newRecipeShopHandler(t)

	player, err := model.NewPlayer(1001, 1, 1, "Crafter", 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}

	client := &GameClient{}
	client.SetActivePlayer(player)

	tests := []struct {
		name    string
		message string
		want    string
	}{
		{"short message", "My Shop", "My Shop"},
		{"max length", "12345678901234567890123456789", "12345678901234567890123456789"},
		{"exceeds max", "123456789012345678901234567890", "12345678901234567890123456789"},
		{"empty message", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgData := encodeUTF16LEString(tt.message)
			buf := make([]byte, 1024)

			_, ok, err := h.handleRequestRecipeShopMessageSet(context.Background(), client, msgData, buf)
			if err != nil {
				t.Fatalf("handleRequestRecipeShopMessageSet() error = %v", err)
			}
			if !ok {
				t.Fatal("handleRequestRecipeShopMessageSet() returned ok=false, want true")
			}

			got := player.StoreMessage()
			if got != tt.want {
				t.Errorf("StoreMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHandleRequestRecipeShopListSet(t *testing.T) {
	t.Run("empty list closes store", func(t *testing.T) {
		h := newRecipeShopHandler(t)
		player := newTestPlayerWithAdena(t, 1001, 1000)
		client := &GameClient{}
		client.SetActivePlayer(player)

		// count=0
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, 0)

		buf := make([]byte, 1024)
		n, ok, err := h.handleRequestRecipeShopListSet(context.Background(), client, data, buf)
		if err != nil {
			t.Fatalf("handleRequestRecipeShopListSet() error = %v", err)
		}
		if !ok {
			t.Fatal("ok=false")
		}
		// Should return ActionFailed
		if n > 0 && buf[0] != 0x25 {
			t.Errorf("expected ActionFailed (0x25), got 0x%02X", buf[0])
		}
		if player.PrivateStoreType() != model.StoreNone {
			t.Errorf("store type = %v, want StoreNone", player.PrivateStoreType())
		}
	})

	t.Run("negative count returns action failed", func(t *testing.T) {
		h := newRecipeShopHandler(t)
		player := newTestPlayerWithAdena(t, 1002, 1000)
		client := &GameClient{}
		client.SetActivePlayer(player)

		// count=-1
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, 0xFFFFFFFF) // -1 as int32
		buf := make([]byte, 1024)

		n, ok, err := h.handleRequestRecipeShopListSet(context.Background(), client, data, buf)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if !ok {
			t.Fatal("ok=false")
		}
		if n > 0 && buf[0] != 0x25 {
			t.Errorf("expected ActionFailed (0x25), got 0x%02X", buf[0])
		}
	})

	t.Run("combat blocks store opening", func(t *testing.T) {
		h := newRecipeShopHandler(t)
		player := newTestPlayerWithAdena(t, 1003, 1000)
		player.MarkAttackStance() // ставим боевую стойку
		client := &GameClient{}
		client.SetActivePlayer(player)

		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, 1)
		buf := make([]byte, 1024)

		n, ok, err := h.handleRequestRecipeShopListSet(context.Background(), client, data, buf)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if !ok {
			t.Fatal("ok=false")
		}
		// Should return ActionFailed
		if n > 0 && buf[0] != 0x25 {
			t.Errorf("expected ActionFailed (0x25), got 0x%02X", buf[0])
		}
	})

	t.Run("recipe not learned is skipped", func(t *testing.T) {
		h := newRecipeShopHandler(t)
		player := newTestPlayerWithAdena(t, 1004, 1000)
		client := &GameClient{}
		client.SetActivePlayer(player)

		// Отправляем 1 рецепт, который игрок НЕ знает
		w := packet.NewWriter(16)
		w.WriteInt(1) // count
		w.WriteInt(42) // recipeID — не выучен
		w.WriteLong(100) // cost

		buf := make([]byte, 1024)
		n, ok, err := h.handleRequestRecipeShopListSet(context.Background(), client, w.Bytes(), buf)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if !ok {
			t.Fatal("ok=false")
		}
		// Все рецепты отклонены — ActionFailed
		if n > 0 && buf[0] != 0x25 {
			t.Errorf("expected ActionFailed (0x25), got 0x%02X", buf[0])
		}
		if player.PrivateStoreType() != model.StoreNone {
			t.Errorf("store type = %v, want StoreNone", player.PrivateStoreType())
		}
	})

	t.Run("valid recipe opens store", func(t *testing.T) {
		h := newRecipeShopHandler(t)
		player := newTestPlayerWithAdena(t, 1005, 1000)
		client := &GameClient{}
		client.SetActivePlayer(player)

		// Выучиваем рецепт
		recipeID := int32(1) // используем рецепт, который может существовать в таблице
		if err := player.LearnRecipe(recipeID, false); err != nil {
			t.Fatalf("LearnRecipe: %v", err)
		}

		w := packet.NewWriter(16)
		w.WriteInt(1)          // count
		w.WriteInt(recipeID)   // recipeID
		w.WriteLong(500)       // cost

		buf := make([]byte, 4096)
		_, ok, err := h.handleRequestRecipeShopListSet(context.Background(), client, w.Bytes(), buf)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if !ok {
			t.Fatal("ok=false")
		}

		if player.PrivateStoreType() != model.StoreManufacture {
			t.Errorf("store type = %v, want StoreManufacture", player.PrivateStoreType())
		}

		items := player.ManufactureItems()
		if len(items) != 1 {
			t.Fatalf("manufacture items count = %d, want 1", len(items))
		}
		if items[0].RecipeID != recipeID {
			t.Errorf("recipeID = %d, want %d", items[0].RecipeID, recipeID)
		}
		if items[0].Cost != 500 {
			t.Errorf("cost = %d, want 500", items[0].Cost)
		}
	})
}

func TestHandleRequestRecipeShopManageQuit(t *testing.T) {
	h := newRecipeShopHandler(t)
	player, _ := model.NewPlayer(1010, 1, 1, "Crafter", 40, 0, 0)
	client := &GameClient{}
	client.SetActivePlayer(player)

	t.Run("manufacture manage resets to none", func(t *testing.T) {
		player.SetPrivateStoreType(model.StoreManufactureManage)

		buf := make([]byte, 1024)
		_, ok, err := h.handleRequestRecipeShopManageQuit(context.Background(), client, nil, buf)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if !ok {
			t.Fatal("ok=false")
		}
		if player.PrivateStoreType() != model.StoreNone {
			t.Errorf("store type = %v, want StoreNone", player.PrivateStoreType())
		}
	})

	t.Run("active store is not closed", func(t *testing.T) {
		player.SetPrivateStoreType(model.StoreManufacture)

		buf := make([]byte, 1024)
		_, ok, err := h.handleRequestRecipeShopManageQuit(context.Background(), client, nil, buf)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if !ok {
			t.Fatal("ok=false")
		}
		// Магазин должен оставаться активным
		if player.PrivateStoreType() != model.StoreManufacture {
			t.Errorf("store type = %v, want StoreManufacture", player.PrivateStoreType())
		}
	})
}

func TestHandleRequestRecipeShopMakeInfo(t *testing.T) {
	t.Run("shop owner not found returns action failed", func(t *testing.T) {
		h := newRecipeShopHandler(t)
		player, _ := model.NewPlayer(2001, 1, 1, "Buyer", 40, 0, 0)
		client := &GameClient{}
		client.SetActivePlayer(player)

		w := packet.NewWriter(8)
		w.WriteInt(9999) // shopObjectID — несуществующий
		w.WriteInt(1)    // recipeID

		buf := make([]byte, 1024)
		n, ok, err := h.handleRequestRecipeShopMakeInfo(context.Background(), client, w.Bytes(), buf)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if !ok {
			t.Fatal("ok=false")
		}
		if n > 0 && buf[0] != 0x25 {
			t.Errorf("expected ActionFailed (0x25), got 0x%02X", buf[0])
		}
	})
}

func TestHandleRequestRecipeShopManagePrev(t *testing.T) {
	h := newRecipeShopHandler(t)
	player, _ := model.NewPlayer(3001, 1, 1, "Crafter", 40, 0, 0)
	client := &GameClient{}
	client.SetActivePlayer(player)

	// Выучиваем рецепты
	_ = player.LearnRecipe(100, false) // common
	_ = player.LearnRecipe(200, true)  // dwarven

	buf := make([]byte, 4096)
	n, ok, err := h.handleRequestRecipeShopManagePrev(context.Background(), client, nil, buf)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if !ok {
		t.Fatal("ok=false")
	}

	if player.PrivateStoreType() != model.StoreManufactureManage {
		t.Errorf("store type = %v, want StoreManufactureManage", player.PrivateStoreType())
	}

	// Должен вернуть пакет 0xD8 (RecipeShopManageList)
	if n == 0 {
		t.Fatal("expected response packet, got n=0")
	}
	if buf[0] != 0xD8 {
		t.Errorf("opcode = 0x%02X, want 0xD8", buf[0])
	}
}

func TestManufactureItemsPlayerModel(t *testing.T) {
	player, _ := model.NewPlayer(4001, 1, 1, "TestCrafter", 40, 0, 0)

	t.Run("initially nil", func(t *testing.T) {
		if items := player.ManufactureItems(); items != nil {
			t.Errorf("initial ManufactureItems = %v, want nil", items)
		}
	})

	t.Run("set and get", func(t *testing.T) {
		items := []*model.ManufactureItem{
			{RecipeID: 10, Cost: 100, IsDwarven: false},
			{RecipeID: 20, Cost: 200, IsDwarven: true},
		}
		player.SetManufactureItems(items)

		got := player.ManufactureItems()
		if len(got) != 2 {
			t.Fatalf("ManufactureItems() len = %d, want 2", len(got))
		}
		if got[0].RecipeID != 10 {
			t.Errorf("got[0].RecipeID = %d, want 10", got[0].RecipeID)
		}
		if got[1].Cost != 200 {
			t.Errorf("got[1].Cost = %d, want 200", got[1].Cost)
		}
	})

	t.Run("find manufacture item", func(t *testing.T) {
		found := player.FindManufactureItem(10)
		if found == nil {
			t.Fatal("FindManufactureItem(10) = nil, want non-nil")
		}
		if found.Cost != 100 {
			t.Errorf("found.Cost = %d, want 100", found.Cost)
		}

		notFound := player.FindManufactureItem(999)
		if notFound != nil {
			t.Errorf("FindManufactureItem(999) = %v, want nil", notFound)
		}
	})

	t.Run("clear", func(t *testing.T) {
		player.ClearManufactureItems()
		if items := player.ManufactureItems(); items != nil {
			t.Errorf("after Clear, ManufactureItems = %v, want nil", items)
		}
	})

	t.Run("close private store clears manufacture", func(t *testing.T) {
		player.SetManufactureItems([]*model.ManufactureItem{
			{RecipeID: 30, Cost: 300, IsDwarven: false},
		})
		player.SetPrivateStoreType(model.StoreManufacture)
		player.SetStoreMessage("Test Shop")

		player.ClosePrivateStore()

		if player.PrivateStoreType() != model.StoreNone {
			t.Errorf("store type after close = %v, want StoreNone", player.PrivateStoreType())
		}
		if items := player.ManufactureItems(); items != nil {
			t.Errorf("manufacture items after close = %v, want nil", items)
		}
		if msg := player.StoreMessage(); msg != "" {
			t.Errorf("store message after close = %q, want empty", msg)
		}
	})
}

func TestRecipeShopPacketSerialization(t *testing.T) {
	t.Run("RecipeShopManageList opcode 0xD8", func(t *testing.T) {
		pkt := &serverpackets.RecipeShopManageList{
			PlayerObjectID: 12345,
			CurrentMP:      500,
			MaxMP:          1000,
			RecipeIDs:      []int32{10, 20, 30},
		}
		data, err := pkt.Write()
		if err != nil {
			t.Fatalf("Write: %v", err)
		}
		if data[0] != 0xD8 {
			t.Errorf("opcode = 0x%02X, want 0xD8", data[0])
		}
		// objectID at offset 1-4
		objID := int32(binary.LittleEndian.Uint32(data[1:5]))
		if objID != 12345 {
			t.Errorf("objectID = %d, want 12345", objID)
		}
		// count at offset 13-16
		count := int32(binary.LittleEndian.Uint32(data[13:17]))
		if count != 3 {
			t.Errorf("count = %d, want 3", count)
		}
	})

	t.Run("RecipeShopSellList opcode 0xD9", func(t *testing.T) {
		pkt := &serverpackets.RecipeShopSellList{
			SellerObjectID: 54321,
			CurrentMP:      300,
			MaxMP:          600,
			Adena:          10000,
			Items: []*model.ManufactureItem{
				{RecipeID: 42, Cost: 1500, IsDwarven: true},
			},
		}
		data, err := pkt.Write()
		if err != nil {
			t.Fatalf("Write: %v", err)
		}
		if data[0] != 0xD9 {
			t.Errorf("opcode = 0x%02X, want 0xD9", data[0])
		}
	})

	t.Run("RecipeShopMsg opcode 0xDB", func(t *testing.T) {
		pkt := &serverpackets.RecipeShopMsg{
			ObjectID: 77777,
			Message:  "Craft!",
		}
		data, err := pkt.Write()
		if err != nil {
			t.Fatalf("Write: %v", err)
		}
		if data[0] != 0xDB {
			t.Errorf("opcode = 0x%02X, want 0xDB", data[0])
		}
		objID := int32(binary.LittleEndian.Uint32(data[1:5]))
		if objID != 77777 {
			t.Errorf("objectID = %d, want 77777", objID)
		}
	})
}

func TestPrivateStoreTypeManufactureManage(t *testing.T) {
	if model.StoreManufactureManage.String() != "ManufactureManage" {
		t.Errorf("String() = %q, want %q", model.StoreManufactureManage.String(), "ManufactureManage")
	}

	// ManufactureManage is NOT a store mode (it's a manage mode)
	if model.StoreManufactureManage.IsInStoreMode() {
		t.Error("ManufactureManage.IsInStoreMode() = true, want false")
	}

	// Manufacture IS a store mode
	if !model.StoreManufacture.IsInStoreMode() {
		t.Error("Manufacture.IsInStoreMode() = false, want true")
	}
}
