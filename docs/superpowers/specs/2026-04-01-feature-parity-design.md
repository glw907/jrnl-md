# Feature Parity: Remaining jrnl Features

## Goal

Implement the 11 remaining jrnl features listed as "not implemented" in `docs/jrnl-compat.md`,
plus add skipped compat test stubs for all of them. After this work, jrnl-md will have 100%
feature parity with jrnl (beyond the documented by-design exceptions).

## Documented Exceptions (not implementing)

Two features are being added as documented exceptions rather than implemented:

- **`timeformat` config key**: jrnl uses Python strftime. jrnl-md already provides equivalent
  functionality via `format.date` + `format.time` using Go time layouts.
- **`--import --format TYPE`**: jrnl only accepts `jrnl` as a value. jrnl-md has one import
  format (markdown), making the flag unnecessary.

## Features to Implement

### 1. `--tagged` (boolean filter flag)

Show only entries that have at least one tag. Inverse of `--not-tagged`.

- Add `Tagged bool` to `Filter` struct
- Add `--tagged` flag to CLI
- Filter: entry matches if `len(entry.Tags) > 0`

### 2. `-year DATE` (date component filter)

Show entries from a specific year.

- Add `Year int` to `Filter` struct
- Parse argument as date string, extract year component
- Accept integer (e.g., `2024`) or date string (e.g., `2024-01-01`)
- Filter: `entry.Date.Year() == filter.Year`

### 3. `-month DATE` (date component filter)

Show entries from a specific month across all years.

- Add `Month int` to `Filter` struct
- Parse argument, extract month component
- Accept integer (1-12), month name (`March`, `Mar`), or date string
- Filter: `entry.Date.Month() == filter.Month`

### 4. `-day DATE` (date component filter)

Show entries on a specific day-of-month across all months.

- Add `Day int` to `Filter` struct
- Parse argument, extract day component
- Accept integer (1-31) or date string
- Filter: `entry.Date.Day() == filter.Day`

### 5. `-today-in-history` (boolean filter flag)

Show entries from today's calendar date in all years.

- Boolean flag, no argument
- Implementation: set Month and Day to today's values before filtering
- Equivalent to `-month <current> -day <current>`

### 6. `--diagnostic` (hidden info flag)

Print version, Go runtime, and OS info for bug reports. Exit immediately.

- Hidden from `--help` (like jrnl)
- Output format:
  ```
  jrnl-md: <version>
  Go: <runtime.Version()>
  OS: <runtime.GOOS> <runtime.GOARCH>
  ```
- Runs before config loading, exits after printing

### 7. `--debug` (verbose logging flag)

Enable debug-level logging during normal operation.

- Add `--debug` boolean flag
- Configure `log/slog` at debug level when set
- Add debug log calls at key points: config load, journal load, filter apply, export

### 8. `display_format` config key

Set the default output format when no `--format`/`--export` is given on CLI.

- Add `DisplayFormat string` to `GeneralConfig`
- In `readEntries()`, if no `--format`/`--export` flag given, use `cfg.General.DisplayFormat`
- Accepts any valid format name: `pretty`, `short`, `json`, `md`, `txt`, `xml`, `yaml`,
  `fancy`, `boxed`, `tags`, `dates`
- Default: `""` (falls through to `pretty`, preserving current behavior)

### 9. `--config-override key value` (repeatable config override)

Override any config key from the command line.

- Repeatable flag taking two tokens: key (dot-notation) and value
- Applied after config file is loaded, before any operations
- Dot-notation traversal: `colors.tags` -> `Config.Colors.Tags`
- Type coercion: attempt bool/int parsing, fall back to string
- Error on unknown keys

### 10. `--format fancy` / `--format boxed` (unicode box-drawing export)

Display entries in Unicode box-drawing cards.

- `fancy` and `boxed` are aliases for the same exporter
- Per-entry card format:
  ```
  ┎─────────────────────────── 2024-01-15 09:00 AM ╮
  ┃ Entry body first line                           │
  ┃ Entry body continues...                         │
  ┖──────────────────────────────────────────────────┘
  ```
- Width controlled by `linewrap` config (default 79)
- Date formatted using `format.date` + `format.time`

### 11. YAML Directory Export

`--format yaml --file dir/` writes one YAML-frontmatter file per entry.

- Requires `--file` to be a directory path (trailing `/` or existing directory)
- Error if `--file` is not a directory
- Per-entry file: `YYYY-MM-DD_HHMMSS_<slug>.md`
- File content:
  ```yaml
  ---
  title: "<first line or empty>"
  date: "2024-01-15T09:00:00"
  starred: false
  tags: ["@project", "@meeting"]
  ---
  Entry body here...
  ```

## Compat Test Strategy

- Add skipped `TestCompat_*` stubs for all 11 features FIRST
- Convert each stub to a real test as the feature is implemented
- Add documented exception tests for `timeformat` and `--import --format`

## Files to Modify

- `cmd/jrnl-md/root.go` — new flags
- `cmd/jrnl-md/args.go` — flag-to-filter mapping, config-override application
- `cmd/jrnl-md/read.go` — display_format default, new export cases, diagnostic output
- `internal/journal/filter.go` — new filter fields and matching logic
- `internal/config/config.go` — new config fields
- `internal/export/` — new fancy and yaml-dir exporters
- `e2e/jrnl_compat_test.go` — new compat tests
- `docs/jrnl-compat.md` — update status of implemented features
