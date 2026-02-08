package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

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
