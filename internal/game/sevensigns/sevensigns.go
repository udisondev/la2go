package sevensigns

// Cabal represents a faction in the Seven Signs system.
type Cabal int32

const (
	CabalNull Cabal = 0
	CabalDusk Cabal = 1
	CabalDawn Cabal = 2
)

// Seal represents one of the three seals.
type Seal int32

const (
	SealNull    Seal = 0
	SealAvarice Seal = 1
	SealGnosis  Seal = 2
	SealStrife  Seal = 3
)

// Period represents the current phase of a Seven Signs cycle.
type Period int32

const (
	PeriodRecruitment   Period = 0
	PeriodCompetition   Period = 1
	PeriodResults       Period = 2
	PeriodSealValidation Period = 3
)

// Seal stone item IDs.
const (
	ItemBlueSealStone  int32 = 6360
	ItemGreenSealStone int32 = 6361
	ItemRedSealStone   int32 = 6362
	ItemRecordOfSevenSigns int32 = 5707
	ItemAncientAdena   int32 = 5575
	ItemFestivalOffering int32 = 5901
)

// Contribution points per stone type.
const (
	BlueContribPoints  int64 = 3
	GreenContribPoints int64 = 5
	RedContribPoints   int64 = 10
)

// Record of Seven Signs cost in Ancient Adena.
const RecordCost int64 = 500

// SealCount is the number of seals in the system.
const SealCount = 3

// FestivalCount is the number of festival tiers.
const FestivalCount = 5

// PlayerData holds per-character Seven Signs participation data.
type PlayerData struct {
	CharID            int64
	Cabal             Cabal
	Seal              Seal
	RedStones         int32
	GreenStones       int32
	BlueStones        int32
	AncientAdena      int64
	ContributionScore int64
}

// ContributionFromStones calculates contribution points from current stone holdings.
func (d *PlayerData) ContributionFromStones() int64 {
	return int64(d.BlueStones)*BlueContribPoints +
		int64(d.GreenStones)*GreenContribPoints +
		int64(d.RedStones)*RedContribPoints
}

// Status holds global Seven Signs state persisted in the database.
type Status struct {
	CurrentCycle    int32
	FestivalCycle   int32
	ActivePeriod    Period
	Date            int64 // unix millis
	PreviousWinner  Cabal
	DawnStoneScore  int64
	DawnFestivalScore int32
	DuskStoneScore  int64
	DuskFestivalScore int32
	AvariceOwner    Cabal
	GnosisOwner     Cabal
	StrifeOwner     Cabal
	AvariceDawnScore int32
	GnosisDawnScore  int32
	StrifeDawnScore  int32
	AvariceDuskScore int32
	GnosisDuskScore  int32
	StrifeDuskScore  int32
	AccumulatedBonuses [FestivalCount]int32
}

// SealOwner returns the cabal that owns the given seal.
func (s *Status) SealOwner(seal Seal) Cabal {
	switch seal {
	case SealAvarice:
		return s.AvariceOwner
	case SealGnosis:
		return s.GnosisOwner
	case SealStrife:
		return s.StrifeOwner
	default:
		return CabalNull
	}
}

// SetSealOwner sets the seal owner.
func (s *Status) SetSealOwner(seal Seal, cabal Cabal) {
	switch seal {
	case SealAvarice:
		s.AvariceOwner = cabal
	case SealGnosis:
		s.GnosisOwner = cabal
	case SealStrife:
		s.StrifeOwner = cabal
	}
}

// SealScore returns the dawn and dusk scores for a given seal.
func (s *Status) SealScore(seal Seal) (dawn, dusk int32) {
	switch seal {
	case SealAvarice:
		return s.AvariceDawnScore, s.AvariceDuskScore
	case SealGnosis:
		return s.GnosisDawnScore, s.GnosisDuskScore
	case SealStrife:
		return s.StrifeDawnScore, s.StrifeDuskScore
	default:
		return 0, 0
	}
}

// CabalShortName returns the database short name for a cabal.
func CabalShortName(c Cabal) string {
	switch c {
	case CabalDawn:
		return "dawn"
	case CabalDusk:
		return "dusk"
	default:
		return ""
	}
}

// ParseCabal parses a cabal from its short name.
func ParseCabal(s string) Cabal {
	switch s {
	case "dawn":
		return CabalDawn
	case "dusk":
		return CabalDusk
	default:
		return CabalNull
	}
}

// FestivalResult holds a festival's high score for one tier and cabal.
type FestivalResult struct {
	FestivalID int32
	Cabal      Cabal
	Cycle      int32
	Date       int64
	Score      int32
	Members    string // comma-separated character names
}
