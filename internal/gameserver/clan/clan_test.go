package clan

import (
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	c := New(1, "TestClan", 100)

	if c.ID() != 1 {
		t.Errorf("ID = %d, want 1", c.ID())
	}
	if c.Name() != "TestClan" {
		t.Errorf("Name = %q, want %q", c.Name(), "TestClan")
	}
	if c.LeaderID() != 100 {
		t.Errorf("LeaderID = %d, want 100", c.LeaderID())
	}
	if c.Level() != 0 {
		t.Errorf("Level = %d, want 0", c.Level())
	}
	if c.MemberCount() != 0 {
		t.Errorf("MemberCount = %d, want 0", c.MemberCount())
	}
}

func TestClan_SettersGetters(t *testing.T) {
	c := New(1, "Test", 100)

	c.SetLeaderID(200)
	if c.LeaderID() != 200 {
		t.Errorf("SetLeaderID: got %d, want 200", c.LeaderID())
	}

	c.SetLevel(5)
	if c.Level() != 5 {
		t.Errorf("SetLevel: got %d, want 5", c.Level())
	}

	c.SetCrestID(42)
	if c.CrestID() != 42 {
		t.Errorf("SetCrestID: got %d, want 42", c.CrestID())
	}

	c.SetLargeCrestID(43)
	if c.LargeCrestID() != 43 {
		t.Errorf("SetLargeCrestID: got %d, want 43", c.LargeCrestID())
	}

	c.SetAllyID(10)
	if c.AllyID() != 10 {
		t.Errorf("SetAllyID: got %d, want 10", c.AllyID())
	}

	c.SetAllyCrestID(11)
	if c.AllyCrestID() != 11 {
		t.Errorf("SetAllyCrestID: got %d, want 11", c.AllyCrestID())
	}

	c.SetAllyName("TestAlly")
	if c.AllyName() != "TestAlly" {
		t.Errorf("SetAllyName: got %q, want %q", c.AllyName(), "TestAlly")
	}
}

func TestClan_Reputation(t *testing.T) {
	c := New(1, "Test", 100)

	c.SetReputation(1000)
	if c.Reputation() != 1000 {
		t.Errorf("SetReputation: got %d, want 1000", c.Reputation())
	}

	result := c.AddReputation(-200)
	if result != 800 {
		t.Errorf("AddReputation(-200) = %d, want 800", result)
	}
	if c.Reputation() != 800 {
		t.Errorf("Reputation after AddReputation = %d, want 800", c.Reputation())
	}
}

func TestClan_MaxMembers(t *testing.T) {
	c := New(1, "Test", 100)

	if c.MaxMembers() != 10 {
		t.Errorf("Level 0 MaxMembers = %d, want 10", c.MaxMembers())
	}

	c.SetLevel(4)
	if c.MaxMembers() != 40 {
		t.Errorf("Level 4 MaxMembers = %d, want 40", c.MaxMembers())
	}
}

func TestClan_AddRemoveMember(t *testing.T) {
	c := New(1, "Test", 100)

	m := NewMember(1, "Player1", 40, 0, PledgeMain, 5)
	if err := c.AddMember(m); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	if c.MemberCount() != 1 {
		t.Errorf("MemberCount = %d, want 1", c.MemberCount())
	}

	// Retrieve member.
	got := c.Member(1)
	if got == nil {
		t.Fatal("Member(1) = nil, want member")
	}
	if got.Name() != "Player1" {
		t.Errorf("Member(1).Name = %q, want %q", got.Name(), "Player1")
	}

	// Remove member.
	removed := c.RemoveMember(1)
	if removed == nil {
		t.Fatal("RemoveMember(1) = nil, want member")
	}
	if c.MemberCount() != 0 {
		t.Errorf("MemberCount = %d after remove, want 0", c.MemberCount())
	}

	// Remove non-existent.
	if c.RemoveMember(999) != nil {
		t.Error("RemoveMember(999) should return nil")
	}
}

func TestClan_AddMember_Full(t *testing.T) {
	c := New(1, "Test", 100)
	// Level 0 = max 10 members.
	for i := int64(1); i <= 10; i++ {
		m := NewMember(i, "P", 1, 0, PledgeMain, 5)
		if err := c.AddMember(m); err != nil {
			t.Fatalf("AddMember(%d): %v", i, err)
		}
	}

	// 11th should fail.
	m := NewMember(11, "P11", 1, 0, PledgeMain, 5)
	if err := c.AddMember(m); err != ErrClanFull {
		t.Errorf("AddMember(11) = %v, want ErrClanFull", err)
	}
}

func TestClan_Members_Snapshot(t *testing.T) {
	c := New(1, "Test", 100)
	for i := int64(1); i <= 3; i++ {
		m := NewMember(i, "P", 1, 0, PledgeMain, 5)
		if err := c.AddMember(m); err != nil {
			t.Fatalf("AddMember: %v", err)
		}
	}

	members := c.Members()
	if len(members) != 3 {
		t.Errorf("Members() returned %d, want 3", len(members))
	}
}

func TestClan_OnlineMemberCount(t *testing.T) {
	c := New(1, "Test", 100)

	m1 := NewMember(1, "P1", 1, 0, PledgeMain, 5)
	m2 := NewMember(2, "P2", 1, 0, PledgeMain, 5)
	m3 := NewMember(3, "P3", 1, 0, PledgeMain, 5)

	m1.SetOnline(true)
	m3.SetOnline(true)

	c.AddMember(m1) //nolint:errcheck
	c.AddMember(m2) //nolint:errcheck
	c.AddMember(m3) //nolint:errcheck

	if c.OnlineMemberCount() != 2 {
		t.Errorf("OnlineMemberCount = %d, want 2", c.OnlineMemberCount())
	}
}

func TestClan_ForEachMember(t *testing.T) {
	c := New(1, "Test", 100)
	for i := int64(1); i <= 5; i++ {
		c.AddMember(NewMember(i, "P", 1, 0, PledgeMain, 5)) //nolint:errcheck
	}

	count := 0
	c.ForEachMember(func(m *Member) bool {
		count++
		return count < 3 // stop after 3
	})
	if count != 3 {
		t.Errorf("ForEachMember stopped after %d, want 3", count)
	}
}

func TestClan_SubPledge(t *testing.T) {
	c := New(1, "Test", 100)
	c.SetLevel(6)

	sp := &SubPledge{ID: PledgeRoyal1, Name: "Royal Guard", LeaderID: 200}
	if err := c.AddSubPledge(sp); err != nil {
		t.Fatalf("AddSubPledge: %v", err)
	}

	got := c.SubPledge(PledgeRoyal1)
	if got == nil {
		t.Fatal("SubPledge(Royal1) = nil")
	}
	if got.Name != "Royal Guard" {
		t.Errorf("SubPledge name = %q, want %q", got.Name, "Royal Guard")
	}

	// Wrong level for Knights.
	c.SetLevel(5)
	spK := &SubPledge{ID: PledgeKnight1, Name: "Knights", LeaderID: 300}
	if err := c.AddSubPledge(spK); err != ErrClanLevelTooLow {
		t.Errorf("AddSubPledge(Knight at level 5) = %v, want ErrClanLevelTooLow", err)
	}

	// SubPledges snapshot.
	sps := c.SubPledges()
	if len(sps) != 1 {
		t.Errorf("SubPledges() = %d items, want 1", len(sps))
	}

	c.RemoveSubPledge(PledgeRoyal1)
	if c.SubPledge(PledgeRoyal1) != nil {
		t.Error("SubPledge should be nil after remove")
	}
}

func TestClan_SubPledgeMemberCount(t *testing.T) {
	c := New(1, "Test", 100)
	c.AddMember(NewMember(1, "P1", 1, 0, PledgeMain, 5))     //nolint:errcheck
	c.AddMember(NewMember(2, "P2", 1, 0, PledgeAcademy, 5))   //nolint:errcheck
	c.AddMember(NewMember(3, "P3", 1, 0, PledgeAcademy, 5))   //nolint:errcheck

	if c.SubPledgeMemberCount(PledgeAcademy) != 2 {
		t.Errorf("SubPledgeMemberCount(Academy) = %d, want 2", c.SubPledgeMemberCount(PledgeAcademy))
	}
	if c.SubPledgeMemberCount(PledgeMain) != 1 {
		t.Errorf("SubPledgeMemberCount(Main) = %d, want 1", c.SubPledgeMemberCount(PledgeMain))
	}
}

func TestClan_Wars(t *testing.T) {
	c := New(1, "Test", 100)

	// Declare war.
	if err := c.DeclareWar(2); err != nil {
		t.Fatalf("DeclareWar(2): %v", err)
	}

	if !c.IsAtWarWith(2) {
		t.Error("IsAtWarWith(2) = false after DeclareWar")
	}

	// Duplicate declare.
	if err := c.DeclareWar(2); err != ErrAlreadyAtWar {
		t.Errorf("DeclareWar(2) duplicate = %v, want ErrAlreadyAtWar", err)
	}

	// Accept war.
	c.AcceptWar(3)
	if !c.IsUnderAttack(3) {
		t.Error("IsUnderAttack(3) = false after AcceptWar")
	}

	// War list.
	wars := c.WarList()
	if len(wars) != 1 || wars[0] != 2 {
		t.Errorf("WarList = %v, want [2]", wars)
	}

	attackers := c.AttackerList()
	if len(attackers) != 1 || attackers[0] != 3 {
		t.Errorf("AttackerList = %v, want [3]", attackers)
	}

	// End war.
	c.EndWar(2)
	if c.IsAtWarWith(2) {
		t.Error("IsAtWarWith(2) = true after EndWar")
	}
}

func TestClan_Skills(t *testing.T) {
	c := New(1, "Test", 100)

	if c.SkillLevel(100) != 0 {
		t.Errorf("SkillLevel(100) = %d, want 0", c.SkillLevel(100))
	}

	c.SetSkill(100, 3)
	if c.SkillLevel(100) != 3 {
		t.Errorf("SkillLevel(100) = %d after SetSkill, want 3", c.SkillLevel(100))
	}

	skills := c.Skills()
	if len(skills) != 1 || skills[100] != 3 {
		t.Errorf("Skills() = %v, want map[100:3]", skills)
	}
}

func TestClan_RankPrivileges(t *testing.T) {
	c := New(1, "Test", 100)

	// Default: grade 1 = all
	if c.RankPrivileges(1) != PrivAll {
		t.Errorf("RankPrivileges(1) = %d, want %d", c.RankPrivileges(1), PrivAll)
	}

	// Set custom privileges.
	c.SetRankPrivileges(3, PrivCHSetFunctions|PrivCHOpenDoor)
	if c.RankPrivileges(3) != PrivCHSetFunctions|PrivCHOpenDoor {
		t.Errorf("RankPrivileges(3) = %d, want %d", c.RankPrivileges(3), PrivCHSetFunctions|PrivCHOpenDoor)
	}

	// All rank privileges.
	all := c.AllRankPrivileges()
	if len(all) != 9 {
		t.Errorf("AllRankPrivileges has %d grades, want 9", len(all))
	}
}

func TestClan_Notice(t *testing.T) {
	c := New(1, "Test", 100)

	c.SetNotice("Hello clan!")
	if c.Notice() != "Hello clan!" {
		t.Errorf("Notice = %q, want %q", c.Notice(), "Hello clan!")
	}

	c.SetNoticeEnabled(true)
	if !c.NoticeEnabled() {
		t.Error("NoticeEnabled = false after SetNoticeEnabled(true)")
	}

	c.SetIntroductionMessage("Welcome")
	if c.IntroductionMessage() != "Welcome" {
		t.Errorf("IntroductionMessage = %q, want %q", c.IntroductionMessage(), "Welcome")
	}
}

func TestClan_Dissolution(t *testing.T) {
	c := New(1, "Test", 100)

	if c.IsDissolving() {
		t.Error("new clan should not be dissolving")
	}

	c.SetDissolutionTime(123456)
	if !c.IsDissolving() {
		t.Error("IsDissolving = false after SetDissolutionTime")
	}
	if c.DissolutionTime() != 123456 {
		t.Errorf("DissolutionTime = %d, want 123456", c.DissolutionTime())
	}

	c.SetDissolutionTime(0)
	if c.IsDissolving() {
		t.Error("IsDissolving = true after SetDissolutionTime(0)")
	}
}

func TestClan_Leader(t *testing.T) {
	c := New(1, "Test", 100)

	// No leader member yet.
	if c.Leader() != nil {
		t.Error("Leader() should be nil when leader not added as member")
	}

	m := NewMember(100, "Leader", 80, 0, PledgeMain, 1)
	c.AddMember(m) //nolint:errcheck

	leader := c.Leader()
	if leader == nil {
		t.Fatal("Leader() = nil after adding leader member")
	}
	if leader.Name() != "Leader" {
		t.Errorf("Leader().Name = %q, want %q", leader.Name(), "Leader")
	}
}

func TestClan_MemberByName(t *testing.T) {
	c := New(1, "Test", 100)
	c.SetLevel(4)

	m1 := NewMember(100, "Alice", 40, 5, PledgeMain, 5)
	m2 := NewMember(200, "Bob", 50, 10, PledgeMain, 5)
	if err := c.AddMember(m1); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	if err := c.AddMember(m2); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	// Exact name
	found := c.MemberByName("Alice")
	if found == nil {
		t.Fatal("MemberByName(Alice) = nil, want non-nil")
	}
	if found.PlayerID() != 100 {
		t.Errorf("MemberByName(Alice).PlayerID() = %d, want 100", found.PlayerID())
	}

	// Case-insensitive
	found = c.MemberByName("BOB")
	if found == nil {
		t.Fatal("MemberByName(BOB) = nil, want non-nil")
	}
	if found.PlayerID() != 200 {
		t.Errorf("MemberByName(BOB).PlayerID() = %d, want 200", found.PlayerID())
	}

	// Not found
	found = c.MemberByName("Charlie")
	if found != nil {
		t.Errorf("MemberByName(Charlie) = %v, want nil", found)
	}
}

func TestClan_ConcurrentAccess(t *testing.T) {
	c := New(1, "Test", 100)
	c.SetLevel(8)

	var wg sync.WaitGroup
	// Concurrent member add/remove.
	for i := int64(1); i <= 20; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			m := NewMember(id, "P", 1, 0, PledgeMain, 5)
			c.AddMember(m) //nolint:errcheck
			_ = c.MemberCount()
			_ = c.Members()
			c.RemoveMember(id)
		}(i)
	}

	// Concurrent war operations.
	for i := int32(100); i <= 110; i++ {
		wg.Add(1)
		go func(clanID int32) {
			defer wg.Done()
			c.DeclareWar(clanID) //nolint:errcheck
			_ = c.IsAtWarWith(clanID)
			_ = c.WarList()
			c.EndWar(clanID)
		}(i)
	}

	// Concurrent reputation.
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.AddReputation(10)
			c.AddReputation(-5)
			_ = c.Reputation()
		}()
	}

	wg.Wait()
	// No race detected = pass.
}
