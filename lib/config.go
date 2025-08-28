package lib

import (
	"encoding/json"
	"os"
)

// Config holds configuration values for Strato DNS and Unifi Site Manager integration.
type Config struct {
	// User: a domain from your domains registered at Strato (example: yourstratodomain.com).
	User string `json:"user"`
	// Password: your Dynamic DNS password setup in Strato.
	Password string `json:"password"`
	// Host: domain or subdomain you want to update (example: test.yourstratodomain.com).
	Host string `json:"host"`
	// UnifiSiteManagerApiKey is your Unifi Site Manager API key.
	UnifiSiteManagerApiKey string `json:"unifiSiteManagerApiKey"`
	// UnifiSiteManagerHostId is your Unifi Site Manager host ID.
	UnifiSiteManagerHostId string `json:"unifiSiteManagerHostId"`
	// UpdateIpv4 specifies whether to update the IPv4 address.
	UpdateIpv4 bool `json:"updateIpv4"`
	// UpdateIpv6 specifies whether to update the IPv6 address.
	UpdateIpv6 bool `json:"updateIpv6"`
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
	if config.User == "" {
		return nil, &ConfigError{"User is not set"}
	}
	if config.Password == "" {
		return nil, &ConfigError{"Password is not set"}
	}
	if config.Host == "" {
		return nil, &ConfigError{"Host is not set"}
	}
	if config.UnifiSiteManagerApiKey == "" {
		return nil, &ConfigError{"UnifiSiteManagerApiKey is not set"}
	}
	if config.UnifiSiteManagerHostId == "" {
		return nil, &ConfigError{"UnifiSiteManagerHostId is not set"}
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
