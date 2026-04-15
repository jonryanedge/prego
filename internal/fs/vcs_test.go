package fs

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsGitRepo(t *testing.T) {
	dir := t.TempDir()
	assert.False(t, IsGitRepo(dir), "temp dir should not be a git repo")
}

func TestIsGitRepoTrue(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	dir := t.TempDir()
	cmd := exec.Command("git", "init", dir)
	require.NoError(t, cmd.Run(), "git init should succeed")
	assert.True(t, IsGitRepo(dir), "initialized dir should be a git repo")
}

func TestGitRemoteURL(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	dir := t.TempDir()
	cmd := exec.Command("git", "init", dir)
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", dir, "remote", "add", "origin", "https://github.com/user/repo.git")
	require.NoError(t, cmd.Run())

	url := GitRemoteURL(dir)
	assert.Equal(t, "https://github.com/user/repo.git", url)
}

func TestGitRemoteURLEmpty(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	dir := t.TempDir()
	cmd := exec.Command("git", "init", dir)
	require.NoError(t, cmd.Run())

	url := GitRemoteURL(dir)
	assert.Empty(t, url, "repo with no remote should return empty string")
}

func TestDetectVCSGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	dir := t.TempDir()
	cmd := exec.Command("git", "init", dir)
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", dir, "remote", "add", "origin", "git@github.com:user/repo.git")
	require.NoError(t, cmd.Run())

	vcs, remote := DetectVCS(dir)
	assert.Equal(t, "git", vcs)
	assert.Equal(t, "git@github.com:user/repo.git", remote)
}

func TestDetectVCSNone(t *testing.T) {
	dir := t.TempDir()
	vcs, remote := DetectVCS(dir)
	assert.Empty(t, vcs)
	assert.Empty(t, remote)
}
