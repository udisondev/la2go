package admin

import "testing"

func TestGetAccessLevel_KnownLevels(t *testing.T) {
	tests := []struct {
		level    int32
		wantName string
		wantGM   bool
	}{
		{0, "User", false},
		{1, "Moderator", true},
		{2, "Game Master", true},
		{100, "Administrator", true},
	}
	for _, tt := range tests {
		al := GetAccessLevel(tt.level)
		if al == nil {
			t.Fatalf("GetAccessLevel(%d) = nil, want %q", tt.level, tt.wantName)
		}
		if al.Name != tt.wantName {
			t.Errorf("GetAccessLevel(%d).Name = %q, want %q", tt.level, al.Name, tt.wantName)
		}
		if al.IsGM != tt.wantGM {
			t.Errorf("GetAccessLevel(%d).IsGM = %v, want %v", tt.level, al.IsGM, tt.wantGM)
		}
	}
}

func TestGetAccessLevel_NegativeIsBanned(t *testing.T) {
	al := GetAccessLevel(-1)
	if al != nil {
		t.Errorf("GetAccessLevel(-1) = %+v, want nil (banned)", al)
	}

	al = GetAccessLevel(-100)
	if al != nil {
		t.Errorf("GetAccessLevel(-100) = %+v, want nil (banned)", al)
	}
}

func TestGetAccessLevel_UnknownPositiveFallback(t *testing.T) {
	// Level 5 should fall back to level 2 (Game Master) — highest known <= 5
	al := GetAccessLevel(5)
	if al == nil {
		t.Fatalf("GetAccessLevel(5) = nil, want fallback to Game Master")
	}
	if al.Name != "Game Master" {
		t.Errorf("GetAccessLevel(5).Name = %q, want %q", al.Name, "Game Master")
	}

	// Level 50 should fall back to level 2
	al = GetAccessLevel(50)
	if al == nil {
		t.Fatalf("GetAccessLevel(50) = nil, want fallback to Game Master")
	}
	if al.Name != "Game Master" {
		t.Errorf("GetAccessLevel(50).Name = %q, want %q", al.Name, "Game Master")
	}

	// Level 200 should fall back to level 100 (Administrator)
	al = GetAccessLevel(200)
	if al == nil {
		t.Fatalf("GetAccessLevel(200) = nil, want fallback to Administrator")
	}
	if al.Name != "Administrator" {
		t.Errorf("GetAccessLevel(200).Name = %q, want %q", al.Name, "Administrator")
	}
}

func TestAccessLevel_Permissions(t *testing.T) {
	// User cannot use admin commands
	user := GetAccessLevel(0)
	if user.CanUseAdminCommands {
		t.Error("User (level 0) should not be able to use admin commands")
	}
	if user.CanBan {
		t.Error("User (level 0) should not be able to ban")
	}

	// Moderator can use admin commands and ban
	mod := GetAccessLevel(1)
	if !mod.CanUseAdminCommands {
		t.Error("Moderator (level 1) should be able to use admin commands")
	}
	if !mod.CanBan {
		t.Error("Moderator (level 1) should be able to ban")
	}

	// Admin can do everything
	adm := GetAccessLevel(100)
	if !adm.AllowPeaceAttack {
		t.Error("Administrator (level 100) should be able to peace attack")
	}
}
