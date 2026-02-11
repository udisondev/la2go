package gameserver

import "testing"

func TestChatType_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		ct    ChatType
		valid bool
	}{
		{"General", ChatGeneral, true},
		{"Shout", ChatShout, true},
		{"Whisper", ChatWhisper, true},
		{"Party", ChatParty, true},
		{"Clan", ChatClan, true},
		{"GM", ChatGM, true},
		{"Trade", ChatTrade, true},
		{"Alliance", ChatAlliance, true},
		{"Announce", ChatAnnounce, true},
		{"HeroVoice", ChatHeroVoice, true},
		{"Invalid_-1", ChatType(-1), false},
		{"Invalid_13", ChatType(13), false},
		{"Invalid_100", ChatType(100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.IsValid(); got != tt.valid {
				t.Errorf("ChatType(%d).IsValid() = %v, want %v", tt.ct, got, tt.valid)
			}
		})
	}
}

func TestChatType_IsMVPChannel(t *testing.T) {
	tests := []struct {
		name string
		ct   ChatType
		mvp  bool
	}{
		{"General", ChatGeneral, true},
		{"Shout", ChatShout, true},
		{"Whisper", ChatWhisper, true},
		{"Trade", ChatTrade, true},
		{"Party_not_mvp", ChatParty, false},
		{"Clan_not_mvp", ChatClan, false},
		{"GM_not_mvp", ChatGM, false},
		{"HeroVoice_not_mvp", ChatHeroVoice, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.IsMVPChannel(); got != tt.mvp {
				t.Errorf("ChatType(%d).IsMVPChannel() = %v, want %v", tt.ct, got, tt.mvp)
			}
		})
	}
}
