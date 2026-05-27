package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	EnvBaseURL = "PAIR_BASE_URL"
	EnvToken   = "PAIR_TOKEN"
)

// Config stores PAIR CLI settings.
type Config struct {
	BaseURL string `json:"base_url,omitempty"`
	Token   string `json:"token,omitempty"`
}

// Overrides are explicit command-line values.
type Overrides struct {
	BaseURL string
	Token   string
}

// Source describes where a resolved value came from.
type Source string

const (
	SourceUnset Source = "unset"
	SourceFlag  Source = "flag"
	SourceEnv   Source = "env"
	SourceFile  Source = "file"
)

// Resolved is a config value plus source metadata for status output and tests.
type Resolved struct {
	Config
	BaseURLSource Source
	TokenSource   Source
	Path          string
}

// DefaultPath returns the XDG config path for pair.
func DefaultPath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "pair", "config.json"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	if home == "" {
		return "", errors.New("find home directory: HOME is empty")
	}

	return filepath.Join(home, ".config", "pair", "config.json"), nil
}

// Load reads a config file. Missing files return an empty config.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if len(data) == 0 {
		return Config{}, nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// Save writes a config file with owner-only permissions.
func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return os.Chmod(path, 0o600)
}

// Resolve applies flag > env > file precedence.
func Resolve(overrides Overrides) (Resolved, error) {
	path, err := DefaultPath()
	if err != nil {
		return Resolved{}, err
	}

	fileCfg, err := Load(path)
	if err != nil {
		return Resolved{}, err
	}

	resolved := Resolved{
		Config:        fileCfg,
		BaseURLSource: sourceFor(fileCfg.BaseURL, SourceFile),
		TokenSource:   sourceFor(fileCfg.Token, SourceFile),
		Path:          path,
	}

	if value := os.Getenv(EnvBaseURL); value != "" {
		resolved.BaseURL = value
		resolved.BaseURLSource = SourceEnv
	}
	if value := os.Getenv(EnvToken); value != "" {
		resolved.Token = value
		resolved.TokenSource = SourceEnv
	}
	if overrides.BaseURL != "" {
		resolved.BaseURL = overrides.BaseURL
		resolved.BaseURLSource = SourceFlag
	}
	if overrides.Token != "" {
		resolved.Token = overrides.Token
		resolved.TokenSource = SourceFlag
	}

	return resolved, nil
}

func sourceFor(value string, source Source) Source {
	if value == "" {
		return SourceUnset
	}
	return source
}
