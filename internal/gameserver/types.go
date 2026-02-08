package gameserver

// GSConnectionState represents the state machine for a GameServer→LoginServer connection.
type GSConnectionState int32

const (
	GSStateConnected   GSConnectionState = iota // Ожидание BlowFishKey
	GSStateBFConnected                          // Ожидание GameServerAuth (Blowfish установлен)
	GSStateAuthed                               // Полноценная работа (аутентифицирован)
)

func (s GSConnectionState) String() string {
	switch s {
	case GSStateConnected:
		return "CONNECTED"
	case GSStateBFConnected:
		return "BF_CONNECTED"
	case GSStateAuthed:
		return "AUTHED"
	default:
		return "UNKNOWN"
	}
}

// ClientConnectionState represents the state machine for a GameClient→GameServer connection.
type ClientConnectionState int

const (
	ClientStateConnected     ClientConnectionState = iota // TCP connected, KeyPacket sent
	ClientStateAuthenticated                              // AuthLogin successful, SessionKey validated
	ClientStateEntering                                   // Character selected, loading world data
	ClientStateInGame                                     // Player spawned in world
	ClientStateDisconnected                               // Connection closed
)

func (s ClientConnectionState) String() string {
	switch s {
	case ClientStateConnected:
		return "CONNECTED"
	case ClientStateAuthenticated:
		return "AUTHENTICATED"
	case ClientStateEntering:
		return "ENTERING"
	case ClientStateInGame:
		return "IN_GAME"
	case ClientStateDisconnected:
		return "DISCONNECTED"
	default:
		return "UNKNOWN"
	}
}

// ServerStatus константы
const (
	StatusAuto   = 0x00
	StatusGood   = 0x01
	StatusNormal = 0x02
	StatusFull   = 0x03
	StatusDown   = 0x04
	StatusGMOnly = 0x05
)

// ServerType константы
const (
	ServerNormal              = 0x01
	ServerRelax               = 0x02
	ServerTest                = 0x04
	ServerNoLabel             = 0x08
	ServerCreationRestricted  = 0x10
	ServerEvent               = 0x20
	ServerFree                = 0x40
)

// ServerAge константы
const (
	ServerAgeAll = 0x00
	ServerAge15  = 0x0F
	ServerAge18  = 0x12
)

// LoginServerFail reason codes
const (
	ReasonIPBanned        = 1
	ReasonIPReserved      = 2
	ReasonWrongHexID      = 3
	ReasonIDReserved      = 4
	ReasonNoFreeID        = 5
	ReasonNotAuthed       = 6
	ReasonAlreadyLoggedIn = 7
)

// ServerStatus attribute types
const (
	ServerListStatus        = 0x01
	ServerType_             = 0x02
	ServerListSquareBracket = 0x03
	MaxPlayers              = 0x04
	ServerAge               = 0x06
)
