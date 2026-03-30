---
name: Pass 2 — Date & Time
description: Date-prefixed inline entries, apply default_hour/default_minute to date-only inputs
type: project
---

# Pass 2: Date & Time

## Goal

Match jrnl behavior for four cases:
1. Inline entry with a date prefix: `jrnl yesterday: Entry text` records the entry at that date/time.
2. Natural-language and date-only inputs respect `default_hour` / `default_minute` from config.
3. `echo "text" | jrnl-md` creates an entry from stdin (when stdin is not a TTY and no text args are given).
4. `--config-file` is the primary flag name (matching jrnl); `--config` is kept as a hidden alias.

---

## 1. Date-prefixed inline entries

**Format:** `jrnl-md [journal:] <date_expression>: <entry text>`

The `: ` separator (colon + space) splits a date expression from the entry body. Date expressions are anything `dateparse.Parse` accepts.

**`cmd/jrnl-md/write.go`** — `writeInline`

Current: joins args, checks for `*`, writes at `now`.

Change: before the starred check, detect a date prefix:

```go
body := strings.Join(text, " ")

// Detect date prefix: "expression: body text"
entryTime := now
if idx := strings.Index(body, ": "); idx > 0 {
    candidate := body[:idx]
    if t, err := dateparse.Parse(candidate, now, cfg.General.DefaultHour, cfg.General.DefaultMinute); err == nil {
        entryTime = t
        body = body[idx+2:]
    }
}

starred := strings.HasSuffix(body, "*") || strings.HasPrefix(body, "*")
if starred { body = strings.Trim(body, "* ") }
```

The parsed time `t` becomes `entryTime`. The journal day file is determined by `entryTime`, not `now`.

`fj.LoadDay` is currently called with `now` in `root.go` before `writeInline`. When a date prefix is detected, the correct day file may differ from today. `writeInline` must either:

- Accept `entryTime` as a return value and have `root.go` re-load, or
- Call `fj.LoadDay(entryTime)` itself before `fj.AddEntry`.

**Design choice:** `writeInline` calls `fj.LoadDay(entryTime)` directly when `entryTime != now`. The FolderJournal LoadDay is idempotent for the same date (re-loading the same file is safe). Signature change:

```go
func writeInline(fj *journal.FolderJournal, text []string, cfg config.Config, now time.Time) error
```

No signature change needed. `writeInline` calls `fj.LoadDay(entryTime)` after detecting the prefix (this replaces the already-loaded today file in memory with the correct date's file).

---

## 2. Apply default_hour / default_minute

**`internal/dateparse/dateparse.go`** — `Parse`

Current: explicit date-only layouts (e.g. `2006-01-02`) return `time.Date(y, m, d, 0, 0, 0, 0, loc)` — midnight. The `when` library returns a time anchored to the current moment for natural-language expressions like "yesterday" or "last tuesday", so those already land at the right time-of-day.

Change: `Parse` needs `defaultHour int` and `defaultMinute int` parameters so callers can pass config values.

New signature:

```go
func Parse(s string, now time.Time, defaultHour, defaultMinute int) (time.Time, error)
```

When an explicit date-only layout matches, replace the zero hour/minute with defaultHour/defaultMinute:

```go
t = time.Date(t.Year(), t.Month(), t.Day(), defaultHour, defaultMinute, 0, 0, t.Location())
```

The `when` library returns a full timestamp; do NOT override its time component — it already reflects the current time of day (or midnight for date-only expressions it parses). Only apply default_hour/default_minute to our own explicit date-only layout matches.

**All callers of `dateparse.Parse`** must be updated to pass defaultHour and defaultMinute:

- `cmd/jrnl-md/read.go` — `buildFilter` uses `dateparse.Parse` for `--from`, `--to`, `--on`
- `cmd/jrnl-md/changetime.go` — uses `dateparse.Parse` for `--change-time`
- `cmd/jrnl-md/write.go` — new date-prefix detection call

For filter/change-time callers, pass `cfg.General.DefaultHour` and `cfg.General.DefaultMinute`. These are available wherever `cfg config.Config` is in scope.

---

## Testing

**`internal/dateparse/dateparse_test.go`**

Add cases:
- Date-only string with non-zero defaultHour/defaultMinute → returns that hour/minute
- Date-only string with defaultHour=0, defaultMinute=0 → returns midnight (unchanged behavior)
- Full datetime string → defaultHour/defaultMinute ignored

**`cmd/jrnl-md/write_test.go`** (or e2e)

- `yesterday: Entry text` → entry lands in yesterday's day file at default_hour:default_minute
- `2025-01-15: Entry text` → entry lands in 2025/01/15.md
- No prefix → entry lands at `now`
- Invalid date prefix (e.g. `foo: bar`) → treated as body text (no prefix detected)

---

---

## 3. Stdin entry writing

When `len(text) == 0` and stdin is not a TTY, read from stdin and use it as the entry body.

**`cmd/jrnl-md/root.go`** — in `runRoot`, in the `len(text) > 0` branch, add a stdin check before it:

```go
if len(text) == 0 && !term.IsTerminal(int(os.Stdin.Fd())) {
    data, err := io.ReadAll(os.Stdin)
    if err != nil {
        return fmt.Errorf("reading stdin: %w", err)
    }
    body := strings.TrimSpace(string(data))
    if body != "" {
        text = []string{body}
    }
}
```

This runs after the passphrase prompt (which itself requires a TTY) but before the `len(text) > 0` branch, so it feeds naturally into `writeInline`. Imports needed: `"io"`, `"golang.org/x/term"`.

---

## 4. `--config-file` flag

Rename the `--config` flag to `--config-file` to match jrnl. Keep `--config` as a hidden alias for backward compatibility.

**`cmd/jrnl-md/root.go`**

In `flags` struct, rename `configFile string` → keep as-is (internal name unchanged).

In `newRootCmd`, replace:

```go
cmd.Flags().StringVar(&f.configFile, "config", "", "Config file path")
```

with:

```go
cmd.Flags().StringVar(&f.configFile, "config-file", "", "Config file path")
cmd.Flags().StringVar(&f.configFile, "config", "", "Config file path (alias for --config-file)")
cmd.Flag("config").Hidden = true
```

---

## Files touched

| File | Change |
|------|--------|
| `internal/dateparse/dateparse.go` | Add defaultHour/defaultMinute params; apply to date-only layouts |
| `cmd/jrnl-md/write.go` | Detect `: ` prefix, parse date, re-load correct day, use entryTime |
| `cmd/jrnl-md/read.go` | Pass defaultHour/defaultMinute to dateparse.Parse calls |
| `cmd/jrnl-md/changetime.go` | Pass defaultHour/defaultMinute to dateparse.Parse calls |
| `cmd/jrnl-md/root.go` | Stdin read before write path; --config-file primary flag |
