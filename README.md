# repomni

test

A CLI tool that injects shared configuration files into multiple repository clones, keeping them invisible to git.

When working with multiple branch copies of the same repo (e.g., running parallel AI agents across branches), you need the same `.envrc`, `.env`, Claude skills, and hooks in every clone. Repomni symlinks them from a single source directory so you configure once and every clone stays in sync.

## What it does

- Symlinks (or copies) files from a central source into target repos
- Hides injected files from git using `.git/info/exclude` (no tracked files modified)
- Supports batch operations across all repos in a directory
- Track workflow state per branch (`active`, `review`, `closed`, `paused`) with color-coded overview
- Interactive TUI for configuration, simple CLI for daily use

### Injected items (defaults)

| Source path | Target path in repo | Type |
|---|---|---|
| `skills/` | `.claude/skills/` | directory |
| `hooks.json` | `.claude/hooks.json` | file |
| `.envrc` | `.envrc` | file |
| `.env` | `.env` | file |

Source paths are configurable via `repomni config global`.

## Installation

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- Git

### From source (recommended)

```sh
git clone https://github.com/ezerfernandes/repomni.git
cd repomni
sudo make install
```

This builds the binary and installs it to `/usr/local/bin/repomni`.

To install to a different location:

```sh
make install PREFIX=$HOME/.local
```

This puts the binary in `~/.local/bin/repomni` — make sure `~/.local/bin` is in your `PATH`.

### With `go install`

```sh
go install github.com/ezerfernandes/repomni@latest
```

This installs to `$GOPATH/bin` (usually `~/go/bin`). Make sure it's in your `PATH`.

### Verify installation

```sh
repomni --version
```

### Uninstall

```sh
# If installed with make install
sudo make uninstall

# If installed with go install
rm $(which repomni)
```

## Quick start

### 1. Set up your source directory

Create a directory with the files you want to share across repos:

```
my-agent-config/
  skills/
    my-skill.md
  hooks.json
  .envrc
  .env
```

### 2. Configure repomni

```sh
repomni config global
```

This walks you through an interactive wizard to set the source directory, injection mode, and which items to inject. Configuration is saved to `~/.config/repomni/config.yaml`.

For scripted setup:

```sh
repomni config global --non-interactive --source /path/to/my-agent-config
```

### 3. Inject into a repo

```sh
cd /path/to/my-repo-clone
repomni inject
```

Or inject into all repos under a parent directory:

```sh
repomni inject --all /path/to/my-clones
```

### 4. Check status

```sh
repomni status --all /path/to/my-clones
```

Output:

```
Repository: /path/to/my-clones/feature-a
  Item                   Present   Current   Excluded
  ─────────────────────  ────────  ────────  ────────
  .claude/skills         Yes       Yes       Yes
  .claude/hooks.json     Yes       Yes       Yes
  .envrc                 Yes       Yes       Yes
  .env                   Yes       Yes       Yes
```

### 5. Clean up when done

```sh
repomni eject /path/to/my-clones/feature-a
```

Or eject from all at once:

```sh
repomni eject --all /path/to/my-clones
```

## Commands

| Command | Description |
|---|---|
| `inject [target]` | Inject files into target repo (default: current dir) |
| `eject [target]` | Remove injected files from target repo |
| `status [target]` | Show injection or git status of target repo(s) |
| `list [dir]` | List git repos in a directory (for scripting) |
| **branch** | |
| `branch create <name>` | Clone parent repo and create a new branch |
| `branch clone <name>` | Clone parent repo and check out an existing remote branch |
| `branch list [dir]` | List branch repos with their workflow states |
| `branch set-state <state>` | Set workflow state for the current branch repo |
| **sync** | |
| `sync [dir]` | Pull code updates and refresh PR/MR status |
| `sync code [dir]` | Pull updates for all repos in a directory |
| `sync state [dir]` | Update workflow states from PR/MR status |
| **config** | |
| `config global` | Interactive global setup wizard |
| `config repo` | Configure injection settings for the current repository |
| `config script` | Manage per-repo setup scripts |

### `config global`

Interactively configure the source directory, injection mode, and items. Saved to `~/.config/repomni/config.yaml`.

```sh
repomni config global
repomni config global --source ~/shared-config
repomni config global --source ~/shared-config --non-interactive
```

| Flag | Description |
|---|---|
| `--source` | Source directory (skip interactive prompt) |
| `--non-interactive` | Use defaults without prompting (requires `--source`) |

### `config repo`

Interactively select which items and directory entries to inject into the current repository. Saved to `.git/repomni/config.yaml`.

```sh
repomni config repo
```

### `inject`

Symlink or copy configured files into the target repo. Injected paths are added to `.git/info/exclude`.

```sh
repomni inject
repomni inject /path/to/repo
repomni inject --all
repomni inject --dry-run
repomni inject --force
```

| Flag | Description |
|---|---|
| `--all` | Inject into all git repos under the target directory |
| `--dry-run` | Show what would be done without making changes |
| `--force` | Overwrite existing regular files |
| `--copy` | Use copy mode for this run |
| `--symlink` | Use symlink mode for this run |

### `eject`

Remove all injected files/symlinks and clean up `.git/info/exclude`.

```sh
repomni eject
repomni eject --all
```

| Flag | Description |
|---|---|
| `--all` | Eject from all git repos under the target directory |

### `status`

Show injection status or git sync status for target repos.

```sh
repomni status
repomni status --all
repomni status --git
repomni status --json
```

| Flag | Description |
|---|---|
| `--all` | Check all git repos under the target directory |
| `--git` | Show git sync status (branch, ahead/behind, dirty) |
| `--no-fetch` | Skip `git fetch` when checking git status |
| `--json` | Output as JSON |

### `branch create`

Clone the parent repository and check out a new branch. Automatically injects configured files, runs the setup script, and sets state to `active`.

```sh
repomni branch create my-feature
repomni branch create my-feature --no-inject
```

| Flag | Description |
|---|---|
| `--no-inject` | Skip automatic injection into the new branch |

### `branch clone`

Clone the parent repository and check out an existing remote branch. Directory name is derived from the branch name (e.g., `feature/my-thing` becomes `feature-my-thing`).

```sh
repomni branch clone feature/my-thing
repomni branch clone feature/my-thing --no-inject
```

| Flag | Description |
|---|---|
| `--no-inject` | Skip automatic injection into the new clone |

### `branch set-state`

Set a workflow state for the current repo. Stored in `.git/repomni/config.yaml`.

Predefined states: `active`, `review`, `approved`, `review-blocked`, `merged`, `closed`, `paused`. Custom states are also accepted (lowercase letters, digits, hyphens).

```sh
repomni branch set-state active
repomni branch set-state review
repomni branch set-state closed
repomni branch set-state my-custom-state
repomni branch set-state --clear
```

| Flag | Description |
|---|---|
| `--clear` | Remove the workflow state |

### `branch list`

List branch repos with state, git branch, and dirty status. States are color-coded: `active` (green), `review` (yellow), `closed` (red), `paused` (blue).

```sh
repomni branch list
repomni branch list /path/to/branches
repomni branch list --state review
repomni branch list --json
```

| Flag | Description |
|---|---|
| `--state` | Filter by workflow state |
| `--json` | Output as JSON |

Example output:

```
  Name           Branch          State    Dirty
  ─────────────  ──────────────  ───────  ─────
  feat-auth      feat-auth       active
  fix-bug-123    fix-bug-123     review   *
  old-feature    old-feature     closed
```

### `sync`

Run `sync code` and `sync state` together. First pulls git updates for all repos, then queries GitHub/GitLab for PR/MR status changes and updates workflow states.

```sh
repomni sync
repomni sync --dry-run
repomni sync --autostash -j 4
repomni sync --json
```

| Flag | Description |
|---|---|
| `--dry-run` | Show what would be done without making changes |
| `--autostash` | Stash dirty working trees before pull |
| `-j`, `--jobs` | Number of parallel sync workers (default: 1) |
| `--no-fetch` | Skip `git fetch` (local status only) |
| `--strategy` | Pull strategy: `ff-only` (default), `rebase`, `merge` |
| `--json` | Output as JSON |

### `sync code`

Fetch and pull updates for all repos under a directory. Dirty repos are skipped unless `--autostash` is used. Diverged repos are always skipped.

```sh
repomni sync code
repomni sync code --dry-run
repomni sync code --autostash
repomni sync code -j 4
repomni sync code --strategy rebase
repomni sync code --json
```

| Flag | Description |
|---|---|
| `--dry-run` | Show what would be done without pulling |
| `--autostash` | Stash dirty working trees before pull |
| `-j`, `--jobs` | Number of parallel sync workers (default: 1) |
| `--no-fetch` | Skip `git fetch` (local status only) |
| `--strategy` | Pull strategy: `ff-only` (default), `rebase`, `merge` |
| `--json` | Output as JSON |

### `sync state`

Query GitHub or GitLab for PR/MR status and update workflow states. Only repos with a stored merge URL and a review-related state (`review`, `approved`, `review-blocked`) are checked.

Requires `gh` (GitHub) or `glab` (GitLab) to be installed and authenticated.

```sh
repomni sync state
repomni sync state --dry-run
repomni sync state --json
```

| Flag | Description |
|---|---|
| `--dry-run` | Show what would change without updating configs |
| `--json` | Output as JSON |

### `list`

List git repositories under a directory, one per line. Useful for piping to other tools.

```sh
repomni list
repomni list --names
```

| Flag | Description |
|---|---|
| `--names` | Output directory names only instead of full paths |

### `config script`

Interactively create or edit a setup script that runs when you create a new branch with `repomni branch create`. Stored in `.git` and never committed.

```sh
repomni config script
```

## How it works

- **Symlink mode** (default): Creates symbolic links from target paths to source files. Changes to source are instantly reflected in all targets.
- **Copy mode**: Copies files from source to target. Targets are independent snapshots.
- **Git exclusion**: All injected paths are added to `.git/info/exclude` inside a managed block. This file is local to each clone and never tracked by git, so `git status` stays clean.
- **Worktree support**: Works with both regular git repos and `git worktree` directories.

## Development

```sh
make build    # Build to bin/repomni
make test     # Run all tests
make run      # Build and run
make clean    # Remove build artifacts
```
