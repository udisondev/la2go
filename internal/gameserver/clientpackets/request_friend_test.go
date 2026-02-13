package clientpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

func writeTestString(s string) []byte {
	w := packet.NewWriter(len(s)*2 + 2)
	w.WriteString(s)
	return w.Bytes()
}

func writeTestInt(v int32) []byte {
	w := packet.NewWriter(4)
	w.WriteInt(v)
	return w.Bytes()
}

func TestParseRequestFriendInvite(t *testing.T) {
	data := writeTestString("Alice")
	pkt, err := ParseRequestFriendInvite(data)
	if err != nil {
		t.Fatalf("ParseRequestFriendInvite() error: %v", err)
	}
	if pkt.Name != "Alice" {
		t.Errorf("Name = %q; want %q", pkt.Name, "Alice")
	}
}

func TestParseRequestFriendInvite_Empty(t *testing.T) {
	_, err := ParseRequestFriendInvite(nil)
	if err == nil {
		t.Error("ParseRequestFriendInvite(nil) = nil; want error")
	}
}

func TestParseRequestAnswerFriendInvite_Accept(t *testing.T) {
	data := writeTestInt(1)
	pkt, err := ParseRequestAnswerFriendInvite(data)
	if err != nil {
		t.Fatalf("ParseRequestAnswerFriendInvite() error: %v", err)
	}
	if pkt.Response != 1 {
		t.Errorf("Response = %d; want 1", pkt.Response)
	}
}

func TestParseRequestAnswerFriendInvite_Decline(t *testing.T) {
	data := writeTestInt(0)
	pkt, err := ParseRequestAnswerFriendInvite(data)
	if err != nil {
		t.Fatalf("ParseRequestAnswerFriendInvite() error: %v", err)
	}
	if pkt.Response != 0 {
		t.Errorf("Response = %d; want 0", pkt.Response)
	}
}

func TestParseRequestFriendDel(t *testing.T) {
	data := writeTestString("Bob")
	pkt, err := ParseRequestFriendDel(data)
	if err != nil {
		t.Fatalf("ParseRequestFriendDel() error: %v", err)
	}
	if pkt.Name != "Bob" {
		t.Errorf("Name = %q; want %q", pkt.Name, "Bob")
	}
}

func TestParseRequestBlock_Block(t *testing.T) {
	w := packet.NewWriter(64)
	w.WriteInt(BlockTypeBlock)
	w.WriteString("Enemy")
	data := w.Bytes()

	pkt, err := ParseRequestBlock(data)
	if err != nil {
		t.Fatalf("ParseRequestBlock() error: %v", err)
	}
	if pkt.Type != BlockTypeBlock {
		t.Errorf("Type = %d; want %d", pkt.Type, BlockTypeBlock)
	}
	if pkt.Name != "Enemy" {
		t.Errorf("Name = %q; want %q", pkt.Name, "Enemy")
	}
}

func TestParseRequestBlock_Unblock(t *testing.T) {
	w := packet.NewWriter(64)
	w.WriteInt(BlockTypeUnblock)
	w.WriteString("Friend")
	data := w.Bytes()

	pkt, err := ParseRequestBlock(data)
	if err != nil {
		t.Fatalf("ParseRequestBlock() error: %v", err)
	}
	if pkt.Type != BlockTypeUnblock {
		t.Errorf("Type = %d; want %d", pkt.Type, BlockTypeUnblock)
	}
	if pkt.Name != "Friend" {
		t.Errorf("Name = %q; want %q", pkt.Name, "Friend")
	}
}

func TestParseRequestBlock_List(t *testing.T) {
	w := packet.NewWriter(4)
	w.WriteInt(BlockTypeList)
	data := w.Bytes()

	pkt, err := ParseRequestBlock(data)
	if err != nil {
		t.Fatalf("ParseRequestBlock() error: %v", err)
	}
	if pkt.Type != BlockTypeList {
		t.Errorf("Type = %d; want %d", pkt.Type, BlockTypeList)
	}
	if pkt.Name != "" {
		t.Errorf("Name = %q; want empty", pkt.Name)
	}
}

func TestParseRequestBlock_AllBlock(t *testing.T) {
	w := packet.NewWriter(4)
	w.WriteInt(BlockTypeAllBlock)
	data := w.Bytes()

	pkt, err := ParseRequestBlock(data)
	if err != nil {
		t.Fatalf("ParseRequestBlock() error: %v", err)
	}
	if pkt.Type != BlockTypeAllBlock {
		t.Errorf("Type = %d; want %d", pkt.Type, BlockTypeAllBlock)
	}
}

func TestParseRequestBlock_AllUnblock(t *testing.T) {
	w := packet.NewWriter(4)
	w.WriteInt(BlockTypeAllUnblock)
	data := w.Bytes()

	pkt, err := ParseRequestBlock(data)
	if err != nil {
		t.Fatalf("ParseRequestBlock() error: %v", err)
	}
	if pkt.Type != BlockTypeAllUnblock {
		t.Errorf("Type = %d; want %d", pkt.Type, BlockTypeAllUnblock)
	}
}

func TestParseRequestBlock_EmptyData(t *testing.T) {
	_, err := ParseRequestBlock(nil)
	if err == nil {
		t.Error("ParseRequestBlock(nil) = nil; want error")
	}
}

func TestOpcodes(t *testing.T) {
	tests := []struct {
		name   string
		opcode byte
		want   byte
	}{
		{"RequestFriendInvite", OpcodeRequestFriendInvite, 0x5E},
		{"RequestAnswerFriendInvite", OpcodeRequestAnswerFriendInvite, 0x5F},
		{"RequestFriendList", OpcodeRequestFriendList, 0x60},
		{"RequestFriendDel", OpcodeRequestFriendDel, 0x61},
		{"RequestBlock", OpcodeRequestBlock, 0xA0},
	}

	for _, tt := range tests {
		if tt.opcode != tt.want {
			t.Errorf("%s opcode = 0x%02X; want 0x%02X", tt.name, tt.opcode, tt.want)
		}
	}
}
