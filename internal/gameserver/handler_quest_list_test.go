package gameserver

import (
	"context"
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/game/quest"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

func newQuestListHandler(t *testing.T, questMgr *quest.Manager) *Handler {
	t.Helper()
	cm := NewClientManager()
	return NewHandler(
		nil, cm, nil, nil, nil, nil, nil,
		nil, nil, nil, questMgr, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
}

func newQuestListPlayer(t *testing.T, objectID uint32) *model.Player {
	t.Helper()
	player, err := model.NewPlayer(objectID, int64(objectID), int64(objectID), "QuestTester", 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))
	return player
}

func TestHandleRequestQuestList_NoQuestManager(t *testing.T) {
	t.Parallel()

	h := newQuestListHandler(t, nil)
	player := newQuestListPlayer(t, 1001)
	client := &GameClient{}
	client.SetActivePlayer(player)

	buf := make([]byte, 1024)
	n, ok, err := h.handleRequestQuestList(context.Background(), client, nil, buf)
	if err != nil {
		t.Fatalf("handleRequestQuestList() error = %v", err)
	}
	if !ok {
		t.Error("ok = false; want true")
	}

	// Should return empty QuestList: opcode(1) + count(2) = 3 bytes
	if n != 3 {
		t.Fatalf("n = %d; want 3", n)
	}
	if buf[0] != serverpackets.OpcodeQuestList {
		t.Errorf("opcode = 0x%02X; want 0x%02X", buf[0], serverpackets.OpcodeQuestList)
	}
	count := int16(binary.LittleEndian.Uint16(buf[1:3]))
	if count != 0 {
		t.Errorf("quest count = %d; want 0", count)
	}
}

func TestHandleRequestQuestList_NoActivePlayer(t *testing.T) {
	t.Parallel()

	h := newQuestListHandler(t, nil)
	client := &GameClient{}

	buf := make([]byte, 1024)
	_, _, err := h.handleRequestQuestList(context.Background(), client, nil, buf)
	if err == nil {
		t.Fatal("expected error for nil player, got nil")
	}
}
