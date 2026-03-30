# Architectural Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the monolithic `root.go` into focused operation files with a closed-over flags struct, add heading helpers to the journal package, move encrypted editing into the editor package, simplify `reencryptJournal`, and add targeted `LoadDay` loading.

**Architecture:** The `cmd/jrnl-md/root.go` file (~700 lines, 15 package-level flag vars) is split into 10 focused files. A local `flags` struct replaces globals, closed over by cobra's `RunE`. The `editor` package gains `LaunchEncrypted`. The `journal` package exports `DayHeading` and `EntryHeading` helpers. `FolderJournal` gains `LoadDay` for write/edit paths.

**Tech Stack:** Go 1.22, cobra, age encryption, stdlib only for new code.

---

## File Structure

### Files to create (cmd/jrnl-md/)

| File            | Responsibility                                          |
| --------------- | ------------------------------------------------------- |
| `write.go`      | `writeInline` function                                  |
| `edit.go`       | `editEntry` dispatcher (plain path only)                |
| `read.go`       | `readEntries`, `showTags`                               |
| `delete.go`     | `deleteEntries` with own confirm loop                   |
| `changetime.go` | `changeTime` with own confirm loop                      |
| `encrypt.go`    | `encryptJournal`, `decryptJournal`, `reencryptJournal`, `promptPassphrase`, `promptNewPassphrase` |

### Files to modify

| File                            | Changes                                                               |
| ------------------------------- | --------------------------------------------------------------------- |
| `cmd/jrnl-md/root.go`          | Replace 15 globals with `flags` struct, slim down to routing only     |
| `cmd/jrnl-md/args.go`          | Move `parseArgs` here from root.go (already has `preprocessArgs`)     |
| `internal/journal/entry.go`    | Add `DayHeading`, `EntryHeading` exported helpers                     |
| `internal/journal/day.go`      | Use `DayHeading`, `EntryHeading` in `Format`                         |
| `internal/journal/folder.go`   | Add `LoadDay` method                                                  |
| `internal/editor/editor.go`    | Add `LaunchEncrypted` function                                        |

### Files unchanged

`cmd/jrnl-md/main.go`, `cmd/jrnl-md/completion.go`, all `internal/` packages not listed above, all test files (they continue to pass).

---

## Task 1: Export Heading Helpers from journal Package

**Files:**
- Modify: `internal/journal/entry.go`
- Modify: `internal/journal/day.go`
- Test: `internal/journal/day_test.go`

- [ ] **Step 1: Write failing tests for DayHeading and EntryHeading**

Add to `internal/journal/entry_test.go`:

```go
func TestDayHeading(t *testing.T) {
	date := time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local)
	got := DayHeading(date, "2006-01-02")
	want := "# 2026-03-29 Sunday"
	if got != want {
		t.Errorf("DayHeading() = %q, want %q", got, want)
	}
}

func TestEntryHeading(t *testing.T) {
	tests := []struct {
		name    string
		date    time.Time
		timeFmt string
		starred bool
		want    string
	}{
		{
			name:    "plain",
			date:    time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local),
			timeFmt: "03:04 PM",
			starred: false,
			want:    "## [02:30 PM]",
		},
		{
			name:    "starred",
			date:    time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
			timeFmt: "03:04 PM",
			starred: true,
			want:    "## [09:00 AM] *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EntryHeading(tt.date, tt.timeFmt, tt.starred)
			if got != tt.want {
				t.Errorf("EntryHeading() = %q, want %q", got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./internal/journal/ -run 'TestDayHeading|TestEntryHeading' -v`
Expected: FAIL — `DayHeading` and `EntryHeading` undefined.

- [ ] **Step 3: Add DayHeading and EntryHeading to entry.go**

Add to the end of `internal/journal/entry.go`:

```go
// DayHeading returns the markdown day heading for a date.
func DayHeading(date time.Time, dateFmt string) string {
	return fmt.Sprintf("# %s %s", date.Format(dateFmt), date.Format("Monday"))
}

// EntryHeading returns the markdown entry heading for a timestamp.
func EntryHeading(date time.Time, timeFmt string, starred bool) string {
	h := fmt.Sprintf("## [%s]", date.Format(timeFmt))
	if starred {
		h += " *"
	}
	return h
}
```

- [ ] **Step 4: Update day.Format to use DayHeading and EntryHeading**

In `internal/journal/day.go`, replace the `Format` method body:

```go
func (d day) Format(dateFmt, timeFmt string) string {
	var b strings.Builder

	b.WriteString(DayHeading(d.date, dateFmt))
	b.WriteString("\n")

	for _, e := range d.entries {
		b.WriteString("\n")
		b.WriteString(e.Format(timeFmt))
	}

	return b.String()
}
```

Update `Entry.Format` to use `EntryHeading`:

```go
func (e Entry) Format(timeFmt string) string {
	heading := EntryHeading(e.Date, timeFmt, e.Starred)

	body := strings.TrimRight(e.Body, "\n ")
	if body != "" {
		return heading + "\n\n" + body + "\n"
	}
	return heading + "\n"
}
```

Remove the now-unused `fmt` import from `day.go` (the `fmt` import moves to `entry.go` which already has it).

- [ ] **Step 5: Run all journal tests**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./internal/journal/ -v`
Expected: ALL PASS (including existing day/entry/folder/filter tests).

- [ ] **Step 6: Commit**

```bash
cd /home/glw907/Projects/jrnl-md-rewrite
git add internal/journal/entry.go internal/journal/entry_test.go internal/journal/day.go
git commit -m "journal: export DayHeading and EntryHeading helpers"
```

---

## Task 2: Add LoadDay to FolderJournal

**Files:**
- Modify: `internal/journal/folder.go`
- Test: `internal/journal/folder_test.go`

- [ ] **Step 1: Write failing tests for LoadDay**

Add to `internal/journal/folder_test.go`:

```go
func TestLoadDay(t *testing.T) {
	dir := t.TempDir()

	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nToday's entry.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	// Also write another day that LoadDay should NOT load.
	content2 := "# 2026-03-28 Saturday\n\n## [10:00 AM]\n\nYesterday.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 28, 0, 0, 0, 0, time.Local), content2, "md")

	fj := NewFolderJournal(dir, testOpts)
	if err := fj.LoadDay(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("LoadDay failed: %v", err)
	}

	entries := fj.AllEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Body != "Today's entry." {
		t.Errorf("body = %q", entries[0].Body)
	}
}

func TestLoadDayMissingFile(t *testing.T) {
	dir := t.TempDir()

	fj := NewFolderJournal(dir, testOpts)
	if err := fj.LoadDay(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("LoadDay on missing file should succeed, got: %v", err)
	}

	entries := fj.AllEntries()
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestLoadDayEncrypted(t *testing.T) {
	dir := t.TempDir()

	// Write an encrypted day file via Save, then LoadDay it.
	opts := testOpts
	opts.Encrypt = true
	opts.Passphrase = "testpass"

	fj := NewFolderJournal(dir, opts)
	fj.AddEntry(time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), "Secret.", false)
	if err := fj.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	fj2 := NewFolderJournal(dir, opts)
	if err := fj2.LoadDay(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("LoadDay encrypted failed: %v", err)
	}

	entries := fj2.AllEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Body != "Secret." {
		t.Errorf("body = %q", entries[0].Body)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./internal/journal/ -run 'TestLoadDay' -v`
Expected: FAIL — `LoadDay` undefined.

- [ ] **Step 3: Implement LoadDay**

Add to `internal/journal/folder.go` after the `Load` method:

```go
// LoadDay reads and parses only the day file for the given date. If the
// file does not exist, LoadDay succeeds with no entries for that day.
func (fj *FolderJournal) LoadDay(date time.Time) error {
	plainExt := "." + fj.opts.FileExt
	encExt := plainExt + ".age"

	base := filepath.Join(
		fj.path,
		fmt.Sprintf("%04d", date.Year()),
		fmt.Sprintf("%02d", int(date.Month())),
		fmt.Sprintf("%02d", date.Day()),
	)

	var path string
	var encrypted bool

	if fj.opts.Encrypt {
		path = base + encExt
		encrypted = true
	} else {
		path = base + plainExt
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	if encrypted {
		data, err = crypto.Decrypt(data, fj.opts.Passphrase)
		if err != nil {
			return fmt.Errorf("decrypting %s: %w", path, err)
		}
	}

	parsed, err := parseDay(string(data), fj.opts.DateFmt, fj.opts.TimeFmt)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	for i := range parsed.entries {
		parsed.entries[i].Tags = fj.tagParser.Parse(parsed.entries[i].Body)
	}

	key := dateKeyFromTime(date)
	fj.days[key] = &parsed

	return nil
}
```

Add `"errors"` to the import block in `folder.go` (it is not currently imported there).

- [ ] **Step 4: Run all journal tests**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./internal/journal/ -v`
Expected: ALL PASS.

- [ ] **Step 5: Commit**

```bash
cd /home/glw907/Projects/jrnl-md-rewrite
git add internal/journal/folder.go internal/journal/folder_test.go
git commit -m "journal: add LoadDay for targeted single-day loading"
```

---

## Task 3: Move LaunchEncrypted into the editor Package

**Files:**
- Modify: `internal/editor/editor.go`
- Test: `internal/editor/editor_test.go`

- [ ] **Step 1: Write failing test for LaunchEncrypted content preparation**

We test the content preparation logic, not the actual editor launch. Add to `internal/editor/editor_test.go`:

```go
func TestPrepareEncryptedNew(t *testing.T) {
	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	content, lineCount := prepareEncryptedContent("", date, "2006-01-02", "03:04 PM", "")

	if !strings.HasPrefix(content, "# 2026-03-29 Sunday") {
		t.Errorf("missing day heading, got: %q", content[:40])
	}
	if !strings.Contains(content, "## [02:30 PM]") {
		t.Errorf("missing entry heading")
	}
	if lineCount < 4 {
		t.Errorf("expected at least 4 lines, got %d", lineCount)
	}
}

func TestPrepareEncryptedExisting(t *testing.T) {
	existing := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMorning entry.\n"
	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	content, _ := prepareEncryptedContent(existing, date, "2006-01-02", "03:04 PM", "")

	if !strings.Contains(content, "Morning entry.") {
		t.Error("lost existing content")
	}
	if !strings.Contains(content, "## [02:30 PM]") {
		t.Error("missing new entry heading")
	}
}

func TestPrepareEncryptedWithTemplate(t *testing.T) {
	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	content, _ := prepareEncryptedContent("", date, "2006-01-02", "03:04 PM", "## Mood\n")

	if !strings.Contains(content, "## Mood") {
		t.Error("missing template content")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./internal/editor/ -run 'TestPrepareEncrypted' -v`
Expected: FAIL — `prepareEncryptedContent` undefined.

- [ ] **Step 3: Implement prepareEncryptedContent and LaunchEncrypted**

Add to `internal/editor/editor.go`. First update imports to add `"github.com/glw907/jrnl-md/internal/crypto"`:

```go
import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/glw907/jrnl-md/internal/atomicfile"
	"github.com/glw907/jrnl-md/internal/crypto"
)
```

Then add after the existing `Launch` function:

```go
// prepareEncryptedContent builds the editor content for an encrypted day file.
// If existing is empty, a new day heading is created. A new entry heading is
// always appended.
func prepareEncryptedContent(existing string, date time.Time, dateFmt, timeFmt, template string) (string, int) {
	if existing == "" {
		existing = fmt.Sprintf("# %s %s\n", date.Format(dateFmt), date.Format("Monday"))
	}
	existing += fmt.Sprintf("\n## [%s]\n\n", date.Format(timeFmt))
	if template != "" {
		existing += template
		if !strings.HasSuffix(template, "\n") {
			existing += "\n"
		}
	}
	return existing, countLines(existing)
}

// LaunchEncrypted decrypts the day file (if it exists), appends an entry
// heading, opens the editor, then re-encrypts and writes atomically.
func LaunchEncrypted(editorCmd, encPath string, date time.Time, dateFmt, timeFmt, passphrase, template string) error {
	var existing string
	data, err := os.ReadFile(encPath)
	if err == nil {
		plain, err := crypto.Decrypt(data, passphrase)
		if err != nil {
			return fmt.Errorf("decrypting %s: %w", encPath, err)
		}
		existing = string(plain)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading %s: %w", encPath, err)
	}

	content, lineCount := prepareEncryptedContent(existing, date, dateFmt, timeFmt, template)

	tmpFile, err := os.CreateTemp("", "jrnl-md-*.md")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := Launch(editorCmd, tmpPath, lineCount); err != nil {
		return fmt.Errorf("launching editor: %w", err)
	}

	edited, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("reading edited file: %w", err)
	}

	dir := filepath.Dir(encPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	enc, err := crypto.Encrypt(edited, passphrase)
	if err != nil {
		return fmt.Errorf("encrypting: %w", err)
	}

	if err := atomicfile.WriteFile(encPath, enc, 0600); err != nil {
		return fmt.Errorf("writing %s: %w", encPath, err)
	}

	return nil
}
```

- [ ] **Step 4: Run all editor tests**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./internal/editor/ -v`
Expected: ALL PASS.

- [ ] **Step 5: Commit**

```bash
cd /home/glw907/Projects/jrnl-md-rewrite
git add internal/editor/editor.go internal/editor/editor_test.go
git commit -m "editor: add LaunchEncrypted for encrypted day file editing"
```

---

## Task 4: Create the flags Struct and Rewrite root.go

This is the largest task. It replaces 15 package-level vars with a struct, rewrites `newRootCmd` to close over it, and slims `root.go` down to routing logic only. All operation functions are extracted to separate files in subsequent tasks.

**Files:**
- Modify: `cmd/jrnl-md/root.go`

- [ ] **Step 1: Define the flags struct and rewrite newRootCmd**

Replace the entire `cmd/jrnl-md/root.go` with:

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

var version = "0.1.0"

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

func newRootCmd() *cobra.Command {
	var f flags

	cmd := &cobra.Command{
		Use:          "jrnl-md [journal:] [text...]",
		Short:        "A markdown-native journal for the command line",
		Long:         "jrnl-md is a journaling CLI that stores entries as markdown day files.",
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRoot(cmd, args, &f)
		},
	}

	cmd.Flags().IntVarP(&f.n, "num", "n", 0, "Show last N entries")
	cmd.Flags().BoolVarP(&f.short, "short", "s", false, "Show short entry list")
	cmd.Flags().BoolVar(&f.starred, "starred", false, "Show only starred entries")
	cmd.Flags().BoolVar(&f.edit, "edit", false, "Open entries in editor")
	cmd.Flags().BoolVar(&f.delete, "delete", false, "Delete entries")
	cmd.Flags().BoolVar(&f.encrypt, "encrypt", false, "Encrypt the journal")
	cmd.Flags().BoolVar(&f.decrypt, "decrypt", false, "Decrypt the journal")
	cmd.Flags().StringVar(&f.changeTime, "change-time", "", "Change entry time")
	cmd.Flag("change-time").NoOptDefVal = "now"
	cmd.Flags().StringVar(&f.from, "from", "", "Show entries from date")
	cmd.Flags().StringVar(&f.to, "to", "", "Show entries to date")
	cmd.Flags().StringVar(&f.on, "on", "", "Show entries on date")
	cmd.Flags().StringVar(&f.contains, "contains", "", "Filter entries containing text")
	cmd.Flags().StringVar(&f.export, "export", "", "Export format (json, md, txt, xml, yaml)")
	cmd.Flags().BoolVar(&f.list, "list", false, "List configured journals")
	cmd.Flags().BoolVar(&f.tags, "tags", false, "List all tags")
	cmd.Flags().BoolVarP(&f.version, "version", "v", false, "Show version")
	cmd.Flags().StringVar(&f.configFile, "config", "", "Config file path")

	cmd.AddCommand(newCompletionCmd())

	return cmd
}

func runRoot(cmd *cobra.Command, args []string, f *flags) error {
	if f.version {
		fmt.Printf("jrnl-md %s\n", version)
		return nil
	}

	configPath := f.configFile
	if configPath == "" {
		var err error
		configPath, err = config.DefaultPath()
		if err != nil {
			return err
		}
	}

	cfg, err := config.Load(configPath)
	if errors.Is(err, os.ErrNotExist) {
		cfg = config.Default()
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("determining home directory: %w", err)
		}
		cfg.Journals["default"] = config.JournalConfig{
			Path: home + "/Documents/Journal/",
		}
		if err := config.Save(cfg, configPath); err != nil {
			return fmt.Errorf("saving default config: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Created config at %s\n", configPath)
	} else if err != nil {
		return fmt.Errorf("loading config %s: %w", configPath, err)
	}

	if f.list {
		return listJournals(cfg)
	}

	journalName, text, tagArgs := parseArgs(args, cfg)

	journalCfg, ok := cfg.Journals[journalName]
	if !ok {
		return fmt.Errorf("journal %q not found in config", journalName)
	}

	path, err := expandPath(journalCfg.Path)
	if err != nil {
		return fmt.Errorf("expanding path for journal %q: %w", journalName, err)
	}

	encrypted := journalEncrypted(journalCfg, cfg)

	if f.encrypt {
		return encryptJournal(path, journalName, cfg, configPath)
	}
	if f.decrypt {
		return decryptJournal(path, journalName, cfg, configPath)
	}

	var passphrase string
	if encrypted {
		passphrase, err = promptPassphrase(fmt.Sprintf("Passphrase for journal %q: ", journalName))
		if err != nil {
			return err
		}
	}

	opts := journalOptions(cfg, encrypted, passphrase)
	now := time.Now()

	// Dispatch: write and edit use LoadDay; read operations use Load.
	if len(text) > 0 {
		fj := journal.NewFolderJournal(path, opts)
		if err := fj.LoadDay(now); err != nil {
			return fmt.Errorf("loading journal: %w", err)
		}
		return writeInline(fj, text, cfg)
	}

	if f.edit || (len(args) == 0 && !hasFilterFlags(f)) {
		fj := journal.NewFolderJournal(path, opts)
		if err := fj.LoadDay(now); err != nil {
			return fmt.Errorf("loading journal: %w", err)
		}
		return editEntry(fj, cfg, encrypted, passphrase)
	}

	fj := journal.NewFolderJournal(path, opts)
	if err := fj.Load(); err != nil {
		return fmt.Errorf("loading journal: %w", err)
	}

	if f.delete {
		return deleteEntries(fj, cfg, f, tagArgs)
	}

	if f.changeTime != "" {
		return changeTime(fj, cfg, f, tagArgs)
	}

	return readEntries(fj, cfg, f, tagArgs)
}

func listJournals(cfg config.Config) error {
	fmt.Println("Journals:")
	for name, j := range cfg.Journals {
		fmt.Printf("  %s -> %s\n", name, j.Path)
	}
	return nil
}

func hasFilterFlags(f *flags) bool {
	return f.n > 0 || f.short || f.starred || f.delete || f.encrypt || f.decrypt || f.changeTime != "" || f.from != "" || f.to != "" || f.on != "" || f.contains != "" || f.tags || f.export != ""
}

func journalEncrypted(jcfg config.JournalConfig, cfg config.Config) bool {
	if jcfg.Encrypt != nil {
		return *jcfg.Encrypt
	}
	return cfg.General.Encrypt
}

func journalOptions(cfg config.Config, encrypted bool, passphrase string) journal.Options {
	return journal.Options{
		DateFmt:    cfg.Format.Date,
		TimeFmt:    cfg.Format.Time,
		TagSymbols: cfg.Format.TagSymbols,
		FileExt:    cfg.Format.FileExtension,
		Encrypt:    encrypted,
		Passphrase: passphrase,
	}
}

func boolPtr(v bool) *bool { return &v }

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("determining home directory: %w", err)
		}
		return home + path[1:], nil
	}
	return path, nil
}

func buildFilter(f *flags, tagArgs []string) (journal.Filter, error) {
	var flt journal.Filter
	flt.N = f.n
	flt.Starred = f.starred

	if len(tagArgs) > 0 {
		flt.Tags = tagArgs
	}

	if f.contains != "" {
		flt.Contains = f.contains
	}

	if f.on != "" {
		start, err := dateparse.Parse(f.on)
		if err != nil {
			return flt, fmt.Errorf("parsing --on date: %w", err)
		}
		startOfDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
		endOfDay := time.Date(start.Year(), start.Month(), start.Day(), 23, 59, 59, 0, time.Local)
		flt.StartDate = &startOfDay
		flt.EndDate = &endOfDay
	}

	if f.from != "" {
		start, err := dateparse.Parse(f.from)
		if err != nil {
			return flt, fmt.Errorf("parsing --from date: %w", err)
		}
		flt.StartDate = &start
	}

	if f.to != "" {
		end, err := dateparse.ParseInclusive(f.to)
		if err != nil {
			return flt, fmt.Errorf("parsing --to date: %w", err)
		}
		flt.EndDate = &end
	}

	return flt, nil
}
```

Note: `buildFilter` needs the `dateparse` import. The full import block for `root.go` is:

```go
import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/dateparse"
	"github.com/glw907/jrnl-md/internal/journal"
)
```

- [ ] **Step 2: Verify it compiles (will fail until operation files exist)**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go vet ./cmd/jrnl-md/`
Expected: errors about undefined functions (`writeInline`, `editEntry`, `readEntries`, `deleteEntries`, `changeTime`, `encryptJournal`, `decryptJournal`, `promptPassphrase`). This is expected — we create those files in the next tasks.

Do NOT commit yet. Continue to Tasks 5-10 to create the operation files before committing.

---

## Task 5: Create write.go

**Files:**
- Create: `cmd/jrnl-md/write.go`

- [ ] **Step 1: Create write.go**

```go
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

func writeInline(fj *journal.FolderJournal, text []string, cfg config.Config) error {
	body := strings.Join(text, " ")
	starred := strings.HasSuffix(body, "*") || strings.HasPrefix(body, "*")
	if starred {
		body = strings.Trim(body, "* ")
	}

	fj.AddEntry(time.Now(), body, starred)

	if err := fj.Save(); err != nil {
		return fmt.Errorf("saving journal: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Entry added.")
	return nil
}
```

---

## Task 6: Create edit.go

**Files:**
- Create: `cmd/jrnl-md/edit.go`

- [ ] **Step 1: Create edit.go**

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/editor"
	"github.com/glw907/jrnl-md/internal/journal"
)

func editEntry(fj *journal.FolderJournal, cfg config.Config, encrypted bool, passphrase string) error {
	if cfg.General.Editor == "" {
		msg := "no editor configured"
		if configPath, err := config.DefaultPath(); err == nil {
			msg = fmt.Sprintf("no editor configured. Set editor in %s", configPath)
		}
		return errors.New(msg)
	}

	now := time.Now()

	var tmpl string
	if cfg.General.Template != "" {
		tmplPath, err := expandPath(cfg.General.Template)
		if err != nil {
			return fmt.Errorf("expanding template path: %w", err)
		}
		data, err := os.ReadFile(tmplPath)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", tmplPath, err)
		}
		tmpl = string(data)
	}

	if encrypted {
		return editor.LaunchEncrypted(cfg.General.Editor, fj.DayFilePath(now), now,
			cfg.Format.Date, cfg.Format.Time, passphrase, tmpl)
	}

	path := fj.DayFilePath(now)
	lineCount, err := editor.PrepareDayFile(path, now, cfg.Format.Date, cfg.Format.Time, tmpl)
	if err != nil {
		return fmt.Errorf("preparing day file: %w", err)
	}

	return editor.Launch(cfg.General.Editor, path, lineCount)
}
```

---

## Task 7: Create read.go

**Files:**
- Create: `cmd/jrnl-md/read.go`

- [ ] **Step 1: Create read.go**

```go
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/display"
	"github.com/glw907/jrnl-md/internal/export"
	"github.com/glw907/jrnl-md/internal/journal"
)

func readEntries(fj *journal.FolderJournal, cfg config.Config, f *flags, tagArgs []string) error {
	entries := fj.AllEntries()

	flt, err := buildFilter(f, tagArgs)
	if err != nil {
		return fmt.Errorf("building filter: %w", err)
	}
	entries = flt.Apply(entries)

	if f.tags {
		return showTags(entries)
	}

	fmt.Fprintf(os.Stderr, "%d entries found\n", len(entries))

	if len(entries) == 0 {
		return nil
	}

	if f.export != "" {
		var output string
		var err error
		switch strings.ToLower(f.export) {
		case "json":
			output, err = export.JSON(entries, cfg)
		case "md", "markdown":
			output, err = export.Markdown(entries, cfg)
		case "txt", "text":
			output, err = export.Text(entries, cfg)
		case "xml":
			output, err = export.XML(entries, cfg)
		case "yaml":
			output, err = export.YAML(entries, cfg)
		default:
			return fmt.Errorf("unknown export format %q (supported: json, md, txt, xml, yaml)", f.export)
		}
		if err != nil {
			return fmt.Errorf("exporting as %s: %w", f.export, err)
		}
		fmt.Print(output)
		return nil
	}

	linewrap := cfg.General.Linewrap
	if linewrap == 0 {
		linewrap = display.TerminalWidth()
	}

	indent := ""
	if cfg.General.IndentCharacter != "" {
		indent = cfg.General.IndentCharacter + " "
	}

	dateColor := display.ColorFunc(cfg.Colors.Date)
	bodyColor := display.ColorFunc(cfg.Colors.Body)

	for _, e := range entries {
		if f.short {
			fmt.Println(e.FormatShort(cfg.Format.Date, cfg.Format.Time))
		} else {
			dateStr := dateColor(e.Date.Format(cfg.Format.Date + " " + cfg.Format.Time))
			fmt.Println(dateStr)
			fmt.Println()

			body := e.Body
			if linewrap > 0 && indent != "" {
				body = display.WrapText(body, linewrap-len(indent))
			} else if linewrap > 0 {
				body = display.WrapText(body, linewrap)
			}
			if indent != "" {
				body = display.IndentBody(body, indent)
			}
			fmt.Println(bodyColor(body))
		}
	}

	return nil
}

func showTags(entries []journal.Entry) error {
	counts := export.TagCounts(entries)

	if len(counts) == 0 {
		fmt.Println("No tags found.")
		return nil
	}

	tags := make([]string, 0, len(counts))
	for tag := range counts {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	for _, tag := range tags {
		fmt.Printf("%-20s : %d\n", tag, counts[tag])
	}
	return nil
}
```

---

## Task 8: Create delete.go

**Files:**
- Create: `cmd/jrnl-md/delete.go`

- [ ] **Step 1: Create delete.go**

```go
package main

import (
	"fmt"
	"os"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/glw907/jrnl-md/internal/prompt"
)

func deleteEntries(fj *journal.FolderJournal, cfg config.Config, f *flags, tagArgs []string) error {
	entries := fj.AllEntries()

	flt, err := buildFilter(f, tagArgs)
	if err != nil {
		return fmt.Errorf("building filter: %w", err)
	}
	entries = flt.Apply(entries)

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No entries to delete, because the search returned no results.")
		return nil
	}

	if len(entries) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry found.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries found.\n", len(entries))
	}

	var toDelete []journal.Entry
	for _, e := range entries {
		msg := fmt.Sprintf("Delete entry '%s'?", e.FormatShort(cfg.Format.Date, cfg.Format.Time))
		if prompt.YesNo(os.Stdin, os.Stderr, msg) {
			toDelete = append(toDelete, e)
		}
	}

	if len(toDelete) == 0 {
		return nil
	}

	fj.DeleteEntries(toDelete)

	if err := fj.Save(); err != nil {
		return fmt.Errorf("saving journal: %w", err)
	}

	if len(toDelete) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry deleted.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries deleted.\n", len(toDelete))
	}

	return nil
}
```

---

## Task 9: Create changetime.go

**Files:**
- Create: `cmd/jrnl-md/changetime.go`

- [ ] **Step 1: Create changetime.go**

```go
package main

import (
	"fmt"
	"os"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/dateparse"
	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/glw907/jrnl-md/internal/prompt"
)

func changeTime(fj *journal.FolderJournal, cfg config.Config, f *flags, tagArgs []string) error {
	newTime, err := dateparse.Parse(f.changeTime)
	if err != nil {
		return fmt.Errorf("parsing --change-time date: %w", err)
	}

	entries := fj.AllEntries()

	flt, err := buildFilter(f, tagArgs)
	if err != nil {
		return fmt.Errorf("building filter: %w", err)
	}
	entries = flt.Apply(entries)

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No entries to modify.")
		return nil
	}

	if len(entries) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry found.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries found.\n", len(entries))
	}

	var toChange []journal.Entry
	for _, e := range entries {
		msg := fmt.Sprintf("Change time for '%s'?", e.FormatShort(cfg.Format.Date, cfg.Format.Time))
		if prompt.YesNo(os.Stdin, os.Stderr, msg) {
			toChange = append(toChange, e)
		}
	}

	if len(toChange) == 0 {
		return nil
	}

	fj.ChangeEntryTimes(toChange, newTime)

	if err := fj.Save(); err != nil {
		return fmt.Errorf("saving journal: %w", err)
	}

	if len(toChange) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry modified.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries modified.\n", len(toChange))
	}

	return nil
}
```

---

## Task 10: Create encrypt.go and Simplify reencryptJournal

**Files:**
- Create: `cmd/jrnl-md/encrypt.go`

- [ ] **Step 1: Create encrypt.go**

```go
package main

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

func promptPassphrase(msg string) (string, error) {
	fmt.Fprint(os.Stderr, msg)
	pass, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("reading passphrase: %w", err)
	}
	return string(pass), nil
}

func promptNewPassphrase() (string, error) {
	pass, err := promptPassphrase("New passphrase: ")
	if err != nil {
		return "", err
	}
	if pass == "" {
		return "", fmt.Errorf("passphrase cannot be empty")
	}
	confirm, err := promptPassphrase("Confirm passphrase: ")
	if err != nil {
		return "", err
	}
	if pass != confirm {
		return "", fmt.Errorf("passphrases do not match")
	}
	return pass, nil
}

func encryptJournal(journalPath, journalName string, cfg config.Config, configPath string) error {
	if journalEncrypted(cfg.Journals[journalName], cfg) {
		return fmt.Errorf("journal %q is already encrypted", journalName)
	}
	passphrase, err := promptNewPassphrase()
	if err != nil {
		return err
	}
	return reencryptJournal(journalPath, journalName, cfg, configPath, false, passphrase, true)
}

func decryptJournal(journalPath, journalName string, cfg config.Config, configPath string) error {
	if !journalEncrypted(cfg.Journals[journalName], cfg) {
		return fmt.Errorf("journal %q is not encrypted", journalName)
	}
	passphrase, err := promptPassphrase(fmt.Sprintf("Passphrase for journal %q: ", journalName))
	if err != nil {
		return err
	}
	return reencryptJournal(journalPath, journalName, cfg, configPath, true, passphrase, false)
}

func reencryptJournal(journalPath, journalName string, cfg config.Config, configPath string, fromEncrypt bool, passphrase string, toEncrypt bool) error {
	fj := journal.NewFolderJournal(journalPath, journalOptions(cfg, fromEncrypt, passphrase))
	if err := fj.Load(); err != nil {
		return fmt.Errorf("loading journal: %w", err)
	}

	oldFiles, err := fj.DayFiles()
	if err != nil {
		return fmt.Errorf("listing day files: %w", err)
	}

	fj.MarkAllModified()
	fj.SetEncryption(toEncrypt, passphrase)
	if err := fj.Save(); err != nil {
		return fmt.Errorf("saving journal: %w", err)
	}

	for _, f := range oldFiles {
		os.Remove(f)
	}

	jcfg := cfg.Journals[journalName]
	jcfg.Encrypt = boolPtr(toEncrypt)
	cfg.Journals[journalName] = jcfg
	if err := config.Save(cfg, configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	verb := "encrypted"
	if !toEncrypt {
		verb = "decrypted"
	}
	fmt.Fprintf(os.Stderr, "Journal %q %s (%d files).\n", journalName, verb, len(oldFiles))
	return nil
}
```

---

## Task 11: Move parseArgs to args.go

**Files:**
- Modify: `cmd/jrnl-md/args.go`

- [ ] **Step 1: Move parseArgs from root.go to args.go**

Append to `cmd/jrnl-md/args.go`:

```go
import (
	"strings"

	"github.com/glw907/jrnl-md/internal/config"
)

func parseArgs(args []string, cfg config.Config) (journalName string, text []string, tagArgs []string) {
	if len(args) == 0 {
		return "default", nil, nil
	}

	first := args[0]
	if strings.HasSuffix(first, ":") {
		name := strings.TrimSuffix(first, ":")
		if _, ok := cfg.Journals[name]; ok {
			return name, args[1:], nil
		}
	}

	if len(cfg.Format.TagSymbols) > 0 {
		allTags := true
		var tags []string
		for _, arg := range args {
			if len(arg) > 1 && strings.ContainsRune(cfg.Format.TagSymbols, rune(arg[0])) {
				tags = append(tags, arg)
			} else {
				allTags = false
				break
			}
		}
		if allTags && len(tags) > 0 {
			return "default", nil, tags
		}
	}

	return "default", args, nil
}
```

The full `args.go` file will have both `preprocessArgs` (existing) and `parseArgs` (moved from root.go).

---

## Task 12: Compile and Run Full Test Suite

**Files:** None (verification only)

- [ ] **Step 1: Verify compilation**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go vet ./...`
Expected: no errors.

- [ ] **Step 2: Run unit tests**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./... -v`
Expected: ALL PASS.

- [ ] **Step 3: Run e2e tests**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./e2e/ -v`
Expected: ALL PASS.

- [ ] **Step 4: Commit the file split (Tasks 4-11)**

```bash
cd /home/glw907/Projects/jrnl-md-rewrite
git add cmd/jrnl-md/root.go cmd/jrnl-md/args.go cmd/jrnl-md/write.go cmd/jrnl-md/edit.go cmd/jrnl-md/read.go cmd/jrnl-md/delete.go cmd/jrnl-md/changetime.go cmd/jrnl-md/encrypt.go
git commit -m "cmd: split root.go into operation files with flags struct"
```

---

## Task 13: Use Heading Helpers in editor Package

Now that `journal.DayHeading` and `journal.EntryHeading` exist, update `editor.go` and `prepareEncryptedContent` to use them instead of inline `fmt.Sprintf`.

**Files:**
- Modify: `internal/editor/editor.go`

- [ ] **Step 1: Update PrepareDayFile to use heading helpers**

In `internal/editor/editor.go`, add `"github.com/glw907/jrnl-md/internal/journal"` to imports.

Replace the heading construction in `PrepareDayFile`:

Old:
```go
	if content == "" {
		content = fmt.Sprintf("# %s %s\n", date.Format(dateFmt), date.Format("Monday"))
	}

	content += fmt.Sprintf("\n## [%s]\n\n", date.Format(timeFmt))
```

New:
```go
	if content == "" {
		content = journal.DayHeading(date, dateFmt) + "\n"
	}

	content += "\n" + journal.EntryHeading(date, timeFmt, false) + "\n\n"
```

- [ ] **Step 2: Update prepareEncryptedContent to use heading helpers**

Old:
```go
	if existing == "" {
		existing = fmt.Sprintf("# %s %s\n", date.Format(dateFmt), date.Format("Monday"))
	}
	existing += fmt.Sprintf("\n## [%s]\n\n", date.Format(timeFmt))
```

New:
```go
	if existing == "" {
		existing = journal.DayHeading(date, dateFmt) + "\n"
	}
	existing += "\n" + journal.EntryHeading(date, timeFmt, false) + "\n\n"
```

- [ ] **Step 3: Run editor and journal tests**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./internal/editor/ ./internal/journal/ -v`
Expected: ALL PASS.

- [ ] **Step 4: Commit**

```bash
cd /home/glw907/Projects/jrnl-md-rewrite
git add internal/editor/editor.go
git commit -m "editor: use journal heading helpers instead of inline formatting"
```

---

## Task 14: Final Verification and Cleanup

**Files:** All modified files.

- [ ] **Step 1: Run full test suite**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./... -v`
Expected: ALL PASS.

- [ ] **Step 2: Run e2e tests**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./e2e/ -v`
Expected: ALL PASS.

- [ ] **Step 3: Run go vet**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go vet ./...`
Expected: no issues.

- [ ] **Step 4: Verify no old code remains in root.go**

Confirm that `root.go` no longer contains: `editEncrypted`, `secureRemove`, `writeInline`, `editEntry`, `readEntries`, `deleteEntries`, `changeTime`, `showTags`, `encryptJournal`, `decryptJournal`, `reencryptJournal`, `promptPassphrase`, `promptNewPassphrase`, or `parseArgs`.

Confirm that `root.go` no longer imports: `atomicfile`, `crypto`, `display`, `editor`, `export`, `prompt`, `golang.org/x/term`, `path/filepath`, `sort`.

- [ ] **Step 5: Verify file structure**

```
cmd/jrnl-md/
  main.go         # unchanged
  args.go          # preprocessArgs + parseArgs
  root.go          # flags struct, newRootCmd, runRoot, routing helpers
  write.go         # writeInline
  edit.go          # editEntry (plain + encrypted dispatch)
  read.go          # readEntries, showTags
  delete.go        # deleteEntries
  changetime.go    # changeTime
  encrypt.go       # encrypt/decrypt/reencrypt, passphrase prompts
  completion.go    # unchanged
```
