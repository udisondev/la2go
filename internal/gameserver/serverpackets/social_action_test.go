package serverpackets

import (
	"encoding/binary"
	"testing"
)

func TestSocialAction_Write(t *testing.T) {
	tests := []struct {
		name     string
		objectID int32
		actionID int32
	}{
		{"greeting", 1001, SocialActionGreeting},
		{"victory", 2002, SocialActionVictory},
		{"charm", 3003, SocialActionCharm},
		{"shyness", 4004, SocialActionShyness},
		{"dance", 5005, SocialActionDance},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := NewSocialAction(tt.objectID, tt.actionID)
			data, err := sa.Write()
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			// opcode(1) + objectID(4) + actionID(4) = 9 bytes
			if len(data) != 9 {
				t.Fatalf("Write() len = %d; want 9", len(data))
			}

			if data[0] != OpcodeSocialAction {
				t.Errorf("opcode = 0x%02X; want 0x%02X", data[0], OpcodeSocialAction)
			}

			gotObjectID := int32(binary.LittleEndian.Uint32(data[1:5]))
			if gotObjectID != tt.objectID {
				t.Errorf("objectID = %d; want %d", gotObjectID, tt.objectID)
			}

			gotActionID := int32(binary.LittleEndian.Uint32(data[5:9]))
			if gotActionID != tt.actionID {
				t.Errorf("actionID = %d; want %d", gotActionID, tt.actionID)
			}
		})
	}
}

func TestSocialAction_Constants(t *testing.T) {
	// Verify opcode
	if OpcodeSocialAction != 0x2D {
		t.Errorf("OpcodeSocialAction = 0x%02X; want 0x2D", OpcodeSocialAction)
	}

	// Verify action ID range
	if MinSocialActionID != 2 {
		t.Errorf("MinSocialActionID = %d; want 2", MinSocialActionID)
	}
	if MaxSocialActionID != 16 {
		t.Errorf("MaxSocialActionID = %d; want 16", MaxSocialActionID)
	}

	// Verify specific action IDs
	actionIDs := map[string]int32{
		"Greeting": SocialActionGreeting,
		"Victory":  SocialActionVictory,
		"Advance":  SocialActionAdvance,
		"Etc":      SocialActionEtc,
		"Yes":      SocialActionYes,
		"No":       SocialActionNo,
		"Bow":      SocialActionBow,
		"Unaware":  SocialActionUnaware,
		"Wait":     SocialActionWait,
		"Laugh":    SocialActionLaugh,
		"Applaud":  SocialActionApplaud,
		"Dance":    SocialActionDance,
		"Sorrow":   SocialActionSorrow,
		"Charm":    SocialActionCharm,
		"Shyness":  SocialActionShyness,
	}

	for name, id := range actionIDs {
		if id < MinSocialActionID || id > MaxSocialActionID {
			t.Errorf("SocialAction%s = %d; out of range [%d, %d]",
				name, id, MinSocialActionID, MaxSocialActionID)
		}
	}
}
