# jrnl-md 2.0 Pass 3: Documentation + Polish — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Write complete user-facing documentation (README.md, docs/config.md, Cobra help strings) and mark Pass 2 done in CLAUDE.md.
**Architecture:** All tasks are documentation writes or in-source string edits; no logic changes. README.md and docs/config.md content is specified in full below so the executing agent can Write the files directly. Cobra help strings are specified with exact field values per subcommand.
**Tech Stack:** markdown, git

---

## Tasks

- [ ] **Task 1** — Write `README.md` (content below)
- [ ] **Task 2** — Write `docs/config.md` (content below)
- [ ] **Task 3** — Update Cobra help strings in source files (strings below)
- [ ] **Task 4** — Mark Pass 2 Done in `CLAUDE.md`
- [ ] **Task 5** — Commit all documentation changes

---

## Task 1 — Write `README.md`

Write the following content verbatim to `/home/glw907/Projects/jrnl-md/README.md`:

```markdown
# jrnl-md

A focused CLI for keeping a journal as plain markdown files. One file per day.
Inspired by [jrnl](https://jrnl.sh), built for people who want their notes to
be readable text files first and foremost — no lock-in, no special format, just
a directory of `YYYY/MM/DD.md` files you can grep, edit, sync, and read in any
markdown viewer.

## Installation

```sh
go install github.com/glw907/jrnl-md/cmd/jrnl-md@latest
```

Requires Go 1.21+. The binary is self-contained with no runtime dependencies.

## Quick Start

**Write an entry:**

```sh
jrnl-md write Went for a run this morning. Feeling good.
```

Appends to today's day file (`~/Documents/Journal/2026/04/06.md`). Creates the
file with a day heading if it doesn't exist. Adds a timestamp heading by default.

**List recent entries:**

```sh
jrnl-md list          # last 10 days (configurable)
jrnl-md list -5       # last 5 days
jrnl-md list @work    # days tagged @work
jrnl-md list --short  # one line per day
```

**Edit today's entry:**

```sh
jrnl-md edit
jrnl-md edit --on 2026-04-01
```

Opens the day file in your configured editor. Creates the file with a day
heading if it doesn't exist.

## Subcommand Reference

### `write`

```sh
jrnl-md write <text>
```

Append text to today's day file. The text is everything after `write` on the
command line — no quoting required.

```sh
jrnl-md write Met with @sarah about the project timeline.
```

With timestamps enabled (default), each write gets a `## HH:MM AM/PM` heading.
With timestamps disabled, consecutive writes are separated by a blank line.

### `list`

```sh
jrnl-md list [flags] [@tag...]
```

Display day files matching the given filters. Defaults to the last
`default_list_count` days (10 by default).

| Flag | Description |
|---|---|
| `-N` | Last N days (e.g. `-5` for last 5) |
| `--all` | Show all days |
| `--short` | One line per day: date + first content line |
| `--from <date>` | Days on or after date |
| `--to <date>` | Days on or before date |
| `--on <date>` | A single day |
| `-year <year>` | All days in a given year |
| `-month <month>` | All days in a given month across all years |
| `-day <N>` | All days on day-of-month N across all months |
| `-today-in-history` | Today's date in all prior years |
| `--and` | Require all specified @tags (default: any) |
| `--not <tag>` | Exclude days containing @tag |
| `--contains <text>` | Days whose body contains text (case-insensitive) |

Dates accept natural language: `yesterday`, `last monday`, `3 days ago`, as
well as `YYYY-MM-DD` and common formats.

Examples:

```sh
jrnl-md list @work @sarah --and   # days tagged both @work and @sarah
jrnl-md list --from "last monday" --to today
jrnl-md list -year 2025 --short
jrnl-md list -today-in-history
jrnl-md list --contains "budget question"
```

### `tags`

```sh
jrnl-md tags [date-filter-flags]
```

List all tags with frequency counts, sorted descending. Accepts the same date
filter flags as `list` (`--from`, `--to`, `--on`, `-year`, `-month`, `-day`,
`-today-in-history`).

```sh
jrnl-md tags
# @work: 42
# @sarah: 17
# @reading: 9

jrnl-md tags -year 2025
```

### `edit`

```sh
jrnl-md edit [--on <date>]
```

Open a day file in your editor. Defaults to today. `--on` selects a specific
date. If the day file doesn't exist, it is created with a day heading before
the editor opens.

```sh
jrnl-md edit
jrnl-md edit --on yesterday
jrnl-md edit --on 2026-01-15
```

### `completion`

```sh
jrnl-md completion {bash,zsh,fish,powershell}
```

Generate shell completion scripts. Follow the output instructions to install.

```sh
# bash
jrnl-md completion bash > /etc/bash_completion.d/jrnl-md

# zsh
jrnl-md completion zsh > "${fpath[1]}/_jrnl-md"

# fish
jrnl-md completion fish > ~/.config/fish/completions/jrnl-md.fish
```

## Configuration

Config file: `~/.config/jrnl-md/config.toml`

Auto-created with defaults on first run. Override location with `--config-file`.

```toml
[general]
editor = ""               # falls back to $VISUAL then $EDITOR
timestamps = true         # add ## time headings on write
linewrap = 79             # wrap body text at N columns in list output
default_list_count = 10   # how many days 'list' shows with no -N flag

[format]
time = "03:04 PM"         # Go time format for timestamp headings
date = "2006-01-02"       # Go time format for date headings
tag_symbols = "@"         # prefix character(s) that mark tags

[colors]
date = "none"             # day heading color
body = "none"             # body text color
tags = "none"             # tag highlight color
# color values: black, red, green, yellow, blue, magenta, cyan, white, none

[journals.default]
path = "~/Documents/Journal/"   # where day files are stored
```

See [docs/config.md](docs/config.md) for the full configuration reference.

## Day File Format

Files are stored at `JOURNALS_PATH/YYYY/MM/DD.md`.

With timestamps enabled (default):

```markdown
# 2026-04-06 Sunday

## 09:00 AM

Went for a morning run. Feeling good about the week ahead.

## 02:30 PM

Met with @sarah about the project timeline. Need to follow
up on the budget question.
```

With timestamps disabled:

```markdown
# 2026-04-06 Sunday

Went for a morning run. Feeling good about the week ahead.

Second write of the day, appended with a blank line separator.
```

- Line 1 is always `# YYYY-MM-DD Weekday`
- Timestamp headings are `## time` (format configurable)
- Consecutive writes are separated by a blank line
- Tags are `@word` in body text — no special handling on write
- Files are plain `.md` — readable in any editor or markdown viewer

## Philosophy

**The day is the atom.** jrnl-md stores one file per calendar day and treats
that file as the unit of work. Every operation loads only what it needs.

The format is deliberately ordinary. Your journal files are readable without
jrnl-md — open them in neovim, VS Code, Obsidian, or any markdown viewer. Sync
them with git, rsync, or Syncthing. Search them with grep. The tool gets out of
the way.

Unix does the rest. jrnl-md does not include export, encryption, templates, or
web viewers. Those are solved problems. Use the tools that already exist.
```

---

## Task 2 — Write `docs/config.md`

Write the following content verbatim to `/home/glw907/Projects/jrnl-md/docs/config.md`:

```markdown
# jrnl-md Configuration Reference

Config file: `~/.config/jrnl-md/config.toml`

Auto-created with defaults on first run. Override the path at runtime:

```sh
jrnl-md --config-file /path/to/config.toml list
```

---

## Full Example Config

```toml
[general]
editor = ""
timestamps = true
linewrap = 79
default_list_count = 10

[format]
time = "03:04 PM"
date = "2006-01-02"
tag_symbols = "@"

[colors]
date = "none"
body = "none"
tags = "cyan"

[journals.default]
path = "~/Documents/Journal/"
```

---

## `[general]`

### `editor`

| | |
|---|---|
| Type | string |
| Default | `""` |

Path or name of the external editor to use for `jrnl-md edit`. Supports any
editor that accepts a file path and optional `+N` line-number argument (vim,
neovim, nano, micro, etc.).

**Resolution order:**

1. `editor` in config (if non-empty)
2. `$VISUAL` environment variable
3. `$EDITOR` environment variable
4. Error: no editor configured

Examples:

```toml
editor = "nvim"
editor = "nvim-journal"      # custom wrapper script
editor = "/usr/bin/nano"     # absolute path
```

### `timestamps`

| | |
|---|---|
| Type | bool |
| Default | `true` |

When `true`, each `write` appends a `## time` heading (formatted per
`[format].time`) before the body text. When `false`, consecutive writes are
separated by a blank line with no heading.

```toml
timestamps = false
```

### `linewrap`

| | |
|---|---|
| Type | int |
| Default | `79` |

Column width at which body text is wrapped in `list` output. Set to `0` to
disable wrapping.

```toml
linewrap = 72
```

### `default_list_count`

| | |
|---|---|
| Type | int |
| Default | `10` |

Number of days shown by `jrnl-md list` when no `-N` or `--all` flag is given.

```toml
default_list_count = 7
```

---

## `[format]`

### `time`

| | |
|---|---|
| Type | string (Go time layout) |
| Default | `"03:04 PM"` |

Format string for timestamp headings in day files. Uses Go's reference time
(`Mon Jan 2 15:04:05 MST 2006`). The value is used both when writing timestamp
headings and when detecting timestamp headings in `list --short` output.

Common values:

| Format string | Example output |
|---|---|
| `"03:04 PM"` | `09:30 AM` (default) |
| `"15:04"` | `09:30` |
| `"3:04pm"` | `9:30am` |

### `date`

| | |
|---|---|
| Type | string (Go time layout) |
| Default | `"2006-01-02"` |

Format string for the date portion of day headings. The weekday name is always
appended after the formatted date (e.g., `# 2026-04-06 Sunday`).

### `tag_symbols`

| | |
|---|---|
| Type | string |
| Default | `"@"` |

One or more characters that prefix a tag. A tag is a run of `tag_symbols` +
word characters (`[a-zA-Z0-9_]`). The default `@` means tags look like
`@work`, `@sarah`.

To support multiple symbols, list them all: `tag_symbols = "@#"` matches both
`@work` and `#work`.

---

## `[colors]`

Controls ANSI color highlighting in `list` and `tags` output. Each key accepts
one of the named colors below or `"none"` to disable.

| Key | What it colors |
|---|---|
| `date` | Day heading lines (`# YYYY-MM-DD Weekday`) |
| `body` | Body text |
| `tags` | Tag occurrences in body text |

**Color values:**

`black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, `none`

Colors map to standard ANSI 8-color codes and respect your terminal's color
scheme. `none` outputs no ANSI codes for that element.

Example — subtle highlighting:

```toml
[colors]
date = "blue"
tags = "cyan"
body = "none"
```

---

## `[journals.default]`

### `path`

| | |
|---|---|
| Type | string |
| Default | `"~/Documents/Journal/"` |

Filesystem path to the journal root directory. Tilde expansion is supported.
jrnl-md creates this directory (and any `YYYY/MM/` subdirectories) on first use.

```toml
[journals.default]
path = "~/Documents/Journal/"
```

The `[journals.default]` table name anticipates future multi-journal support.
In 2.0 only `default` is used.

---

## Notes

- **File extension** is always `.md` and is not configurable.
- **`default_hour` / `default_minute`** from jrnl 1.x are removed. Writes
  always use the current system time.
- **`highlight`** toggle from 1.x is removed. Tag highlighting is always on
  when a color is configured (and always off when `tags = "none"`).
- **`[journals.default].path`** is the only journal in 2.0. The schema shape
  supports future multi-journal work without config migration.
```

---

## Task 3 — Update Cobra Help Strings

For each source file, update the `Use`, `Short`, `Long`, and `Example` fields
of the Cobra command struct. The file locations are under `cmd/jrnl-md/`.

### `write.go`

```go
Use:   "write <text>",
Short: "Append text to today's day file",
Long: `Append text to today's day file.

Creates the file with a day heading if it does not exist. With timestamps
enabled (the default), each write gets a ## time heading before the body.
With timestamps disabled, consecutive writes are separated by a blank line.

The text argument is everything after "write" on the command line.`,
Example: `  jrnl-md write Went for a morning run. Feeling good.
  jrnl-md write Met with @sarah about the project timeline.`,
```

### `list.go`

```go
Use:   "list [flags] [@tag...]",
Short: "Display day files matching filters",
Long: `Display day files matching the given filters.

Defaults to the last default_list_count days (10 by default). Positional
@tag arguments filter to days containing those tags. Date arguments accept
natural language (yesterday, last monday, 3 days ago) as well as YYYY-MM-DD.

Body text is wrapped at the configured linewrap width. Tags are highlighted
in color when a color is configured.`,
Example: `  jrnl-md list                          # last 10 days
  jrnl-md list -5                        # last 5 days
  jrnl-md list --all                     # all days
  jrnl-md list --short                   # one line per day
  jrnl-md list @work                     # days tagged @work
  jrnl-md list @work @sarah --and        # days tagged both
  jrnl-md list --from "last monday"
  jrnl-md list --on 2026-04-01
  jrnl-md list -year 2025
  jrnl-md list -today-in-history
  jrnl-md list --contains "budget question"`,
```

### `edit.go`

```go
Use:   "edit [--on <date>]",
Short: "Open a day file in your editor",
Long: `Open a day file in your editor.

Defaults to today. --on selects a specific date. If the day file does not
exist, it is created with a day heading before the editor opens.

The editor is resolved from the config file, then $VISUAL, then $EDITOR.`,
Example: `  jrnl-md edit
  jrnl-md edit --on yesterday
  jrnl-md edit --on 2026-01-15`,
```

### `tags.go`

```go
Use:   "tags [date-filter-flags]",
Short: "List all tags with frequency counts",
Long: `List all tags with frequency counts, sorted descending.

Accepts the same date filter flags as list: --from, --to, --on, -year,
-month, -day, -today-in-history.`,
Example: `  jrnl-md tags
  jrnl-md tags -year 2025
  jrnl-md tags --from "last month"`,
```

### `completion.go`

The `completion` subcommand is generated by Cobra (`rootCmd.GenBashCompletion`
etc. or `cobra.GenBashCompletion`). If the project uses
`rootCmd.AddCommand(completionCmd)` with a manually defined `completionCmd`,
set:

```go
Use:   "completion [bash|zsh|fish|powershell]",
Short: "Generate shell completion scripts",
Long: `Generate shell completion scripts for the specified shell.

Follow the instructions in the output to install the completion script
for your shell.`,
Example: `  # bash
  jrnl-md completion bash > /etc/bash_completion.d/jrnl-md

  # zsh
  jrnl-md completion zsh > "${fpath[1]}/_jrnl-md"

  # fish
  jrnl-md completion fish > ~/.config/fish/completions/jrnl-md.fish`,
```

If the project uses Cobra's built-in `completionCmd` (added automatically when
`rootCmd.CompletionOptions` is set), update the `Short` on the root command to
reflect that completion is available, and leave the generated subcommand as-is.

### Root command (`main.go`)

```go
Use:   "jrnl-md",
Short: "A focused CLI for a markdown day-file journal",
Long: `jrnl-md keeps a journal as plain markdown files — one file per calendar day.

Write entries, list and filter by date or tag, and edit day files directly
in your editor. Your journal is a directory of readable .md files.

Configuration: ~/.config/jrnl-md/config.toml (auto-created on first run)
Day files:     JOURNAL_PATH/YYYY/MM/DD.md`,
```

---

## Task 4 — Mark Pass 2 Done in `CLAUDE.md`

In `/home/glw907/Projects/jrnl-md/CLAUDE.md`, find the Implementation Passes
table (or equivalent section tracking pass status) and update Pass 2 from
`Pending` (or `In Progress`) to `Done`. Update Pass 3 from `Pending` to
`In Progress` (or `Done` if this plan is being executed as Pass 3).

If CLAUDE.md was rewritten for 2.0 in Pass 1 and no longer has a passes table,
add a brief status block:

```markdown
## Implementation Status

| Pass | Status |
|---|---|
| Pass 1: Infrastructure | Done |
| Pass 2: Core Build | Done |
| Pass 3: Documentation + Polish | Done |
```

---

## Task 5 — Commit

Stage and commit all documentation changes:

```sh
git add README.md docs/config.md CLAUDE.md
git add cmd/jrnl-md/write.go cmd/jrnl-md/list.go cmd/jrnl-md/edit.go
git add cmd/jrnl-md/tags.go cmd/jrnl-md/main.go
# add completion.go if it exists and was modified
git commit -m "docs: add README, config reference, and Cobra help strings for 2.0

Co-Authored-By: Claude <noreply@anthropic.com>"
```

Do not push. Verify with `git log --oneline -3` that the commit is present.
