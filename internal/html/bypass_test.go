package html

import (
	"testing"
)

func TestParseNpcBypass_Basic(t *testing.T) {
	tests := []struct {
		name     string
		bypass   string
		wantObjID uint32
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "Shop",
			bypass:   "npc_12345_Shop",
			wantObjID: 12345,
			wantCmd:  "Shop",
		},
		{
			name:     "Sell",
			bypass:   "npc_99_Sell",
			wantObjID: 99,
			wantCmd:  "Sell",
		},
		{
			name:     "DepositP",
			bypass:   "npc_500_DepositP",
			wantObjID: 500,
			wantCmd:  "DepositP",
		},
		{
			name:     "WithdrawP",
			bypass:   "npc_500_WithdrawP",
			wantObjID: 500,
			wantCmd:  "WithdrawP",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := ParseNpcBypass(tc.bypass)
			if err != nil {
				t.Fatalf("ParseNpcBypass(%q): %v", tc.bypass, err)
			}
			if cmd.ObjectID != tc.wantObjID {
				t.Errorf("ObjectID = %d, want %d", cmd.ObjectID, tc.wantObjID)
			}
			if cmd.Command != tc.wantCmd {
				t.Errorf("Command = %q, want %q", cmd.Command, tc.wantCmd)
			}
			if len(cmd.Args) != 0 {
				t.Errorf("Args = %v, want empty", cmd.Args)
			}
		})
	}
}

func TestParseNpcBypass_WithArgs(t *testing.T) {
	tests := []struct {
		name     string
		bypass   string
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "Chat with page",
			bypass:   "npc_12345_Chat 1",
			wantCmd:  "Chat",
			wantArgs: []string{"1"},
		},
		{
			name:     "Multisell with ID",
			bypass:   "npc_12345_Multisell 123",
			wantCmd:  "Multisell",
			wantArgs: []string{"123"},
		},
		{
			name:     "Chat with multiple args",
			bypass:   "npc_100_Chat 3 extra",
			wantCmd:  "Chat",
			wantArgs: []string{"3", "extra"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := ParseNpcBypass(tc.bypass)
			if err != nil {
				t.Fatalf("ParseNpcBypass(%q): %v", tc.bypass, err)
			}
			if cmd.Command != tc.wantCmd {
				t.Errorf("Command = %q, want %q", cmd.Command, tc.wantCmd)
			}
			if len(cmd.Args) != len(tc.wantArgs) {
				t.Fatalf("Args length = %d, want %d (args=%v)", len(cmd.Args), len(tc.wantArgs), cmd.Args)
			}
			for i, arg := range cmd.Args {
				if arg != tc.wantArgs[i] {
					t.Errorf("Args[%d] = %q, want %q", i, arg, tc.wantArgs[i])
				}
			}
		})
	}
}

func TestParseNpcBypass_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		bypass string
	}{
		{"not npc prefix", "player_123_Action"},
		{"no command", "npc_123"},
		{"empty", ""},
		{"bad objectID", "npc_abc_Shop"},
		{"unknown command", "npc_123_FlyToMoon"},
		{"just npc_", "npc_"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseNpcBypass(tc.bypass)
			if err == nil {
				t.Fatalf("expected error for bypass %q", tc.bypass)
			}
		})
	}
}
