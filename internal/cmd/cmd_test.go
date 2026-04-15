package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})
	err := rootCmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "prego v")
}

func TestCheckCommandWithValidConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".pregorc.yml")

	content := []byte("version: 1\ndirs:\n  core:\n    root: \"~\"\n    entries:\n      - path: \"~/.config\"\n        mode: 0700\n")
	err := os.WriteFile(cfgPath, content, 0600)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"-c", cfgPath, "check"})
	err = rootCmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "config is valid")
}

func TestCheckCommandWithInvalidConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".pregorc.yml")

	content := []byte("version: 99\ndirs:\n  core:\n    root: \"~\"\n    entries:\n      - path: \"~/.config\"\n")
	err := os.WriteFile(cfgPath, content, 0600)
	require.NoError(t, err)

	rootCmd.SetArgs([]string{"-c", cfgPath, "check"})
	err = rootCmd.Execute()
	assert.Error(t, err)
}

func TestCheckCommandMissingFile(t *testing.T) {
	rootCmd.SetArgs([]string{"-c", "/nonexistent/.pregorc.yml", "check"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}
