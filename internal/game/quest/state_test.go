package quest

import (
	"testing"
)

func TestNewQuestState(t *testing.T) {
	qs := NewQuestState(303, "Q00303_CollectArrowheads", 1001, StateCreated)

	if qs.QuestID() != 303 {
		t.Errorf("QuestID() = %d, want 303", qs.QuestID())
	}
	if qs.QuestName() != "Q00303_CollectArrowheads" {
		t.Errorf("QuestName() = %q, want Q00303_CollectArrowheads", qs.QuestName())
	}
	if qs.CharID() != 1001 {
		t.Errorf("CharID() = %d, want 1001", qs.CharID())
	}
	if qs.State() != StateCreated {
		t.Errorf("State() = %d, want %d (StateCreated)", qs.State(), StateCreated)
	}
	if qs.IsStarted() {
		t.Error("IsStarted() = true, want false")
	}
	if qs.IsCompleted() {
		t.Error("IsCompleted() = true, want false")
	}
}

func TestQuestState_SetState(t *testing.T) {
	qs := NewQuestState(1, "test", 100, StateCreated)

	qs.SetState(StateStarted)
	if !qs.IsStarted() {
		t.Error("after SetState(Started): IsStarted() = false, want true")
	}
	if !qs.IsChanged() {
		t.Error("after SetState: IsChanged() = false, want true")
	}

	qs.SetState(StateCompleted)
	if !qs.IsCompleted() {
		t.Error("after SetState(Completed): IsCompleted() = false, want true")
	}
}

func TestQuestState_Cond(t *testing.T) {
	qs := NewQuestState(1, "test", 100, StateStarted)

	if qs.GetCond() != 0 {
		t.Errorf("initial GetCond() = %d, want 0", qs.GetCond())
	}

	qs.SetCond(3)
	if qs.GetCond() != 3 {
		t.Errorf("after SetCond(3): GetCond() = %d, want 3", qs.GetCond())
	}

	qs.SetCond(0)
	if qs.GetCond() != 0 {
		t.Errorf("after SetCond(0): GetCond() = %d, want 0", qs.GetCond())
	}
}

func TestQuestState_Variables(t *testing.T) {
	qs := NewQuestState(1, "test", 100, StateStarted)

	qs.Set("kills", "5")
	if got := qs.Get("kills"); got != "5" {
		t.Errorf("Get(kills) = %q, want 5", got)
	}

	qs.Set("kills", "10")
	if got := qs.Get("kills"); got != "10" {
		t.Errorf("Get(kills) after update = %q, want 10", got)
	}

	qs.Unset("kills")
	if got := qs.Get("kills"); got != "" {
		t.Errorf("Get(kills) after Unset = %q, want empty", got)
	}
}

func TestQuestState_Vars_Snapshot(t *testing.T) {
	qs := NewQuestState(1, "test", 100, StateStarted)
	qs.Set("a", "1")
	qs.Set("b", "2")

	snapshot := qs.Vars()
	if len(snapshot) != 2 {
		t.Fatalf("Vars() returned %d entries, want 2", len(snapshot))
	}

	// Мутация снапшота не должна влиять на оригинал
	snapshot["c"] = "3"
	if qs.Get("c") != "" {
		t.Error("modifying snapshot affected original state")
	}
}

func TestQuestState_ClearChanged(t *testing.T) {
	qs := NewQuestState(1, "test", 100, StateCreated)
	if qs.IsChanged() {
		t.Error("newly created state should not be changed")
	}

	qs.Set("x", "1")
	if !qs.IsChanged() {
		t.Error("after Set: IsChanged() = false, want true")
	}

	qs.ClearChanged()
	if qs.IsChanged() {
		t.Error("after ClearChanged: IsChanged() = true, want false")
	}
}

func TestIntToString(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{-7, "-7"},
		{1000000, "1000000"},
	}

	for _, tc := range tests {
		got := intToString(tc.input)
		if got != tc.want {
			t.Errorf("intToString(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
