package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the CLI configuration
type Config struct {
	APIKey       string `json:"api_key"`
	DefaultModel string `json:"default_model"`
	SafeMode     bool   `json:"safe_mode"`
	StreamOutput bool   `json:"stream_output"`
	Theme        string `json:"theme"`
}

var current *Config

// Default returns default configuration
func Default() *Config {
	return &Config{
		DefaultModel: "llama-3.3-70b-versatile",
		SafeMode:     true,
		StreamOutput: true,
		Theme:        "dark",
	}
}

// Get returns the current configuration
func Get() *Config {
	if current == nil {
		current = Default()
	}
	return current
}

// Load reads configuration from disk
func Load() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			current = Default()
			return nil
		}
		return fmt.Errorf("failed to read config: %w", err)
	}

	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Allow env var to override
	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		cfg.APIKey = key
	}

	current = cfg
	return nil
}

// Save writes configuration to disk
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	current = cfg
	return nil
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}
	return filepath.Join(home, ".config", "groq-cli", "config.json"), nil
}

// ConfigPath returns the path to the config file (public)
func ConfigPath() (string, error) {
	return configPath()
}
