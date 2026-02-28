// Package config manages user configuration for clipet.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the user configuration.
type Config struct {
	Language         string `json:"language"`
	FallbackLanguage string `json:"fallback_language"`
	Version          string `json:"version"`
}

// Default configuration values.
const (
	DefaultLanguage         = "zh-CN"
	DefaultFallbackLanguage = "en-US"
	DefaultVersion          = "1.0"
)

// Load loads the configuration from disk, creating defaults if needed.
// Environment variables take precedence over config file settings.
func Load() (*Config, error) {
	cfgPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	var cfg Config

	// Try to load existing config
	if data, err := os.ReadFile(cfgPath); err == nil {
		if err := json.Unmarshal(data, &cfg); err == nil {
			// Validate loaded config
			if cfg.Language == "" {
				cfg.Language = DefaultLanguage
			}
			if cfg.FallbackLanguage == "" {
				cfg.FallbackLanguage = DefaultFallbackLanguage
			}
		}
	} else {
		// Create default config if file doesn't exist
		cfg = Config{
			Language:         detectLanguage(),
			FallbackLanguage: DefaultFallbackLanguage,
			Version:          DefaultVersion,
		}

		// Save default config
		if err := cfg.Save(); err != nil {
			// Log warning but don't fail - config is optional
			fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
		}

		return &cfg, nil
	}

	// Environment variables override config file (highest priority)
	if envLang := detectLanguageFromEnv(); envLang != "" {
		cfg.Language = envLang
	}

	return &cfg, nil
}

// Save writes the configuration to disk.
func (c *Config) Save() error {
	cfgPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	cfgDir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// SetLanguage updates the language setting.
func (c *Config) SetLanguage(lang string) error {
	c.Language = lang
	return c.Save()
}

// getConfigPath returns the path to the configuration file.
func getConfigPath() (string, error) {
	// Check for XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "clipet", "config.json"), nil
	}

	// Fall back to HOME/.config
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".config", "clipet", "config.json"), nil
}

// detectLanguage detects the system language with fallback chain.
func detectLanguage() string {
	// Check environment variables first
	if lang := detectLanguageFromEnv(); lang != "" {
		return lang
	}

	// Default
	return DefaultLanguage
}

// detectLanguageFromEnv checks only environment variables for language setting.
// Returns empty string if no env var is set.
func detectLanguageFromEnv() string {
	// Priority 1: CLIPET_LANG environment variable
	if lang := os.Getenv("CLIPET_LANG"); lang != "" {
		return normalizeLanguage(lang)
	}

	// Priority 2: LANG environment variable
	if lang := os.Getenv("LANG"); lang != "" {
		return normalizeLanguage(lang)
	}

	// Priority 3: LC_ALL environment variable
	if lang := os.Getenv("LC_ALL"); lang != "" {
		return normalizeLanguage(lang)
	}

	return ""
}

// normalizeLanguage normalizes a language code to standard format.
// Examples:
//   - "zh_CN.UTF-8" -> "zh-CN"
//   - "en_US" -> "en-US"
//   - "en" -> "en"
func normalizeLanguage(lang string) string {
	// Remove encoding suffix (e.g., ".UTF-8")
	if idx := strings.Index(lang, "."); idx > 0 {
		lang = lang[:idx]
	}

	// Replace underscore with hyphen (e.g., "zh_CN" -> "zh-CN")
	lang = strings.ReplaceAll(lang, "_", "-")

	// Validate format (should be xx or xx-XX)
	parts := strings.Split(lang, "-")
	if len(parts) == 1 {
		// Just language code (e.g., "en")
		return lang
	}

	// Language + region (e.g., "en-US")
	return lang
}
