package clientpackets

// Extended packet sub-opcode constants for stubs (not yet fully implemented).

const (
	// SubOpcodeRequestExPledgeCrestLarge is 0xD0:0x10 — request large clan crest image.
	SubOpcodeRequestExPledgeCrestLarge int16 = 0x10

	// SubOpcodeRequestExSetPledgeCrestLarge is 0xD0:0x11 — upload large clan crest.
	SubOpcodeRequestExSetPledgeCrestLarge int16 = 0x11

	// SubOpcodeRequestOlympiadObserverEnd is 0xD0:0x12 — leave olympiad observer.
	SubOpcodeRequestOlympiadObserverEnd int16 = 0x12

	// SubOpcodeRequestOlympiadMatchList is 0xD0:0x13 — request olympiad match list.
	SubOpcodeRequestOlympiadMatchList int16 = 0x13

	// SubOpcodeRequestExMPCCShowPartyMembersInfo is 0xD0:0x26 — MPCC party members info.
	SubOpcodeRequestExMPCCShowPartyMembersInfo int16 = 0x26

	// SubOpcodeRequestSetSeed (0x0A) and SubOpcodeRequestSetCrop (0x0B)
	// are declared in request_set_seed.go and request_set_crop.go respectively.

	// SubOpcodeRequestExShowManorSeedInfo is 0xD0:0x0C — display seeds for sale.
	SubOpcodeRequestExShowManorSeedInfo int16 = 0x0C

	// SubOpcodeRequestExShowCropInfo is 0xD0:0x0D — display crops for procure.
	SubOpcodeRequestExShowCropInfo int16 = 0x0D

	// SubOpcodeRequestExShowSeedSetting is 0xD0:0x0E — display seed production settings.
	SubOpcodeRequestExShowSeedSetting int16 = 0x0E

	// SubOpcodeRequestExShowCropSetting is 0xD0:0x0F — display crop procurement settings.
	SubOpcodeRequestExShowCropSetting int16 = 0x0F

	// SubOpcodeRequestOustFromPartyRoom is 0xD0:0x00 — kick from party room.
	SubOpcodeRequestOustFromPartyRoom int16 = 0x00

	// SubOpcodeRequestDismissPartyRoom is 0xD0:0x01 — dismiss party room.
	SubOpcodeRequestDismissPartyRoom int16 = 0x01

	// SubOpcodeRequestWithdrawPartyRoom is 0xD0:0x02 — leave party room.
	SubOpcodeRequestWithdrawPartyRoom int16 = 0x02

	// SubOpcodeRequestListPartyMatchingWaitingRoom is 0xD0:0x03 — party matching list.
	SubOpcodeRequestListPartyMatchingWaitingRoom int16 = 0x03

	// SubOpcodeRequestAskJoinPartyRoom is 0xD0:0x14 — invite to party room.
	SubOpcodeRequestAskJoinPartyRoom int16 = 0x14

	// SubOpcodeConfirmJoinPartyRoom is 0xD0:0x15 — accept party room invite.
	SubOpcodeConfirmJoinPartyRoom int16 = 0x15

	// SubOpcodeRequestListPartyMatching is 0xD0:0x16 — party matching search.
	SubOpcodeRequestListPartyMatching int16 = 0x16
)
