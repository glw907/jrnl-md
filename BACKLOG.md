# BACKLOG

> Project issue tracker. Managed by `/log-issue`.

## High

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

- [x] **#5** Validate and normalize after --edit `#improvement` `#edit` *(2026-03-30)*
  Resolved by Pass 7: ParseError structured type, post-edit validation loop with re-open, CleanupDayContent normalization.
- [x] **#6** Abort --edit when editor returns empty content `#bug` `#edit` *(2026-03-30)*
  Resolved by Pass 7: IsEmptyContent check on all edit paths, backup restore on direct edit.
- [x] **#7** Simplify editFiltered to work with day files directly `#improvement` `#edit` *(2026-03-30)*
  Resolved by Pass 7: single-day full-match redirects to editDayFile (direct day file edit).
- [x] **#9** Use LoadDay instead of Load for single-day operations `#improvement` `#journal` *(2026-03-30)*
  Resolved by Pass 6: `Entries()` uses directory-driven file selection via `listDayFiles`.
- [x] **#8** Replace ReplaceEntries delete+re-add with direct day-level update `#improvement` `#journal` *(2026-03-30)*
  Resolved by Pass 6: `ReplaceEntries` removed; `DeleteEntry`/`UpdateEntry`/`AddEntry` operate per day file.
- [x] **#0** `--format pretty/short/tags/dates` display mode aliases `#feature` `#cli` *(2026-03-30)*
