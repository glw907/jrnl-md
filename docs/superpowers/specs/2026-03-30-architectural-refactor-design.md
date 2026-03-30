# Architectural Refactor Design

## Goal

Clean package boundaries, eliminate shared mutable state, remove
duplication, and add targeted loading — while preserving the identical
CLI interface.

## 1. Flags Struct (root.go)

Replace 15 package-level `var` declarations with a local struct:

```go
type flags struct {
    n          int
    short      bool
    starred    bool
    edit       bool
    delete     bool
    encrypt    bool
    decrypt    bool
    changeTime string
    from       string
    to         string
    on         string
    contains   string
    export     string
    list       bool
    tags       bool
    version    bool
    configFile string
}
```

Created in `newRootCmd()`, bound via `cmd.Flags().IntVarP(&f.n, ...)`,
closed over in `RunE`. All operation functions receive the struct or
specific fields — no globals.

## 2. File Split (cmd/jrnl-md/)

| File            | Contents                                                                                                                              |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `main.go`       | `main()` — unchanged                                                                                                                 |
| `args.go`       | `preprocessArgs`, `parseArgs`                                                                                                        |
| `root.go`       | `flags` struct, `newRootCmd`, `runRoot`, `buildFilter`, `hasFilterFlags`, `listJournals`, `journalOptions`, `journalEncrypted`, `expandPath`, `boolPtr` |
| `write.go`      | `writeInline`                                                                                                                        |
| `edit.go`       | `editEntry` — thin dispatcher, delegates to `editor` package                                                                         |
| `read.go`       | `readEntries`, `showTags`                                                                                                            |
| `delete.go`     | `deleteEntries` — self-contained with own confirm loop                                                                               |
| `changetime.go` | `changeTime` — self-contained with own confirm loop                                                                                  |
| `encrypt.go`    | `encryptJournal`, `decryptJournal`, `reencryptJournal`, `promptPassphrase`, `promptNewPassphrase`                                    |
| `completion.go` | unchanged                                                                                                                            |

## 3. Editor Package Owns Encrypted Editing

Move `editEncrypted` logic into the `editor` package. New function:

```go
func LaunchEncrypted(editorCmd, encPath string, date time.Time,
    dateFmt, timeFmt, passphrase, template string) error
```

Handles: read existing encrypted file, decrypt, build content with
headings, write temp file, launch editor, read back, encrypt, atomic
write.

Root.go's `edit.go` becomes a thin dispatcher:

```go
if encrypted {
    return editor.LaunchEncrypted(cfg.General.Editor, fj.DayFilePath(now), now, ...)
}
// Plain path: existing PrepareDayFile + Launch sequence
path := fj.DayFilePath(now)
lineCount, err := editor.PrepareDayFile(path, now, ...)
if err != nil { return err }
return editor.Launch(cfg.General.Editor, path, lineCount)
```

The `editor` package gains `crypto` and `atomicfile` as dependencies.
Root.go drops both imports.

`secureRemove` is replaced by `defer os.Remove(tmpPath)` inside the
editor package.

## 4. Heading Helpers in journal Package

Export two helpers from `journal`:

```go
func DayHeading(date time.Time, dateFmt string) string {
    return fmt.Sprintf("# %s %s", date.Format(dateFmt), date.Format("Monday"))
}

func EntryHeading(date time.Time, timeFmt string, starred bool) string {
    h := fmt.Sprintf("## [%s]", date.Format(timeFmt))
    if starred {
        h += " *"
    }
    return h
}
```

Used by: `day.Format`, `Entry.Format`, `editor.PrepareDayFile`,
`editor.LaunchEncrypted`. Eliminates three sites of duplicated heading
construction.

## 5. Simplify reencryptJournal

Drop `successMsg` parameter — derive from `toEncrypt`:

```go
func reencryptJournal(journalPath, journalName string, cfg config.Config,
    configPath string, fromEncrypt bool, passphrase string, toEncrypt bool) error {
    // ...
    verb := "encrypted"
    if !toEncrypt {
        verb = "decrypted"
    }
    fmt.Fprintf(os.Stderr, "Journal %q %s (%d files).\n", journalName, verb, len(oldFiles))
}
```

7 params instead of 8.

## 6. Targeted Loading (LoadDay)

Add to `FolderJournal`:

```go
func (fj *FolderJournal) LoadDay(date time.Time) error
```

Reads and parses only the single day file for the given date. Creates
the day in `fj.days` if the file exists, otherwise does nothing (new
day will be created on `AddEntry`).

Usage in dispatch (`runRoot`):

| Operation          | Load method          |
| ------------------ | -------------------- |
| `writeInline`      | `LoadDay(time.Now())` |
| `editEntry`        | `LoadDay(now)`       |
| `readEntries`      | `Load()`             |
| `deleteEntries`    | `Load()`             |
| `changeTime`       | `Load()`             |
| `encryptJournal`   | `Load()`             |
| `decryptJournal`   | `Load()`             |

Each operation function loads what it needs. `runRoot` no longer calls
`Load()` unconditionally. `NewFolderJournal(...)` stays in `runRoot`,
loading moves into the operation functions.

## 7. Testing Impact

- Existing tests pass unchanged — `ParseTags`, `Filter.Apply`,
  `day.Format`/`parseDay`, all export tests, all e2e tests
- New unit tests needed for: `LoadDay`, `DayHeading`, `EntryHeading`,
  `editor.LaunchEncrypted` (test content prep, not the actual editor
  launch)
- cmd tests — `TestCompletionSubcommands` and `TestPreprocessArgs`
  continue to work since `newRootCmd()` still returns a valid cobra
  command

## 8. Summary of Changes by Package

| Package          | Changes                                                                                              |
| ---------------- | ---------------------------------------------------------------------------------------------------- |
| `cmd/jrnl-md`    | Split into 10 files, flags struct, per-operation loading, simplified `reencryptJournal`               |
| `internal/journal` | Add `LoadDay`, export `DayHeading` + `EntryHeading`                                                |
| `internal/editor`  | Add `LaunchEncrypted`, use heading helpers, gain `crypto` + `atomicfile` deps                      |
