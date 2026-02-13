package clientpackets

// OpcodeRequestDismissAlly is the C2S opcode 0x86.
// Client sends this when the alliance leader dissolves the entire alliance.
// No data to parse -- just an opcode.
const OpcodeRequestDismissAlly byte = 0x86
