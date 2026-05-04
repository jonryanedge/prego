package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"tilde expansion", "~/foo", filepath.Join(home, "foo")},
		{"absolute path", "/usr/local", "/usr/local"},
		{"relative path", "foo", "foo"},
		{"tilde only", "~", home},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadAndSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".pregorc.yml")

	cfg := &Config{
		Version: Version,
		General: General{Color: true},
		System: System{
			Machine: Machine{Name: "test", OS: "darwin"},
			Hooks:   Hooks{},
		},
		Directory: map[string]DirCategory{
			"core": {
				Root: "~",
				Entries: []DirEntry{
					{Path: "~/.config", Mode: 0700},
					{Path: "~/.local/bin", Mode: 0700},
				},
			},
			"documents": {
				Root: "~/Documents",
				Entries: []DirEntry{
					{Path: "~/Documents/projects", Mode: 0755},
				},
			},
			"repos": {
				Root: "~/repos",
				Entries: []DirEntry{
					{Path: "~/repos/personal", Mode: 0755, VCS: "git"},
				},
			},
		},
	}

	err := Save(path, cfg)
	require.NoError(t, err)

	loaded, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, cfg.Version, loaded.Version)
	assert.Equal(t, cfg.System.Machine.Name, loaded.System.Machine.Name)
	assert.Equal(t, cfg.System.Machine.OS, loaded.System.Machine.OS)
	assert.Len(t, loaded.Directory["core"].Entries, 2)
	assert.Equal(t, "~/.config", loaded.Directory["core"].Entries[0].Path)
	assert.Equal(t, uint32(0700), loaded.Directory["core"].Entries[0].Mode)
	assert.Equal(t, "~/repos/personal", loaded.Directory["repos"].Entries[0].Path)
	assert.Equal(t, "git", loaded.Directory["repos"].Entries[0].VCS)
}

func TestLoadNonexistent(t *testing.T) {
	_, err := Load("/nonexistent/path/.pregorc.yml")
	assert.Error(t, err)
}

func TestSaveCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "dir", ".pregorc.yml")

	cfg := NewDefault()
	err := Save(path, cfg)
	require.NoError(t, err)

	_, err = Load(path)
	require.NoError(t, err)
}

func TestContractPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"home subdir", filepath.Join(home, "repos"), "~/repos"},
		{"home itself", home, "~"},
		{"absolute non-home", "/usr/local", "/usr/local"},
		{"nested path", filepath.Join(home, "repos", "project"), "~/repos/project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContractPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandContractRoundtrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"tilde path", "~/repos/project"},
		{"home only", "~"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expanded := ExpandPath(tt.input)
			contracted := ContractPath(expanded)
			assert.Equal(t, tt.input, contracted)
		})
	}
}

func TestNewDefault(t *testing.T) {
	cfg := NewDefault()
	assert.Equal(t, Version, cfg.Version)
	assert.Contains(t, cfg.Directory, "core")
	assert.Contains(t, cfg.Directory, "documents")
	assert.Contains(t, cfg.Directory, "repos")
}

func TestResolveEntryPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		path     string
		root     string
		expected string
	}{
		{"absolute path", "/usr/local/bin", "/opt", "/usr/local/bin"},
		{"tilde path", "~/repos/project", "/any", filepath.Join(home, "repos", "project")},
		{"relative path with dot root", "project", "/home/user/repos", "/home/user/repos/project"},
		{"relative path with tilde root", "project", "~/repos", filepath.Join(home, "repos", "project")},
		{"nested relative", "src/components", "/home/user/repos", "/home/user/repos/src/components"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveEntryPath(tt.path, ResolveRoot(tt.root))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveRoot(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		root     string
		expected string
	}{
		{"tilde expansion", "~/repos", filepath.Join(home, "repos")},
		{"home only", "~", home},
		{"absolute path", "/usr/local", "/usr/local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveRoot(tt.root)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveRootDot(t *testing.T) {
	result := ResolveRoot(".")
	cwd, _ := os.Getwd()
	assert.Equal(t, cwd, result)
}

func TestRelPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		root     string
		expected string
	}{
		{"under home", "~/repos/project", "~/repos", "project"},
		{"absolute under root", "/home/user/repos/project", "/home/user/repos", "project"},
		{"outside root", "/other/path", "/home/user/repos", "/other/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RelPath(tt.path, tt.root)
			assert.Equal(t, tt.expected, result)
		})
	}
}
