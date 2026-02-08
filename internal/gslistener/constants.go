package gslistener

// NOTE: Packet structure constants (PacketHeaderSize, PacketChecksumSize, etc.) are now in internal/constants package
// NOTE: ProtocolRevision and RSA512ModulusSize are now in internal/constants package

// GameServer→LoginServer packet opcodes
const (
	OpcodeGSBlowFishKey        = 0x00
	OpcodeGSGameServerAuth     = 0x01
	OpcodeGSPlayerInGame       = 0x02
	OpcodeGSPlayerLogout       = 0x03
	OpcodeGSPlayerAuthRequest  = 0x05
	OpcodeGSServerStatus       = 0x06
	OpcodeGSPlayerTracert      = 0x07
	OpcodeGSReplyCharacters    = 0x08
	OpcodeGSChangePasswordResp = 0x06 // duplicate with ServerStatus in Java
)

// LoginServer→GameServer packet opcodes
const (
	OpcodeLSInitLS              = 0x00
	OpcodeLSLoginServerFail     = 0x01
	OpcodeLSAuthResponse        = 0x02
	OpcodeLSPlayerAuthResponse  = 0x03
	OpcodeLSKickPlayer          = 0x04
	OpcodeLSRequestCharacters   = 0x05
	OpcodeLSChangePasswordResp2 = 0x06 // duplicate naming in Java
)
