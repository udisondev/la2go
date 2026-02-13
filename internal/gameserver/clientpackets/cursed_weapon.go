package clientpackets

// Sub-opcodes for cursed weapon extended packets (0xD0 + sub).
// Java: ExPacket.REQUEST_CURSED_WEAPON_LIST, REQUEST_CURSED_WEAPON_LOCATION.
const (
	SubOpcodeRequestCursedWeaponList     int16 = 0x22
	SubOpcodeRequestCursedWeaponLocation int16 = 0x23
)

// RequestCursedWeaponList (C2S 0xD0:0x22) requests the list of cursed weapon IDs.
// Java: RequestCursedWeaponList.java — no body data, just the opcode.
type RequestCursedWeaponList struct{}

// ParseRequestCursedWeaponList parses the packet. No data to parse.
func ParseRequestCursedWeaponList(_ []byte) (*RequestCursedWeaponList, error) {
	return &RequestCursedWeaponList{}, nil
}

// RequestCursedWeaponLocation (C2S 0xD0:0x23) requests positions of active cursed weapons.
// Java: RequestCursedWeaponLocation.java — no body data, just the opcode.
type RequestCursedWeaponLocation struct{}

// ParseRequestCursedWeaponLocation parses the packet. No data to parse.
func ParseRequestCursedWeaponLocation(_ []byte) (*RequestCursedWeaponLocation, error) {
	return &RequestCursedWeaponLocation{}, nil
}
