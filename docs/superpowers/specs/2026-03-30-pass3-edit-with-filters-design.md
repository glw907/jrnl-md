---
name: Pass 3 — --edit with Filters
description: Open filtered entries in editor temp file, parse changes back to journal
type: project
---

# Pass 3: `--edit` with Filters

## Goal

`--edit` always opens entries in a temp file for editing, matching jrnl's behavior exactly:

- `--edit` with no filters: opens all entries (or last N if `-n` is set) in a temp file.
- `--edit` with filter flags: opens only the matching entries.

On save-and-exit, parse the edited content back and update the journal files. This replaces the current behavior of opening today's raw day file directly.

---

## Current behavior

`root.go` sends to `editEntry(fj, cfg, configPath, passphrase)` when `f.edit` is true or when no args and no filter flags. `editEntry` always opens today's day file. This does not match jrnl, which opens all/last-N entries regardless of date.

---

## New behavior

In `root.go` — `runRoot`:

```
if f.edit:
    → always use editFiltered (load full journal, apply filter — empty filter = all entries)
```

The `editEntry` function (today's-file path) is removed. `editFiltered` handles both the filtered and unfiltered cases. With an empty filter and no `-n`, all entries are opened.

The filtered edit path:
1. `fj.Load()` — load full journal
2. Apply filter → get matched `[]journal.Entry`
3. Serialize matched entries to a temp file as markdown (entry heading + body)
4. Open temp file in editor
5. On exit, parse the temp file back into entries
6. Diff old vs new entries by their original timestamps
7. Write changed entries back to their day files; delete removed entries

---

## Serialization format

Each entry in the temp file:

```
## [HH:MM AM/PM]
[blank line]
body text

```

Starred entries:

```
## [HH:MM AM/PM] *
[blank line]
body text

```

This mirrors the existing day file format. Parsing uses the same entry-heading regex already in `internal/journal/parser.go`.

The day-level heading (`# date weekday`) is NOT included in the temp file. Entries from multiple days are concatenated. The original timestamp is the stable key for matching edits back.

---

## Diff and write-back

After editing:
1. Parse temp file → `[]journal.Entry` (new state). Use the existing parser.
2. Match new entries to originals by timestamp (exact match).
3. For each original entry:
   - **Modified** (body or starred changed): update the entry in memory.
   - **Deleted** (no matching timestamp in new state): remove from the in-memory day.
4. For each new entry with an unrecognized timestamp: add as a new entry (allows user to create entries while editing).
5. Save all modified day files via `fj.SaveDay(date)` — a new method (see below).

---

## New journal methods needed

**`internal/journal/folder.go`**

```go
// FilteredEntries returns entries matching f across all loaded days.
// Requires fj.Load() to have been called first.
func (fj *FolderJournal) FilteredEntries(f filter.Filter) []Entry

// UpdateEntry replaces an existing entry (matched by timestamp) with updated.
// Returns false if no entry with that timestamp was found.
func (fj *FolderJournal) UpdateEntry(updated Entry) bool

// DeleteEntry removes an entry matched by timestamp.
// Returns false if not found.
func (fj *FolderJournal) DeleteEntry(ts time.Time) bool

// SaveDay writes a single day file by date.
func (fj *FolderJournal) SaveDay(date time.Time) error
```

`SaveDay` writes the day file at the path derived from `date`, using the same format as `Save()` but for a single day.

---

## Encrypted journals

Encrypted journals follow the same pattern used in the existing `editEntry`:
- Decrypt to temp file → edit → encrypt from temp file.
- For filtered edit with encryption: decrypt each relevant day file to a single concatenated temp file; on exit, re-encrypt and write back individual day files.
- This is the same `editor.LaunchEncrypted` pattern extended to multiple files.

---

## `cmd/jrnl-md/edit.go`

Add a new function:

```go
func editFiltered(fj *journal.FolderJournal, cfg config.Config, f *flags, tagArgs []string) error
```

Steps:
1. Build `filter.Filter` from flags (same as `buildFilter` in read.go — extract to shared helper or duplicate).
2. `entries := fj.FilteredEntries(filter)`
3. If `len(entries) == 0`: print "No entries found." and return nil.
4. Serialize entries to temp file.
5. Open temp file in editor (plaintext path; encryption handled separately).
6. Parse temp file → new entries.
7. Compute diff: updated, deleted.
8. Apply updates and deletes via `fj.UpdateEntry` / `fj.DeleteEntry`.
9. Save modified day files.

---

## Encryption handling for filtered edit

When `fj.Encrypted()`:
- Each day file must be decrypted before inclusion.
- Use `crypto.Decrypt(passphrase, encryptedBytes)` per day.
- Write decrypted content to a single temp file (entries only, no day headings).
- After edit, re-encrypt updated day files via `crypto.Encrypt(passphrase, plaintext)`.

The existing `editor.LaunchEncrypted` handles single-file encrypted editing. For filtered multi-file, inline the decrypt/encrypt logic in `editFiltered` rather than reusing LaunchEncrypted (different shape).

---

## Testing

- `folder_test.go`: FilteredEntries, UpdateEntry, DeleteEntry, SaveDay
- `edit_test.go` (or e2e): filtered edit updates correct entries; deleted entries removed; no-match prints message
- Encrypted path: unit test decrypt→edit→encrypt round-trip

---

## Files touched

| File | Change |
|------|--------|
| `cmd/jrnl-md/root.go` | Route all --edit cases to editFiltered; remove editEntry call |
| `cmd/jrnl-md/edit.go` | Add editFiltered; remove or repurpose editEntry |
| `internal/journal/folder.go` | FilteredEntries, UpdateEntry, DeleteEntry, SaveDay |
