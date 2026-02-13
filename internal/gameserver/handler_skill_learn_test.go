package gameserver

import (
	"context"
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
)

// setupTestSkillTrees initializes ClassSkillTrees for testing.
func setupTestSkillTrees(t *testing.T) {
	t.Helper()

	data.ClassSkillTrees = map[int32][]*data.SkillLearn{
		0: { // Human Fighter
			{SkillID: 56, SkillLevel: 1, MinLevel: 5, SpCost: 100, ClassID: 0},
			{SkillID: 56, SkillLevel: 2, MinLevel: 10, SpCost: 200, ClassID: 0},
			{SkillID: 56, SkillLevel: 3, MinLevel: 15, SpCost: 500, ClassID: 0},
			{SkillID: 100, SkillLevel: 1, MinLevel: 20, SpCost: 1000, ClassID: 0, Items: []data.ItemReq{
				{ItemID: 57, Count: 10},
			}},
		},
	}

	data.SpecialSkillTrees = map[string][]*data.SkillLearn{
		"fishingSkillTree": {
			{SkillID: 1315, SkillLevel: 1, MinLevel: 10, SpCost: 500},
		},
		"pledgeSkillTree": {
			{SkillID: 300, SkillLevel: 1, MinLevel: 1, SpCost: 0},
		},
	}
}

func newTestClientWithPlayer(t *testing.T, player *model.Player) *GameClient {
	t.Helper()
	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}
	client, err := NewGameClient(conn, key, nil, 0, 0)
	if err != nil {
		t.Fatalf("NewGameClient: %v", err)
	}
	client.SetActivePlayer(player)
	return client
}

// --- handleRequestAcquireSkillInfo tests ---

func TestHandleRequestAcquireSkillInfo_ClassSkill(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 10000, 20)
	client := newTestClientWithPlayer(t, player)
	buf := make([]byte, 65536)

	// Build packet data for skillID=56, level=1, type=CLASS
	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 56)
	binary.LittleEndian.PutUint32(pktData[4:], 1)
	binary.LittleEndian.PutUint32(pktData[8:], 0) // CLASS

	n, ok, err := h.handleRequestAcquireSkillInfo(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n == 0 {
		t.Fatal("expected bytes written > 0")
	}

	// Verify opcode
	if buf[0] != serverpackets.OpcodeAcquireSkillInfo {
		t.Errorf("opcode = 0x%02X; want 0x%02X", buf[0], serverpackets.OpcodeAcquireSkillInfo)
	}

	// Verify skillID in response
	skillID := int32(binary.LittleEndian.Uint32(buf[1:5]))
	if skillID != 56 {
		t.Errorf("SkillID = %d; want 56", skillID)
	}

	// Verify spCost
	spCost := int32(binary.LittleEndian.Uint32(buf[9:13]))
	if spCost != 100 {
		t.Errorf("SpCost = %d; want 100", spCost)
	}
}

func TestHandleRequestAcquireSkillInfo_SkillNotInTree(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 10001, 20)
	client := newTestClientWithPlayer(t, player)
	buf := make([]byte, 65536)

	// Non-existent skill
	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 9999)
	binary.LittleEndian.PutUint32(pktData[4:], 1)
	binary.LittleEndian.PutUint32(pktData[8:], 0)

	n, ok, err := h.handleRequestAcquireSkillInfo(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n != 0 {
		t.Errorf("expected 0 bytes written for missing skill; got %d", n)
	}
}

func TestHandleRequestAcquireSkillInfo_FishingSkill(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 10002, 20)
	client := newTestClientWithPlayer(t, player)
	buf := make([]byte, 65536)

	// Fishing skill
	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 1315)
	binary.LittleEndian.PutUint32(pktData[4:], 1)
	binary.LittleEndian.PutUint32(pktData[8:], 1) // FISHING

	n, ok, err := h.handleRequestAcquireSkillInfo(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n == 0 {
		t.Fatal("expected bytes written > 0 for fishing skill")
	}

	// Verify opcode
	if buf[0] != serverpackets.OpcodeAcquireSkillInfo {
		t.Errorf("opcode = 0x%02X; want 0x%02X", buf[0], serverpackets.OpcodeAcquireSkillInfo)
	}

	// Verify skillType=FISHING in response
	skillType := int32(binary.LittleEndian.Uint32(buf[13:17]))
	if skillType != 1 {
		t.Errorf("SkillType = %d; want 1 (FISHING)", skillType)
	}
}

func TestHandleRequestAcquireSkillInfo_InvalidParams(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 10003, 20)
	client := newTestClientWithPlayer(t, player)
	buf := make([]byte, 65536)

	// skillID=0 (invalid)
	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 0) // invalid
	binary.LittleEndian.PutUint32(pktData[4:], 1)
	binary.LittleEndian.PutUint32(pktData[8:], 0)

	n, ok, err := h.handleRequestAcquireSkillInfo(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n != 0 {
		t.Errorf("expected 0 bytes for invalid skillID; got %d", n)
	}
}

func TestHandleRequestAcquireSkillInfo_WithItemReqs(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 10004, 20)
	client := newTestClientWithPlayer(t, player)
	buf := make([]byte, 65536)

	// Skill 100 has item requirements
	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 100)
	binary.LittleEndian.PutUint32(pktData[4:], 1)
	binary.LittleEndian.PutUint32(pktData[8:], 0)

	n, ok, err := h.handleRequestAcquireSkillInfo(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n == 0 {
		t.Fatal("expected bytes > 0")
	}

	// ReqCount at offset 17
	reqCount := int32(binary.LittleEndian.Uint32(buf[17:21]))
	if reqCount != 1 {
		t.Errorf("ReqCount = %d; want 1", reqCount)
	}

	// ItemID at offset 25 (21 + type=4)
	reqItemID := int32(binary.LittleEndian.Uint32(buf[25:29]))
	if reqItemID != 57 {
		t.Errorf("Req.ItemID = %d; want 57", reqItemID)
	}
}

// --- handleRequestAcquireSkill tests ---

func TestHandleRequestAcquireSkill_Success(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 10100, 20)
	player.SetSP(500)
	client := newTestClientWithPlayer(t, player)
	buf := make([]byte, 65536)

	// Learn skillID=56, level=1
	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 56)
	binary.LittleEndian.PutUint32(pktData[4:], 1)
	binary.LittleEndian.PutUint32(pktData[8:], 0) // CLASS

	n, ok, err := h.handleRequestAcquireSkill(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n == 0 {
		t.Fatal("expected bytes written > 0")
	}

	// Verify skill was learned
	learned := player.GetSkillLevel(56)
	if learned != 1 {
		t.Errorf("GetSkillLevel(56) = %d; want 1", learned)
	}

	// Verify SP was deducted
	if player.SP() != 400 {
		t.Errorf("SP = %d; want 400 (500 - 100)", player.SP())
	}
}

func TestHandleRequestAcquireSkill_InsufficientSP(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 10101, 20)
	player.SetSP(50) // not enough (need 100)
	client := newTestClientWithPlayer(t, player)
	buf := make([]byte, 65536)

	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 56)
	binary.LittleEndian.PutUint32(pktData[4:], 1)
	binary.LittleEndian.PutUint32(pktData[8:], 0)

	n, ok, err := h.handleRequestAcquireSkill(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n != 0 {
		t.Errorf("expected 0 bytes for insufficient SP; got %d", n)
	}

	// Skill should NOT be learned
	if player.GetSkillLevel(56) != 0 {
		t.Error("skill should not be learned with insufficient SP")
	}

	// SP should not be deducted
	if player.SP() != 50 {
		t.Errorf("SP should be unchanged; got %d", player.SP())
	}
}

func TestHandleRequestAcquireSkill_LevelTooLow(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 10102, 3) // level 3, needs 5
	player.SetSP(1000)
	client := newTestClientWithPlayer(t, player)
	buf := make([]byte, 65536)

	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 56)
	binary.LittleEndian.PutUint32(pktData[4:], 1)
	binary.LittleEndian.PutUint32(pktData[8:], 0)

	n, ok, err := h.handleRequestAcquireSkill(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n != 0 {
		t.Errorf("expected 0 bytes for level too low; got %d", n)
	}
	if player.GetSkillLevel(56) != 0 {
		t.Error("skill should not be learned at low level")
	}
}

func TestHandleRequestAcquireSkill_WrongSkillLevel(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 10103, 20)
	player.SetSP(5000)
	client := newTestClientWithPlayer(t, player)
	buf := make([]byte, 65536)

	// Try to learn level 2 without knowing level 1
	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 56)
	binary.LittleEndian.PutUint32(pktData[4:], 2) // level 2 without level 1
	binary.LittleEndian.PutUint32(pktData[8:], 0)

	n, ok, err := h.handleRequestAcquireSkill(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n != 0 {
		t.Errorf("expected 0 bytes for wrong skill level; got %d", n)
	}
}

func TestHandleRequestAcquireSkill_SequentialLevels(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 10104, 20)
	player.SetSP(5000)
	client := newTestClientWithPlayer(t, player)
	buf := make([]byte, 65536)

	// Learn level 1
	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 56)
	binary.LittleEndian.PutUint32(pktData[4:], 1)
	binary.LittleEndian.PutUint32(pktData[8:], 0)

	_, _, err := h.handleRequestAcquireSkill(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("learn level 1: %v", err)
	}
	if player.GetSkillLevel(56) != 1 {
		t.Fatalf("expected skill level 1; got %d", player.GetSkillLevel(56))
	}

	// Now learn level 2
	binary.LittleEndian.PutUint32(pktData[4:], 2)
	_, _, err = h.handleRequestAcquireSkill(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("learn level 2: %v", err)
	}
	if player.GetSkillLevel(56) != 2 {
		t.Errorf("expected skill level 2; got %d", player.GetSkillLevel(56))
	}

	// SP should be deducted: 5000 - 100 - 200 = 4700
	if player.SP() != 4700 {
		t.Errorf("SP = %d; want 4700", player.SP())
	}
}

func TestHandleRequestAcquireSkill_MissingItems(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 10105, 20)
	player.SetSP(5000)
	client := newTestClientWithPlayer(t, player)
	buf := make([]byte, 65536)

	// Skill 100 requires item 57 x10, but player has no items
	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 100)
	binary.LittleEndian.PutUint32(pktData[4:], 1)
	binary.LittleEndian.PutUint32(pktData[8:], 0)

	n, ok, err := h.handleRequestAcquireSkill(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n != 0 {
		t.Errorf("expected 0 bytes for missing items; got %d", n)
	}
	if player.GetSkillLevel(100) != 0 {
		t.Error("skill should not be learned without items")
	}
	if player.SP() != 5000 {
		t.Errorf("SP should be unchanged; got %d", player.SP())
	}
}

func TestHandleRequestAcquireSkill_WithItemConsume(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 10106, 20)
	player.SetSP(5000)

	// Add required items (itemID=57, count=15)
	tmpl := &model.ItemTemplate{
		ItemID: 57,
		Name:   "Adena",
		Type:   model.ItemTypeEtcItem,
	}
	item, err := model.NewItem(20001, 57, 1, 15, tmpl) // ownerID=1, count=15
	if err != nil {
		t.Fatalf("NewItem: %v", err)
	}
	if err := player.Inventory().AddItem(item); err != nil {
		t.Fatalf("AddItem: %v", err)
	}

	client := newTestClientWithPlayer(t, player)
	buf := make([]byte, 65536)

	// Learn skill 100 (requires 10x item 57)
	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 100)
	binary.LittleEndian.PutUint32(pktData[4:], 1)
	binary.LittleEndian.PutUint32(pktData[8:], 0)

	n, ok, err := h.handleRequestAcquireSkill(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n == 0 {
		t.Fatal("expected bytes written > 0")
	}

	// Skill should be learned
	if player.GetSkillLevel(100) != 1 {
		t.Errorf("GetSkillLevel(100) = %d; want 1", player.GetSkillLevel(100))
	}

	// SP should be deducted: 5000 - 1000 = 4000
	if player.SP() != 4000 {
		t.Errorf("SP = %d; want 4000", player.SP())
	}
}

// --- handleDlgAnswer tests ---

func TestHandleDlgAnswer_Success(t *testing.T) {
	t.Parallel()

	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 10200, 20)
	client := newTestClientWithPlayer(t, player)
	buf := make([]byte, 65536)

	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 1024) // messageID
	binary.LittleEndian.PutUint32(pktData[4:], 1)    // answer = Yes
	binary.LittleEndian.PutUint32(pktData[8:], 500)  // requesterID

	n, ok, err := h.handleDlgAnswer(context.Background(), client, pktData, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	// DlgAnswer just logs and acknowledges; n=0 is expected
	if n != 0 {
		t.Errorf("expected 0 bytes written; got %d", n)
	}
}

func TestHandleDlgAnswer_NoPlayer(t *testing.T) {
	t.Parallel()

	h := newTestHandler(t)
	conn := testutil.NewMockConn()
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}
	client, err := NewGameClient(conn, key, nil, 0, 0)
	if err != nil {
		t.Fatalf("NewGameClient: %v", err)
	}
	// No active player set
	buf := make([]byte, 65536)

	pktData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pktData[0:], 1024)
	binary.LittleEndian.PutUint32(pktData[4:], 1)
	binary.LittleEndian.PutUint32(pktData[8:], 500)

	_, _, err = h.handleDlgAnswer(context.Background(), client, pktData, buf)
	if err == nil {
		t.Error("expected error for no active player")
	}
}

// --- buildAcquireSkillList / buildAcquireSkillListByType tests ---

func TestBuildAcquireSkillList_ClassSkills(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	player := newTestPlayerAtLevel(t, 10300, 20)
	player.SetSP(5000)

	pkt := buildAcquireSkillList(player)
	if pkt.SkillType != 0 {
		t.Errorf("SkillType = %d; want 0 (CLASS)", pkt.SkillType)
	}

	// Level 20: all 4 skills should be learnable (56 lvl1-3 + 100 lvl1)
	if len(pkt.Skills) != 4 {
		t.Errorf("len(Skills) = %d; want 4", len(pkt.Skills))
	}
}

func TestBuildAcquireSkillList_WithKnownSkills(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	player := newTestPlayerAtLevel(t, 10301, 20)
	// Player already knows skill 56 level 1
	player.AddSkill(56, 1, false)

	pkt := buildAcquireSkillList(player)

	// Should show: 56 lvl2, 56 lvl3 is skipped (need lvl2 first), 100 lvl1 = 2 entries
	// Actually GetLearnableSkills only shows next level: 56 lvl2 + 100 lvl1 = 2
	if len(pkt.Skills) != 2 {
		t.Errorf("len(Skills) = %d; want 2 (56 lvl2 + 100 lvl1)", len(pkt.Skills))
	}
}

func TestBuildAcquireSkillListByType_Fishing(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	player := newTestPlayerAtLevel(t, 10302, 20)

	pkt := buildAcquireSkillListByType(player, clientpackets.AcquireSkillTypeFishing)
	if pkt.SkillType != 1 {
		t.Errorf("SkillType = %d; want 1 (FISHING)", pkt.SkillType)
	}
	if len(pkt.Skills) != 1 {
		t.Errorf("len(Skills) = %d; want 1", len(pkt.Skills))
	}
}

func TestBuildAcquireSkillListByType_Pledge(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	player := newTestPlayerAtLevel(t, 10303, 20)

	pkt := buildAcquireSkillListByType(player, clientpackets.AcquireSkillTypePledge)
	if pkt.SkillType != 2 {
		t.Errorf("SkillType = %d; want 2 (PLEDGE)", pkt.SkillType)
	}
	if len(pkt.Skills) != 1 {
		t.Errorf("len(Skills) = %d; want 1", len(pkt.Skills))
	}
}

// --- resolveSkillLearn tests ---

func TestResolveSkillLearn_ClassSkill(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	sl := resolveSkillLearn(0, 56, 1, clientpackets.AcquireSkillTypeClass)
	if sl == nil {
		t.Fatal("expected non-nil SkillLearn for class skill")
	}
	if sl.SkillID != 56 || sl.SkillLevel != 1 {
		t.Errorf("SkillLearn = {%d, %d}; want {56, 1}", sl.SkillID, sl.SkillLevel)
	}
}

func TestResolveSkillLearn_FishingSkill(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	sl := resolveSkillLearn(0, 1315, 1, clientpackets.AcquireSkillTypeFishing)
	if sl == nil {
		t.Fatal("expected non-nil SkillLearn for fishing skill")
	}
	if sl.SkillID != 1315 {
		t.Errorf("SkillID = %d; want 1315", sl.SkillID)
	}
}

func TestResolveSkillLearn_PledgeSkill(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	sl := resolveSkillLearn(0, 300, 1, clientpackets.AcquireSkillTypePledge)
	if sl == nil {
		t.Fatal("expected non-nil SkillLearn for pledge skill")
	}
	if sl.SkillID != 300 {
		t.Errorf("SkillID = %d; want 300", sl.SkillID)
	}
}

func TestResolveSkillLearn_UnknownType(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	sl := resolveSkillLearn(0, 56, 1, 99)
	if sl != nil {
		t.Error("expected nil for unknown skill type")
	}
}

func TestResolveSkillLearn_NotFound(t *testing.T) {
	t.Parallel()
	setupTestSkillTrees(t)

	sl := resolveSkillLearn(0, 9999, 1, clientpackets.AcquireSkillTypeClass)
	if sl != nil {
		t.Error("expected nil for non-existent skill")
	}
}
