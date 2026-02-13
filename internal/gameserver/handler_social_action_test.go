package gameserver

import (
	"context"
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

func newSocialActionHandler(t *testing.T) *Handler {
	t.Helper()
	cm := NewClientManager()
	return NewHandler(nil, cm, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
}

func newSocialActionPlayer(t *testing.T, objectID uint32) *model.Player {
	t.Helper()
	player, err := model.NewPlayer(objectID, int64(objectID), int64(objectID), "TestPlayer", 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))
	return player
}

func encodeSocialAction(actionID int32) []byte {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(actionID))
	return data
}

func TestHandleRequestSocialAction_ValidActions(t *testing.T) {
	h := newSocialActionHandler(t)

	tests := []struct {
		name     string
		actionID int32
	}{
		{"greeting", serverpackets.SocialActionGreeting},
		{"victory", serverpackets.SocialActionVictory},
		{"advance", serverpackets.SocialActionAdvance},
		{"etc", serverpackets.SocialActionEtc},
		{"yes", serverpackets.SocialActionYes},
		{"no", serverpackets.SocialActionNo},
		{"bow", serverpackets.SocialActionBow},
		{"unaware", serverpackets.SocialActionUnaware},
		{"wait", serverpackets.SocialActionWait},
		{"laugh", serverpackets.SocialActionLaugh},
		{"applaud", serverpackets.SocialActionApplaud},
		{"dance", serverpackets.SocialActionDance},
		{"sorrow", serverpackets.SocialActionSorrow},
		{"shyness", serverpackets.SocialActionShyness},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := newSocialActionPlayer(t, 1001)
			client := &GameClient{}
			client.SetActivePlayer(player)

			data := encodeSocialAction(tt.actionID)
			buf := make([]byte, 1024)

			n, ok, err := h.handleRequestSocialAction(context.Background(), client, data, buf)
			if err != nil {
				t.Fatalf("handleRequestSocialAction() error = %v", err)
			}
			// Broadcast-only — no data in buf for the sender
			if ok {
				t.Errorf("ok = true; want false (broadcast only)")
			}
			if n != 0 {
				t.Errorf("n = %d; want 0", n)
			}
		})
	}
}

func TestHandleRequestSocialAction_InvalidActionID(t *testing.T) {
	h := newSocialActionHandler(t)

	tests := []struct {
		name     string
		actionID int32
	}{
		{"below_min", 1},
		{"zero", 0},
		{"negative", -1},
		{"above_max", 17},
		{"very_high", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := newSocialActionPlayer(t, 2001)
			client := &GameClient{}
			client.SetActivePlayer(player)

			data := encodeSocialAction(tt.actionID)
			buf := make([]byte, 1024)

			n, ok, err := h.handleRequestSocialAction(context.Background(), client, data, buf)
			if err != nil {
				t.Fatalf("handleRequestSocialAction() error = %v", err)
			}
			if !ok {
				t.Error("ok = false; want true (ActionFailed response)")
			}
			if n == 0 {
				t.Fatal("n = 0; want > 0 (ActionFailed packet)")
			}
			if buf[0] != serverpackets.OpcodeActionFailed {
				t.Errorf("opcode = 0x%02X; want 0x%02X (ActionFailed)",
					buf[0], serverpackets.OpcodeActionFailed)
			}
		})
	}
}

func TestHandleRequestSocialAction_DeadPlayer(t *testing.T) {
	h := newSocialActionHandler(t)
	player := newSocialActionPlayer(t, 3001)
	player.SetCurrentHP(0) // мертвый

	client := &GameClient{}
	client.SetActivePlayer(player)

	data := encodeSocialAction(serverpackets.SocialActionGreeting)
	buf := make([]byte, 1024)

	n, ok, err := h.handleRequestSocialAction(context.Background(), client, data, buf)
	if err != nil {
		t.Fatalf("handleRequestSocialAction() error = %v", err)
	}
	if !ok {
		t.Error("ok = false; want true (ActionFailed for dead player)")
	}
	if n == 0 {
		t.Fatal("n = 0; want > 0 (ActionFailed)")
	}
	if buf[0] != serverpackets.OpcodeActionFailed {
		t.Errorf("opcode = 0x%02X; want 0x%02X", buf[0], serverpackets.OpcodeActionFailed)
	}
}

func TestHandleRequestSocialAction_InStoreMode(t *testing.T) {
	h := newSocialActionHandler(t)
	player := newSocialActionPlayer(t, 4001)
	player.SetPrivateStoreType(model.StoreSell)

	client := &GameClient{}
	client.SetActivePlayer(player)

	data := encodeSocialAction(serverpackets.SocialActionDance)
	buf := make([]byte, 1024)

	n, ok, err := h.handleRequestSocialAction(context.Background(), client, data, buf)
	if err != nil {
		t.Fatalf("handleRequestSocialAction() error = %v", err)
	}
	if !ok {
		t.Error("ok = false; want true (ActionFailed for store mode)")
	}
	if n == 0 {
		t.Fatal("n = 0; want > 0")
	}
	if buf[0] != serverpackets.OpcodeActionFailed {
		t.Errorf("opcode = 0x%02X; want 0x%02X", buf[0], serverpackets.OpcodeActionFailed)
	}
}

func TestHandleRequestSocialAction_Fishing(t *testing.T) {
	h := newSocialActionHandler(t)
	player := newSocialActionPlayer(t, 5001)
	player.SetFishing(true)

	client := &GameClient{}
	client.SetActivePlayer(player)

	data := encodeSocialAction(serverpackets.SocialActionLaugh)
	buf := make([]byte, 1024)

	n, ok, err := h.handleRequestSocialAction(context.Background(), client, data, buf)
	if err != nil {
		t.Fatalf("handleRequestSocialAction() error = %v", err)
	}
	if !ok {
		t.Error("ok = false; want true (ActionFailed for fishing)")
	}
	if n == 0 {
		t.Fatal("n = 0; want > 0")
	}
	if buf[0] != serverpackets.OpcodeActionFailed {
		t.Errorf("opcode = 0x%02X; want 0x%02X", buf[0], serverpackets.OpcodeActionFailed)
	}
}

func TestHandleRequestSocialAction_CharmRequiresHero(t *testing.T) {
	h := newSocialActionHandler(t)

	t.Run("non-hero denied", func(t *testing.T) {
		player := newSocialActionPlayer(t, 6001)
		// player.IsHero() == false by default
		client := &GameClient{}
		client.SetActivePlayer(player)

		data := encodeSocialAction(serverpackets.SocialActionCharm)
		buf := make([]byte, 1024)

		n, ok, err := h.handleRequestSocialAction(context.Background(), client, data, buf)
		if err != nil {
			t.Fatalf("handleRequestSocialAction() error = %v", err)
		}
		if !ok {
			t.Error("ok = false; want true (ActionFailed for non-hero charm)")
		}
		if n == 0 {
			t.Fatal("n = 0; want > 0")
		}
		if buf[0] != serverpackets.OpcodeActionFailed {
			t.Errorf("opcode = 0x%02X; want 0x%02X", buf[0], serverpackets.OpcodeActionFailed)
		}
	})

	t.Run("hero allowed", func(t *testing.T) {
		player := newSocialActionPlayer(t, 6002)
		player.SetHero(true)
		client := &GameClient{}
		client.SetActivePlayer(player)

		data := encodeSocialAction(serverpackets.SocialActionCharm)
		buf := make([]byte, 1024)

		n, ok, err := h.handleRequestSocialAction(context.Background(), client, data, buf)
		if err != nil {
			t.Fatalf("handleRequestSocialAction() error = %v", err)
		}
		if ok {
			t.Error("ok = true; want false (broadcast only for hero charm)")
		}
		if n != 0 {
			t.Errorf("n = %d; want 0", n)
		}
	})
}

func TestHandleRequestSocialAction_NoActivePlayer(t *testing.T) {
	h := newSocialActionHandler(t)
	client := &GameClient{}
	// no active player set

	data := encodeSocialAction(serverpackets.SocialActionGreeting)
	buf := make([]byte, 1024)

	_, _, err := h.handleRequestSocialAction(context.Background(), client, data, buf)
	if err == nil {
		t.Fatal("handleRequestSocialAction() expected error for no active player, got nil")
	}
}

func TestHandleRequestSocialAction_MalformedPacket(t *testing.T) {
	h := newSocialActionHandler(t)
	player := newSocialActionPlayer(t, 7001)
	client := &GameClient{}
	client.SetActivePlayer(player)

	// Слишком короткие данные — не хватает для int32
	data := make([]byte, 2)
	buf := make([]byte, 1024)

	_, _, err := h.handleRequestSocialAction(context.Background(), client, data, buf)
	if err == nil {
		t.Fatal("handleRequestSocialAction() expected error for malformed packet, got nil")
	}
}

func TestHandleRequestSocialAction_StoreManageNotBlocked(t *testing.T) {
	// StoreSellManage is NOT "in store mode" — it's just the manage UI
	h := newSocialActionHandler(t)
	player := newSocialActionPlayer(t, 8001)
	player.SetPrivateStoreType(model.StoreSellManage)

	client := &GameClient{}
	client.SetActivePlayer(player)

	data := encodeSocialAction(serverpackets.SocialActionGreeting)
	buf := make([]byte, 1024)

	n, ok, err := h.handleRequestSocialAction(context.Background(), client, data, buf)
	if err != nil {
		t.Fatalf("handleRequestSocialAction() error = %v", err)
	}
	// StoreSellManage is NOT IsInStoreMode() — should succeed
	if ok {
		t.Error("ok = true; want false (manage mode should not block)")
	}
	if n != 0 {
		t.Errorf("n = %d; want 0", n)
	}
}
