package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeSystemOnly(t *testing.T) {
	system := &Config{
		Version: Version,
		General: General{Color: true},
		System: System{
			Machine: Machine{Name: "desktop", OS: "darwin"},
			Hooks:   Hooks{PostCreate: []string{"echo hello"}},
		},
		Directory: map[string]DirCategory{
			"core": {
				Root: "~",
				Entries: []DirEntry{
					{Path: "~/.config", Mode: 0700},
				},
			},
		},
	}

	result := Merge(system, nil)
	assert.Equal(t, "desktop", result.System.Machine.Name)
	assert.Len(t, result.Directory["core"].Entries, 1)
}

func TestMergeLocalOnly(t *testing.T) {
	local := &Config{
		Version: Version,
		General: General{Verbose: true},
		System: System{
			Machine: Machine{Name: "laptop", OS: "linux"},
		},
		Directory: map[string]DirCategory{
			"repos": {
				Root: "~/repos",
				Entries: []DirEntry{
					{Path: "~/repos/work", Mode: 0755},
				},
			},
		},
	}

	result := Merge(nil, local)
	assert.Equal(t, "laptop", result.System.Machine.Name)
	assert.Len(t, result.Directory["repos"].Entries, 1)
}

func TestMergeOverrides(t *testing.T) {
	system := &Config{
		Version: Version,
		System: System{
			Machine: Machine{Name: "desktop", OS: "darwin"},
			Hooks:   Hooks{PostCreate: []string{"echo system"}},
		},
		Directory: map[string]DirCategory{
			"core": {
				Root: "~",
				Entries: []DirEntry{
					{Path: "~/.config", Mode: 0755},
				},
			},
		},
	}

	local := &Config{
		Version: Version,
		System: System{
			Machine: Machine{Name: "laptop", OS: "linux"},
			Hooks:   Hooks{PostCreate: []string{"echo local"}},
		},
		Directory: map[string]DirCategory{
			"core": {
				Root: "~",
				Entries: []DirEntry{
					{Path: "~/.config", Mode: 0700},
				},
			},
		},
	}

	result := Merge(system, local)
	assert.Equal(t, "laptop", result.System.Machine.Name)
	assert.Equal(t, "linux", result.System.Machine.OS)
	assert.Len(t, result.System.Hooks.PostCreate, 2)
	assert.Equal(t, uint32(0700), result.Directory["core"].Entries[0].Mode)
}

func TestMergeAddsCategory(t *testing.T) {
	system := &Config{
		Version: Version,
		System:  System{Hooks: Hooks{}},
		Directory: map[string]DirCategory{
			"core": {
				Root:    "~",
				Entries: []DirEntry{{Path: "~/.config", Mode: 0755}},
			},
		},
	}

	local := &Config{
		Version: Version,
		Directory: map[string]DirCategory{
			"repos": {
				Root:    "~/repos",
				Entries: []DirEntry{{Path: "~/repos/work", Mode: 0755}},
			},
		},
	}

	result := Merge(system, local)
	assert.Contains(t, result.Directory, "core")
	assert.Contains(t, result.Directory, "repos")
	assert.Len(t, result.Directory["core"].Entries, 1)
	assert.Len(t, result.Directory["repos"].Entries, 1)
}

func TestMergeAppendsHooks(t *testing.T) {
	system := &Config{
		Version: Version,
		System: System{
			Hooks: Hooks{PostCreate: []string{"echo system"}},
		},
		Directory: map[string]DirCategory{},
	}

	local := &Config{
		Version:   Version,
		Directory: map[string]DirCategory{},
		System: System{
			Hooks: Hooks{PostCreate: []string{"echo local"}},
		},
	}

	result := Merge(system, local)
	assert.Len(t, result.System.Hooks.PostCreate, 2)
	assert.Equal(t, "echo system", result.System.Hooks.PostCreate[0])
	assert.Equal(t, "echo local", result.System.Hooks.PostCreate[1])
}

func TestDiscoverConfigNoFiles(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	origWd, _ := os.Getwd()
	chdir := t.TempDir()
	require.NoError(t, os.Chdir(chdir))
	defer os.Chdir(origWd)

	cfg, err := DiscoverConfig(DefaultConfigPath)
	require.NoError(t, err)
	assert.Equal(t, Version, cfg.Version)
	assert.Contains(t, cfg.Directory, "core")
}

func TestDiscoverConfigSystemOnly(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".pregorc.yml")

	systemCfg := &Config{
		Version:   Version,
		General:   General{Color: true},
		System:    System{Machine: Machine{Name: "testhost", OS: "darwin"}},
		Directory: map[string]DirCategory{},
	}
	require.NoError(t, Save(cfgPath, systemCfg))

	cfg, err := DiscoverConfig(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, "testhost", cfg.System.Machine.Name)
}

func TestNewDefaultHasFields(t *testing.T) {
	cfg := NewDefault()
	assert.Equal(t, Version, cfg.Version)
	assert.True(t, cfg.General.Color)
	assert.Contains(t, cfg.Directory, "core")
	assert.Contains(t, cfg.Directory, "documents")
	assert.Contains(t, cfg.Directory, "repos")
	assert.NotEmpty(t, cfg.System.Machine.Name)
	assert.NotEmpty(t, cfg.System.Machine.OS)
}
