package clan

import "testing"

func TestMaxSubPledgeMembers(t *testing.T) {
	tests := []struct {
		pledgeType int32
		want       int32
	}{
		{PledgeAcademy, 20},
		{PledgeRoyal1, 20},
		{PledgeRoyal2, 20},
		{PledgeKnight1, 10},
		{PledgeKnight2, 10},
		{PledgeKnight3, 10},
		{PledgeKnight4, 10},
		{PledgeMain, 0},
		{999, 0},
	}
	for _, tt := range tests {
		if got := MaxSubPledgeMembers(tt.pledgeType); got != tt.want {
			t.Errorf("MaxSubPledgeMembers(%d) = %d, want %d", tt.pledgeType, got, tt.want)
		}
	}
}

func TestIsSubPledge(t *testing.T) {
	tests := []struct {
		pledgeType int32
		want       bool
	}{
		{PledgeMain, false},
		{PledgeAcademy, true},
		{PledgeRoyal1, true},
		{PledgeRoyal2, true},
		{PledgeKnight1, true},
		{PledgeKnight2, true},
		{PledgeKnight3, true},
		{PledgeKnight4, true},
		{42, false},
	}
	for _, tt := range tests {
		if got := IsSubPledge(tt.pledgeType); got != tt.want {
			t.Errorf("IsSubPledge(%d) = %v, want %v", tt.pledgeType, got, tt.want)
		}
	}
}

func TestSubPledgeRequiredLevel(t *testing.T) {
	tests := []struct {
		pledgeType int32
		want       int32
	}{
		{PledgeAcademy, 5},
		{PledgeRoyal1, 6},
		{PledgeRoyal2, 6},
		{PledgeKnight1, 7},
		{PledgeKnight2, 7},
		{PledgeKnight3, 8},
		{PledgeKnight4, 8},
		{PledgeMain, 0},
	}
	for _, tt := range tests {
		if got := SubPledgeRequiredLevel(tt.pledgeType); got != tt.want {
			t.Errorf("SubPledgeRequiredLevel(%d) = %d, want %d", tt.pledgeType, got, tt.want)
		}
	}
}
