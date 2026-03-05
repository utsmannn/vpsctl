package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the main configuration structure
type Config struct {
	LXD    LXDConfig    `mapstructure:"lxd"`
	API    APIConfig    `mapstructure:"api"`
	Output OutputConfig `mapstructure:"output"`
}

// LXDConfig holds LXD-specific configuration
type LXDConfig struct {
	Socket string `mapstructure:"socket"` // LXD unix socket path
}

// APIConfig holds API server configuration
type APIConfig struct {
	Port   int    `mapstructure:"port"`
	Socket string `mapstructure:"socket"`
	Auth   bool   `mapstructure:"auth"`
	Token  string `mapstructure:"token"`
	CORS   bool   `mapstructure:"cors"`
}

// OutputConfig holds output formatting configuration
type OutputConfig struct {
	Format string `mapstructure:"format"` // table, json, csv
}

// Default config values
var defaults = map[string]interface{}{
	"lxd.socket":     "/var/snap/lxd/common/lxd/unix.socket",
	"api.port":       8080,
	"api.auth":       false,
	"api.token":      "",
	"api.cors":       true,
	"output.format":  "table",
}

// Load loads configuration from default locations
func Load() (*Config, error) {
	// Set default values
	for key, value := range defaults {
		viper.SetDefault(key, value)
	}

	// Set config file name and paths
	viper.SetConfigName(".vpsctl")
	viper.SetConfigType("yaml")

	// Add config search paths
	configPath := GetConfigPath()
	viper.AddConfigPath(filepath.Dir(configPath))
	viper.AddConfigPath(".")

	// Read from environment variables
	viper.SetEnvPrefix("VPSCTL")
	viper.AutomaticEnv()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// LoadFromFile loads configuration from a specific file path
func LoadFromFile(path string) (*Config, error) {
	viper.SetConfigFile(path)

	// Set default values
	for key, value := range defaults {
		viper.SetDefault(key, value)
	}

	// Read from environment variables
	viper.SetEnvPrefix("VPSCTL")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", path, err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// Save saves configuration to the specified file path
func Save(cfg *Config, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Set values in viper
	viper.Set("lxd", cfg.LXD)
	viper.Set("api", cfg.API)
	viper.Set("output", cfg.Output)

	// Write config file
	if err := viper.WriteConfigAs(path); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the default config file path (~/.vpsctl.yaml)
func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return ".vpsctl.yaml"
	}
	return filepath.Join(home, ".vpsctl.yaml")
}

// GetSocketPath returns the LXD socket path from config or default
func GetSocketPath() string {
	return viper.GetString("lxd.socket")
}

// GetAPIPort returns the API port from config or default
func GetAPIPort() int {
	return viper.GetInt("api.port")
}

// GetOutputFormat returns the output format from config or default
func GetOutputFormat() string {
	return viper.GetString("output.format")
}

// IsAuthEnabled returns whether API authentication is enabled
func IsAuthEnabled() bool {
	return viper.GetBool("api.auth")
}

// GetAPIToken returns the API token
func GetAPIToken() string {
	return viper.GetString("api.token")
}

// IsCORSEnabled returns whether CORS is enabled
func IsCORSEnabled() bool {
	return viper.GetBool("api.cors")
}
