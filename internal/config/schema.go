package config

type Config struct {
	Version     string     `yaml:"version"`
	ScriptsHome string     `yaml:"scripts_home"`
	HelpFlag    string     `yaml:"help_flag"`
	Categories  []Category `yaml:"categories"`
	Scripts     []Script   `yaml:"scripts"`
	Source      string     `yaml:"-"` // "global" | "local" — set at load time
}

type Category struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	ScriptsPath string `yaml:"scripts_path"`
	HelpFlag    string `yaml:"help_flag"`
}

type Script struct {
	Name        string `yaml:"name"`
	Category    string `yaml:"category"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"` // "script" | "executable"
	Path        string `yaml:"path"`
	HelpFlag    string `yaml:"help_flag"`
}
