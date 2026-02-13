package serverpackets

import (
	"encoding/binary"
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func TestFriendListPacket_Empty(t *testing.T) {
	pkt := NewFriendListPacket(nil)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if len(data) < 5 {
		t.Fatalf("packet too short: %d bytes", len(data))
	}

	if data[0] != OpcodeFriendList {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeFriendList)
	}

	count := int32(binary.LittleEndian.Uint32(data[1:5]))
	if count != 0 {
		t.Errorf("friend count = %d; want 0", count)
	}
}

func TestFriendListPacket_WithFriends(t *testing.T) {
	friends := []FriendInfo{
		{ObjectID: 42, Name: "Alice", IsOnline: true},
		{ObjectID: 99, Name: "Bob", IsOnline: false},
	}
	pkt := NewFriendListPacket(friends)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != OpcodeFriendList {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeFriendList)
	}

	count := int32(binary.LittleEndian.Uint32(data[1:5]))
	if count != 2 {
		t.Errorf("friend count = %d; want 2", count)
	}

	// Проверяем что данные парсятся обратно
	r := packet.NewReader(data[5:])

	// Friend 1: Alice (online)
	objID1, err := r.ReadInt()
	if err != nil {
		t.Fatalf("read objID1: %v", err)
	}
	if objID1 != 42 {
		t.Errorf("friend1 objectID = %d; want 42", objID1)
	}

	name1, err := r.ReadString()
	if err != nil {
		t.Fatalf("read name1: %v", err)
	}
	if name1 != "Alice" {
		t.Errorf("friend1 name = %q; want %q", name1, "Alice")
	}

	online1, err := r.ReadInt()
	if err != nil {
		t.Fatalf("read online1: %v", err)
	}
	if online1 != 1 {
		t.Errorf("friend1 online = %d; want 1", online1)
	}

	objIDOnline1, err := r.ReadInt()
	if err != nil {
		t.Fatalf("read objIDOnline1: %v", err)
	}
	if objIDOnline1 != 42 {
		t.Errorf("friend1 online objectID = %d; want 42", objIDOnline1)
	}

	// Friend 2: Bob (offline)
	objID2, err := r.ReadInt()
	if err != nil {
		t.Fatalf("read objID2: %v", err)
	}
	if objID2 != 99 {
		t.Errorf("friend2 objectID = %d; want 99", objID2)
	}

	name2, err := r.ReadString()
	if err != nil {
		t.Fatalf("read name2: %v", err)
	}
	if name2 != "Bob" {
		t.Errorf("friend2 name = %q; want %q", name2, "Bob")
	}

	online2, err := r.ReadInt()
	if err != nil {
		t.Fatalf("read online2: %v", err)
	}
	if online2 != 0 {
		t.Errorf("friend2 online = %d; want 0", online2)
	}

	objIDOnline2, err := r.ReadInt()
	if err != nil {
		t.Fatalf("read objIDOnline2: %v", err)
	}
	if objIDOnline2 != 0 {
		t.Errorf("friend2 offline objectID = %d; want 0", objIDOnline2)
	}
}

func TestL2FriendPacket_Add(t *testing.T) {
	pkt := NewL2FriendPacket(FriendActionAdd, "Alice", true, 42)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != OpcodeL2Friend {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeL2Friend)
	}

	r := packet.NewReader(data[1:])

	action, err := r.ReadInt()
	if err != nil {
		t.Fatalf("read action: %v", err)
	}
	if action != FriendActionAdd {
		t.Errorf("action = %d; want %d", action, FriendActionAdd)
	}

	// unknown
	_, err = r.ReadInt()
	if err != nil {
		t.Fatalf("read unknown: %v", err)
	}

	name, err := r.ReadString()
	if err != nil {
		t.Fatalf("read name: %v", err)
	}
	if name != "Alice" {
		t.Errorf("name = %q; want %q", name, "Alice")
	}

	online, err := r.ReadInt()
	if err != nil {
		t.Fatalf("read online: %v", err)
	}
	if online != 1 {
		t.Errorf("online = %d; want 1", online)
	}

	objID, err := r.ReadInt()
	if err != nil {
		t.Fatalf("read objectID: %v", err)
	}
	if objID != 42 {
		t.Errorf("objectID = %d; want 42", objID)
	}
}

func TestL2FriendPacket_Remove(t *testing.T) {
	pkt := NewL2FriendPacket(FriendActionRemove, "Bob", false, 99)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	r := packet.NewReader(data[1:])

	action, err := r.ReadInt()
	if err != nil {
		t.Fatalf("read action: %v", err)
	}
	if action != FriendActionRemove {
		t.Errorf("action = %d; want %d", action, FriendActionRemove)
	}
}

func TestFriendAddRequestPacket(t *testing.T) {
	pkt := NewFriendAddRequest("Charlie")
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if data[0] != OpcodeFriendAddRequest {
		t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeFriendAddRequest)
	}

	r := packet.NewReader(data[1:])
	name, err := r.ReadString()
	if err != nil {
		t.Fatalf("read name: %v", err)
	}
	if name != "Charlie" {
		t.Errorf("name = %q; want %q", name, "Charlie")
	}

	unknown, err := r.ReadInt()
	if err != nil {
		t.Fatalf("read unknown: %v", err)
	}
	if unknown != 0 {
		t.Errorf("unknown = %d; want 0", unknown)
	}
}
