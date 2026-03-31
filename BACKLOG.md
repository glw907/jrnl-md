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
  Both --edit paths should validate and normalize after the editor exits. Spec:
  **Filtered --edit** (temp file): if `ParseMultiDay` returns a parse error, show the
  specific error to stderr and offer to re-open the editor (the temp file still exists).
  If empty result, abort with warning per #6. If fewer entries, accept and print count
  (already works).
  **Direct --edit** (raw day file): back up the pre-edit content. After the editor exits,
  re-parse with `parseDay`. If it parses, re-serialize with `day.Format()` to normalize
  spacing. If it fails, show the parse error and offer to re-open. If the user declines,
  restore the backup. Never silently accept broken structure — that defers the problem to
  the next `Load()`.
  Key locations: `cmd/jrnl-md/edit.go`, `internal/journal/day.go:22-34`
- [ ] **#1** Compat test suite audit `#improvement` `#docs` *(2026-03-30)*
  Cross-reference every feature in `docs/jrnl-compat.md` against `e2e/jrnl_compat_test.go`.
  Confirm each has a real assertion. Add `TestCompat_*` tests for gaps. Known gaps:
  `--format pretty/short/tags/dates`, `--export xml/yaml`, `linewrap`, `indent_character`,
  starred-entry write syntax, `-N` shorthand, encrypt/decrypt.

## Medium

- [ ] **#4** Skip malformed day files during Load instead of aborting `#bug` `#journal` *(2026-03-30)*
  A single file with a bad date or time heading causes `Load()` to abort — the entire journal
  becomes unreadable. Spec: log a warning to stderr with the file path and specific parse
  error, skip the file, continue loading everything else. The user can then go fix that one
  file manually. jrnl's folder backend has the same crash behavior (arguably a bug), so this
  is a reasonable deviation to document in `docs/jrnl-compat.md`.
  `internal/journal/folder.go:120-123`

## Low

- [ ] **#3** `--config-override key=value` flag `#feature` `#config` *(2026-03-30)*
  Power-user escape hatch to override individual config keys at the command line.
- [ ] **#2** `--debug` flag `#feature` `#cli` *(2026-03-30)*
  Verbose diagnostic output for troubleshooting. No functional gap for normal use.

## Done

- [x] **#0** `--format pretty/short/tags/dates` display mode aliases `#feature` `#cli` *(2026-03-30)*
