package serverpackets

const LoginFailOpcode = 0x01

// LoginFailReason codes sent to the client.
const (
	ReasonNoMessage            byte = 0x00
	ReasonSystemError          byte = 0x01
	ReasonUserOrPassWrong      byte = 0x02
	ReasonAccessFailedTryLater byte = 0x04
	ReasonAccountInfoIncorrect byte = 0x05
	ReasonNotAuthed            byte = 0x06
	ReasonAccountInUse         byte = 0x07
	ReasonServerOverloaded     byte = 0x0F
	ReasonServerMaintenance    byte = 0x10
	ReasonAccessFailed         byte = 0x15
	ReasonRestrictedIP         byte = 0x16
	ReasonDualBox              byte = 0x23
)

// LoginFail writes the LoginFail packet (opcode 0x01) into buf.
// Returns the number of bytes written.
func LoginFail(buf []byte, reason byte) int {
	buf[0] = LoginFailOpcode
	buf[1] = reason
	return 2
}
