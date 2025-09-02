package lib

import (
	"context"
	"encoding/json"
	"os"
)

// Registrar interface for DNS update operations
type Registrar interface {
	UpdateDNS(ctx context.Context, config *Config, ipv4, ipv6 string) error
}

// Config holds configuration values for DNS update, supporting multiple registrars.
type Config struct {
	Registrar string `json:"registrar"` // e.g., "strato", "dyndns", "noip"
	User      string `json:"user"`
	Password  string `json:"password"`
	Host      string `json:"host"`
	// Registrar-specific options
	Options map[string]string `json:"options"`
	// Unifi fields (optional)
	UnifiSiteManagerApiKey string `json:"unifiSiteManagerApiKey"`
	UnifiSiteManagerHostId string `json:"unifiSiteManagerHostId"`
	UpdateIpv4             bool   `json:"updateIpv4"`
	UpdateIpv6             bool   `json:"updateIpv6"`
}

// LoadConfig loads configuration from the specified JSON file path.
// It returns a pointer to a Config struct and an error if loading or decoding fails.
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	// Validate required fields
	if config.Registrar == "" {
		return nil, &ConfigError{"Registrar is not set"}
	}
	if config.User == "" {
		return nil, &ConfigError{"User is not set"}
	}
	if config.Password == "" {
		return nil, &ConfigError{"Password is not set"}
	}
	if config.Host == "" {
		return nil, &ConfigError{"Host is not set"}
	}
	if !config.UpdateIpv4 && !config.UpdateIpv6 {
		return nil, &ConfigError{"UpdateIpv4 and UpdateIpv6 cannot both be false"}
	}

	return &config, nil
}

// ConfigError represents a configuration validation error.
type ConfigError struct {
	Msg string
}

func (e *ConfigError) Error() string {
	return "Config error: " + e.Msg
}
