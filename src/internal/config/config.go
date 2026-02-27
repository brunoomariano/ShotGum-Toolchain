// Package config handles loading, validating, and persisting ShotGum
// configuration files. Global config lives at ~/.config/shotgum/config.yaml;
// local config is discovered by walking up from the working directory.
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

// expandPath expands ~ and environment variables in a single path string.
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

// expandPaths expands ~ and env vars in all path fields of cfg in-place.
func expandPaths(cfg *Config) {
	cfg.ScriptsHome = expandPath(cfg.ScriptsHome)
	cfg.DefaultExec = expandPath(cfg.DefaultExec)
	for i := range cfg.Categories {
		cfg.Categories[i].ScriptsPath = expandPath(cfg.Categories[i].ScriptsPath)
	}
	for i := range cfg.Scripts {
		cfg.Scripts[i].Path = expandPath(cfg.Scripts[i].Path)
		cfg.Scripts[i].Executable = expandPath(cfg.Scripts[i].Executable)
	}
}

func applyDefaults(cfg *Config) {
	if cfg.MakefileImport == nil {
		defaultEnabled := true
		cfg.MakefileImport = &defaultEnabled
	}
	if cfg.MakefileImportMode == "" {
		cfg.MakefileImportMode = "all"
	}
	if cfg.MakefileImportSource == "" {
		cfg.MakefileImportSource = "parser"
	}
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
	applyDefaults(&cfg)
	expandPaths(&cfg)
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
			applyDefaults(&cfg)
			expandPaths(&cfg)
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
		defaultEnabled := true
		cfg := &Config{
			Version:              "1",
			ScriptsHome:          "~/" + defaultScriptsDir,
			HelpFlag:             "--help",
			DefaultExec:          "/bin/sh",
			MakefileImport:       &defaultEnabled,
			MakefileImportMode:   "all",
			MakefileImportSource: "parser",
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

// EffectiveMakefileImport returns the effective makefile import toggle.
func EffectiveMakefileImport(cfg *Config) bool {
	if cfg == nil || cfg.MakefileImport == nil {
		return true
	}
	return *cfg.MakefileImport
}

// EffectiveMakefileImportMode returns the effective import mode.
func EffectiveMakefileImportMode(cfg *Config) string {
	if cfg == nil || cfg.MakefileImportMode == "" {
		return "all"
	}
	return cfg.MakefileImportMode
}

// EffectiveMakefileImportSource returns the effective import source.
func EffectiveMakefileImportSource(cfg *Config) string {
	if cfg == nil || cfg.MakefileImportSource == "" {
		return "parser"
	}
	return cfg.MakefileImportSource
}
