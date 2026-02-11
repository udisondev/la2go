package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Rates holds server rate multipliers for drop, XP, SP, etc.
type Rates struct {
	DeathDropChanceMultiplier float64 `yaml:"death_drop_chance_multiplier"`
	DeathDropAmountMultiplier float64 `yaml:"death_drop_amount_multiplier"`
	ItemAutoDestroyTime       int     `yaml:"item_auto_destroy_time"` // seconds
}

// DefaultRates returns Rates with x1 multipliers and 60s auto-destroy.
func DefaultRates() Rates {
	return Rates{
		DeathDropChanceMultiplier: 1.0,
		DeathDropAmountMultiplier: 1.0,
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

	// Write queue / timeouts (Phase 7.0)
	WriteTimeout  time.Duration `yaml:"write_timeout"`    // per-write deadline (default: 5s)
	ReadTimeout   time.Duration `yaml:"read_timeout"`     // idle client disconnect (default: 120s)
	SendQueueSize int           `yaml:"send_queue_size"`  // per-client outbox capacity (default: 256)

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
		FloodProtection:     true,
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
