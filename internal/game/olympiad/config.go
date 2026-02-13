package olympiad

import "time"

// Конфигурация олимпиады. Значения из OlympiadConfig.java.
const (
	// StartPoints — начальное количество очков при регистрации в олимпиаде.
	StartPoints = 18

	// WeeklyPoints — еженедельное пополнение очков.
	WeeklyPoints = 3

	// MaxPoints — максимум очков за один бой (±10).
	MaxPoints = 10

	// MinClassedParticipants — минимум для class-based матча.
	MinClassedParticipants = 5

	// MinNonClassedParticipants — минимум для non-class-based матча.
	MinNonClassedParticipants = 9

	// BattleDuration — длительность боя.
	BattleDuration = 6 * time.Minute

	// CompPeriodDuration — длительность периода соревнований (6 часов).
	CompPeriodDuration = 6 * time.Hour

	// WeeklyPeriod — интервал еженедельного пополнения очков.
	WeeklyPeriod = 7 * 24 * time.Hour

	// ValidationPeriod — длительность периода валидации (24 часа).
	ValidationPeriod = 24 * time.Hour

	// CompStartHour — час начала соревнований (18:00).
	CompStartHour = 18

	// WaitTime — задержка перед началом боя (секунды).
	WaitTime = 120 * time.Second

	// ClassedScoreDiv — делитель для подсчёта очков class-based.
	ClassedScoreDiv = 3

	// NonClassedScoreDiv — делитель для подсчёта очков non-class-based.
	NonClassedScoreDiv = 5

	// DrawPenaltyDiv — делитель штрафа при ничьей.
	DrawPenaltyDiv = 5

	// DefaultPenaltyDiv — делитель штрафа при неявке.
	DefaultPenaltyDiv = 3

	// HeroMinMatches — минимум матчей для получения героя.
	HeroMinMatches = 9

	// HeroMinWins — минимум побед для получения героя.
	HeroMinWins = 1

	// RewardItemID — Olympiad Token (ID предмета награды).
	RewardItemID = 6651

	// ClassedRewardCount — награда за classed победу (токены).
	ClassedRewardCount = 50

	// NonClassedRewardCount — награда за non-classed победу (токены).
	NonClassedRewardCount = 30

	// MatchmakingInterval — интервал между циклами matchmaking.
	MatchmakingInterval = 30 * time.Second

	// GameStartDelay — задержка между стартами игр.
	GameStartDelay = 1 * time.Second
)

// Period определяет текущий период олимпиады.
type Period int32

const (
	// PeriodCompetition — период соревнований (основной).
	PeriodCompetition Period = 0
	// PeriodValidation — период валидации (подсчёт героев).
	PeriodValidation Period = 1
)

// String возвращает текстовое представление периода.
func (p Period) String() string {
	switch p {
	case PeriodCompetition:
		return "COMPETITION"
	case PeriodValidation:
		return "VALIDATION"
	default:
		return "UNKNOWN"
	}
}

// Rank определяет ранг в олимпиаде (top 1%, 10%, 25%, 50%, rest).
type Rank int32

const (
	Rank1 Rank = 1 // Top 1%
	Rank2 Rank = 2 // Top 10%
	Rank3 Rank = 3 // Top 25%
	Rank4 Rank = 4 // Top 50%
	Rank5 Rank = 5 // Rest
)

// HeroClassIDs — классы, участвующие в отборе героев (35 классов).
// Java: 88-118, 131-134 (3rd class transfer IDs).
var HeroClassIDs = []int32{
	88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98,
	99, 100, 101, 102, 103, 104, 105, 106, 107, 108,
	109, 110, 111, 112, 113, 114, 115, 116, 117, 118,
	131, 132, 133, 134,
}
