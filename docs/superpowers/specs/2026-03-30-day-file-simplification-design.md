# Day-File-First Simplification

**Date:** 2026-03-30
**Status:** Draft
**Scope:** Full restructuring of `FolderJournal` around the per-day-file storage model

## Motivation

jrnl-md inherited several patterns from jrnl's single-file storage model: load everything
into memory, flatten into a sorted list, filter, mutate the flat list, re-partition into
day buckets, save all modified days. This is unnecessary complexity given that jrnl-md's
storage format is one markdown file per calendar day (`YYYY/MM/DD.md`).

The day file should be the unit of work. The directory structure (`YYYY/MM/`) should serve
as the primary index. Operations should load only the files they need, mutate them directly,
and save immediately.

## Design

### 1. Filter decomposition

Split the current single-pass `Filter` into two layers:

**Date-range filter** (drives file selection):
- Derived from `--from`, `--to`, `--on`, and `-n`
- Used by `FolderJournal.Entries()` to determine which day files to read
- The `YYYY/MM/` directory structure allows skipping entire months/years

**Content filter** (runs per-entry after load):
- `--starred`, `--not-starred`, `--not-tagged`, `--contains`, `--and`, tag matching, `--not`
- Applied to entries from each loaded day file
- `-n` applied last as a tail slice on collected results

The `Filter` struct stays. It gains a `DateRange() (start, end *time.Time)` method that the
journal uses to decide what to load. `Apply()` still runs content predicates but no longer
checks dates — those are handled by file selection.

### 2. FolderJournal API

Replace the current API with a day-centric surface:

**Loading:**

- `Entries(f *Filter) ([]Entry, error)` — load only needed files based on filter's date
  range, apply content filter, return results. Replaces `Load()` + `AllEntries()` +
  `Filter.Apply()`.
- `LoadDay(date time.Time) (*day, error)` — load a single day file, return the day struct.
  `day` stays unexported — `LoadDay` is used internally by mutation methods and by the edit
  paths (which live in `cmd/jrnl-md/`). The edit paths access it via a new exported method
  `DayEntries(date time.Time) ([]Entry, error)` when they need to check whether a filtered
  edit covers all entries in a day file (to decide between direct-file and temp-file editing).

**Mutations (each loads, modifies, and saves the target day file immediately):**

- `AddEntry(date time.Time, body string, starred bool) error` — LoadDay + append + write.
- `DeleteEntry(e Entry) error` — route by `e.Date`, remove from day, write (or delete file
  if no entries remain).
- `UpdateEntry(old Entry, new Entry) error` — route by `old.Date`, replace in-place, write.
  Handles cross-day moves (change-time) by removing from old day and adding to new day.
- `ImportEntry(e Entry) error` — LoadDay + duplicate timestamp check + append + write.

**Lifecycle:**

- `DayFilePath(date time.Time) string` — unchanged.
- `Save() error` — retained only for encrypt/decrypt which re-writes all files. Regular
  mutations save immediately per-file.

**Removed:**

- `Load()` — replaced by `Entries(f)` and `LoadDay`
- `AllEntries()` — replaced by `Entries(f)`
- `ReplaceEntries()` — replaced by `UpdateEntry`
- `ChangeEntryTimes()` — callers use `UpdateEntry` with a new date
- `DeleteEntries()` (batch) — callers use `DeleteEntry` (single) in a loop
- `MarkAllModified()` — encrypt/decrypt gets its own walk

### 3. Directory-driven file selection

`Entries(f)` uses the directory structure as an index rather than loading everything:

- **`--on 2026-03-15`**: compute path directly as `2026/03/15.md`. No directory walk.
- **`--from 2026-02-01 --to 2026-03-15`**: list only `2026/02/` and `2026/03/` directories,
  filter filenames by day range. Skip all other year/month dirs.
- **`--from 2026-03-01`** (no end): list from `2026/03/` through current month.
- **`--to 2026-02-28`** (no start): walk backward from `2026/02/` to earliest year directory.
- **No date range**: full walk (same as current behavior). Only happens for fully unfiltered
  reads.
- **`-n 5` with no date range**: walk backward from today, loading day files until 5 entries
  are collected. Stop early rather than loading the entire journal.

Directory inspection uses `os.ReadDir` at each level. Year dirs are 4-digit names, month
dirs are 2-digit names, day files are `DD.md` (or `DD.md.age`). Filenames are naturally
sortable since they're zero-padded.

### 4. Edit paths

**Direct `--edit` (no filter flags):**
Unchanged — opens today's day file in place. Add post-edit validation: re-parse with
`parseDay`, lightweight whitespace cleanup if valid, actionable error + re-open offer if
invalid.

**Filtered `--edit` with single-day result (all entries from that day):**
Open the day file directly in the editor. No temp file, no serialize round-trip. After
editor exits, re-parse and validate.

If the filter excludes some entries from that day (e.g. `--edit --on 2026-03-15 --starred`),
fall through to the multi-day path below — the user would otherwise see unfiltered entries.

**Filtered `--edit` with multi-day or partial-day result:**
The one case where the serialize-to-temp-file round-trip is genuinely needed. Keep
`FormatEntries` and `ParseMultiDay` for this path. After parsing, update each affected
day file individually via `UpdateEntry`. Add validation.

**Encrypted `--edit`:**
Same as direct path, with decrypt-before/re-encrypt-after wrapper. Validation happens on
plaintext before re-encrypting.

### 5. Mutation save strategy

Each mutation method loads the target day file, modifies it, and writes it back atomically
in one call. No deferred-save pattern.

**`writeDay` (private):**
Serializes one `day` struct via `day.Format()`, writes with `atomicfile.WriteFile`. If the
day has zero entries after mutation, delete the file and clean up empty parent directories
(`MM/`, `YYYY/`).

**Batch `Save()` for encrypt/decrypt only:**
Walk all files, decrypt/re-encrypt, write new files, then delete old files. Write all new
files before deleting any old files — failure leaves both copies (safe) rather than no
copies.

### 6. Internal state changes

- `fj.days` no longer accumulates all loaded days in memory. Each operation loads what it
  needs, operates, writes, done. For `Entries(f)`, days are loaded, filtered, and the
  entries returned — no need to hold day structs after.
- The `modified` flag on `day` structs is removed. Mutations write immediately, so there
  is no deferred-save tracking.
- Empty day files are deleted rather than persisting as heading-only stubs.
- Empty `MM/` and `YYYY/` directories are removed when their last day file is deleted.

### 7. Caller impact

**`root.go` dispatch:**
Every path that did `fj.Load()` + `buildFilter` + `flt.Apply(fj.AllEntries())` collapses
to `fj.Entries(&flt)`.

| Operation | Before | After |
|-----------|--------|-------|
| inline write | `LoadDay` + `AddEntry` + `Save` | `AddEntry` (saves internally) |
| `--edit` (no filter) | `LoadDay` + open file | unchanged |
| `--edit` (filtered) | `Load` + filter + serialize + parse + `ReplaceEntries` + `Save` | `Entries(&flt)` + day-file edit + `UpdateEntry` |
| `--delete` | `Load` + filter + `DeleteEntries` + `Save` | `Entries(&flt)` + loop `DeleteEntry` |
| `--change-time` | `Load` + filter + `ChangeEntryTimes` + `Save` | `Entries(&flt)` + loop `UpdateEntry` |
| read/export | `Load` + filter + display | `Entries(&flt)` + display |
| `--import` | `ParseMultiDay` + loop `ImportEntry` + `Save` | `ParseMultiDay` + loop `ImportEntry` (saves internally) |
| `--encrypt`/`--decrypt` | `Load` + `MarkAllModified` + `SetEncryption` + `Save` | dedicated walk-all-files path |

**Files with no changes:** export functions, display package, config, crypto, atomicfile,
prompt, dateparse.

### 8. Functions kept but scoped

- `FormatEntries` — used by multi-day/partial-day filtered edit and `--export md`
- `ParseMultiDay` — used by `--import` and multi-day/partial-day filtered edit

### 9. Error handling (from backlog #4, #5, #6)

**During `Entries()` (load):**
If a day file fails to parse, log a warning to stderr with the file path, line number, bad
value, and expected format. Skip the file, continue loading. Example:
`warning: 2026/03/30.md: line 3: can't parse time "3:59pm" (expected "## [03:04 PM]") -- skipping file`

**After `--edit` (all paths):**
Re-parse to validate. If invalid, show actionable error and offer to re-open editor. For
direct edit, restore backup if user declines. For encrypted edit, discard edits (original
encrypted file untouched). For filtered edit, abort without saving (temp file preserved).

**Empty editor buffer:**
Abort with warning matching jrnl's message: "Were you trying to delete all the entries?
This seems a bit drastic, so the operation was cancelled."
