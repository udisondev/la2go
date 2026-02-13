package gameserver

import (
	"context"
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/login"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/testutil"
)

// mockFriendStore реализует FriendStore для тестов (no-op).
type mockFriendStore struct{}

func (m *mockFriendStore) InsertFriend(_ context.Context, _ int64, _ int32) error { return nil }
func (m *mockFriendStore) DeleteFriend(_ context.Context, _ int64, _ int32) error { return nil }
func (m *mockFriendStore) InsertBlock(_ context.Context, _ int64, _ int32) error  { return nil }
func (m *mockFriendStore) DeleteBlock(_ context.Context, _ int64, _ int32) error  { return nil }

// setupFriendMsgTest создаёт Handler, ClientManager и двух игроков для тестирования friend PM.
// Возвращает handler, senderClient, targetClient, sender, target.
func setupFriendMsgTest(t *testing.T) (*Handler, *GameClient, *GameClient, *model.Player, *model.Player) {
	t.Helper()

	sessionManager := login.NewSessionManager()
	clientManager := NewClientManager()
	handler := NewHandler(
		sessionManager, clientManager,
		&mockCharacterRepository{}, &mockPlayerPersister{},
		nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		&mockFriendStore{},
	)

	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	// Sender
	senderConn := testutil.NewMockConn()
	senderClient, err := NewGameClient(senderConn, key, nil, 0, 0)
	if err != nil {
		t.Fatalf("creating sender client: %v", err)
	}
	senderClient.SetState(ClientStateInGame)
	sender, err := model.NewPlayer(1001, 100, 1, "Sender", 1, 0, 0)
	if err != nil {
		t.Fatalf("creating sender player: %v", err)
	}
	senderClient.SetActivePlayer(sender)
	clientManager.RegisterPlayer(sender, senderClient)

	// Target
	targetConn := testutil.NewMockConn()
	targetClient, err := NewGameClient(targetConn, key, nil, 0, 0)
	if err != nil {
		t.Fatalf("creating target client: %v", err)
	}
	targetClient.SetState(ClientStateInGame)
	target, err := model.NewPlayer(1002, 200, 2, "Target", 1, 0, 0)
	if err != nil {
		t.Fatalf("creating target player: %v", err)
	}
	targetClient.SetActivePlayer(target)
	clientManager.RegisterPlayer(target, targetClient)

	return handler, senderClient, targetClient, sender, target
}

// buildFriendMsgPacket собирает данные пакета RequestSendFriendMsg (без opcode).
func buildFriendMsgPacket(message, receiver string) []byte {
	w := packet.NewWriter(len(message)*2 + len(receiver)*2 + 8)
	w.WriteString(message)
	w.WriteString(receiver)
	return w.Bytes()
}

// drainSendCh вычитывает все пакеты из sendCh клиента.
func drainSendCh(client *GameClient) [][]byte {
	var packets [][]byte
	for {
		select {
		case pkt := <-client.sendCh:
			packets = append(packets, pkt)
		default:
			return packets
		}
	}
}

// findSystemMessage ищет SystemMessage пакет с заданным messageID в отправленных пакетах.
func findSystemMessage(packets [][]byte, messageID int32) bool {
	for _, pkt := range packets {
		if len(pkt) < 5 {
			continue
		}
		if pkt[0] == serverpackets.OpcodeSystemMessage {
			gotID := int32(binary.LittleEndian.Uint32(pkt[1:5]))
			if gotID == messageID {
				return true
			}
		}
	}
	return false
}

func TestHandleRequestSendFriendMsg_Success(t *testing.T) {
	handler, senderClient, targetClient, sender, target := setupFriendMsgTest(t)

	// Делаем друзьями
	sender.AddFriend(int32(target.ObjectID()))
	target.AddFriend(int32(sender.ObjectID()))

	data := buildFriendMsgPacket("Hello friend!", "Target")
	buf := make([]byte, 4096)

	_, _, err := handler.handleRequestSendFriendMsg(context.Background(), senderClient, data, buf)
	if err != nil {
		t.Fatalf("handleRequestSendFriendMsg() error: %v", err)
	}

	// Проверяем, что target получил L2FriendSay
	packets := drainSendCh(targetClient)
	if len(packets) == 0 {
		t.Error("target received no packets; want L2FriendSay")
	}

	found := false
	for _, pkt := range packets {
		if len(pkt) > 0 && pkt[0] == serverpackets.OpcodeL2FriendSay {
			found = true
			break
		}
	}
	if !found {
		t.Error("target did not receive L2FriendSay packet")
	}
}

func TestHandleRequestSendFriendMsg_TargetNotOnline(t *testing.T) {
	handler, senderClient, _, sender, _ := setupFriendMsgTest(t)

	// Отправляем сообщение игроку, которого нет онлайн
	sender.AddFriend(999) // фейковый друг
	data := buildFriendMsgPacket("Hello!", "NonExistent")
	buf := make([]byte, 4096)

	_, _, err := handler.handleRequestSendFriendMsg(context.Background(), senderClient, data, buf)
	if err != nil {
		t.Fatalf("handleRequestSendFriendMsg() error: %v", err)
	}

	// Проверяем, что sender получил SystemMessage(145) — "That player is not online."
	packets := drainSendCh(senderClient)
	if !findSystemMessage(packets, sysMsgTargetPlayerNotFound) {
		t.Errorf("sender did not receive SystemMessage(%d); want 'That player is not online.'", sysMsgTargetPlayerNotFound)
	}
}

func TestHandleRequestSendFriendMsg_TargetBlocked(t *testing.T) {
	handler, senderClient, targetClient, sender, target := setupFriendMsgTest(t)

	// Target заблокировал sender
	target.AddBlock(int32(sender.ObjectID()))
	// Но они друзья (блок имеет приоритет над доставкой)
	sender.AddFriend(int32(target.ObjectID()))
	target.AddFriend(int32(sender.ObjectID()))

	data := buildFriendMsgPacket("Hello!", "Target")
	buf := make([]byte, 4096)

	_, _, err := handler.handleRequestSendFriendMsg(context.Background(), senderClient, data, buf)
	if err != nil {
		t.Fatalf("handleRequestSendFriendMsg() error: %v", err)
	}

	// Проверяем, что sender получил SystemMessage(176) — "That person is in message refusal mode."
	senderPackets := drainSendCh(senderClient)
	if !findSystemMessage(senderPackets, sysMsgTargetInRefusalMode) {
		t.Errorf("sender did not receive SystemMessage(%d); want 'That person is in message refusal mode.'", sysMsgTargetInRefusalMode)
	}

	// Проверяем, что target НЕ получил L2FriendSay
	targetPackets := drainSendCh(targetClient)
	for _, pkt := range targetPackets {
		if len(pkt) > 0 && pkt[0] == serverpackets.OpcodeL2FriendSay {
			t.Error("target received L2FriendSay despite blocking sender")
		}
	}
}

func TestHandleRequestSendFriendMsg_MessageRefusal(t *testing.T) {
	handler, senderClient, targetClient, sender, target := setupFriendMsgTest(t)

	// Target включил отказ от сообщений
	target.SetMessageRefusal(true)
	sender.AddFriend(int32(target.ObjectID()))
	target.AddFriend(int32(sender.ObjectID()))

	data := buildFriendMsgPacket("Hello!", "Target")
	buf := make([]byte, 4096)

	_, _, err := handler.handleRequestSendFriendMsg(context.Background(), senderClient, data, buf)
	if err != nil {
		t.Fatalf("handleRequestSendFriendMsg() error: %v", err)
	}

	// Проверяем, что sender получил SystemMessage(176)
	senderPackets := drainSendCh(senderClient)
	if !findSystemMessage(senderPackets, sysMsgTargetInRefusalMode) {
		t.Errorf("sender did not receive SystemMessage(%d); want 'That person is in message refusal mode.'", sysMsgTargetInRefusalMode)
	}

	// Проверяем, что target НЕ получил L2FriendSay
	targetPackets := drainSendCh(targetClient)
	for _, pkt := range targetPackets {
		if len(pkt) > 0 && pkt[0] == serverpackets.OpcodeL2FriendSay {
			t.Error("target received L2FriendSay despite message refusal mode")
		}
	}
}

func TestHandleRequestSendFriendMsg_NotFriends(t *testing.T) {
	handler, senderClient, targetClient, _, _ := setupFriendMsgTest(t)

	// Не добавляем в друзья
	data := buildFriendMsgPacket("Hello!", "Target")
	buf := make([]byte, 4096)

	_, _, err := handler.handleRequestSendFriendMsg(context.Background(), senderClient, data, buf)
	if err != nil {
		t.Fatalf("handleRequestSendFriendMsg() error: %v", err)
	}

	// Проверяем, что target НЕ получил L2FriendSay
	targetPackets := drainSendCh(targetClient)
	for _, pkt := range targetPackets {
		if len(pkt) > 0 && pkt[0] == serverpackets.OpcodeL2FriendSay {
			t.Error("target received L2FriendSay despite not being friends")
		}
	}
}

func TestHandleRequestSendFriendMsg_EmptyMessage(t *testing.T) {
	handler, senderClient, _, _, _ := setupFriendMsgTest(t)

	data := buildFriendMsgPacket("", "Target")
	buf := make([]byte, 4096)

	_, ok, err := handler.handleRequestSendFriendMsg(context.Background(), senderClient, data, buf)
	if err != nil {
		t.Fatalf("handleRequestSendFriendMsg() error: %v", err)
	}
	if !ok {
		t.Error("ok = false; want true (connection should stay open)")
	}
}

func TestHandleRequestSendFriendMsg_MessageTooLong(t *testing.T) {
	handler, senderClient, _, _, _ := setupFriendMsgTest(t)

	// 301 символ — превышает лимит 300
	longMsg := make([]byte, 301)
	for i := range longMsg {
		longMsg[i] = 'A'
	}
	data := buildFriendMsgPacket(string(longMsg), "Target")
	buf := make([]byte, 8192)

	_, ok, err := handler.handleRequestSendFriendMsg(context.Background(), senderClient, data, buf)
	if err != nil {
		t.Fatalf("handleRequestSendFriendMsg() error: %v", err)
	}
	if !ok {
		t.Error("ok = false; want true (connection should stay open)")
	}
}
