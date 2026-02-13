package data

import "math"

// Stat bonus tables (indexed by stat value 0-100).
// Each stat uses its OWN formula (from statBonus.xml):
//   STR: 1.036^(stat - 34.845)
//   INT: 1.020^(stat - 31.375)
//   DEX: 1.009^(stat - 19.360)
//   WIT: 1.050^(stat - 20.000)
//   CON: 1.030^(stat - 27.632)
//   MEN: 1.010^(stat +  0.060)
//
// Phase 5.4: Character Templates & Stats System.
// Java reference: BaseStat.java, statBonus.xml
var (
	// STRBonus — Physical Attack modifier (base=1.036, offset=34.845)
	STRBonus [101]float64

	// INTBonus — Magic Attack modifier (base=1.020, offset=31.375)
	INTBonus [101]float64

	// DEXBonus — Accuracy, Evasion, Attack Speed modifier (base=1.009, offset=19.360)
	DEXBonus [101]float64

	// WITBonus — Magic Speed, Critical Rate modifier (base=1.050, offset=20.000)
	WITBonus [101]float64

	// CONBonus — HP, HP Regen, Stun/Poison Resist modifier (base=1.030, offset=27.632)
	CONBonus [101]float64

	// MENBonus — MP, MP Regen, Magic Defense modifier (base=1.010, offset=-0.060)
	MENBonus [101]float64
)

// InitStatBonuses инициализирует таблицы бонусов от атрибутов.
// Вызывается при старте сервера (cmd/gameserver/main.go).
//
// Каждый стат использует свою формулу — base^(stat - offset).
// Java reference: BaseStat.java, statBonus.xml
func InitStatBonuses() {
	for i := 0; i <= 100; i++ {
		stat := float64(i)
		STRBonus[i] = math.Pow(1.036, stat-34.845)
		INTBonus[i] = math.Pow(1.020, stat-31.375)
		DEXBonus[i] = math.Pow(1.009, stat-19.360)
		WITBonus[i] = math.Pow(1.050, stat-20.000)
		CONBonus[i] = math.Pow(1.030, stat-27.632)
		MENBonus[i] = math.Pow(1.010, stat+0.060)
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
