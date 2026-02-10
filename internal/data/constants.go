package data

// Equipment slot constants.
// Соответствуют L2J Mobius Inventory.PAPERDOLL_* constants.
//
// Phase 5.4: Character Templates & Stats System.
// Java reference: Inventory.java (PAPERDOLL_CHEST = 5, etc.)
const (
	SlotUnderwear   uint8 = 0
	SlotHead        uint8 = 1
	SlotRightHand   uint8 = 2
	SlotLeftHand    uint8 = 3
	SlotGloves      uint8 = 4
	SlotChest       uint8 = 5
	SlotLegs        uint8 = 6
	SlotFeet        uint8 = 7
	SlotCloak       uint8 = 8
	SlotNeck        uint8 = 9
	SlotRightEar    uint8 = 10
	SlotLeftEar     uint8 = 11
	SlotRightFinger uint8 = 12
	SlotLeftFinger  uint8 = 13
)

// SlotNames — человекочитаемые названия слотов (для отладки).
var SlotNames = map[uint8]string{
	SlotUnderwear:   "Underwear",
	SlotHead:        "Head",
	SlotRightHand:   "Right Hand",
	SlotLeftHand:    "Left Hand",
	SlotGloves:      "Gloves",
	SlotChest:       "Chest",
	SlotLegs:        "Legs",
	SlotFeet:        "Feet",
	SlotCloak:       "Cloak",
	SlotNeck:        "Neck",
	SlotRightEar:    "Right Ear",
	SlotLeftEar:     "Left Ear",
	SlotRightFinger: "Right Finger",
	SlotLeftFinger:  "Left Finger",
}
