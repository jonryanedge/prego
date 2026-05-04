package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath = "~/.pregorc.yml"
	LocalConfigName   = ".pregorc.yml"
	DefaultMode       = 0755
	Version           = 2
)

type General struct {
	Color   bool `yaml:"color,omitempty"`
	Verbose bool `yaml:"verbose,omitempty"`
}

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

type System struct {
	Machine Machine `yaml:"machine,omitempty"`
	Hooks   Hooks   `yaml:"hooks,omitempty"`
}

type Config struct {
	Version   int                    `yaml:"version"`
	General   General                `yaml:"general,omitempty"`
	System    System                 `yaml:"system,omitempty"`
	Directory map[string]DirCategory `yaml:"directory"`
}

func ExpandPath(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func ContractPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if len(path) > len(home) && path[:len(home)+1] == home+string(filepath.Separator) {
		return "~" + path[len(home):]
	}
	if path == home {
		return "~"
	}
	return path
}

func RelPath(path, root string) string {
	expanded := ExpandPath(path)
	rootExpanded := ExpandPath(root)
	rel, err := filepath.Rel(rootExpanded, expanded)
	if err != nil {
		return path
	}
	if strings.HasPrefix(rel, "..") {
		return path
	}
	return rel
}

func ResolveEntryPath(entryPath, root string) string {
	if strings.HasPrefix(entryPath, "/") || strings.HasPrefix(entryPath, "~/") {
		return ExpandPath(entryPath)
	}
	return filepath.Join(ExpandPath(root), entryPath)
}

func ResolveRoot(root string) string {
	if root == "." {
		abs, err := filepath.Abs(".")
		if err != nil {
			return root
		}
		return abs
	}
	return ExpandPath(root)
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

func Merge(system, local *Config) *Config {
	if system == nil {
		return local
	}
	if local == nil {
		return system
	}

	merged := *system
	merged.Version = local.Version

	if local.General.Color {
		merged.General.Color = true
	}
	if local.General.Verbose {
		merged.General.Verbose = true
	}

	if local.System.Machine.Name != "" {
		merged.System.Machine.Name = local.System.Machine.Name
	}
	if local.System.Machine.OS != "" {
		merged.System.Machine.OS = local.System.Machine.OS
	}
	merged.System.Hooks.PostCreate = append(merged.System.Hooks.PostCreate, local.System.Hooks.PostCreate...)

	if merged.Directory == nil {
		merged.Directory = map[string]DirCategory{}
	}
	for cat, localCat := range local.Directory {
		systemCat, exists := merged.Directory[cat]
		if !exists {
			merged.Directory[cat] = localCat
			continue
		}

		existingPaths := make(map[string]int)
		for i, e := range systemCat.Entries {
			existingPaths[ExpandPath(e.Path)] = i
		}
		for _, e := range localCat.Entries {
			expanded := ExpandPath(e.Path)
			if idx, ok := existingPaths[expanded]; ok {
				systemCat.Entries[idx] = e
			} else {
				systemCat.Entries = append(systemCat.Entries, e)
			}
		}

		existingLinks := make(map[string]int)
		for i, sl := range systemCat.Symlinks {
			existingLinks[sl.To] = i
		}
		for _, sl := range localCat.Symlinks {
			if idx, ok := existingLinks[sl.To]; ok {
				systemCat.Symlinks[idx] = sl
			} else {
				systemCat.Symlinks = append(systemCat.Symlinks, sl)
			}
		}

		if localCat.Root != "" {
			systemCat.Root = localCat.Root
		}
		merged.Directory[cat] = systemCat
	}

	return &merged
}

func DiscoverConfig(cfgPath string) (*Config, error) {
	systemPath := ExpandPath(cfgPath)

	systemCfg, err := Load(systemPath)
	if err != nil {
		systemCfg = nil
	}

	localPath := filepath.Join(".", LocalConfigName)
	localCfg, err := Load(localPath)
	if err != nil {
		localCfg = nil
	}

	if localCfg != nil {
		localAbs, absErr := filepath.Abs(localPath)
		if absErr == nil {
			localDir := filepath.Dir(localAbs)
			for cat, dirCat := range localCfg.Directory {
				if dirCat.Root == "." {
					dirCat.Root = localDir
					localCfg.Directory[cat] = dirCat
				}
			}
		}
	}

	if systemCfg == nil && localCfg == nil {
		return NewDefault(), nil
	}

	return Merge(systemCfg, localCfg), nil
}

func NewDefault() *Config {
	hostname, _ := os.Hostname()
	return &Config{
		Version: Version,
		General: General{
			Color:   true,
			Verbose: false,
		},
		System: System{
			Machine: Machine{
				Name: hostname,
				OS:   runtime.GOOS,
			},
			Hooks: Hooks{},
		},
		Directory: map[string]DirCategory{
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
	}
}
