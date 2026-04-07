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
