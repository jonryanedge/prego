package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/jonryanedge/prego/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	resetFlags()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})
	err := rootCmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "prego v")
}

func TestCheckCommandWithValidConfig(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".pregorc.yml")

	cfg := &config.Config{
		Version: config.Version,
		System:  config.System{Hooks: config.Hooks{}},
		Directory: map[string]config.DirCategory{
			"core": {
				Root: "~",
				Entries: []config.DirEntry{
					{Path: "~/.config", Mode: 0700},
				},
			},
		},
	}
	require.NoError(t, config.Save(cfgPath, cfg))

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"-c", cfgPath, "check"})
	err := rootCmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "config is valid")
}

func TestCheckCommandWithInvalidConfig(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".pregorc.yml")

	content := []byte("version: 99\ndirectory:\n  core:\n    root: \"~\"\n    entries:\n      - path: \"~/.config\"\n")
	err := os.WriteFile(cfgPath, content, 0600)
	require.NoError(t, err)

	rootCmd.SetArgs([]string{"-c", cfgPath, "check"})
	err = rootCmd.Execute()
	assert.Error(t, err)
}

func TestCheckCommandNoFileReturnsDefault(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".pregorc.yml")

	rootCmd.SetArgs([]string{"-c", cfgPath, "check"})
	err := rootCmd.Execute()
	require.NoError(t, err)
}

func TestInitCommandSystem(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".pregorc.yml")

	rootCmd.SetArgs([]string{"-c", cfgPath, "init"})
	require.NoError(t, rootCmd.Execute())

	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, config.Version, cfg.Version)
	assert.Contains(t, cfg.Directory, "core")
}

func TestInitCommandAlreadyExists(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".pregorc.yml")
	require.NoError(t, os.WriteFile(cfgPath, []byte("exists"), 0600))

	rootCmd.SetArgs([]string{"-c", cfgPath, "init"})
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestInitCommandLocal(t *testing.T) {
	resetFlags()
	origDir, _ := os.Getwd()
	chdir := t.TempDir()
	require.NoError(t, os.Chdir(chdir))
	defer os.Chdir(origDir)

	rootCmd.SetArgs([]string{"-c", filepath.Join(chdir, "system.yml"), "init", "--local"})
	require.NoError(t, rootCmd.Execute())

	cfg, err := config.Load(filepath.Join(chdir, ".pregorc.yml"))
	require.NoError(t, err)
	assert.Equal(t, config.Version, cfg.Version)
}
