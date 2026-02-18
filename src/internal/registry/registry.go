// Package registry merges the global and local ShotGum configurations into a
// single unified view. Local entries always take precedence over global ones of
// the same name. The registry also resolves script paths, executables, and help
// flags according to the configured cascade rules.
package registry

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
)

// CategoryEntry is a category with its source (user/local).
type CategoryEntry struct {
	config.Category
	Source string // "user" | "local"
}

// ScriptEntry is a script with its source and resolved paths.
type ScriptEntry struct {
	config.Script
	Source string // "user" | "local"
}

// Registry holds the merged view of global and local configs.
type Registry struct {
	global *config.Config
	local  *config.Config
}

// Load loads and merges global + local configs.
func Load() (*Registry, error) {
	global, err := config.LoadGlobal()
	if err != nil {
		return nil, fmt.Errorf("loading global config: %w", err)
	}
	local, err := config.LoadLocal()
	if err != nil {
		return nil, fmt.Errorf("loading local config: %w", err)
	}
	return &Registry{global: global, local: local}, nil
}

// GetCategories returns merged categories. Local entries override global ones with the same name.
func (r *Registry) GetCategories() []CategoryEntry {
	seen := map[string]bool{}
	var result []CategoryEntry

	// Local takes precedence
	if r.local != nil {
		for _, c := range r.local.Categories {
			seen[c.Name] = true
			result = append(result, CategoryEntry{Category: c, Source: "local"})
		}
	}

	if r.global != nil {
		for _, c := range r.global.Categories {
			if !seen[c.Name] {
				result = append(result, CategoryEntry{Category: c, Source: "user"})
			}
		}
	}

	return result
}

// GetScripts returns scripts for a given category name.
func (r *Registry) GetScripts(category string) []ScriptEntry {
	seen := map[string]bool{}
	var result []ScriptEntry

	if r.local != nil {
		for _, s := range r.local.Scripts {
			if s.Category == category {
				seen[s.Name] = true
				result = append(result, ScriptEntry{Script: s, Source: "local"})
			}
		}
	}

	if r.global != nil {
		for _, s := range r.global.Scripts {
			if s.Category == category && !seen[s.Name] {
				result = append(result, ScriptEntry{Script: s, Source: "user"})
			}
		}
	}

	return result
}

// FindScript returns the ScriptEntry for the given category and script name.
func (r *Registry) FindScript(category, name string) (*ScriptEntry, error) {
	for _, s := range r.GetScripts(category) {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("script %q not found in category %q", name, category)
}

// ResolveScriptPath returns the absolute path for a script.
func (r *Registry) ResolveScriptPath(entry ScriptEntry) string {
	p := entry.Path

	// Already absolute
	if filepath.IsAbs(p) {
		return p
	}

	// Resolve relative to category scripts_path first
	cat := r.findCategory(entry.Category, entry.Source)
	if cat != nil && cat.ScriptsPath != "" {
		return filepath.Join(cat.ScriptsPath, p)
	}

	// Fall back to scripts_home/category/
	scriptsHome := r.scriptsHome(entry.Source)
	return filepath.Join(scriptsHome, entry.Category, p)
}

// ResolveHelpFlag returns the effective help flag: script → category → config → default.
func (r *Registry) ResolveHelpFlag(entry ScriptEntry) string {
	if entry.HelpFlag != "" {
		return entry.HelpFlag
	}
	cat := r.findCategory(entry.Category, entry.Source)
	if cat != nil && cat.HelpFlag != "" {
		return cat.HelpFlag
	}
	if cfg := r.configFor(entry.Source); cfg != nil && cfg.HelpFlag != "" {
		return cfg.HelpFlag
	}
	return "--help"
}

// ResolveExecutable returns the effective executable for a script:
// script.executable -> config.default_executable (same source) ->
// global.default_executable -> /bin/sh.
func (r *Registry) ResolveExecutable(entry ScriptEntry) string {
	if entry.Executable != "" {
		return resolveExecutablePath(entry.Executable)
	}
	if cfg := r.configFor(entry.Source); cfg != nil && cfg.DefaultExec != "" {
		return resolveExecutablePath(cfg.DefaultExec)
	}
	if r.global != nil && r.global.DefaultExec != "" {
		return resolveExecutablePath(r.global.DefaultExec)
	}
	return "/bin/sh"
}

// GlobalConfig returns the global config (may be nil if not found).
func (r *Registry) GlobalConfig() *config.Config {
	return r.global
}

// LocalConfig returns the local config (may be nil if not found).
func (r *Registry) LocalConfig() *config.Config {
	return r.local
}

func (r *Registry) findCategory(name, source string) *config.Category {
	cfg := r.configFor(source)
	if cfg == nil {
		return nil
	}
	for i := range cfg.Categories {
		if cfg.Categories[i].Name == name {
			return &cfg.Categories[i]
		}
	}
	return nil
}

func (r *Registry) configFor(source string) *config.Config {
	if source == "local" {
		return r.local
	}
	return r.global
}

func (r *Registry) scriptsHome(source string) string {
	cfg := r.configFor(source)
	if cfg != nil && cfg.ScriptsHome != "" {
		return cfg.ScriptsHome
	}
	if r.global != nil && r.global.ScriptsHome != "" {
		return r.global.ScriptsHome
	}
	return ""
}

func resolveExecutablePath(executable string) string {
	if executable == "" {
		return "/bin/sh"
	}
	if filepath.IsAbs(executable) {
		return executable
	}
	if resolved, err := exec.LookPath(executable); err == nil {
		return resolved
	}
	return executable
}
