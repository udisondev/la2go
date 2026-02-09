package config

import (
	"fmt"
	"os"
	"strings"

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

	// Logging
	LogLevel string `yaml:"log_level"` // debug, info, warn, error (default: info)

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

	// Connection pool parameters (optional, defaults from pgxpool apply if not set)
	MaxConns          int32  `yaml:"max_conns"`            // default: max(4, NumCPU)
	MinConns          int32  `yaml:"min_conns"`            // default: 0
	MinIdleConns      int32  `yaml:"min_idle_conns"`       // default: 0
	MaxConnLifetime   string `yaml:"max_conn_lifetime"`    // duration, e.g. "1h"
	MaxConnIdleTime   string `yaml:"max_conn_idle_time"`   // duration, e.g. "30m"
	HealthCheckPeriod string `yaml:"health_check_period"`  // duration, e.g. "1m"
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	base := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)

	// Append pool parameters if set (non-zero/non-empty)
	var params []string
	if d.MaxConns > 0 {
		params = append(params, fmt.Sprintf("pool_max_conns=%d", d.MaxConns))
	}
	if d.MinConns > 0 {
		params = append(params, fmt.Sprintf("pool_min_conns=%d", d.MinConns))
	}
	if d.MaxConnLifetime != "" {
		params = append(params, fmt.Sprintf("pool_max_conn_lifetime=%s", d.MaxConnLifetime))
	}
	if d.MaxConnIdleTime != "" {
		params = append(params, fmt.Sprintf("pool_max_conn_idle_time=%s", d.MaxConnIdleTime))
	}
	if d.HealthCheckPeriod != "" {
		params = append(params, fmt.Sprintf("pool_health_check_period=%s", d.HealthCheckPeriod))
	}

	if len(params) > 0 {
		return base + "&" + strings.Join(params, "&")
	}
	return base
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
		LogLevel:            "info",
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
