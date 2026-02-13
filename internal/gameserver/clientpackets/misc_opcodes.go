package clientpackets

// Miscellaneous opcodes (Phase 50).

const (
	// OpcodeRequestGiveNickName is 0x55 — set clan member title or self title for noblesse.
	OpcodeRequestGiveNickName byte = 0x55

	// OpcodeRequestRecordInfo is 0xCF — client requests UserInfo refresh.
	OpcodeRequestRecordInfo byte = 0xCF

	// OpcodeRequestShowMiniMap is 0xCD — client opens minimap (server-side no-op).
	OpcodeRequestShowMiniMap byte = 0xCD

	// OpcodeObserverReturn is 0xB8 — exit observer mode.
	OpcodeObserverReturn byte = 0xB8

	// OpcodeRequestRecipeBookDestroy is 0xAD — delete recipe from recipe book.
	OpcodeRequestRecipeBookDestroy byte = 0xAD

	// OpcodeRequestEvaluate is 0xB9 — give recommendation to another player.
	OpcodeRequestEvaluate byte = 0xB9

	// OpcodeRequestPartyMatchConfig is 0x6F — party matching window config.
	OpcodeRequestPartyMatchConfig byte = 0x6F

	// OpcodeRequestPartyMatchList is 0x70 — party matching list.
	OpcodeRequestPartyMatchList byte = 0x70

	// OpcodeRequestPartyMatchDetail is 0x71 — party matching detail.
	OpcodeRequestPartyMatchDetail byte = 0x71
)
