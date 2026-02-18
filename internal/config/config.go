package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	globalConfigDir   = ".config/shotgum"
	globalConfigFile  = "config.yaml"
	localConfigFile   = ".shotgum.yaml"
	defaultScriptsDir = ".shotgum/scripts"
)

func globalConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, globalConfigDir, globalConfigFile)
}

func defaultScriptsHome() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, defaultScriptsDir)
}

// expandPath expands ~ and environment variables in a path.
func expandPath(path string) string {
	if path == "" {
		return path
	}
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	return os.ExpandEnv(path)
}

// LoadGlobal reads ~/.config/shotgum/config.yaml.
func LoadGlobal() (*Config, error) {
	path := globalConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading global config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing global config: %w", err)
	}
	cfg.Source = "global"
	cfg.ScriptsHome = expandPath(cfg.ScriptsHome)
	for i := range cfg.Categories {
		cfg.Categories[i].ScriptsPath = expandPath(cfg.Categories[i].ScriptsPath)
	}
	for i := range cfg.Scripts {
		cfg.Scripts[i].Path = expandPath(cfg.Scripts[i].Path)
	}
	return &cfg, nil
}

// LoadLocal walks CWD upward looking for .shotgum.yaml.
func LoadLocal() (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting cwd: %w", err)
	}

	dir := cwd
	for {
		candidate := filepath.Join(dir, localConfigFile)
		data, err := os.ReadFile(candidate)
		if err == nil {
			var cfg Config
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("parsing local config %s: %w", candidate, err)
			}
			cfg.Source = "local"
			cfg.ScriptsHome = expandPath(cfg.ScriptsHome)
			for i := range cfg.Categories {
				cfg.Categories[i].ScriptsPath = expandPath(cfg.Categories[i].ScriptsPath)
			}
			for i := range cfg.Scripts {
				cfg.Scripts[i].Path = expandPath(cfg.Scripts[i].Path)
			}
			return &cfg, nil
		}
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading local config: %w", err)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}
	return nil, nil
}

// Save writes a config to the given path, creating parent directories as needed.
func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// EnsureDefault creates the global config and scripts directory on first run.
func EnsureDefault() error {
	cfgPath := globalConfigPath()
	scriptsHome := defaultScriptsHome()

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfg := &Config{
			Version:     "1",
			ScriptsHome: "~/" + defaultScriptsDir,
			HelpFlag:    "--help",
		}
		if err := Save(cfg, cfgPath); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(scriptsHome, 0755); err != nil {
		return fmt.Errorf("creating scripts home: %w", err)
	}
	return nil
}

// GlobalConfigPath returns the path to the global config file (exported for commands).
func GlobalConfigPath() string {
	return globalConfigPath()
}
