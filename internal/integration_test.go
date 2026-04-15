package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jonryanedge/prego/internal/config"
	"github.com/jonryanedge/prego/internal/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestConfig(t *testing.T, dir string) string {
	t.Helper()
	cfgPath := filepath.Join(dir, ".pregorc.yml")

	cfg := &config.Config{
		Version: config.Version,
		Machine: config.Machine{Name: "test", OS: "darwin"},
		Dirs: map[string]config.DirCategory{
			"core": {
				Root: "~",
				Entries: []config.DirEntry{
					{Path: filepath.Join(dir, "core", ".config"), Mode: 0700},
					{Path: filepath.Join(dir, "core", ".ssh"), Mode: 0700},
				},
				Symlinks: []config.Symlink{
					{From: filepath.Join(dir, "core", ".ssh"), To: filepath.Join(dir, "core", "ssh-link")},
				},
			},
			"documents": {
				Root: filepath.Join(dir, "docs"),
				Entries: []config.DirEntry{
					{Path: filepath.Join(dir, "docs", "projects"), Mode: 0755},
					{Path: filepath.Join(dir, "docs", "notes"), Mode: 0755},
				},
			},
		},
		Hooks: config.Hooks{},
	}

	err := config.Save(cfgPath, cfg)
	require.NoError(t, err)
	return cfgPath
}

func TestApplyCreatesDirs(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	cfgPath := writeTestConfig(t, dir)

	rootCmd.SetArgs([]string{"-c", cfgPath, "apply"})
	err := rootCmd.Execute()
	require.NoError(t, err)

	assert.DirExists(t, filepath.Join(dir, "core", ".config"))
	assert.DirExists(t, filepath.Join(dir, "core", ".ssh"))
	assert.DirExists(t, filepath.Join(dir, "docs", "projects"))
	assert.DirExists(t, filepath.Join(dir, "docs", "notes"))
}

func TestApplyIdempotent(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	cfgPath := writeTestConfig(t, dir)

	rootCmd.SetArgs([]string{"-c", cfgPath, "apply"})
	require.NoError(t, rootCmd.Execute())

	rootCmd.SetArgs([]string{"-c", cfgPath, "apply"})
	require.NoError(t, rootCmd.Execute())
}

func TestApplyDryRun(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	cfgPath := writeTestConfig(t, dir)

	rootCmd.SetArgs([]string{"-c", cfgPath, "apply", "--dry-run"})
	require.NoError(t, rootCmd.Execute())

	_, err := os.Stat(filepath.Join(dir, "core", ".config"))
	assert.True(t, os.IsNotExist(err), "dry-run should not create directories")
}

func TestApplyCreatesSymlinks(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	cfgPath := writeTestConfig(t, dir)

	rootCmd.SetArgs([]string{"-c", cfgPath, "apply"})
	require.NoError(t, rootCmd.Execute())

	link := filepath.Join(dir, "core", "ssh-link")
	target, err := os.Readlink(link)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "core", ".ssh"), target)
}

func TestApplySetsPermissions(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	cfgPath := writeTestConfig(t, dir)

	rootCmd.SetArgs([]string{"-c", cfgPath, "apply"})
	require.NoError(t, rootCmd.Execute())

	info, err := os.Stat(filepath.Join(dir, "core", ".ssh"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
}

func TestDiffNoDriftAfterApply(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	cfgPath := writeTestConfig(t, dir)

	rootCmd.SetArgs([]string{"-c", cfgPath, "apply"})
	require.NoError(t, rootCmd.Execute())

	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)

	drifts := fs.Diff(cfg)
	assert.Empty(t, drifts, "expected no drift after apply, got %d drifts", len(drifts))
}

func TestDiffDetectsMissingDirs(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	cfgPath := writeTestConfig(t, dir)

	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)

	drifts := fs.Diff(cfg)
	assert.NotEmpty(t, drifts, "expected drift for missing directories")

	found := false
	for _, d := range drifts {
		if d.Type == fs.MissingDir {
			found = true
		}
	}
	assert.True(t, found, "expected MissingDir drift")
}

func TestDiffDetectsModeMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "somedir")
	require.NoError(t, os.MkdirAll(path, 0755))

	cfg := &config.Config{
		Version: config.Version,
		Dirs: map[string]config.DirCategory{
			"core": {
				Root: dir,
				Entries: []config.DirEntry{
					{Path: path, Mode: 0700},
				},
			},
		},
	}

	drifts := fs.Diff(cfg)
	found := false
	for _, d := range drifts {
		if d.Type == fs.ModeMismatch {
			found = true
		}
	}
	assert.True(t, found, "expected ModeMismatch drift")
}

func TestScanCommand(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub1"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub2"), 0755))

	rootCmd.SetArgs([]string{"scan", dir})
	require.NoError(t, rootCmd.Execute())
}

func TestScanWithCategory(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs", "projects"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs", "notes"), 0755))

	cfgPath := filepath.Join(dir, ".pregorc.yml")
	cfg := &config.Config{
		Version: config.Version,
		Dirs: map[string]config.DirCategory{
			"documents": {
				Root:    filepath.Join(dir, "docs"),
				Entries: []config.DirEntry{},
			},
		},
	}
	require.NoError(t, config.Save(cfgPath, cfg))

	rootCmd.SetArgs([]string{"-c", cfgPath, "scan", "--category", "documents", "--depth", "1"})
	require.NoError(t, rootCmd.Execute())
}

func TestScanNoRootError(t *testing.T) {
	resetFlags()
	rootCmd.SetArgs([]string{"scan"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}
