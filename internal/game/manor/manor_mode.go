package manor

// Mode represents the current state of the castle manor system.
// The cycle is: APPROVED → MAINTENANCE → MODIFIABLE → APPROVED.
type Mode int32

const (
	// ModeDisabled means manor system is turned off.
	ModeDisabled Mode = iota
	// ModeModifiable means clan leaders can modify seed/crop settings.
	ModeModifiable
	// ModeMaintenance means transition period: crops are processed, periods rotate.
	ModeMaintenance
	// ModeApproved means manor settings are locked for the current period.
	ModeApproved
)

// String returns the human-readable name of the manor mode.
func (m Mode) String() string {
	switch m {
	case ModeDisabled:
		return "DISABLED"
	case ModeModifiable:
		return "MODIFIABLE"
	case ModeMaintenance:
		return "MAINTENANCE"
	case ModeApproved:
		return "APPROVED"
	default:
		return "UNKNOWN"
	}
}
