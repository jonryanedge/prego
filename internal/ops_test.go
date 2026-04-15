package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMkdirAllCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "foo", "bar")

	err := MkdirAll(path, 0755)
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

func TestMkdirAllCustomPerm(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret")

	err := MkdirAll(path, 0700)
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
}

func TestMkdirAllDefaultPerm(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "defaults")

	err := MkdirAll(path, 0)
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

func TestMkdirAllExistingDir(t *testing.T) {
	dir := t.TempDir()

	err := MkdirAll(dir, 0755)
	require.NoError(t, err)
}

func TestMkdirAllFileConflict(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file")
	require.NoError(t, os.WriteFile(path, []byte("hi"), 0600))

	err := MkdirAll(path, 0755)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestSymlinkCreates(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	require.NoError(t, os.MkdirAll(target, 0755))
	link := filepath.Join(dir, "link")

	err := Symlink(target, link)
	require.NoError(t, err)

	actual, err := os.Readlink(link)
	require.NoError(t, err)
	assert.Equal(t, target, actual)
}

func TestSymlinkIdempotent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	require.NoError(t, os.MkdirAll(target, 0755))
	link := filepath.Join(dir, "link")

	err := Symlink(target, link)
	require.NoError(t, err)

	err = Symlink(target, link)
	require.NoError(t, err)
}

func TestSymlinkConflictDifferentTarget(t *testing.T) {
	dir := t.TempDir()
	targetA := filepath.Join(dir, "targetA")
	targetB := filepath.Join(dir, "targetB")
	require.NoError(t, os.MkdirAll(targetA, 0755))
	require.NoError(t, os.MkdirAll(targetB, 0755))
	link := filepath.Join(dir, "link")

	err := Symlink(targetA, link)
	require.NoError(t, err)

	err = Symlink(targetB, link)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestSymlinkFileConflict(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	require.NoError(t, os.MkdirAll(target, 0755))
	link := filepath.Join(dir, "link")
	require.NoError(t, os.WriteFile(link, []byte("data"), 0600))

	err := Symlink(target, link)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a symlink")
}

func TestSymlinkCreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	require.NoError(t, os.MkdirAll(target, 0755))
	link := filepath.Join(dir, "sub", "dir", "link")

	err := Symlink(target, link)
	require.NoError(t, err)

	actual, err := os.Readlink(link)
	require.NoError(t, err)
	assert.Equal(t, target, actual)
}
