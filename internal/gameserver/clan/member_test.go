package clan

import "testing"

func TestNewMember(t *testing.T) {
	m := NewMember(100, "TestChar", 40, 5, PledgeMain, 3)

	if m.PlayerID() != 100 {
		t.Errorf("PlayerID = %d, want 100", m.PlayerID())
	}
	if m.Name() != "TestChar" {
		t.Errorf("Name = %q, want %q", m.Name(), "TestChar")
	}
	if m.Level() != 40 {
		t.Errorf("Level = %d, want 40", m.Level())
	}
	if m.ClassID() != 5 {
		t.Errorf("ClassID = %d, want 5", m.ClassID())
	}
	if m.PledgeType() != PledgeMain {
		t.Errorf("PledgeType = %d, want %d", m.PledgeType(), PledgeMain)
	}
	if m.PowerGrade() != 3 {
		t.Errorf("PowerGrade = %d, want 3", m.PowerGrade())
	}
	if m.Privileges() != DefaultRankPrivileges(3) {
		t.Errorf("Privileges = %d, want %d", m.Privileges(), DefaultRankPrivileges(3))
	}
	if m.Online() {
		t.Error("new member should not be online")
	}
}

func TestMember_SettersGetters(t *testing.T) {
	m := NewMember(1, "A", 1, 0, PledgeMain, 5)

	m.SetName("B")
	if m.Name() != "B" {
		t.Errorf("SetName: got %q, want %q", m.Name(), "B")
	}

	m.SetLevel(80)
	if m.Level() != 80 {
		t.Errorf("SetLevel: got %d, want 80", m.Level())
	}

	m.SetClassID(10)
	if m.ClassID() != 10 {
		t.Errorf("SetClassID: got %d, want 10", m.ClassID())
	}

	m.SetPledgeType(PledgeAcademy)
	if m.PledgeType() != PledgeAcademy {
		t.Errorf("SetPledgeType: got %d, want %d", m.PledgeType(), PledgeAcademy)
	}

	m.SetTitle("Warlord")
	if m.Title() != "Warlord" {
		t.Errorf("SetTitle: got %q, want %q", m.Title(), "Warlord")
	}

	m.SetOnline(true)
	if !m.Online() {
		t.Error("SetOnline(true) should make online")
	}

	m.SetSponsorID(500)
	if m.SponsorID() != 500 {
		t.Errorf("SetSponsorID: got %d, want 500", m.SponsorID())
	}

	m.SetApprentice(600)
	if m.Apprentice() != 600 {
		t.Errorf("SetApprentice: got %d, want 600", m.Apprentice())
	}
}

func TestMember_SetPowerGrade_UpdatesPrivileges(t *testing.T) {
	m := NewMember(1, "A", 1, 0, PledgeMain, 5)

	if m.Privileges() != PrivNone {
		t.Errorf("grade 5 should have PrivNone, got %d", m.Privileges())
	}

	m.SetPowerGrade(1)
	if m.Privileges() != PrivAll {
		t.Errorf("grade 1 should have PrivAll, got %d", m.Privileges())
	}
}

func TestMember_HasPrivilege(t *testing.T) {
	m := NewMember(1, "A", 1, 0, PledgeMain, 2)

	if !m.HasPrivilege(PrivCLJoinClan) {
		t.Error("grade 2 should have join privilege")
	}
	if m.HasPrivilege(PrivCSManageSiege) {
		t.Error("grade 2 should not have manage siege privilege")
	}
}

func TestMember_SetPrivileges(t *testing.T) {
	m := NewMember(1, "A", 1, 0, PledgeMain, 5)
	m.SetPrivileges(PrivCHSetFunctions | PrivCHOpenDoor)

	if !m.HasPrivilege(PrivCHSetFunctions) {
		t.Error("should have CHEnter after SetPrivileges")
	}
	if !m.HasPrivilege(PrivCHOpenDoor) {
		t.Error("should have CHOpenDoor after SetPrivileges")
	}
	if m.HasPrivilege(PrivCLJoinClan) {
		t.Error("should not have CLJoinClan after SetPrivileges")
	}
}
