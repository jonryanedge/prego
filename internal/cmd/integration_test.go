package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jonryanedge/prego/internal/config"
	"github.com/jonryanedge/prego/internal/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func execLookPath(name string) (string, error) {
	return exec.LookPath(name)
}

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

func TestScanWriteToNewConfig(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "repos", "personal"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "repos", "work"), 0755))

	cfgPath := filepath.Join(dir, ".pregorc.yml")

	rootCmd.SetArgs([]string{"-c", cfgPath, "scan", dir, "--category", "repos", "--write"})
	require.NoError(t, rootCmd.Execute())

	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)
	require.NoError(t, config.Validate(cfg))

	assert.Contains(t, cfg.Dirs, "repos")
	found := false
	for _, e := range cfg.Dirs["repos"].Entries {
		if filepath.Base(e.Path) == "personal" {
			found = true
		}
	}
	assert.True(t, found, "expected 'personal' entry in repos category")
}

func TestScanWriteMergesExisting(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "repos", "personal"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "repos", "work"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "repos", "oss"), 0755))

	cfgPath := filepath.Join(dir, ".pregorc.yml")
	personalPath := filepath.Join(dir, "repos", "personal")

	cfg := &config.Config{
		Version: config.Version,
		Machine: config.Machine{Name: "test", OS: "darwin"},
		Dirs: map[string]config.DirCategory{
			"repos": {
				Root:    dir,
				Entries: []config.DirEntry{{Path: personalPath, Mode: 0755}},
			},
		},
		Hooks: config.Hooks{},
	}
	require.NoError(t, config.Save(cfgPath, cfg))

	rootCmd.SetArgs([]string{"-c", cfgPath, "scan", dir, "--category", "repos", "--write"})
	require.NoError(t, rootCmd.Execute())

	loaded, err := config.Load(cfgPath)
	require.NoError(t, err)

	personalCount := 0
	hasWork := false
	hasOss := false
	for _, e := range loaded.Dirs["repos"].Entries {
		if e.Path == personalPath {
			personalCount++
		}
		if filepath.Base(e.Path) == "work" {
			hasWork = true
		}
		if filepath.Base(e.Path) == "oss" {
			hasOss = true
		}
	}
	assert.Equal(t, 1, personalCount, "existing entry should not be duplicated")
	assert.True(t, hasWork, "work entry should be added")
	assert.True(t, hasOss, "oss entry should be added")
}

func TestScanWriteNoDuplicates(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "repos", "personal"), 0755))

	cfgPath := filepath.Join(dir, ".pregorc.yml")

	rootCmd.SetArgs([]string{"-c", cfgPath, "scan", dir, "--category", "repos", "--write"})
	require.NoError(t, rootCmd.Execute())

	resetFlags()
	rootCmd.SetArgs([]string{"-c", cfgPath, "scan", dir, "--category", "repos", "--write"})
	require.NoError(t, rootCmd.Execute())

	loaded, err := config.Load(cfgPath)
	require.NoError(t, err)
	count := 0
	for _, e := range loaded.Dirs["repos"].Entries {
		if filepath.Base(e.Path) == "personal" {
			count++
		}
	}
	assert.Equal(t, 1, count, "should not duplicate entries on repeated writes")
}

func TestScanWriteWithoutFlagDoesNotModify(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "repos", "personal"), 0755))

	cfgPath := filepath.Join(dir, ".pregorc.yml")
	_, err := os.Stat(cfgPath)
	require.True(t, os.IsNotExist(err))

	rootCmd.SetArgs([]string{"-c", cfgPath, "scan", dir, "--depth", "1"})
	require.NoError(t, rootCmd.Execute())

	_, err = os.Stat(cfgPath)
	assert.True(t, os.IsNotExist(err), "scan without --write should not create config file")
}

func TestScanWriteDetectsGitRepo(t *testing.T) {
	if _, err := execLookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	resetFlags()
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "my-repo")
	require.NoError(t, os.MkdirAll(repoDir, 0755))

	cmd := exec.Command("git", "init", repoDir)
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", repoDir, "remote", "add", "origin", "https://github.com/user/repo.git")
	require.NoError(t, cmd.Run())

	cfgPath := filepath.Join(dir, ".pregorc.yml")

	rootCmd.SetArgs([]string{"-c", cfgPath, "scan", dir, "--category", "repos", "--write"})
	require.NoError(t, rootCmd.Execute())

	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)

	found := false
	for _, e := range cfg.Dirs["repos"].Entries {
		if filepath.Base(e.Path) == "my-repo" {
			assert.Equal(t, "git", e.VCS)
			assert.Equal(t, "https://github.com/user/repo.git", e.Remote)
			found = true
		}
	}
	assert.True(t, found, "expected my-repo entry with VCS info")
}

func TestScanWriteGitRepoNoRemote(t *testing.T) {
	if _, err := execLookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	resetFlags()
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "bare-repo")
	require.NoError(t, os.MkdirAll(repoDir, 0755))

	cmd := exec.Command("git", "init", repoDir)
	require.NoError(t, cmd.Run())

	cfgPath := filepath.Join(dir, ".pregorc.yml")

	rootCmd.SetArgs([]string{"-c", cfgPath, "scan", dir, "--category", "repos", "--write"})
	require.NoError(t, rootCmd.Execute())

	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)

	found := false
	for _, e := range cfg.Dirs["repos"].Entries {
		if filepath.Base(e.Path) == "bare-repo" {
			assert.Equal(t, "git", e.VCS)
			assert.Empty(t, e.Remote, "repo with no remote should have empty remote")
			found = true
		}
	}
	assert.True(t, found, "expected bare-repo entry")
}

func TestBuildCreatesDirsAndClones(t *testing.T) {
	if _, err := execLookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	resetFlags()
	dir := t.TempDir()

	localRepo := filepath.Join(dir, "source-repo")
	require.NoError(t, os.MkdirAll(localRepo, 0755))
	cmd := exec.Command("git", "init", localRepo)
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", localRepo, "remote", "add", "origin", "https://example.com/repo.git")
	require.NoError(t, cmd.Run())

	commitFile := filepath.Join(localRepo, "README.md")
	require.NoError(t, os.WriteFile(commitFile, []byte("hello"), 0644))
	cmd = exec.Command("git", "-C", localRepo, "add", ".")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", localRepo, "commit", "-m", "init")
	require.NoError(t, cmd.Run())

	cloneTarget := filepath.Join(dir, "repos", "cloned-repo")
	targetDir := filepath.Join(dir, "repos")

	cfgPath := filepath.Join(dir, ".pregorc.yml")
	cfg := &config.Config{
		Version: config.Version,
		Machine: config.Machine{Name: "test", OS: "darwin"},
		Dirs: map[string]config.DirCategory{
			"core": {
				Root: dir,
				Entries: []config.DirEntry{
					{Path: filepath.Join(dir, "configs"), Mode: 0755},
				},
			},
			"repos": {
				Root: targetDir,
				Entries: []config.DirEntry{
					{Path: cloneTarget, Mode: 0755, VCS: "git", Remote: localRepo},
				},
			},
		},
		Hooks: config.Hooks{},
	}
	require.NoError(t, config.Save(cfgPath, cfg))

	rootCmd.SetArgs([]string{"-c", cfgPath, "build"})
	require.NoError(t, rootCmd.Execute())

	assert.DirExists(t, filepath.Join(dir, "configs"))
	assert.DirExists(t, cloneTarget)
	assert.True(t, fs.IsGitRepo(cloneTarget), "cloned directory should be a git repo")
}

func TestBuildDryRun(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".pregorc.yml")

	cfg := &config.Config{
		Version: config.Version,
		Machine: config.Machine{Name: "test", OS: "darwin"},
		Dirs: map[string]config.DirCategory{
			"core": {
				Root: dir,
				Entries: []config.DirEntry{
					{Path: filepath.Join(dir, "drydir"), Mode: 0755},
				},
			},
		},
		Hooks: config.Hooks{},
	}
	require.NoError(t, config.Save(cfgPath, cfg))

	rootCmd.SetArgs([]string{"-c", cfgPath, "build", "--dry-run"})
	require.NoError(t, rootCmd.Execute())

	_, err := os.Stat(filepath.Join(dir, "drydir"))
	assert.True(t, os.IsNotExist(err), "dry-run should not create directories")
}

func TestBuildSkipsExistingRepo(t *testing.T) {
	if _, err := execLookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	resetFlags()
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "repos", "existing")
	require.NoError(t, os.MkdirAll(repoDir, 0755))

	cmd := exec.Command("git", "init", repoDir)
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", repoDir, "remote", "add", "origin", "https://example.com/repo.git")
	require.NoError(t, cmd.Run())

	cfgPath := filepath.Join(dir, ".pregorc.yml")
	cfg := &config.Config{
		Version: config.Version,
		Machine: config.Machine{Name: "test", OS: "darwin"},
		Dirs: map[string]config.DirCategory{
			"repos": {
				Root:    filepath.Join(dir, "repos"),
				Entries: []config.DirEntry{{Path: repoDir, Mode: 0755, VCS: "git", Remote: "https://example.com/repo.git"}},
			},
		},
		Hooks: config.Hooks{},
	}
	require.NoError(t, config.Save(cfgPath, cfg))

	rootCmd.SetArgs([]string{"-c", cfgPath, "build"})
	require.NoError(t, rootCmd.Execute())

	assert.True(t, fs.IsGitRepo(repoDir), "existing git repo should remain intact")
}
