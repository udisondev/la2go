package olympiad

import (
	"sort"
	"sync"
	"time"
)

// Hero skill IDs from heroSkillTree.xml.
const (
	SkillHeroicMiracle  int32 = 395
	SkillHeroicBerserker int32 = 396
	SkillHeroicValor    int32 = 1374
	SkillHeroicGrandeur int32 = 1375
	SkillHeroicDread    int32 = 1376
)

// HeroSkillIDs — все скиллы героя.
var HeroSkillIDs = []int32{
	SkillHeroicMiracle,
	SkillHeroicBerserker,
	SkillHeroicValor,
	SkillHeroicGrandeur,
	SkillHeroicDread,
}

// Hero item IDs.
const (
	ItemWingsOfDestiny int32 = 6842  // Circlet
	ItemInfinityFirst  int32 = 6611  // Infinity Blade (first weapon)
	ItemInfinityLast   int32 = 6621  // Infinity Spear (last weapon)
)

// IsHeroItem checks if an item is a hero-exclusive item.
func IsHeroItem(itemID int32) bool {
	return itemID == ItemWingsOfDestiny ||
		(itemID >= ItemInfinityFirst && itemID <= ItemInfinityLast)
}

// DiaryAction — тип события в дневнике героя.
type DiaryAction int32

const (
	DiaryRaidKilled  DiaryAction = 1
	DiaryHeroGained  DiaryAction = 2
	DiariCastleTaken DiaryAction = 3
)

// DiaryEntry — запись в дневнике героя.
type DiaryEntry struct {
	Time   time.Time
	Action DiaryAction
	Param  int32 // npcID или castleID
}

// FightResult — результат олимпийского боя для Hero stats.
type FightResult int32

const (
	FightVictory FightResult = 1
	FightDraw    FightResult = 0
	FightLoss    FightResult = 2
)

// FightRecord — запись о бое для Hero history.
type FightRecord struct {
	OpponentCharID int64
	OpponentClassID int32
	OpponentName   string
	StartTime      time.Time
	Duration       time.Duration
	Classed        bool
	Result         FightResult
}

// HeroData хранит информацию о герое.
type HeroData struct {
	CharID  int64
	ClassID int32
	Name    string
	Count   int32 // сколько раз был героем
	Played  bool  // текущий герой
	Claimed bool  // claimed статус

	ClanName  string
	ClanCrest int32
	AllyName  string
	AllyCrest int32

	Message string // hero message
}

// HeroCandidate — кандидат в герои из олимпиады.
type HeroCandidate struct {
	CharID  int64
	ClassID int32
	Name    string
	Points  int32
	CompDone int32
	CompWon  int32
}

// HeroTable управляет героями.
// Потокобезопасна через sync.RWMutex.
type HeroTable struct {
	mu sync.RWMutex

	heroes         map[int64]*HeroData   // текущие герои (played=true)
	completeHeroes map[int64]*HeroData   // все герои за всё время
	diary          map[int64][]DiaryEntry
	fights         map[int64][]FightRecord
}

// NewHeroTable creates an empty HeroTable.
func NewHeroTable() *HeroTable {
	return &HeroTable{
		heroes:         make(map[int64]*HeroData),
		completeHeroes: make(map[int64]*HeroData),
		diary:          make(map[int64][]DiaryEntry),
		fights:         make(map[int64][]FightRecord),
	}
}

// IsHero checks if a character is a current hero.
func (ht *HeroTable) IsHero(charID int64) bool {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	h, ok := ht.heroes[charID]
	return ok && h.Played
}

// GetHero returns hero data (nil if not a hero).
func (ht *HeroTable) GetHero(charID int64) *HeroData {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	return ht.heroes[charID]
}

// AllHeroes returns a snapshot of all current heroes.
func (ht *HeroTable) AllHeroes() []*HeroData {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	result := make([]*HeroData, 0, len(ht.heroes))
	for _, h := range ht.heroes {
		result = append(result, h)
	}
	return result
}

// ComputeNewHeroes устанавливает новых героев из кандидатов.
// Сбрасывает старых героев, устанавливает новых.
func (ht *HeroTable) ComputeNewHeroes(candidates []*HeroCandidate) {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	// Фаза 1: сброс старых героев
	for _, h := range ht.heroes {
		h.Played = false
	}

	// Очистить текущих героев
	oldHeroes := ht.heroes
	ht.heroes = make(map[int64]*HeroData, len(candidates))

	// Фаза 2: установить новых героев
	for _, c := range candidates {
		if existing, ok := ht.completeHeroes[c.CharID]; ok {
			// Уже был героем ранее — инкремент
			existing.Count++
			existing.Played = true
			existing.Claimed = false
			existing.ClassID = c.ClassID
			existing.Name = c.Name
			ht.heroes[c.CharID] = existing
		} else {
			// Первый раз герой
			newHero := &HeroData{
				CharID:  c.CharID,
				ClassID: c.ClassID,
				Name:    c.Name,
				Count:   1,
				Played:  true,
				Claimed: false,
			}
			ht.heroes[c.CharID] = newHero
			ht.completeHeroes[c.CharID] = newHero
		}
	}

	// Перенести старых героев в complete (если ещё не там)
	for id, h := range oldHeroes {
		if _, ok := ht.completeHeroes[id]; !ok {
			ht.completeHeroes[id] = h
		}
	}
}

// ClaimHero отмечает героя как claimed.
func (ht *HeroTable) ClaimHero(charID int64) bool {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	h, ok := ht.heroes[charID]
	if !ok || !h.Played {
		return false
	}

	if h.Claimed {
		return false // уже claimed
	}

	h.Claimed = true
	return true
}

// SetHeroMessage устанавливает персональное сообщение героя.
func (ht *HeroTable) SetHeroMessage(charID int64, message string) bool {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	h, ok := ht.heroes[charID]
	if !ok || !h.Played {
		return false
	}

	h.Message = message
	return true
}

// HeroMessage returns hero's personal message.
func (ht *HeroTable) HeroMessage(charID int64) string {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	h, ok := ht.heroes[charID]
	if !ok {
		return ""
	}
	return h.Message
}

// AddDiaryEntry добавляет запись в дневник героя.
func (ht *HeroTable) AddDiaryEntry(charID int64, action DiaryAction, param int32) {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	entry := DiaryEntry{
		Time:   time.Now(),
		Action: action,
		Param:  param,
	}
	ht.diary[charID] = append(ht.diary[charID], entry)
}

// Diary returns hero's diary entries.
func (ht *HeroTable) Diary(charID int64) []DiaryEntry {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	entries := ht.diary[charID]
	result := make([]DiaryEntry, len(entries))
	copy(result, entries)
	return result
}

// AddFightRecord добавляет запись о бое.
func (ht *HeroTable) AddFightRecord(charID int64, record FightRecord) {
	ht.mu.Lock()
	defer ht.mu.Unlock()
	ht.fights[charID] = append(ht.fights[charID], record)
}

// Fights returns hero's fight history.
func (ht *HeroTable) Fights(charID int64) []FightRecord {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	records := ht.fights[charID]
	result := make([]FightRecord, len(records))
	copy(result, records)
	return result
}

// LoadHero loads a hero from DB data.
func (ht *HeroTable) LoadHero(data *HeroData) {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	ht.completeHeroes[data.CharID] = data
	if data.Played {
		ht.heroes[data.CharID] = data
	}
}

// SelectHeroes выбирает героев из NobleTable по классам.
// Для каждого класса из HeroClassIDs выбирается top-1 noble:
//   - минимум HeroMinMatches матчей
//   - минимум HeroMinWins побед
//   - сортировка: points DESC → compDone DESC → compWon DESC
func SelectHeroes(nobles *NobleTable) []*HeroCandidate {
	var candidates []*HeroCandidate

	// Soulhound кандидаты (классы 132/133 — 1 герой на обоих)
	var soulhounds []*HeroCandidate

	for _, classID := range HeroClassIDs {
		classNobles := nobles.ByClassID(classID)

		// Фильтр: ≥ HeroMinMatches и ≥ HeroMinWins
		var eligible []NobleStats
		for _, n := range classNobles {
			stats := n.Stats()
			if stats.CompDone >= HeroMinMatches && stats.CompWon >= HeroMinWins {
				eligible = append(eligible, stats)
			}
		}

		if len(eligible) == 0 {
			continue
		}

		// Сортировка: points DESC → compDone DESC → compWon DESC
		sort.Slice(eligible, func(i, j int) bool {
			if eligible[i].Points != eligible[j].Points {
				return eligible[i].Points > eligible[j].Points
			}
			if eligible[i].CompDone != eligible[j].CompDone {
				return eligible[i].CompDone > eligible[j].CompDone
			}
			return eligible[i].CompWon > eligible[j].CompWon
		})

		top := eligible[0]
		candidate := &HeroCandidate{
			CharID:   top.CharID,
			ClassID:  top.ClassID,
			Points:   top.Points,
			CompDone: top.CompDone,
			CompWon:  top.CompWon,
		}

		// Soulhound: male (132) / female (133) — 1 герой на два класса
		if classID == 132 || classID == 133 {
			soulhounds = append(soulhounds, candidate)
		} else {
			candidates = append(candidates, candidate)
		}
	}

	// Soulhound: выбрать лучшего из 2 кандидатов
	if len(soulhounds) == 1 {
		candidates = append(candidates, soulhounds[0])
	} else if len(soulhounds) >= 2 {
		// Сравнить points → compDone → compWon
		best := soulhounds[0]
		for _, sh := range soulhounds[1:] {
			if sh.Points > best.Points ||
				(sh.Points == best.Points && sh.CompDone > best.CompDone) ||
				(sh.Points == best.Points && sh.CompDone == best.CompDone && sh.CompWon > best.CompWon) {
				best = sh
			}
		}
		candidates = append(candidates, best)
	}

	return candidates
}

// CalculateRanks рассчитывает ранги для noble snapshot.
// Ранги: top 1%→1, top 10%→2, top 25%→3, top 50%→4, rest→5.
// Требование: ≥ HeroMinMatches матчей.
func CalculateRanks(stats []NobleStats) map[int64]Rank {
	// Фильтр: только с ≥9 матчей
	var eligible []NobleStats
	for _, s := range stats {
		if s.CompDone >= HeroMinMatches {
			eligible = append(eligible, s)
		}
	}

	if len(eligible) == 0 {
		return nil
	}

	// Сортировка: points DESC → compDone DESC → compWon DESC
	sort.Slice(eligible, func(i, j int) bool {
		if eligible[i].Points != eligible[j].Points {
			return eligible[i].Points > eligible[j].Points
		}
		if eligible[i].CompDone != eligible[j].CompDone {
			return eligible[i].CompDone > eligible[j].CompDone
		}
		return eligible[i].CompWon > eligible[j].CompWon
	})

	total := len(eligible)
	rank1 := max(1, int(float64(total)*0.01+0.5))
	rank2 := max(rank1+1, int(float64(total)*0.10+0.5))
	rank3 := max(rank2+1, int(float64(total)*0.25+0.5))
	rank4 := max(rank3+1, int(float64(total)*0.50+0.5))

	ranks := make(map[int64]Rank, total)
	for i, s := range eligible {
		place := i + 1
		switch {
		case place <= rank1:
			ranks[s.CharID] = Rank1
		case place <= rank2:
			ranks[s.CharID] = Rank2
		case place <= rank3:
			ranks[s.CharID] = Rank3
		case place <= rank4:
			ranks[s.CharID] = Rank4
		default:
			ranks[s.CharID] = Rank5
		}
	}

	return ranks
}
