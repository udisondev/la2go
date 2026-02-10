package data

import "math"

// Stat bonus tables (indexed by stat value 0-100).
// Формула: 1.036^(stat - 34.845)
//
// Phase 5.4: Character Templates & Stats System.
// Java reference: BaseStat.java (line 244), statBonus.xml
var (
	// STRBonus — Physical Attack modifier
	// STR=40 → 1.20, STR=50 → 1.71, STR=60 → 2.43
	STRBonus [101]float64

	// INTBonus — Magic Attack modifier
	INTBonus [101]float64

	// DEXBonus — Accuracy, Evasion, Attack Speed modifier
	DEXBonus [101]float64

	// WITBonus — Magic Speed, Critical Rate modifier
	WITBonus [101]float64

	// CONBonus — HP, HP Regen, Stun/Poison Resist modifier
	CONBonus [101]float64

	// MENBonus — MP, MP Regen, Magic Defense modifier
	MENBonus [101]float64
)

// InitStatBonuses инициализирует таблицы бонусов от атрибутов.
// Вызывается при старте сервера (cmd/gameserver/main.go).
//
// Формула: 1.036^(stat - 34.845)
// Базовая точка: stat=35 → bonus=1.0 (нейтральный модификатор)
//
// Phase 5.4: Character Templates & Stats System.
// Java reference: BaseStat.java:244, statBonus.xml
func InitStatBonuses() {
	const base = 34.845
	const multiplier = 1.036

	for i := 0; i <= 100; i++ {
		stat := float64(i)
		bonus := math.Pow(multiplier, stat-base)

		STRBonus[i] = bonus
		INTBonus[i] = bonus
		DEXBonus[i] = bonus
		WITBonus[i] = bonus
		CONBonus[i] = bonus
		MENBonus[i] = bonus
	}
}

// GetSTRBonus возвращает STR bonus с bounds checking.
// stat вне диапазона [0..100] → возвращает граничное значение.
func GetSTRBonus(stat uint8) float64 {
	if stat > 100 {
		return STRBonus[100]
	}
	return STRBonus[stat]
}

// GetINTBonus возвращает INT bonus с bounds checking.
func GetINTBonus(stat uint8) float64 {
	if stat > 100 {
		return INTBonus[100]
	}
	return INTBonus[stat]
}

// GetDEXBonus возвращает DEX bonus с bounds checking.
func GetDEXBonus(stat uint8) float64 {
	if stat > 100 {
		return DEXBonus[100]
	}
	return DEXBonus[stat]
}

// GetWITBonus возвращает WIT bonus с bounds checking.
func GetWITBonus(stat uint8) float64 {
	if stat > 100 {
		return WITBonus[100]
	}
	return WITBonus[stat]
}

// GetCONBonus возвращает CON bonus с bounds checking.
func GetCONBonus(stat uint8) float64 {
	if stat > 100 {
		return CONBonus[100]
	}
	return CONBonus[stat]
}

// GetMENBonus возвращает MEN bonus с bounds checking.
func GetMENBonus(stat uint8) float64 {
	if stat > 100 {
		return MENBonus[100]
	}
	return MENBonus[stat]
}
