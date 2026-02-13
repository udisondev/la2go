package clientpackets

// OpcodeRequestQuestList is the C2S opcode 0x63 — client opens quest journal.
// This packet has no body; only the opcode is sent.
//
// Java reference: RequestQuestList.java — readImpl() is empty.
const OpcodeRequestQuestList byte = 0x63
