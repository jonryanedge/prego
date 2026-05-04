package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNosauceFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".nosauce")

	content := `# comment
node_modules
build
.cache

*.pyc
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	patterns, err := parseNosauceFile(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"node_modules", "build", ".cache", "*.pyc"}, patterns)
}

func TestParseNosauceFileMissing(t *testing.T) {
	patterns, err := parseNosauceFile("/nonexistent/.nosauce")
	assert.Error(t, err)
	assert.Nil(t, patterns)
}

func TestParseNosauceFileEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".nosauce")

	require.NoError(t, os.WriteFile(path, []byte(""), 0644))

	patterns, err := parseNosauceFile(path)
	require.NoError(t, err)
	assert.Nil(t, patterns)
}

func TestParsePattern(t *testing.T) {
	tests := []struct {
		raw       string
		base      string
		isDirOnly bool
		isNegated bool
		hasSlash  bool
	}{
		{"node_modules", "node_modules", false, false, false},
		{"staging/", "staging", true, false, false},
		{"!keepme", "keepme", false, true, false},
		{"build/output", "build/output", false, false, true},
		{"docs/", "docs", true, false, false},
		{"**/temp", "**/temp", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			pp := parsePattern(tt.raw)
			assert.Equal(t, tt.base, pp.base)
			assert.Equal(t, tt.isDirOnly, pp.isDirOnly)
			assert.Equal(t, tt.isNegated, pp.isNegated)
			assert.Equal(t, tt.hasSlash, pp.hasSlash)
		})
	}
}

func TestMatchNamePattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		entry   string
		match   bool
	}{
		{"exact match", "node_modules", "node_modules", true},
		{"no match", "node_modules", "src", false},
		{"glob star", "*.pyc", "foo.pyc", true},
		{"glob no match", "*.pyc", "foo.js", false},
		{"single char", "cache?", "cache1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchNamePattern(tt.entry, tt.pattern)
			assert.Equal(t, tt.match, result)
		})
	}
}

func TestShouldIgnore(t *testing.T) {
	dir := t.TempDir()
	rules := []ignoreRule{
		{dir: dir, patterns: []string{"node_modules", ".cache", "build"}},
	}

	tests := []struct {
		name   string
		path   string
		expect bool
	}{
		{"ignored dir", filepath.Join(dir, "node_modules"), true},
		{"not ignored", filepath.Join(dir, "src"), false},
		{"ignored dotcache", filepath.Join(dir, ".cache"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIgnore(tt.path, true, rules)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestShouldIgnoreTrailingSlash(t *testing.T) {
	dir := t.TempDir()
	rules := []ignoreRule{
		{dir: dir, patterns: []string{"staging/"}},
	}

	tests := []struct {
		name   string
		path   string
		isDir  bool
		expect bool
	}{
		{"dir matches staging/", filepath.Join(dir, "staging"), true, true},
		{"file does not match staging/", filepath.Join(dir, "staging"), false, false},
		{"other dir not matched", filepath.Join(dir, "other"), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIgnore(tt.path, tt.isDir, rules)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestShouldIgnoreRelativePath(t *testing.T) {
	dir := t.TempDir()
	rules := []ignoreRule{
		{dir: dir, patterns: []string{"build/output", "dist/assets"}},
	}

	tests := []struct {
		name   string
		path   string
		expect bool
	}{
		{"exact relative match", filepath.Join(dir, "build", "output"), true},
		{"partial match no", filepath.Join(dir, "build"), false},
		{"other relative", filepath.Join(dir, "dist", "assets"), true},
		{"no match", filepath.Join(dir, "other"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIgnore(tt.path, true, rules)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestScanIgnoreNosauce(t *testing.T) {
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

	names := make(map[string]bool)
	for _, e := range result.Entries {
		names[filepath.Base(e.Path)] = true
	}
	assert.True(t, names["src"], "src should be included")
	assert.False(t, names["node_modules"], "node_modules should be ignored")
	assert.False(t, names["build"], "build should be ignored")
}

func TestScanIgnoreNosauceWithTrailingSlash(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "staging", "subdir"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "production", "subdir"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "src"), 0755))

	nosauceContent := `staging/
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".nosauce"), []byte(nosauceContent), 0644))

	result, err := Scan(dir, 0)
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, e := range result.Entries {
		names[e.Path] = true
	}
	assert.True(t, names[filepath.Join(dir, "src")], "src should be included")
	assert.True(t, names[filepath.Join(dir, "production")], "production should be included")
	assert.False(t, names[filepath.Join(dir, "staging")], "staging should be ignored by trailing /")
	assert.False(t, names[filepath.Join(dir, "staging", "subdir")], "subdirs of staging should not appear")

	ignored := make(map[string]bool)
	for _, ig := range result.Ignored {
		ignored[filepath.Base(ig.Path)] = true
	}
	assert.True(t, ignored["staging"], "staging should appear in ignored list")
}

func TestScanIgnoreNosauceNested(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "project", "node_modules", "pkg"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "project", "src"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "project", ".nosauce"), []byte{}, 0644))

	nosauceContent := `node_modules
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "project", ".nosauce"), []byte(nosauceContent), 0644))

	result, err := Scan(dir, 0)
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, e := range result.Entries {
		names[filepath.Base(e.Path)] = true
	}
	assert.True(t, names["project"], "project dir should be included")
	assert.True(t, names["src"], "src should be included")
	assert.False(t, names["node_modules"], "nested node_modules should be ignored")
}

func TestScanNosauceSkipsGlob(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "dist"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "src"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".cache"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "data"), 0755))

	nosauceContent := `.cache
d*
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".nosauce"), []byte(nosauceContent), 0644))

	result, err := Scan(dir, 0)
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, e := range result.Entries {
		names[filepath.Base(e.Path)] = true
	}
	assert.True(t, names["src"], "src should be included")
	assert.False(t, names["dist"], "dist should match d* and be ignored")
	assert.False(t, names[".cache"], ".cache should be ignored")
	assert.False(t, names["data"], "data should match d* and be ignored")
}

func TestScanNosauceDoesNotCreateEntry(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "src"), 0755))

	require.NoError(t, os.WriteFile(filepath.Join(dir, ".nosauce"), []byte("src\n"), 0644))

	result, err := Scan(dir, 0)
	require.NoError(t, err)
	assert.Empty(t, result.Entries, "entries matching .nosauce patterns should not be scanned")
}

func TestScanNosauceDoubleStar(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "a", "temp"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "b", "temp"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "src"), 0755))

	nosauceContent := `**/temp
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".nosauce"), []byte(nosauceContent), 0644))

	result, err := Scan(dir, 0)
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, e := range result.Entries {
		names[filepath.Base(e.Path)] = true
	}
	assert.True(t, names["src"])
	assert.True(t, names["a"])
	assert.True(t, names["b"])
	assert.False(t, names["temp"], "**/temp should match temp at any depth")
}
