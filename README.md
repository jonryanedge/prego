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

Prego reads `~/.pregorc.yml` as the system config and `.pregorc.yml` in the current directory as a local config. The local config overrides the system config (entries with the same path are replaced, hooks are appended). If neither exists, prego uses sensible defaults. Override the system config path with `-c` on any command.

### `prego init`

Create a config file with default values:

```bash
prego init                    # create ~/.pregorc.yml (system config)
prego init --local            # create .pregorc.yml in current directory
prego init -c ~/my-config.yml # create config at custom path
```

### Sample Config

```yaml
version: 2

general:
  color: true
  verbose: false

system:
  machine:
    name: work-laptop
    os: darwin
  hooks:
    post_create:
      - "chmod 700 ~/.ssh"
      - "chmod 700 ~/.gnupg"

directory:
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
| `version` | int | yes | Config schema version (currently `2`) |
| `general.color` | bool | no | Enable colored output (default: true) |
| `general.verbose` | bool | no | Enable verbose output — shows additional detail about filesystem operations, skipped items, and ignored entries (default: false) |
| `system.machine.name` | string | no | Human-readable machine identifier |
| `system.machine.os` | string | no | `darwin`, `linux`, `windows` |
| `system.hooks.post_create` | list | no | Shell commands to run after creation / build |
| `directory.<category>.root` | string | yes | Base path for the category. Use `.` in a local config to mean "the directory containing this config file" |
| `directory.<category>.entries[]` | list | yes | Directory entries |
| `directory.<category>.entries[].path` | string | yes | Absolute path, `~/`-relative path, or path relative to the category root |
| `directory.<category>.entries[].mode` | octal | no | Unix permissions (default `0755`) |
| `directory.<category>.entries[].vcs` | string | no | VCS type, e.g. `git` (auto-detected by scan) |
| `directory.<category>.entries[].remote` | string | no | Remote URL for cloning (e.g. `https://github.com/user/repo.git`) |
| `directory.<category>.symlinks[]` | list | no | Symlink declarations (core only) |

#### `general.color`

Controls whether prego uses ANSI color codes in its terminal output. When `true` (the default), commands like `apply`, `build`, `diff`, and `scan` use color to distinguish created, skipped, and error items. Set to `false` to disable color — useful when piping output to a file or running in environments that don't support ANSI codes. This setting is inherited from the system config unless the local config overrides it.

#### `general.verbose`

When `true`, prego prints additional detail during operations:

- **`scan`** — shows each entry that was ignored by a `.nosauce` file, including which pattern matched
- **`apply` / `build`** — shows skipped entries and symlink targets that already exist
- **`diff`** — shows detailed drift information including expected vs actual modes

When `false` (the default), only created items and errors are shown. This setting is inherited from the system config unless the local config overrides it.

### Path Resolution

Entry paths can be written in three styles:

| Style | Example | Where it resolves to |
|---|---|---|
| `~/`-prefixed | `~/repos/project` | `$HOME/repos/project` |
| Absolute | `/opt/data/logs` | `/opt/data/logs` |
| Relative to root | `project` | `<root>/project` |

In a **system config** (`~/.pregorc.yml`), entries typically use `~/` paths:

```yaml
directory:
  repos:
    root: ~/repos
    entries:
      - path: ~/repos/my-project
```

In a **local config** (`.pregorc.yml`), entries use paths relative to `root`, and `root` is set to `.` (meaning the directory containing the config file):

```yaml
directory:
  repos:
    root: .
    entries:
      - path: my-project
      - path: another-repo
        vcs: git
        remote: https://github.com/user/repo.git
```

When `prego apply` or `prego build` runs, it resolves relative entry paths by joining them with the resolved root. A local config with `root: .` resolves to the config file's parent directory, so the config is completely portable — copy the project folder to another machine and the same `.pregorc.yml` works without edits.

### Config Hierarchy

Prego loads configs in this order (later overrides earlier):

1. **System config**: `~/.pregorc.yml` (or path specified by `-c`)
2. **Local config**: `.pregorc.yml` in the current directory

When merging:
- `general` settings: local overrides system
- `system.machine`: local overrides system
- `system.hooks.post_create`: local appends to system
- `directory`: categories merge; entries with the same path are overridden by local

If no config files exist, prego uses sensible defaults and works without error.

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

# Write scanned entries into the system config file
prego scan ~/repos -C repos --write

# Write to a local .pregorc.yml (uses relative paths and root: .)
prego scan . --write --local

# Write to a specific config file
prego scan ~/repos -C repos --write -c ~/my-pregorc.yml
```

Without `--write`, scan only prints results. With `--write`, it merges entries into the config file, creating a new config if one doesn't exist yet. Duplicate entries are skipped. Git repositories are auto-detected: directories containing `.git` get `vcs: git` and their `origin` remote URL captured.

When `--local --write` is used, scan writes a local `.pregorc.yml` with `root: .` and relative entry paths, replacing the entire category rather than merging. This makes the config portable — commit it alongside your project and it works on any machine.

| Flag | Description |
|---|---|
| `-C`, `--category` | Config category to write into (default: `repos`) |
| `-d`, `--depth` | Max traversal depth (0 = unlimited, default: 0) |
| `--write` | Write scanned entries into the config file |
| `--local` | Write to `.pregorc.yml` in current directory (requires `--write`) |

### `.nosauce` files

Place a `.nosauce` file in any directory to prevent `prego scan` from descending into matching subdirectories. The format follows `.gitignore` conventions:

```gitignore
# this is a comment
node_modules      # exact name match at any depth
build/            # trailing / matches directories only
*.pyc             # glob patterns
dist/output       # relative path from this .nosauce file
**/temp           # match "temp" at any depth below this directory
```

- Patterns apply to the directory tree below the `.nosauce` file
- `#` lines and blank lines are ignored
- Trailing `/` means "match directories only" (the same as `.gitignore`)
- `**/` prefix matches at any nesting depth
- `.nosauce` files are hierarchical — each directory can have its own

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
| `prego add` | Add a directory entry to the config |
| `prego rm` | Remove a directory entry from the config |
| `prego list` | Print all tracked directories |

## License

MIT