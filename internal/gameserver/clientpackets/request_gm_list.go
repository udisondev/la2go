package clientpackets

// OpcodeRequestGmList is the C2S opcode 0x81.
// Client requests list of online GMs (/gmlist command).
// No data to parse — just an opcode.
const OpcodeRequestGmList byte = 0x81
