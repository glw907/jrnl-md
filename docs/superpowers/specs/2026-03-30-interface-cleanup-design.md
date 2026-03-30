# Interface Cleanup Design

## Goal

Clean up package interfaces identified during the simplify review:
eliminate stringly-typed export formats, reduce parameter sprawl in the
editor package, remove redundant state passing between CLI and journal,
and eliminate the double directory walk in reencryptJournal.

No user-facing behavior changes.

## 1. Export Format Constants

Add string constants to the `export` package:

```go
const (
    FormatJSON     = "json"
    FormatMarkdown = "md"
    FormatText     = "txt"
    FormatXML      = "xml"
    FormatYAML     = "yaml"
)
```

`read.go` switches on these constants instead of raw strings. The flag
description and error message derive from the constants. `"markdown"`
and `"text"` remain as aliases in the switch — they are user-facing
convenience, not canonical format names.

## 2. Editor Config Struct

Replace the 7 individual parameters on `LaunchEncrypted` and the
repeated `dateFmt, timeFmt` pairs on `PrepareDayFile` with a struct:

```go
// in internal/editor/editor.go
type Config struct {
    Command    string // editor command (e.g. "vim", "code")
    DateFmt    string
    TimeFmt    string
    Passphrase string
    Template   string
}
```

Function signatures become:

```go
func PrepareDayFile(path string, date time.Time, cfg Config) (int, error)
func LaunchEncrypted(encPath string, date time.Time, cfg Config) error
func Launch(cmd string, path string, line int) error  // unchanged
```

`Launch` stays as-is — it is a generic "open file in editor" utility
that does not need journal context. `PrepareDayFile` and
`LaunchEncrypted` are journal-specific and benefit from the struct.

The CLI layer builds `editor.Config` once from `config.Config` and
passes it down. `Passphrase` is empty for unencrypted journals;
`LaunchEncrypted` uses it, `PrepareDayFile` ignores it.

## 3. Expose Encrypted() on FolderJournal

Add a one-line method:

```go
func (fj *FolderJournal) Encrypted() bool { return fj.opts.Encrypt }
```

`editEntry` currently receives `encrypted bool` as a separate parameter
alongside the `*FolderJournal` that already carries the same
information. With this method, the signature simplifies:

```go
// before
func editEntry(fj *journal.FolderJournal, cfg config.Config,
    configPath string, encrypted bool, passphrase string) error

// after
func editEntry(fj *journal.FolderJournal, cfg config.Config,
    configPath string, passphrase string) error
```

The `passphrase` parameter stays — it comes from user input (terminal
prompt), not from the journal. The journal stores it in `opts` for its
own Load/Save, but the editor needs it independently for the temp-file
encrypt cycle.

## 4. Eliminate Double WalkDir in reencryptJournal

`reencryptJournal` calls `fj.Load()` (WalkDir #1), then `fj.DayFiles()`
(WalkDir #2) to get the old file paths for removal. Both walk the same
tree.

Fix: have `Load` record the paths it visits. Add a field to
`FolderJournal`:

```go
loadedPaths []string  // paths read by Load/LoadDay
```

`Load` appends each file path as it processes it. `LoadDay` appends the
single path. A new method exposes them:

```go
func (fj *FolderJournal) LoadedPaths() []string { return fj.loadedPaths }
```

`reencryptJournal` replaces `fj.DayFiles()` with `fj.LoadedPaths()`.
`DayFiles()` is removed — its only callers are `reencryptJournal` and
two tests (`TestEncryptDecryptConversion`), which switch to
`LoadedPaths()`.

## 5. Testing Impact

- All existing tests pass unchanged — the changes are signature
  refactors, not behavior changes
- `editor_test.go`: `TestPrepareEncryptedNew/Existing/WithTemplate`
  update to use `editor.Config` instead of individual params.
  `TestPrepareDayFile*` tests update similarly.
- `folder_test.go`: Add a test verifying `LoadedPaths` returns correct
  paths after `Load` and after `LoadDay`.
  `TestEncryptDecryptConversion` switches from `DayFiles()` to
  `LoadedPaths()`.
- `cmd/` tests: `TestCompletionSubcommands` and `TestPreprocessArgs`
  unchanged — they do not touch the refactored signatures
- e2e tests: unchanged — they test the binary's CLI interface, not
  internal signatures

## 6. Summary of Changes by Package

| Package            | Changes                                                                    |
| ------------------ | -------------------------------------------------------------------------- |
| `internal/export`  | Add format constants (`FormatJSON`, etc.)                                  |
| `internal/editor`  | Add `Config` struct, update `PrepareDayFile` and `LaunchEncrypted` sigs    |
| `internal/journal` | Add `Encrypted()`, add `loadedPaths` + `LoadedPaths()`, remove `DayFiles` |
| `cmd/jrnl-md/read.go`    | Switch on export constants                                          |
| `cmd/jrnl-md/edit.go`    | Build `editor.Config`, drop `encrypted` param, use `fj.Encrypted()` |
| `cmd/jrnl-md/root.go`    | Drop `encrypted` from `editEntry` call                              |
| `cmd/jrnl-md/encrypt.go` | Use `LoadedPaths()` instead of `DayFiles()`                         |
