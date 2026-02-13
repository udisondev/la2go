package clientpackets

// Pet/Summon client packet opcodes.
// Phase 19: Pets/Summons System.
const (
	OpcodeRequestChangePetName   = 0x89
	OpcodeRequestPetUseItem      = 0x8A
	OpcodeRequestGiveItemToPet   = 0x8B
	OpcodeRequestGetItemFromPet  = 0x8C
	OpcodeRequestPetGetItem      = 0x8F
	OpcodeRequestActionUse       = 0x45
)

// Pet command action IDs used in RequestActionUse (0x45).
// Phase 19: Pets/Summons System.
const (
	ActionPetFollow     = 15 // Pet: follow owner
	ActionPetAttack     = 16 // Pet: attack target
	ActionPetStop       = 17 // Pet: stop
	ActionPetUnsummon   = 19 // Pet: unsummon
	ActionSrvFollow     = 21 // Servitor: follow owner
	ActionSrvAttack     = 22 // Servitor: attack target
	ActionSrvStop       = 23 // Servitor: stop
	ActionSrvUnsummon   = 38 // Servitor: unsummon
)
