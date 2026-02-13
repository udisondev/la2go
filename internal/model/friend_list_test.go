package model

import (
	"sort"
	"testing"
)

func TestPlayer_FriendList_Empty(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	got := p.FriendList()
	if got != nil {
		t.Errorf("FriendList() = %v; want nil", got)
	}
}

func TestPlayer_AddFriend(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	p.AddFriend(42)

	if !p.IsFriend(42) {
		t.Error("IsFriend(42) = false; want true")
	}
	if p.IsFriend(99) {
		t.Error("IsFriend(99) = true; want false")
	}

	got := p.FriendList()
	if len(got) != 1 || got[0] != 42 {
		t.Errorf("FriendList() = %v; want [42]", got)
	}
}

func TestPlayer_RemoveFriend(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	p.AddFriend(42)
	p.AddFriend(43)
	p.RemoveFriend(42)

	if p.IsFriend(42) {
		t.Error("IsFriend(42) = true after remove; want false")
	}
	if !p.IsFriend(43) {
		t.Error("IsFriend(43) = false; want true")
	}
}

func TestPlayer_SetFriendList(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	p.SetFriendList([]int32{10, 20, 30})

	got := p.FriendList()
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })

	want := []int32{10, 20, 30}
	if len(got) != len(want) {
		t.Fatalf("FriendList() len = %d; want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("FriendList()[%d] = %d; want %d", i, got[i], want[i])
		}
	}
}

func TestPlayer_BlockList_Empty(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	got := p.BlockList()
	if got != nil {
		t.Errorf("BlockList() = %v; want nil", got)
	}
}

func TestPlayer_AddBlock(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	p.AddBlock(55)

	if !p.IsBlocked(55) {
		t.Error("IsBlocked(55) = false; want true")
	}
	if p.IsBlocked(99) {
		t.Error("IsBlocked(99) = true; want false")
	}
}

func TestPlayer_RemoveBlock(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	p.AddBlock(55)
	p.AddBlock(56)
	p.RemoveBlock(55)

	if p.IsBlocked(55) {
		t.Error("IsBlocked(55) = true after remove; want false")
	}
	if !p.IsBlocked(56) {
		t.Error("IsBlocked(56) = false; want true")
	}
}

func TestPlayer_SetBlockList(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	p.SetBlockList([]int32{100, 200})

	got := p.BlockList()
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })

	want := []int32{100, 200}
	if len(got) != len(want) {
		t.Fatalf("BlockList() len = %d; want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("BlockList()[%d] = %d; want %d", i, got[i], want[i])
		}
	}
}

func TestPlayer_MessageRefusal(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	if p.MessageRefusal() {
		t.Error("MessageRefusal() = true; want false (default)")
	}

	p.SetMessageRefusal(true)
	if !p.MessageRefusal() {
		t.Error("MessageRefusal() = false after set true; want true")
	}

	p.SetMessageRefusal(false)
	if p.MessageRefusal() {
		t.Error("MessageRefusal() = true after set false; want false")
	}
}

func TestPlayer_FriendAndBlock_Independent(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Друг и блокировка — разные списки
	p.AddFriend(42)
	p.AddBlock(99)

	if p.IsFriend(99) {
		t.Error("IsFriend(99) = true (blocked, not friend); want false")
	}
	if p.IsBlocked(42) {
		t.Error("IsBlocked(42) = true (friend, not blocked); want false")
	}

	// Удаление друга не затрагивает блок-лист
	p.RemoveFriend(42)
	if !p.IsBlocked(99) {
		t.Error("IsBlocked(99) = false after RemoveFriend; want true")
	}
}
