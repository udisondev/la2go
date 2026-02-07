package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoginServer holds all configuration for the login server.
type LoginServer struct {
	// Network
	BindAddress string `yaml:"bind_address"`
	Port        int    `yaml:"port"`

	// GameServer listener
	GSListenHost string `yaml:"gs_listen_host"`
	GSListenPort int    `yaml:"gs_listen_port"`

	// Database
	Database DatabaseConfig `yaml:"database"`

	// Security
	AutoCreateAccounts bool `yaml:"auto_create_accounts"`
	ShowLicence        bool `yaml:"show_licence"`
	LoginTryBeforeBan  int  `yaml:"login_try_before_ban"`
	LoginBlockAfterBan int  `yaml:"login_block_after_ban"` // seconds

	// Flood protection
	FloodProtection     bool `yaml:"flood_protection"`
	FastConnectionLimit int  `yaml:"fast_connection_limit"`
	NormalConnectionTime int  `yaml:"normal_connection_time"` // ms
	FastConnectionTime  int  `yaml:"fast_connection_time"`   // ms
	MaxConnectionPerIP  int  `yaml:"max_connection_per_ip"`

	// Game servers (static list for Phase 2)
	GameServers []GameServerEntry `yaml:"game_servers"`
}

// DatabaseConfig holds PostgreSQL connection parameters.
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

// GameServerEntry represents a known game server in the config.
type GameServerEntry struct {
	ID   int    `yaml:"id"`
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// DefaultLoginServer returns LoginServer config with sensible defaults.
func DefaultLoginServer() LoginServer {
	return LoginServer{
		BindAddress:         "0.0.0.0",
		Port:                2106,
		GSListenHost:        "127.0.0.1",
		GSListenPort:        9013,
		AutoCreateAccounts:  true,
		ShowLicence:         true,
		LoginTryBeforeBan:   5,
		LoginBlockAfterBan:  900,
		FloodProtection:     true,
		FastConnectionLimit: 15,
		NormalConnectionTime: 700,
		FastConnectionTime:  350,
		MaxConnectionPerIP:  50,
		Database: DatabaseConfig{
			Host:    "127.0.0.1",
			Port:    5432,
			User:    "la2go",
			Password: "la2go",
			DBName:  "la2go",
			SSLMode: "disable",
		},
		GameServers: []GameServerEntry{
			{
				ID:   1,
				Name: "Bartz",
				Host: "127.0.0.1",
				Port: 7777,
			},
		},
	}
}

// LoadLoginServer loads login server config from a YAML file.
// If the file doesn't exist, returns defaults.
func LoadLoginServer(path string) (LoginServer, error) {
	cfg := DefaultLoginServer()

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
