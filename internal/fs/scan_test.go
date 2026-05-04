package fs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func execLookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func TestScanBasic(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "a", "b"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "c"), 0755))

	result, err := Scan(dir, 0)
	require.NoError(t, err)
	assert.Len(t, result.Entries, 3)

	paths := make(map[string]bool)
	for _, e := range result.Entries {
		paths[e.Path] = true
	}
	assert.True(t, paths[filepath.Join(dir, "a")])
	assert.True(t, paths[filepath.Join(dir, "a", "b")])
	assert.True(t, paths[filepath.Join(dir, "c")])
}

func TestScanDepth1(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "a", "b", "c"), 0755))

	result, err := Scan(dir, 1)
	require.NoError(t, err)

	hasA := false
	hasAB := false
	hasABC := false
	for _, e := range result.Entries {
		if filepath.Base(e.Path) == "a" {
			hasA = true
		}
		if filepath.Base(e.Path) == "b" {
			hasAB = true
		}
		if filepath.Base(e.Path) == "c" {
			hasABC = true
		}
	}
	assert.True(t, hasA)
	assert.False(t, hasAB)
	assert.False(t, hasABC)
}

func TestScanDepth2(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "a", "b", "c"), 0755))

	result, err := Scan(dir, 2)
	require.NoError(t, err)

	hasA := false
	hasAB := false
	hasABC := false
	for _, e := range result.Entries {
		if filepath.Base(e.Path) == "a" {
			hasA = true
		}
		if filepath.Base(e.Path) == "b" {
			hasAB = true
		}
		if filepath.Base(e.Path) == "c" {
			hasABC = true
		}
	}
	assert.True(t, hasA)
	assert.True(t, hasAB)
	assert.False(t, hasABC)
}

func TestScanUnlimitedDepth(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "a", "b", "c"), 0755))

	result, err := Scan(dir, 0)
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, e := range result.Entries {
		names[filepath.Base(e.Path)] = true
	}
	assert.True(t, names["a"])
	assert.True(t, names["b"])
	assert.True(t, names["c"])
}

func TestScanEmpty(t *testing.T) {
	dir := t.TempDir()

	result, err := Scan(dir, 0)
	require.NoError(t, err)
	assert.Empty(t, result.Entries)
}

func TestScanNonexistent(t *testing.T) {
	_, err := Scan("/nonexistent/path/that/does/not/exist", 0)
	assert.Error(t, err)
}

func TestScanSkipsFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "subdir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hi"), 0644))

	result, err := Scan(dir, 0)
	require.NoError(t, err)
	assert.Len(t, result.Entries, 1)
	assert.Equal(t, filepath.Join(dir, "subdir"), result.Entries[0].Path)
}

func TestScanModeCaptured(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "restricted")
	require.NoError(t, os.MkdirAll(sub, 0700))

	result, err := Scan(dir, 0)
	require.NoError(t, err)
	assert.Len(t, result.Entries, 1)
	assert.Equal(t, uint32(0700), result.Entries[0].Mode)
}

func TestScanSkipsGitInternals(t *testing.T) {
	if _, err := execLookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	dir := t.TempDir()
	repo := filepath.Join(dir, "my-repo")
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "src"), 0755))
	cmd := exec.Command("git", "init", repo)
	require.NoError(t, cmd.Run())

	result, err := Scan(dir, 0)
	require.NoError(t, err)

	for _, e := range result.Entries {
		assert.NotContains(t, e.Path, ".git", "should not include .git directories")
	}

	found := false
	for _, e := range result.Entries {
		if filepath.Base(e.Path) == "my-repo" {
			found = true
			assert.Equal(t, "git", e.VCS)
		}
	}
	assert.True(t, found, "should include the repo directory itself")
}

func TestScanStopsAtGitRepoBoundary(t *testing.T) {
	if _, err := execLookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	dir := t.TempDir()
	repo := filepath.Join(dir, "my-repo")
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "deep", "nested"), 0755))
	cmd := exec.Command("git", "init", repo)
	require.NoError(t, cmd.Run())

	result, err := Scan(dir, 0)
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, e := range result.Entries {
		names[filepath.Base(e.Path)] = true
	}
	assert.True(t, names["my-repo"], "should include repo directory")
	assert.False(t, names["deep"], "should not include subdirectories inside a git repo")
	assert.False(t, names["nested"], "should not include nested directories inside a git repo")
}

func TestScanReportsIgnoredEntries(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "src"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "node_modules", "pkg"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "build"), 0755))

	nosauceContent := `node_modules
build
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".nosauce"), []byte(nosauceContent), 0644))

	result, err := Scan(dir, 0)
	require.NoError(t, err)

	assert.Len(t, result.Ignored, 2, "should report 2 ignored entries")

	ignoredNames := make(map[string]string)
	for _, ig := range result.Ignored {
		ignoredNames[filepath.Base(ig.Path)] = ig.Pattern
	}
	assert.Equal(t, "node_modules", ignoredNames["node_modules"])
	assert.Equal(t, "build", ignoredNames["build"])
	assert.Equal(t, ".nosauce", result.Ignored[0].Source)

	names := make(map[string]bool)
	for _, e := range result.Entries {
		names[filepath.Base(e.Path)] = true
	}
	assert.True(t, names["src"])
}
