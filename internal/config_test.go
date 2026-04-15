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
		{"tilde only", "~", "~"},
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
		Machine: Machine{Name: "test", OS: "darwin"},
		Dirs: map[string]DirCategory{
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
	assert.Equal(t, cfg.Machine.Name, loaded.Machine.Name)
	assert.Equal(t, cfg.Machine.OS, loaded.Machine.OS)
	assert.Len(t, loaded.Dirs["core"].Entries, 2)
	assert.Equal(t, "~/.config", loaded.Dirs["core"].Entries[0].Path)
	assert.Equal(t, uint32(0700), loaded.Dirs["core"].Entries[0].Mode)
	assert.Equal(t, "~/repos/personal", loaded.Dirs["repos"].Entries[0].Path)
	assert.Equal(t, "git", loaded.Dirs["repos"].Entries[0].VCS)
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

func TestNewDefault(t *testing.T) {
	cfg := NewDefault()
	assert.Equal(t, Version, cfg.Version)
	assert.Contains(t, cfg.Dirs, "core")
	assert.Contains(t, cfg.Dirs, "documents")
	assert.Contains(t, cfg.Dirs, "repos")
}
