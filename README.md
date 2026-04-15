# prego

Save and reproduce directory structures across machines.

Prego captures your directory layout — core files, documents, and code repos — into a single dotfile that travels with you. Clone it on a new machine, run `prego apply`, and your familiar structure is back.

## Install

Build from source (requires Go 1.22+):

```bash
git clone https://github.com/jonryanedge/prego.git
cd prego
go build -o /usr/local/bin/prego ./cmd/prego
```

Or with `go install`:

```bash
go install github.com/jonryanedge/prego/cmd/prego@latest
```

## Quick Start

```bash
# validate an existing config
prego check

# validate a config at a custom path
prego check -c ~/my-pregorc.yml

# print version
prego version
```

## Config File

Prego reads `~/.pregorc.yml` by default. Override with `-c` on any command.

### Sample Config

```yaml
version: 1

machine:
  name: work-laptop
  os: darwin

dirs:
  core:
    root: "~"
    entries:
      - path: "~/.config"
        mode: 0700
      - path: "~/.local/bin"
        mode: 0700
      - path: "~/.ssh"
        mode: 0700
      - path: "~/.gnupg"
        mode: 0700
    symlinks:
      - from: "~/.config/git/config"
        to: "~/.dotfiles/git/config"

  documents:
    root: "~/Documents"
    entries:
      - path: "~/Documents/projects"
        mode: 0755
      - path: "~/Documents/notes"
        mode: 0755
      - path: "~/Documents/archive"
        mode: 0755

  repos:
    root: "~/repos"
    entries:
      - path: "~/repos/personal"
        mode: 0755
        vcs: git
        remote: "https://github.com/youruser"
      - path: "~/repos/work"
        mode: 0755
        vcs: git
        remote: "https://github.com/yourorg"
      - path: "~/repos/oss"
        mode: 0755
        vcs: git

hooks:
  post_create:
    - "chmod 700 ~/.ssh"
    - "chmod 700 ~/.gnupg"
```

### Config Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `version` | int | yes | Config schema version (currently `1`) |
| `machine.name` | string | no | Human-readable machine identifier |
| `machine.os` | string | no | `darwin`, `linux`, `windows` |
| `dirs.<category>.root` | string | yes | Base path for the category |
| `dirs.<category>.entries[]` | list | yes | Directory entries |
| `dirs.<category>.entries[].path` | string | yes | Absolute or `~/`-relative path |
| `dirs.<category>.entries[].mode` | octal | no | Unix permissions (default `0755`) |
| `dirs.<category>.entries[].vcs` | string | no | VCS type, e.g. `git` (repos only) |
| `dirs.<category>.entries[].remote` | string | no | Default remote URL pattern |
| `dirs.<category>.symlinks[]` | list | no | Symlink declarations (core only) |
| `hooks.post_create` | list | no | Shell commands to run after creation |

### Categories

| Category | Purpose | Typical Root |
|---|---|---|
| `core` | Dotfiles, configs, SSH/GPG dirs, symlinks | `~` |
| `documents` | Documents, notes, archives | `~/Documents` |
| `repos` | Code repositories | `~/repos` |

## Commands

### `prego check`

Validate the config file. Reports structural errors, invalid paths, duplicate entries, and bad modes.

```bash
prego check                  # uses ~/.pregorc.yml
prego check -c /path/to/rc  # custom path
```

Exit code `0` if valid, non-zero with error details if invalid.

### `prego version`

Print the current version.

## Global Flags

| Flag | Description |
|---|---|
| `-c`, `--config` | Path to config file (default `~/.pregorc.yml`) |
| `-h`, `--help` | Help for any command |

## Project Structure

```
prego/
├── cmd/prego/main.go          # Entrypoint
├── internal/
│   ├── cmd/                    # Cobra commands (root, check, version)
│   ├── config/                 # Config structs, Load, Save, Validate
│   ├── fs/                     # Filesystem operations (scan, diff, apply)
│   └── prompt/                 # Interactive prompts (init)
├── PRD.md                      # Product requirements
├── PLAN.md                     # Build plan
├── go.mod
└── go.sum
```

## Development

### Run Tests

```bash
go test ./... -v
```

### Build

```bash
go build -o bin/prego ./cmd/prego
```

### Lint

```bash
go vet ./...
```

## Roadmap

Planned commands not yet implemented:

| Command | Description |
|---|---|
| `prego init` | Generate config by scanning the current machine |
| `prego scan` | Walk a directory tree and output entries as YAML |
| `prego add` | Add a directory entry to the config |
| `prego rm` | Remove a directory entry from the config |
| `prego list` | Print all tracked directories |
| `prego apply` | Create all directories and symlinks from the config |
| `prego diff` | Compare local filesystem against the config |

## License

MIT