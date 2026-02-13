package clan

import "testing"

func TestPrivilege_Has(t *testing.T) {
	tests := []struct {
		name string
		mask Privilege
		priv Privilege
		want bool
	}{
		{"all has join", PrivAll, PrivCLJoinClan, true},
		{"all has enter", PrivAll, PrivCHOpenDoor, true},
		{"none has nothing", PrivNone, PrivCHOpenDoor, false},
		{"single flag", PrivCHOpenDoor, PrivCHOpenDoor, true},
		{"single flag miss", PrivCHOpenDoor, PrivCLJoinClan, false},
		{"combined has subset", PrivCHOpenDoor | PrivCLJoinClan, PrivCLJoinClan, true},
		{"subset check both", PrivCHOpenDoor | PrivCLJoinClan, PrivCHOpenDoor | PrivCLJoinClan, true},
		{"subset check miss", PrivCHOpenDoor, PrivCHOpenDoor | PrivCLJoinClan, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mask.Has(tt.priv); got != tt.want {
				t.Errorf("(%d).Has(%d) = %v, want %v", tt.mask, tt.priv, got, tt.want)
			}
		})
	}
}

func TestPrivilege_Add(t *testing.T) {
	p := PrivNone
	p = p.Add(PrivCHOpenDoor)
	if !p.Has(PrivCHOpenDoor) {
		t.Error("Add(PrivCHOpenDoor) should set CHEnter flag")
	}
	p = p.Add(PrivCLJoinClan)
	if !p.Has(PrivCHOpenDoor) || !p.Has(PrivCLJoinClan) {
		t.Error("Add should preserve existing flags")
	}
}

func TestPrivilege_Remove(t *testing.T) {
	p := PrivCHOpenDoor | PrivCLJoinClan | PrivCLDismiss
	p = p.Remove(PrivCLJoinClan)
	if p.Has(PrivCLJoinClan) {
		t.Error("Remove(PrivCLJoinClan) should clear the flag")
	}
	if !p.Has(PrivCHOpenDoor) {
		t.Error("Remove should not clear unrelated flags")
	}
	if !p.Has(PrivCLDismiss) {
		t.Error("Remove should not clear unrelated flags")
	}
}

func TestDefaultRankPrivileges(t *testing.T) {
	// Leader (grade 1) = all
	if priv := DefaultRankPrivileges(1); priv != PrivAll {
		t.Errorf("DefaultRankPrivileges(1) = %d, want %d (PrivAll)", priv, PrivAll)
	}

	// Grade 2 has join/titles/warehouse/ranks/war/dismiss/crest/use warehouse
	g2 := DefaultRankPrivileges(2)
	if !g2.Has(PrivCLJoinClan) {
		t.Error("Grade 2 should have PrivCLJoinClan")
	}
	if !g2.Has(PrivCLPledgeWar) {
		t.Error("Grade 2 should have PrivCLPledgeWar")
	}

	// Grade 5+ = none
	if priv := DefaultRankPrivileges(5); priv != PrivNone {
		t.Errorf("DefaultRankPrivileges(5) = %d, want 0", priv)
	}
}

func TestPrivAll_Has24Bits(t *testing.T) {
	// PrivAll should have exactly 24 bits set
	count := 0
	v := int32(PrivAll)
	for v != 0 {
		count += int(v & 1)
		v >>= 1
	}
	if count != 24 {
		t.Errorf("PrivAll has %d bits set, want 24", count)
	}
}
