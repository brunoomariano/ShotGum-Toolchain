package config

// Config is the top-level structure for both global (~/.config/shotgum/config.yaml)
// and local (.shotgum.yaml) configuration files.
type Config struct {
	Version     string     `yaml:"version"`
	ScriptsHome string     `yaml:"scripts_home"`
	HelpFlag    string     `yaml:"help_flag"` // default flag passed to scripts for help output
	DefaultExec string     `yaml:"default_executable"`
	Categories  []Category `yaml:"categories"`
	Scripts     []Script   `yaml:"scripts"`
	Source      string     `yaml:"-"` // "global" | "local" — populated at load time, not persisted
}

// Category groups related scripts under a named folder.
type Category struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	ScriptsPath string `yaml:"scripts_path"` // directory containing the scripts
	HelpFlag    string `yaml:"help_flag"`    // overrides Config.HelpFlag for all scripts in this category
}

// Script is a single registered script entry.
type Script struct {
	Name        string `yaml:"name"`
	Category    string `yaml:"category"`
	Description string `yaml:"description"`
	Executable  string `yaml:"executable"` // overrides Config.DefaultExec for this script
	Path        string `yaml:"path"`
	HelpFlag    string `yaml:"help_flag"` // overrides Category.HelpFlag for this script
}
