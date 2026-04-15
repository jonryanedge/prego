package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func validConfig() *Config {
	return &Config{
		Version: Version,
		Dirs: map[string]DirCategory{
			"core": {
				Root: "~",
				Entries: []DirEntry{
					{Path: "~/.config", Mode: 0700},
				},
			},
			"documents": {
				Root: "~/Documents",
				Entries: []DirEntry{
					{Path: "~/Documents/projects", Mode: 0755},
				},
			},
			"repos": {
				Root: "~/repos",
				Entries: []DirEntry{
					{Path: "~/repos/personal", Mode: 0755, VCS: "git"},
				},
			},
		},
	}
}

func TestValidateValid(t *testing.T) {
	cfg := validConfig()
	assert.NoError(t, Validate(cfg))
}

func TestValidateWrongVersion(t *testing.T) {
	cfg := validConfig()
	cfg.Version = 99
	assert.Error(t, Validate(cfg))
}

func TestValidateNoDirs(t *testing.T) {
	cfg := &Config{Version: Version, Dirs: map[string]DirCategory{}}
	assert.Error(t, Validate(cfg))
}

func TestValidateMissingRoot(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Dirs["core"]
	cat.Root = ""
	cfg.Dirs["core"] = cat
	assert.Error(t, Validate(cfg))
}

func TestValidateEmptyPath(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Dirs["core"]
	cat.Entries = append(cat.Entries, DirEntry{Path: ""})
	cfg.Dirs["core"] = cat
	assert.Error(t, Validate(cfg))
}

func TestValidateRelativePath(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Dirs["core"]
	cat.Entries = append(cat.Entries, DirEntry{Path: "no/slash/prefix"})
	cfg.Dirs["core"] = cat
	assert.Error(t, Validate(cfg))
}

func TestValidateAbsolutePath(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Dirs["core"]
	cat.Entries = append(cat.Entries, DirEntry{Path: "/absolute/is/ok", Mode: 0755})
	cfg.Dirs["core"] = cat
	assert.NoError(t, Validate(cfg))
}

func TestValidateInvalidMode(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Dirs["core"]
	cat.Entries = append(cat.Entries, DirEntry{Path: "~/bad", Mode: 0777 + 1})
	cfg.Dirs["core"] = cat
	assert.Error(t, Validate(cfg))
}

func TestValidateDuplicatePath(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Dirs["core"]
	cat.Entries = append(cat.Entries, DirEntry{Path: "~/.config", Mode: 0755})
	cfg.Dirs["core"] = cat
	assert.Error(t, Validate(cfg))
}

func TestValidateEmptySymlinkFrom(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Dirs["core"]
	cat.Symlinks = []Symlink{{From: "", To: "~/target"}}
	cfg.Dirs["core"] = cat
	assert.Error(t, Validate(cfg))
}

func TestValidateEmptySymlinkTo(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Dirs["core"]
	cat.Symlinks = []Symlink{{From: "~/source", To: ""}}
	cfg.Dirs["core"] = cat
	assert.Error(t, Validate(cfg))
}

func TestValidateEmptyHookCommand(t *testing.T) {
	cfg := validConfig()
	cfg.Hooks.PostCreate = []string{"   "}
	assert.Error(t, Validate(cfg))
}

func TestValidateValidHooks(t *testing.T) {
	cfg := validConfig()
	cfg.Hooks.PostCreate = []string{"chmod 700 ~/.ssh", "echo done"}
	assert.NoError(t, Validate(cfg))
}

func TestValidateZeroModeIsOK(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Dirs["core"]
	cat.Entries = append(cat.Entries, DirEntry{Path: "~/no-mode"})
	cfg.Dirs["core"] = cat
	assert.NoError(t, Validate(cfg))
}
