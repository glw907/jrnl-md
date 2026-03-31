# BACKLOG

> Project issue tracker. Managed by `/log-issue`.

## High

- [ ] **#6** Abort --edit when editor returns empty content `#bug` `#edit` *(2026-03-30)*
  jrnl-md silently deletes all selected entries via `ReplaceEntries(old, nil)` when the user
  empties the editor buffer. jrnl aborts with a warning: "Were you trying to delete all the
  entries? This seems a bit drastic, so the operation was cancelled." Match that behavior.
  For partial deletions (some entries removed), jrnl accepts silently and prints a post-hoc
  count — match that too. `cmd/jrnl-md/edit.go:73-78`, `internal/journal/day.go:127-129`
- [ ] **#5** Normalize spacing in day files after direct --edit `#improvement` `#edit` *(2026-03-30)*
  The direct `--edit` path (no filter flags) edits the raw day file in place with no re-parse.
  If the user introduces extra blank lines, trailing whitespace, or missing blank lines around
  `## [time]` headings, the damage persists. The filtered `--edit` path already normalizes via
  its parse→format round-trip. Add a post-edit normalize step to the direct path: re-read the
  file, parse it, and re-serialize with `day.Format()` to enforce canonical spacing. This is
  safe because the parser already `TrimSpace`s bodies. `cmd/jrnl-md/edit.go:45-51`,
  `internal/journal/day.go:22-34`
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
