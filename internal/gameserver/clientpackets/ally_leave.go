package clientpackets

// OpcodeAllyLeave is the C2S opcode 0x84.
// Client sends this when a clan leader withdraws their clan from an alliance.
// No data to parse -- just an opcode.
const OpcodeAllyLeave byte = 0x84
