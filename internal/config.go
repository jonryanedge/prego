package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath = "~/.pregorc.yml"
	DefaultMode       = 0755
	Version           = 1
)

type Machine struct {
	Name string `yaml:"name,omitempty"`
	OS   string `yaml:"os,omitempty"`
}

type DirEntry struct {
	Path   string `yaml:"path"`
	Mode   uint32 `yaml:"mode,omitempty"`
	VCS    string `yaml:"vcs,omitempty"`
	Remote string `yaml:"remote,omitempty"`
}

type Symlink struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

type DirCategory struct {
	Root     string     `yaml:"root"`
	Entries  []DirEntry `yaml:"entries"`
	Symlinks []Symlink  `yaml:"symlinks,omitempty"`
}

type Hooks struct {
	PostCreate []string `yaml:"post_create,omitempty"`
}

type Config struct {
	Version int                    `yaml:"version"`
	Machine Machine                `yaml:"machine,omitempty"`
	Dirs    map[string]DirCategory `yaml:"dirs"`
	Hooks   Hooks                  `yaml:"hooks,omitempty"`
}

func ExpandPath(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func Load(path string) (*Config, error) {
	expanded := ExpandPath(path)
	data, err := os.ReadFile(expanded)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Save(path string, cfg *Config) error {
	expanded := ExpandPath(path)
	dir := filepath.Dir(expanded)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(expanded, data, 0600)
}

func NewDefault() *Config {
	return &Config{
		Version: Version,
		Machine: Machine{},
		Dirs: map[string]DirCategory{
			"core": {
				Root:    "~",
				Entries: []DirEntry{},
			},
			"documents": {
				Root:    "~/Documents",
				Entries: []DirEntry{},
			},
			"repos": {
				Root:    "~/repos",
				Entries: []DirEntry{},
			},
		},
		Hooks: Hooks{},
	}
}
