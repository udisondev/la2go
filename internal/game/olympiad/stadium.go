package olympiad

import "github.com/udisondev/la2go/internal/model"

// Stadium represents an Olympiad arena where matches take place.
// L2J reference: OlympiadStadium.java
type Stadium struct {
	id       int32
	location model.Location
	inUse    bool
}

// ID returns the stadium identifier (0-21).
func (s *Stadium) ID() int32 { return s.id }

// Location returns the center point of the stadium.
func (s *Stadium) Location() model.Location { return s.location }

// InUse reports whether the stadium is currently occupied by a match.
func (s *Stadium) InUse() bool { return s.inUse }

// SetInUse marks the stadium as occupied or free.
func (s *Stadium) SetInUse(v bool) { s.inUse = v }

// CompetitionType определяет тип олимпийского матча.
type CompetitionType int32

const (
	// CompClassed — матч с ограничением по классу (≥5 игроков одного класса).
	CompClassed CompetitionType = iota
	// CompNonClassed — матч без ограничения по классу (≥9 любых игроков).
	CompNonClassed
)

// String возвращает текстовое представление типа.
func (ct CompetitionType) String() string {
	switch ct {
	case CompClassed:
		return "CLASSED"
	case CompNonClassed:
		return "NON_CLASSED"
	default:
		return "UNKNOWN"
	}
}

const (
	// StadiumCount — общее количество стадионов (22).
	StadiumCount = 22

	// NonClassedStadiumStart — первый стадион для NON_CLASSED (0).
	NonClassedStadiumStart = 0
	// NonClassedStadiumEnd — последний стадион для NON_CLASSED (10, inclusive).
	NonClassedStadiumEnd = 10

	// ClassedStadiumStart — первый стадион для CLASSED (11).
	ClassedStadiumStart = 11
	// ClassedStadiumEnd — последний стадион для CLASSED (21, inclusive).
	ClassedStadiumEnd = 21
)

// stadiumCoordinates содержит координаты центра каждого из 22 стадионов.
// Источник: OlympiadManager.java
var stadiumCoordinates = [StadiumCount]model.Location{
	model.NewLocation(-20814, -21189, -3030, 0),   // 0
	model.NewLocation(-120324, -225077, -3331, 0),  // 1
	model.NewLocation(-102495, -209023, -3331, 0),  // 2
	model.NewLocation(-120156, -207378, -3331, 0),  // 3
	model.NewLocation(-87628, -225021, -3331, 0),   // 4
	model.NewLocation(-81705, -213209, -3331, 0),   // 5
	model.NewLocation(-87593, -207339, -3331, 0),   // 6
	model.NewLocation(-93709, -218304, -3331, 0),   // 7
	model.NewLocation(-77157, -218608, -3331, 0),   // 8
	model.NewLocation(-69682, -209027, -3331, 0),   // 9
	model.NewLocation(-76887, -201256, -3331, 0),   // 10
	model.NewLocation(-109985, -218701, -3331, 0),  // 11
	model.NewLocation(-126367, -218228, -3331, 0),  // 12
	model.NewLocation(-109629, -201292, -3331, 0),  // 13
	model.NewLocation(-87523, -240169, -3331, 0),   // 14
	model.NewLocation(-81748, -245950, -3331, 0),   // 15
	model.NewLocation(-77123, -251473, -3331, 0),   // 16
	model.NewLocation(-69778, -241801, -3331, 0),   // 17
	model.NewLocation(-76754, -234014, -3331, 0),   // 18
	model.NewLocation(-93742, -251032, -3331, 0),   // 19
	model.NewLocation(-87466, -257752, -3331, 0),   // 20
	model.NewLocation(-114413, -213241, -3331, 0),  // 21
}

// NewStadiums создаёт массив всех 22 стадионов.
func NewStadiums() [StadiumCount]*Stadium {
	var stadiums [StadiumCount]*Stadium
	for i := range StadiumCount {
		stadiums[i] = &Stadium{
			id:       int32(i),
			location: stadiumCoordinates[i],
		}
	}
	return stadiums
}

// TeleportOffset — смещение по оси X при телепортации на стадион.
// Игроки телепортируются в центр ±900 по X.
const TeleportOffset = 900
