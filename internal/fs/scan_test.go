package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanBasic(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "a", "b"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "c"), 0755))

	entries, err := Scan(dir, 0)
	require.NoError(t, err)
	assert.Len(t, entries, 3)

	paths := make(map[string]bool)
	for _, e := range entries {
		paths[e.Path] = true
	}
	assert.True(t, paths[filepath.Join(dir, "a")])
	assert.True(t, paths[filepath.Join(dir, "a", "b")])
	assert.True(t, paths[filepath.Join(dir, "c")])
}

func TestScanDepth1(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "a", "b", "c"), 0755))

	entries, err := Scan(dir, 1)
	require.NoError(t, err)

	hasA := false
	hasAB := false
	hasABC := false
	for _, e := range entries {
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

	entries, err := Scan(dir, 2)
	require.NoError(t, err)

	hasA := false
	hasAB := false
	hasABC := false
	for _, e := range entries {
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

	entries, err := Scan(dir, 0)
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, e := range entries {
		names[filepath.Base(e.Path)] = true
	}
	assert.True(t, names["a"])
	assert.True(t, names["b"])
	assert.True(t, names["c"])
}

func TestScanEmpty(t *testing.T) {
	dir := t.TempDir()

	entries, err := Scan(dir, 0)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestScanNonexistent(t *testing.T) {
	_, err := Scan("/nonexistent/path/that/does/not/exist", 0)
	assert.Error(t, err)
}

func TestScanSkipsFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "subdir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hi"), 0644))

	entries, err := Scan(dir, 0)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, filepath.Join(dir, "subdir"), entries[0].Path)
}

func TestScanModeCaptured(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "restricted")
	require.NoError(t, os.MkdirAll(sub, 0700))

	entries, err := Scan(dir, 0)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, uint32(0700), entries[0].Mode)
}
