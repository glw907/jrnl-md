---
name: Pass 4 — Import
description: --import FILE and stdin import for loading external content into the journal
type: project
---

# Pass 4: Import

## Goal

Add `--import` to load entries from an external file or stdin into the journal. Matches jrnl's `--import` behavior: read a file (or stdin) containing journal-formatted entries and merge them into the appropriate day files.

---

## jrnl behavior reference

From jrnl docs:
- `jrnl --import ./old.txt` — import entries from a jrnl-format text file
- `jrnl --import -` — read from stdin
- Entries in the import file use the same format as export output (one entry per heading)
- Duplicate entries (same timestamp) are skipped

---

## Input format

jrnl-md's markdown format for a day file:

```markdown
# 2025-01-15 Wednesday

## [09:00 AM]

Entry body here.

## [02:30 PM]

Another entry.
```

Import accepts this format. The day heading (`# date weekday`) provides the date context for entries without full timestamps. Entries with `## [HH:MM AM/PM]` headings are parsed normally.

Also accept jrnl's plain-text export format for migration use:

```
2025-01-15 09:00 | Entry body here.
```

Detection: if the file starts with `# ` followed by a date, treat as markdown format. Otherwise try plain-text format.

**Primary format is markdown** (our native format). Plain-text is a migration convenience; implementation may defer to a best-effort parse without full fidelity.

---

## Implementation

**`cmd/jrnl-md/root.go`**

Add to `flags`:

```go
importFile string
```

Register:

```go
cmd.Flags().StringVar(&f.importFile, "import", "", "Import entries from file (use - for stdin)")
```

Add `f.importFile != ""` to `hasFilterFlags` — prevents the default "open editor" path.

In `runRoot`, after encryption checks, before the `len(text) > 0` branch:

```go
if f.importFile != "" {
    fj := journal.NewFolderJournal(path, opts)
    if err := fj.Load(); err != nil {
        return fmt.Errorf("loading journal: %w", err)
    }
    return importEntries(fj, f.importFile)
}
```

**`cmd/jrnl-md/import.go`** (new file)

```go
func importEntries(fj *journal.FolderJournal, source string) error
```

Steps:
1. Read source: if source == "-", read from `os.Stdin`; otherwise read the named file.
2. Parse input as markdown day files using existing parser.
3. For each parsed entry, call `fj.ImportEntry(entry)` — a new journal method.
4. Save all modified day files.
5. Print to stderr: "Imported N entries. Skipped M duplicates."

**`internal/journal/folder.go`** — new method:

```go
// ImportEntry adds entry to the journal if no entry with the same timestamp exists.
// Returns true if added, false if skipped (duplicate).
func (fj *FolderJournal) ImportEntry(e Entry) (bool, error)
```

Steps:
1. Determine day from `e.Time`.
2. Load the day if not already loaded (`fj.LoadDay(e.Time)` — idempotent).
3. Check for duplicate: if any existing entry has the same timestamp, return false, nil.
4. `fj.AddEntry(e.Time, e.Body, e.Starred)`.
5. Return true, nil.

Save is called by the caller (`importEntries`) after all entries are processed — not per-entry.

---

## Parsing

The existing `internal/journal/parser.go` parses day files into entries. `importEntries` uses the same parser.

For multi-day import files: the file may contain multiple `# date weekday` headings. The parser already handles this since `fj.Load()` walks multiple files and parses each. For import, concatenate all day sections and parse as a single logical stream — or parse section-by-section by splitting on `# ` headings.

**Preferred approach:** split input on lines matching `^# \d{4}-\d{2}-\d{2}` to get per-day sections, parse each section using the existing day parser with the date from the heading.

---

## Encrypted journals

Import to an encrypted journal:
- Load encrypted day files (already handled by `fj.Load()` with passphrase).
- `ImportEntry` works on decrypted in-memory state.
- `fj.Save()` re-encrypts on write.

No special handling needed beyond what `fj.Load()` and `fj.Save()` already do.

---

## Testing

- `import_test.go`: import from file, import from stdin (pipe), duplicate skipping, multi-day file, count reporting
- `folder_test.go`: ImportEntry adds entry, ImportEntry skips duplicate, ImportEntry with multiple days

---

## Files touched

| File | Change |
|------|--------|
| `cmd/jrnl-md/root.go` | Add --import flag; route to importEntries; add to hasFilterFlags |
| `cmd/jrnl-md/import.go` | New file: importEntries function |
| `internal/journal/folder.go` | ImportEntry method |
