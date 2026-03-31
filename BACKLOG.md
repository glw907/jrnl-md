# BACKLOG

> Project issue tracker. Managed by `/log-issue`.

## High

- [ ] **#6** Abort --edit when editor returns empty content `#bug` `#edit` *(2026-03-30)*
  jrnl-md silently deletes all selected entries via `ReplaceEntries(old, nil)` when the user
  empties the editor buffer. jrnl aborts with a warning: "Were you trying to delete all the
  entries? This seems a bit drastic, so the operation was cancelled." Match that behavior.
  For partial deletions (some entries removed), jrnl accepts silently and prints a post-hoc
  count — match that too. `cmd/jrnl-md/edit.go:73-78`, `internal/journal/day.go:127-129`
- [ ] **#5** Normalize trailing whitespace after --edit `#improvement` `#edit` *(2026-03-30)*
  After re-parsing edited content, trim trailing whitespace from lines and ensure consistent
  blank-line spacing between entries before saving. Don't rewrite semantic content — only fix
  spacing issues that could cause parse drift over time.
- [ ] **#1** Compat test suite audit `#improvement` `#docs` *(2026-03-30)*
  Cross-reference every feature in `docs/jrnl-compat.md` against `e2e/jrnl_compat_test.go`.
  Confirm each has a real assertion. Add `TestCompat_*` tests for gaps. Known gaps:
  `--format pretty/short/tags/dates`, `--export xml/yaml`, `linewrap`, `indent_character`,
  starred-entry write syntax, `-N` shorthand, encrypt/decrypt.

## Medium

- [ ] **#4** Skip malformed day files during Load instead of aborting `#bug` `#journal` *(2026-03-30)*
  A single file with a bad date or time heading causes `Load()` to abort — the entire journal
  becomes unreadable. Log a warning and skip unparseable files instead. jrnl's folder backend
  has the same crash behavior (arguably a bug), so this is a reasonable deviation to document
  in `docs/jrnl-compat.md`. `internal/journal/folder.go:120-123`

## Low

- [ ] **#3** `--config-override key=value` flag `#feature` `#config` *(2026-03-30)*
  Power-user escape hatch to override individual config keys at the command line.
- [ ] **#2** `--debug` flag `#feature` `#cli` *(2026-03-30)*
  Verbose diagnostic output for troubleshooting. No functional gap for normal use.

## Done

- [x] **#0** `--format pretty/short/tags/dates` display mode aliases `#feature` `#cli` *(2026-03-30)*
