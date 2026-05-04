package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func validConfig() *Config {
	return &Config{
		Version: Version,
		System: System{
			Hooks: Hooks{},
		},
		Directory: map[string]DirCategory{
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
	cfg := &Config{Version: Version, Directory: map[string]DirCategory{}}
	assert.Error(t, Validate(cfg))
}

func TestValidateMissingRoot(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Directory["core"]
	cat.Root = ""
	cfg.Directory["core"] = cat
	assert.Error(t, Validate(cfg))
}

func TestValidateEmptyPath(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Directory["core"]
	cat.Entries = append(cat.Entries, DirEntry{Path: ""})
	cfg.Directory["core"] = cat
	assert.Error(t, Validate(cfg))
}

func TestValidateRelativePath(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Directory["core"]
	cat.Entries = append(cat.Entries, DirEntry{Path: "project", Mode: 0755})
	cfg.Directory["core"] = cat
	assert.NoError(t, Validate(cfg), "relative paths within a root should be valid")
}

func TestValidatePathEscapesRoot(t *testing.T) {
	cfg := &Config{
		Version: Version,
		System:  System{Hooks: Hooks{}},
		Directory: map[string]DirCategory{
			"repos": {
				Root:    ".",
				Entries: []DirEntry{{Path: "../etc/passwd", Mode: 0755}},
			},
		},
	}
	assert.Error(t, Validate(cfg), "paths escaping root should be invalid")
}

func TestValidateAbsolutePath(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Directory["core"]
	cat.Entries = append(cat.Entries, DirEntry{Path: "/absolute/is/ok", Mode: 0755})
	cfg.Directory["core"] = cat
	assert.NoError(t, Validate(cfg))
}

func TestValidateInvalidMode(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Directory["core"]
	cat.Entries = append(cat.Entries, DirEntry{Path: "~/bad", Mode: 0777 + 1})
	cfg.Directory["core"] = cat
	assert.Error(t, Validate(cfg))
}

func TestValidateDuplicatePath(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Directory["core"]
	cat.Entries = append(cat.Entries, DirEntry{Path: "~/.config", Mode: 0755})
	cfg.Directory["core"] = cat
	assert.Error(t, Validate(cfg))
}

func TestValidateEmptySymlinkFrom(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Directory["core"]
	cat.Symlinks = []Symlink{{From: "", To: "~/target"}}
	cfg.Directory["core"] = cat
	assert.Error(t, Validate(cfg))
}

func TestValidateEmptySymlinkTo(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Directory["core"]
	cat.Symlinks = []Symlink{{From: "~/source", To: ""}}
	cfg.Directory["core"] = cat
	assert.Error(t, Validate(cfg))
}

func TestValidateEmptyHookCommand(t *testing.T) {
	cfg := validConfig()
	cfg.System.Hooks.PostCreate = []string{"   "}
	assert.Error(t, Validate(cfg))
}

func TestValidateValidHooks(t *testing.T) {
	cfg := validConfig()
	cfg.System.Hooks.PostCreate = []string{"chmod 700 ~/.ssh", "echo done"}
	assert.NoError(t, Validate(cfg))
}

func TestValidateZeroModeIsOK(t *testing.T) {
	cfg := validConfig()
	cat := cfg.Directory["core"]
	cat.Entries = append(cat.Entries, DirEntry{Path: "~/no-mode"})
	cfg.Directory["core"] = cat
	assert.NoError(t, Validate(cfg))
}
