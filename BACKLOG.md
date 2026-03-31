# BACKLOG

> Project issue tracker. Managed by `/log-issue`.

## High

- [ ] **#6** Abort --edit when editor returns empty content `#bug` `#edit` *(2026-03-30)*
  jrnl-md silently deletes all selected entries via `ReplaceEntries(old, nil)` when the user
  empties the editor buffer. jrnl aborts with a warning: "Were you trying to delete all the
  entries? This seems a bit drastic, so the operation was cancelled." Match that behavior.
  For partial deletions (some entries removed), jrnl accepts silently and prints a post-hoc
  count — match that too. `cmd/jrnl-md/edit.go:73-78`, `internal/journal/day.go:127-129`
- [ ] **#5** Validate and normalize after --edit `#improvement` `#edit` *(2026-03-30)*
  All three --edit paths should validate and normalize after the editor exits. Spec:
  **Filtered --edit** (temp file): if `ParseMultiDay` returns a parse error, show the
  specific error to stderr and offer to re-open the editor (the temp file still exists).
  If empty result, abort with warning per #6. If fewer entries, accept and print count
  (already works).
  **Direct --edit** (raw day file): back up the pre-edit content. After the editor exits,
  re-parse with `parseDay` to validate only — do not re-serialize. If valid, do a lightweight
  whitespace cleanup in place (normalize blank lines around `##` headings, trim trailing
  spaces). If invalid, show the parse error and offer to re-open. If the user declines,
  restore the backup. No full serialize→parse round-trip — that pattern is a hold-over from
  jrnl's single-file format and unnecessary when we're editing a per-day file directly.
  **Encrypted --edit** (decrypt → temp file → re-encrypt): same validate-only pattern as
  direct --edit. After the editor exits, parse the edited content to validate before
  re-encrypting. If invalid, show the error and offer to re-open. If the user declines,
  discard edits (the original encrypted file is untouched since `LaunchEncrypted` only writes
  after success). `internal/editor/editor.go:143-178`
  **Error messages must be actionable.** Don't just say what failed — say what was expected.
  Examples: `line 1: missing day heading (expected "# 2026-03-30 Sunday")`,
  `line 5: can't parse time "3:59pm" (expected format "03:04 PM", e.g. "## [03:59 PM]")`,
  `line 1: missing day heading (expected "# YYYY-MM-DD Weekday")`. Include the line number,
  the bad value, and the expected format so the user can fix it without guessing.
  Key locations: `cmd/jrnl-md/edit.go`, `internal/journal/day.go:22-34`
- [ ] **#7** Simplify editFiltered to work with day files directly `#improvement` `#edit` *(2026-03-30)*
  `editFiltered` inherits jrnl's single-file pattern: serialize all matching entries into one
  multi-day blob → temp file → parse back → delete+re-add via `ReplaceEntries`. For single-day
  filters (the common case), just open the day file directly. For multi-day filters, open each
  affected day file in sequence. This eliminates the `FormatEntries` → `ParseMultiDay` →
  `ReplaceEntries` pipeline from the edit flow. `FormatEntries`/`ParseMultiDay` stay for
  `--import` only. `cmd/jrnl-md/edit.go:56-95`, `internal/journal/day.go:97-146`
- [ ] **#1** Compat test suite audit `#improvement` `#docs` *(2026-03-30)*
  Cross-reference every feature in `docs/jrnl-compat.md` against `e2e/jrnl_compat_test.go`.
  Confirm each has a real assertion. Add `TestCompat_*` tests for gaps. Known gaps:
  `--format pretty/short/tags/dates`, `--export xml/yaml`, `linewrap`, `indent_character`,
  starred-entry write syntax, `-N` shorthand, encrypt/decrypt.

## Medium

- [ ] **#4** Skip malformed day files during Load instead of aborting `#bug` `#journal` *(2026-03-30)*
  A single file with a bad date or time heading causes `Load()` to abort — the entire journal
  becomes unreadable. Spec: log a warning to stderr with the file path and specific parse
  error, skip the file, continue loading everything else. Error messages must be actionable —
  include the file path, line number, bad value, and expected format, e.g.:
  `warning: 2026/03/30.md: line 3: can't parse time "3:59pm" (expected "## [03:04 PM]") — skipping file`.
  The user can then go fix that one file manually. jrnl's folder backend has the same crash
  behavior (arguably a bug), so this is a reasonable deviation to document in
  `docs/jrnl-compat.md`. `internal/journal/folder.go:120-123`

## Low

- [ ] **#3** `--config-override key=value` flag `#feature` `#config` *(2026-03-30)*
  Power-user escape hatch to override individual config keys at the command line.
- [ ] **#2** `--debug` flag `#feature` `#cli` *(2026-03-30)*
  Verbose diagnostic output for troubleshooting. No functional gap for normal use.

## Done

- [x] **#9** Use LoadDay instead of Load for single-day operations `#improvement` `#journal` *(2026-03-30)*
  Resolved by Pass 6: `Entries()` uses directory-driven file selection via `listDayFiles`.
- [x] **#8** Replace ReplaceEntries delete+re-add with direct day-level update `#improvement` `#journal` *(2026-03-30)*
  Resolved by Pass 6: `ReplaceEntries` removed; `DeleteEntry`/`UpdateEntry`/`AddEntry` operate per day file.
- [x] **#0** `--format pretty/short/tags/dates` display mode aliases `#feature` `#cli` *(2026-03-30)*
