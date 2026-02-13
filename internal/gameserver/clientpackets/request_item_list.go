package clientpackets

// OpcodeRequestItemList is the opcode for RequestItemList (C2S 0x0F).
// Java reference: ClientPackets.REQUEST_ITEM_LIST(0x0F).
const OpcodeRequestItemList = 0x0F

// RequestItemList has no body fields.
type RequestItemList struct{}

// ParseRequestItemList parses RequestItemList packet (no fields).
func ParseRequestItemList(_ []byte) (*RequestItemList, error) {
	return &RequestItemList{}, nil
}
