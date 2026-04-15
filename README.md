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
        remote: "https://github.com/youruser/personal.git"
      - path: "~/repos/work"
        mode: 0755
        vcs: git
        remote: "git@github.com:yourorg/work.git"
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
| `dirs.<category>.entries[].vcs` | string | no | VCS type, e.g. `git` (auto-detected by scan) |
| `dirs.<category>.entries[].remote` | string | no | Actual remote URL (e.g. `https://github.com/user/repo.git`) |
| `dirs.<category>.symlinks[]` | list | no | Symlink declarations (core only) |
| `hooks.post_create` | list | no | Shell commands to run after creation |

### Categories

| Category | Purpose | Typical Root |
|---|---|---|
| `core` | Dotfiles, configs, SSH/GPG dirs, symlinks | `~` |
| `documents` | Documents, notes, archives | `~/Documents` |
| `repos` | Code repositories | `~/repos` |

## Commands

### `prego scan`

Scan a directory tree and print discovered entries, or write them into the config file.

```bash
# Preview entries (dry run, prints to stdout)
prego scan ~/repos

# Limit depth
prego scan ~/repos -d 2

# Scan using a category root from the config
prego scan -C repos

# Write scanned entries into the config file
prego scan ~/repos -C repos --write

# Write to a specific config file
prego scan ~/repos -C repos --write -c ~/my-pregorc.yml
```

Without `--write`, scan only prints results. With `--write`, it merges entries into the config file, creating a new config if one doesn't exist yet. Duplicate entries are skipped. Git repositories are auto-detected: directories containing `.git` get `vcs: git` and their `origin` remote URL captured.

| Flag | Description |
|---|---|
| `-C`, `--category` | Config category to write into (default: `repos`) |
| `-d`, `--depth` | Max traversal depth (0 = unlimited, default: 0) |
| `--write` | Write scanned entries into the config file |

### `prego apply`

Create all directories and symlinks declared in the config. Does **not** clone git repos — use `build` for that. Idempotent — safe to run multiple times.

```bash
prego apply                  # create all directories and symlinks
prego apply --dry-run        # preview without making changes
prego apply -c /path/to/rc  # use a different config file
```

### `prego build`

Apply directory structure **and** clone git repos. Runs the same steps as `apply`, then clones any entries with `vcs: git` and a `remote` URL that don't already exist on disk. Skips repos that are already cloned and non-empty directories.

```bash
prego build                  # create dirs, symlinks, and clone repos
prego build --dry-run        # preview without making changes
prego build -c /path/to/rc  # use a different config file
```

### `prego diff`

Compare the config against the local filesystem. Reports missing directories, permission mismatches, and symlink drift.

```bash
prego diff                   # check for drift
prego diff -c /path/to/rc   # use a different config file
```

Exit code `0` if no drift, `1` if drift found.

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
│   ├── cmd/                    # Cobra commands (root, apply, check, diff, scan, version)
│   ├── config/                 # Config structs, Load, Save, Validate
│   ├── fs/                     # Filesystem operations (ops, scan, diff, vcs)
│   └── prompt/                 # Interactive prompts (init)
├── Makefile
├── PRD.md                      # Product requirements
├── PLAN.md                     # Build plan
├── go.mod
└── go.sum
```

## Development

### Build & Test

```bash
make            # lint + test + build
make test       # run tests with race detector
make build      # compile binary
make coverage   # generate HTML coverage report
make help        # show all targets
```

### Lint

```bash
make lint       # go vet + golangci-lint
```

## Roadmap

Planned commands not yet implemented:

| Command | Description |
|---|---|
| `prego init` | Generate config interactively by scanning the current machine |
| `prego add` | Add a directory entry to the config |
| `prego rm` | Remove a directory entry from the config |
| `prego list` | Print all tracked directories |

## License

MIT