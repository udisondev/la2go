package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Rates holds server rate multipliers for drop, XP, SP, Adena, etc.
type Rates struct {
	// XP/SP
	XP float64 `yaml:"xp"` // XP multiplier (default 1.0)
	SP float64 `yaml:"sp"` // SP multiplier (default 1.0)

	// Drop
	DeathDropChanceMultiplier float64 `yaml:"death_drop_chance_multiplier"`
	DeathDropAmountMultiplier float64 `yaml:"death_drop_amount_multiplier"`
	QuestDropChance           float64 `yaml:"quest_drop_chance"` // Quest item drop (default 1.0)
	QuestReward               float64 `yaml:"quest_reward"`      // Quest XP/Adena reward (default 1.0)
	Adena                     float64 `yaml:"adena"`             // Adena drop multiplier (default 1.0)

	// Items
	ItemAutoDestroyTime int `yaml:"item_auto_destroy_time"` // seconds
}

// EnchantConfig holds enchant rates and limits.
type EnchantConfig struct {
	SafeEnchant    int32   `yaml:"safe_enchant"`     // Safe enchant level (default 3)
	MaxEnchant     int32   `yaml:"max_enchant"`      // Max enchant level (default 25)
	ChanceWeapon   float64 `yaml:"chance_weapon"`    // Base weapon enchant chance (default 0.66)
	ChanceArmor    float64 `yaml:"chance_armor"`     // Base armor enchant chance (default 0.66)
	ChanceJewelry  float64 `yaml:"chance_jewelry"`   // Base jewelry enchant chance (default 0.66)
	BlessedChance  float64 `yaml:"blessed_chance"`   // Blessed enchant scroll chance (default 0.66)
	DestroyOnFail  bool    `yaml:"destroy_on_fail"`  // Destroy item on enchant fail (default true)
	CrystalOnFail  bool    `yaml:"crystal_on_fail"`  // Give crystals on enchant fail (default true)
}

// PvPConfig holds PvP-related settings.
type PvPConfig struct {
	FlagDuration      int `yaml:"flag_duration"`      // PvP flag duration in seconds (default 15)
	KarmaMinKills     int `yaml:"karma_min_kills"`    // Kills before karma (default 1)
	KarmaDecayRate    int `yaml:"karma_decay_rate"`   // Karma points lost per minute (default 4)
	PvPDamageMulti    float64 `yaml:"pvp_damage_multi"`   // PvP damage multiplier (default 1.0)
	ProtectionLevel   int     `yaml:"protection_level"`   // Newbie protection max level (default 25)
}

// SiegeConfig holds siege timing and rules.
type SiegeConfig struct {
	IntervalDays int    `yaml:"interval_days"` // Days between sieges (default 14)
	StartHour    int    `yaml:"start_hour"`    // Siege start hour UTC (default 20)
	Duration     int    `yaml:"duration"`      // Siege duration in minutes (default 120)
	MaxDefenders int    `yaml:"max_defenders"` // Max defender clans (default 500)
	MaxAttackers int    `yaml:"max_attackers"` // Max attacker clans (default 500)
}

// ManorConfig holds manor system timing.
type ManorConfig struct {
	ApprovalHour      int `yaml:"approval_hour"`       // Period switch hour (default 6)
	MaintenanceMin    int `yaml:"maintenance_min"`      // Maintenance duration minutes (default 3)
	ModifiableHour    int `yaml:"modifiable_hour"`      // Switch to modifiable (default 20)
	CropMatureReward  int `yaml:"crop_mature_reward"`   // Crop reward % to castle treasury (default 90)
}

// DefaultRates returns Rates with x1 multipliers.
func DefaultRates() Rates {
	return Rates{
		XP:                       1.0,
		SP:                       1.0,
		DeathDropChanceMultiplier: 1.0,
		DeathDropAmountMultiplier: 1.0,
		QuestDropChance:           1.0,
		QuestReward:               1.0,
		Adena:                     1.0,
		ItemAutoDestroyTime:       60,
	}
}

// GameServer holds all configuration for the game server.
type GameServer struct {
	// Network
	BindAddress string `yaml:"bind_address"`
	Port        int    `yaml:"port"`

	// LoginServer connection
	LoginHost string `yaml:"login_host"`
	LoginPort int    `yaml:"login_port"`

	// Server identity
	ServerID int    `yaml:"server_id"`
	HexID    string `yaml:"hex_id"`

	// Database
	Database DatabaseConfig `yaml:"database"`

	// Rates
	Rates Rates `yaml:"rates"`

	// Enchant
	Enchant EnchantConfig `yaml:"enchant"`

	// PvP
	PvP PvPConfig `yaml:"pvp"`

	// Siege
	Siege SiegeConfig `yaml:"siege"`

	// Manor
	Manor ManorConfig `yaml:"manor"`

	// Write queue / timeouts (Phase 7.0)
	WriteTimeout  time.Duration `yaml:"write_timeout"`    // per-write deadline (default: 5s)
	ReadTimeout   time.Duration `yaml:"read_timeout"`     // idle client disconnect (default: 120s)
	SendQueueSize int           `yaml:"send_queue_size"`  // per-client outbox capacity (default: 256)

	// Geodata (Phase 7.1)
	GeodataDir string `yaml:"geodata_dir"` // path to .l2j files (optional, empty = no pathfinding)

	// NPC HTML dialogs (Phase 11)
	HtmlDir      string `yaml:"html_dir"`       // path to data/html/ (default: "data/html/")
	LazyHtmlLoad bool   `yaml:"lazy_html_load"` // false = pre-load all templates at startup

	// Offline Trade (Phase 31)
	OfflineTradeEnabled          bool          `yaml:"offline_trade_enabled"`           // allow offline trade mode
	OfflineMaxDays               int           `yaml:"offline_max_days"`                // max offline duration in days (0 = unlimited)
	OfflineDisconnectFinished    bool          `yaml:"offline_disconnect_finished"`     // remove trader when all items sold
	OfflineSetNameColor          bool          `yaml:"offline_set_name_color"`          // change name color for offline traders
	OfflineNameColor             int32         `yaml:"offline_name_color"`              // name color for offline traders (RGB)
	OfflineTradeRealtimeSave     bool          `yaml:"offline_trade_realtime_save"`     // save to DB after each transaction
	OfflineRestoreOnStartup      bool          `yaml:"offline_restore_on_startup"`      // restore offline traders on server start

	// Flood protection
	FloodProtection     bool `yaml:"flood_protection"`
	FastConnectionLimit int  `yaml:"fast_connection_limit"`
	NormalConnectionTime int  `yaml:"normal_connection_time"` // ms
	FastConnectionTime  int  `yaml:"fast_connection_time"`   // ms
	MaxConnectionPerIP  int  `yaml:"max_connection_per_ip"`
}

// DefaultGameServer returns GameServer config with sensible defaults.
func DefaultGameServer() GameServer {
	return GameServer{
		BindAddress:         "0.0.0.0",
		Port:                7777,
		LoginHost:           "127.0.0.1",
		LoginPort:           9013,
		ServerID:            1,
		HexID:               "c0a80001", // 192.168.0.1
		WriteTimeout:        5 * time.Second,
		ReadTimeout:         120 * time.Second,
		SendQueueSize:       256,
		HtmlDir:             "data/html/",
		OfflineDisconnectFinished: true,
		OfflineSetNameColor:       true,
		OfflineNameColor:          0x999999,
		FloodProtection:           true,
		FastConnectionLimit: 15,
		NormalConnectionTime: 700,
		FastConnectionTime:  350,
		MaxConnectionPerIP:  50,
		Database: DatabaseConfig{
			Host:     "127.0.0.1",
			Port:     5432,
			User:     "la2go",
			Password: "la2go",
			DBName:   "la2go",
			SSLMode:  "disable",
		},
		Rates: DefaultRates(),
		Enchant: EnchantConfig{
			SafeEnchant:   3,
			MaxEnchant:    25,
			ChanceWeapon:  0.66,
			ChanceArmor:   0.66,
			ChanceJewelry: 0.66,
			BlessedChance: 0.66,
			DestroyOnFail: true,
			CrystalOnFail: true,
		},
		PvP: PvPConfig{
			FlagDuration:    15,
			KarmaMinKills:   1,
			KarmaDecayRate:  4,
			PvPDamageMulti:  1.0,
			ProtectionLevel: 25,
		},
		Siege: SiegeConfig{
			IntervalDays: 14,
			StartHour:    20,
			Duration:     120,
			MaxDefenders: 500,
			MaxAttackers: 500,
		},
		Manor: ManorConfig{
			ApprovalHour:     6,
			MaintenanceMin:   3,
			ModifiableHour:   20,
			CropMatureReward: 90,
		},
	}
}

// LoadGameServer loads game server config from a YAML file.
// If the file doesn't exist, returns defaults.
func LoadGameServer(path string) (GameServer, error) {
	cfg := DefaultGameServer()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return cfg, nil
}
