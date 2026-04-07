# jrnl-md

A focused CLI for keeping a journal as plain markdown files. One file per day.
Inspired by [jrnl](https://jrnl.sh), built for people who want their notes to
be readable text files first and foremost — no lock-in, no special format, just
a directory of `YYYY/MM/YYYY-MM-DD.md` files you can grep, edit, sync, and read
in any markdown viewer.

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

Appends to today's day file (`~/Documents/Journal/2026/04/2026-04-06.md`).
Creates the file with a day heading if it doesn't exist. Adds a timestamp
heading by default.

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
| `--year <year>` | All days in a given year |
| `--month <month>` | All days in a given month across all years |
| `--day <N>` | All days on day-of-month N across all months |
| `--today-in-history` | Today's date in all prior years |
| `--and` | Require all specified @tags (default: any) |
| `--not <tag>` | Exclude days containing @tag |
| `--contains <text>` | Days whose body contains text (case-insensitive) |

Dates accept natural language: `yesterday`, `last monday`, `3 days ago`, as
well as `YYYY-MM-DD` and common formats.

Examples:

```sh
jrnl-md list @work @sarah --and   # days tagged both @work and @sarah
jrnl-md list --from "last monday" --to today
jrnl-md list --year 2025 --short
jrnl-md list --today-in-history
jrnl-md list --contains "budget question"
```

### `tags`

```sh
jrnl-md tags [date-filter-flags]
```

List all tags with frequency counts, sorted descending. Accepts the same date
filter flags as `list` (`--from`, `--to`, `--on`, `-year`, `-month`, `-day`,
`--today-in-history`).

```sh
jrnl-md tags
# @work: 42
# @sarah: 17
# @reading: 9

jrnl-md tags --year 2025
```

### `edit`

```sh
jrnl-md edit [--on <date>]
```

Open a day file in your editor. Defaults to today. `--on` selects a specific
date. If the day file doesn't exist, it is created with a day heading (and a
timestamp heading for today, if timestamps are enabled) before the editor opens.
The cursor is positioned at the end of the file, ready for a new paragraph.

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

Files are stored at `JOURNAL_PATH/YYYY/MM/YYYY-MM-DD.md`.

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

## License

[GPL-3.0](LICENSE)
