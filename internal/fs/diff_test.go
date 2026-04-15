package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jonryanedge/prego/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	dirs := []struct {
		path string
		mode os.FileMode
	}{
		{filepath.Join(dir, "dir1"), 0755},
		{filepath.Join(dir, "dir2"), 0700},
		{filepath.Join(dir, "dir3", "sub"), 0755},
	}
	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(d.path, d.mode))
		require.NoError(t, os.Chmod(d.path, d.mode))
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file1"), []byte("data"), 0644))

	linkTarget := filepath.Join(dir, "dir1")
	linkPath := filepath.Join(dir, "link1")
	require.NoError(t, os.Symlink(linkTarget, linkPath))

	return dir
}

func testConfig(dir string) *config.Config {
	return &config.Config{
		Version: config.Version,
		Dirs: map[string]config.DirCategory{
			"core": {
				Root: dir,
				Entries: []config.DirEntry{
					{Path: filepath.Join(dir, "dir1"), Mode: 0755},
					{Path: filepath.Join(dir, "dir2"), Mode: 0700},
					{Path: filepath.Join(dir, "missing"), Mode: 0755},
				},
				Symlinks: []config.Symlink{
					{From: filepath.Join(dir, "dir1"), To: filepath.Join(dir, "link1")},
					{From: filepath.Join(dir, "dir1"), To: filepath.Join(dir, "missing_link")},
				},
			},
		},
	}
}

func TestDiffMissingDir(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		Version: config.Version,
		Dirs: map[string]config.DirCategory{
			"core": {
				Root: dir,
				Entries: []config.DirEntry{
					{Path: filepath.Join(dir, "nope"), Mode: 0755},
				},
			},
		},
	}

	drifts := Diff(cfg)
	assert.NotEmpty(t, drifts)
	found := false
	for _, d := range drifts {
		if d.Type == MissingDir {
			found = true
		}
	}
	assert.True(t, found, "expected MissingDir drift")
}

func TestDiffModeMismatch(t *testing.T) {
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

	drifts := Diff(cfg)
	found := false
	for _, d := range drifts {
		if d.Type == ModeMismatch {
			found = true
			assert.Equal(t, "0700", d.Expected)
			assert.Equal(t, "0755", d.Actual)
		}
	}
	assert.True(t, found, "expected ModeMismatch drift")
}

func TestDiffNoDrift(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "perfect")
	require.NoError(t, os.MkdirAll(path, 0755))
	require.NoError(t, os.Chmod(path, 0755))

	cfg := &config.Config{
		Version: config.Version,
		Dirs: map[string]config.DirCategory{
			"core": {
				Root: dir,
				Entries: []config.DirEntry{
					{Path: path, Mode: 0755},
				},
			},
		},
	}

	drifts := Diff(cfg)
	assert.Empty(t, drifts)
}

func TestDiffFileInsteadOfDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file")
	require.NoError(t, os.WriteFile(path, []byte("data"), 0644))

	cfg := &config.Config{
		Version: config.Version,
		Dirs: map[string]config.DirCategory{
			"core": {
				Root: dir,
				Entries: []config.DirEntry{
					{Path: path, Mode: 0755},
				},
			},
		},
	}

	drifts := Diff(cfg)
	found := false
	for _, d := range drifts {
		if d.Type == ExtraDir && d.Actual == "file" {
			found = true
		}
	}
	assert.True(t, found, "expected ExtraDir drift with Actual=file")
}

func TestDiffSymlinkMissing(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		Version: config.Version,
		Dirs: map[string]config.DirCategory{
			"core": {
				Root:    dir,
				Entries: []config.DirEntry{},
				Symlinks: []config.Symlink{
					{From: "/nowhere", To: filepath.Join(dir, "nonexistent_link")},
				},
			},
		},
	}

	drifts := Diff(cfg)
	found := false
	for _, d := range drifts {
		if d.Type == LinkMissing {
			found = true
		}
	}
	assert.True(t, found, "expected LinkMissing drift")
}

func TestDiffSymlinkMismatch(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	require.NoError(t, os.MkdirAll(target, 0755))
	link := filepath.Join(dir, "link")
	require.NoError(t, os.Symlink("/wrong/target", link))

	cfg := &config.Config{
		Version: config.Version,
		Dirs: map[string]config.DirCategory{
			"core": {
				Root:    dir,
				Entries: []config.DirEntry{},
				Symlinks: []config.Symlink{
					{From: target, To: link},
				},
			},
		},
	}

	drifts := Diff(cfg)
	found := false
	for _, d := range drifts {
		if d.Type == LinkMismatch {
			found = true
		}
	}
	assert.True(t, found, "expected LinkMismatch drift")
}

func TestDiffSymlinkCorrect(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	require.NoError(t, os.MkdirAll(target, 0755))
	link := filepath.Join(dir, "link")
	require.NoError(t, os.Symlink(target, link))

	cfg := &config.Config{
		Version: config.Version,
		Dirs: map[string]config.DirCategory{
			"core": {
				Root:    dir,
				Entries: []config.DirEntry{},
				Symlinks: []config.Symlink{
					{From: target, To: link},
				},
			},
		},
	}

	drifts := Diff(cfg)
	assert.Empty(t, drifts)
}
