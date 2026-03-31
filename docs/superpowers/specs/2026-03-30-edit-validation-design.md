# Pass 7: Edit Validation & Day-File Edit

**Date:** 2026-03-30
**Status:** Draft
**Scope:** Post-edit validation on all edit paths, single-day filtered edits open day files
directly, empty buffer abort

## Motivation

The three edit paths (`editEntry`, `editFiltered`, `LaunchEncrypted`) have no post-edit
validation. Malformed edits are silently accepted or produce cryptic parse errors. Emptying
the editor buffer silently deletes all entries. Single-day filtered edits do an unnecessary
serialize→parse round-trip when they could open the day file directly.

Backlog items: #5 (validate after edit), #6 (empty buffer abort), #7 (day-file direct edit).

## Design

### 1. Edit path taxonomy

| Trigger | Path | Description |
|---------|------|-------------|
| `jrnl-md` (no args/flags) | Direct | Open today's day file |
| `jrnl-md --edit` (no filter) | Direct | Open today's day file |
| `jrnl-md --edit --on DATE` (all entries match) | Direct | Open that date's day file |
| `jrnl-md --edit @tag` (single day, partial match) | Filtered | Temp file round-trip |
| `jrnl-md --edit --from X --to Y` (multi-day) | Filtered | Temp file round-trip |
| Encrypted variants of the above | Same paths | Decrypt before, re-encrypt after |

**Single-day redirect (#7):** When a filtered `--edit` result spans exactly one calendar day
AND includes all entries from that day, use the direct path instead of the temp-file
round-trip. Detection: compare `len(filteredEntries)` against `len(fj.DayEntries(date))`.
If lengths match and all dates fall on the same day, redirect.

If the filter excludes some entries from that day (partial match), fall through to the
filtered path — the user would otherwise see unfiltered entries.

### 2. Direct edit flow

```
1. Read pre-edit file content (backup)
2. If bare edit (no args, no filter): append entry heading + template via PrepareDayFile
   (Do NOT append for redirected single-day filtered edits — the user is editing
   existing entries, not creating a new one.)
3. Open editor
4. Read file back
5. Empty check: if content is empty/whitespace-only → abort, restore backup
6. Validate: parseDay on the content
   - If error → show actionable message, offer re-open
   - If user declines → leave file as-is, print warning with path
   - If user accepts → goto 3
7. Light cleanup: strip trailing empty entry headings, normalize blank
   lines before ## headings, trim trailing whitespace
8. Write cleaned content back atomically (only if cleanup changed anything)
```

**Bare edit behavior:** `PrepareDayFile` appends a `## [time]` heading before opening the
editor. If the user leaves this heading empty (no body text), the cleanup step in (7) strips
it rather than persisting a stub entry.

**Encrypted variant:** Decrypt before step 1, re-encrypt after step 7. On re-open decline,
discard edits — the original encrypted file is untouched since `LaunchEncrypted` only writes
after validation succeeds.

### 3. Filtered edit flow

```
1. Check if result is single-day + full-day → redirect to direct path (§1)
2. FormatEntries → WriteTempAndEdit
3. Read edited content
4. Empty check → abort with warning (temp file preserved for recovery)
5. Validate: ParseMultiDay
   - If error → show actionable message, offer re-open (temp file still exists)
   - If user declines → abort, leave journal unchanged, print temp file path
   - If user accepts → re-open temp file, goto 3
6. DeleteEntries(old) + AddEntries(new)
7. Report count
```

### 4. Empty buffer abort (#6)

All paths check for empty/whitespace-only content after the editor exits.

Message:
```
No entries found after editing. Were you trying to delete all entries?
Aborting — no changes made.
```

**Per-path behavior on empty:**
- Direct: restore backup (pre-edit file content)
- Encrypted: discard (original encrypted file untouched)
- Filtered: abort without modifying journal, preserve temp file

**Partial deletions** (some entries removed in filtered edit): accept silently and print a
count — this matches jrnl's behavior and is already implemented.

### 5. Actionable error messages

Parse errors include: file/source identifier, line number, the bad value, and the expected
format.

Examples:
```
2026/03/30.md: line 1: missing day heading (expected "# YYYY-MM-DD Weekday")
2026/03/30.md: line 5: can't parse time "3:59pm" (expected "## [03:04 PM]")
edited content: line 12: missing day heading between entries on different dates
```

**Implementation:** `parseDay` and `ParseMultiDay` return a `ParseError` type with structured
fields (`File`, `Line`, `Message`) rather than bare `fmt.Errorf` strings. Callers that need
the raw error (like `Entries` which logs warnings) format it to string. Callers in the edit
path use the structured fields for the re-open prompt.

```go
type ParseError struct {
    File    string // file path or "edited content"
    Line    int    // 1-based line number
    Value   string // the bad value found
    Expected string // what was expected
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("%s: line %d: %s (expected %s)", e.File, e.Line, e.Value, e.Expected)
}
```

`parseDay` currently accepts `(text, dateFmt, timeFmt string)`. It gains an optional `file`
parameter (or the caller sets `ParseError.File` after the fact) for error context.

### 6. Light cleanup (direct edits only)

After successful `parseDay` validation on a direct-edit file:

1. **Strip trailing empty entry headings** — a `## [time]` heading followed by only
   whitespace or EOF. This handles the bare-edit case where `PrepareDayFile` appended a
   heading the user didn't use.
2. **Normalize blank lines before `##` headings** — ensure exactly one blank line before
   each `## [time]` heading.
3. **Trim trailing whitespace** from each line.
4. **Ensure single trailing newline** at end of file.

Write back atomically only if cleanup changed the content (byte comparison with pre-cleanup
version). This avoids unnecessary file writes and preserves the file's mtime when no cleanup
was needed.

Cleanup does NOT re-serialize through `day.Format`. The user's body text, paragraph spacing,
and content formatting are preserved exactly. Only the structural whitespace around headings
is normalized.

### 7. Re-open prompt

```
Error in edited content:
  line 5: can't parse time "3:59pm" (expected "## [03:04 PM]")
Re-open editor? [y/N]
```

Unlimited re-opens until the user declines or the content validates. Each re-open returns
to the editor with the file/temp-file as the user left it (including their malformed edits)
so they can fix the specific issue.

### 8. Decline behavior (per path)

| Path | On decline |
|------|-----------|
| Direct (plain) | Leave file as-is, print warning: `"Warning: <path> may contain invalid entries"` |
| Direct (encrypted) | Discard edits, original encrypted file untouched |
| Filtered (temp file) | Abort, journal unchanged, print: `"Edits preserved in <temp-path>"` |

### 9. Files changed

| File | Change |
|------|--------|
| `internal/journal/day.go` | `parseDay`/`ParseMultiDay` return `*ParseError` with line info |
| `internal/journal/cleanup.go` | New: `CleanupDayContent(text) string` — light cleanup functions |
| `internal/editor/editor.go` | `ValidateAndCleanup`, `validateLoop` with re-open prompt |
| `cmd/jrnl-md/edit.go` | Single-day redirect logic, updated flows with validation |

**No changes to:** `folder.go`, `filter.go`, `root.go`, export, display, config, crypto.
