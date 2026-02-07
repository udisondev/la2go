package login

// ConnectionState represents the state machine for a login connection.
type ConnectionState int

const (
	StateConnected  ConnectionState = iota // TCP connected, Init sent
	StateAuthedGG                          // GameGuard verified
	StateAuthedLogin                       // Login/password accepted
)

func (s ConnectionState) String() string {
	switch s {
	case StateConnected:
		return "CONNECTED"
	case StateAuthedGG:
		return "AUTHED_GG"
	case StateAuthedLogin:
		return "AUTHED_LOGIN"
	default:
		return "UNKNOWN"
	}
}
