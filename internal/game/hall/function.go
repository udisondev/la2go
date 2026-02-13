package hall

import "time"

// FunctionType represents the type of clan hall function.
type FunctionType int32

// Function types matching Java ClanHall constants.
const (
	FuncTeleport   FunctionType = 1 // Teleport function
	FuncItemCreate FunctionType = 2 // Item creation (curtains, platforms)
	FuncRestoreHP  FunctionType = 3 // HP restoration
	FuncRestoreMP  FunctionType = 4 // MP restoration
	FuncRestoreExp FunctionType = 5 // Experience restoration
	FuncSupport    FunctionType = 6 // Support magic buffs
)

// Function represents a clan hall function upgrade.
// Each function has a type, level, lease cost, and renewal period.
type Function struct {
	Type    FunctionType
	Level   int32         // Upgrade level (determines effect strength)
	Lease   int64         // Cost per period (Adena)
	Rate    time.Duration // Renewal period
	EndTime time.Time     // When the current lease expires
}

// IsActive returns true if the function lease hasn't expired.
func (f *Function) IsActive() bool {
	return f.EndTime.After(time.Now())
}

// FunctionLevel maps a percentage to the restore/buff effect level.
// In Interlude, typical levels are: 15%, 25%, 35%, 40%, 50%, 65%.
type FunctionLevel struct {
	Level   int32 // Display level (1, 2, 3, ...)
	Percent int32 // Effect percentage
	Lease   int64 // Cost per period
}

// Default function configurations.
// HP/MP restore: 20%, 40%, 60%
// Support: level 1-8 (different buff sets)
// Teleport: level 1, 2
var (
	HPRestoreLevels = []FunctionLevel{
		{Level: 1, Percent: 20, Lease: 2_000},
		{Level: 2, Percent: 40, Lease: 6_500},
		{Level: 3, Percent: 60, Lease: 10_000},
		{Level: 4, Percent: 80, Lease: 15_000},
		{Level: 5, Percent: 100, Lease: 21_000},
		{Level: 6, Percent: 120, Lease: 30_000},
		{Level: 7, Percent: 140, Lease: 37_000},
		{Level: 8, Percent: 160, Lease: 52_000},
	}

	MPRestoreLevels = []FunctionLevel{
		{Level: 1, Percent: 20, Lease: 2_000},
		{Level: 2, Percent: 40, Lease: 6_500},
		{Level: 3, Percent: 60, Lease: 10_000},
		{Level: 4, Percent: 80, Lease: 15_000},
		{Level: 5, Percent: 100, Lease: 21_000},
		{Level: 6, Percent: 120, Lease: 30_000},
		{Level: 7, Percent: 140, Lease: 37_000},
		{Level: 8, Percent: 160, Lease: 52_000},
	}

	ExpRestoreLevels = []FunctionLevel{
		{Level: 1, Percent: 25, Lease: 4_000},
		{Level: 2, Percent: 50, Lease: 17_000},
		{Level: 3, Percent: 75, Lease: 30_000},
	}

	TeleportLevels = []FunctionLevel{
		{Level: 1, Percent: 0, Lease: 2_000},
		{Level: 2, Percent: 0, Lease: 9_000},
	}

	SupportLevels = []FunctionLevel{
		{Level: 1, Percent: 0, Lease: 2_000},
		{Level: 2, Percent: 0, Lease: 6_500},
		{Level: 3, Percent: 0, Lease: 10_000},
		{Level: 4, Percent: 0, Lease: 15_000},
		{Level: 5, Percent: 0, Lease: 21_000},
		{Level: 6, Percent: 0, Lease: 30_000},
		{Level: 7, Percent: 0, Lease: 37_000},
		{Level: 8, Percent: 0, Lease: 52_000},
	}

	ItemCreateLevels = []FunctionLevel{
		{Level: 1, Percent: 0, Lease: 2_000},
		{Level: 2, Percent: 0, Lease: 6_000},
		{Level: 3, Percent: 0, Lease: 12_000},
	}
)
