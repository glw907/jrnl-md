# Interface Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Clean up package interfaces: export format constants, editor config struct, expose Encrypted() on FolderJournal, replace DayFiles() with LoadedPaths().

**Architecture:** Four independent interface improvements that each tighten a package boundary. Tasks are ordered so that internal packages change first, then callers update. Each task produces a compilable, passing codebase.

**Tech Stack:** Go 1.22, cobra, age encryption, stdlib only.

---

## File Structure

### Files to modify

| File | Changes |
|------|---------|
| `internal/export/tags.go` | Add format constants |
| `internal/editor/editor.go` | Add `Config` struct, update `PrepareDayFile` and `LaunchEncrypted` signatures |
| `internal/editor/editor_test.go` | Update all tests to use `Config` struct |
| `internal/journal/folder.go` | Add `loadedPaths` field, `Encrypted()`, `LoadedPaths()`; record paths in `Load`/`LoadDay`; remove `DayFiles()` |
| `internal/journal/folder_test.go` | Add `LoadedPaths` tests; update `TestEncryptDecryptConversion` |
| `cmd/jrnl-md/read.go` | Switch on export constants |
| `cmd/jrnl-md/edit.go` | Build `editor.Config`, drop `encrypted` param, use `fj.Encrypted()` |
| `cmd/jrnl-md/root.go` | Drop `encrypted` from `editEntry` call |
| `cmd/jrnl-md/encrypt.go` | Use `LoadedPaths()` instead of `DayFiles()` |

---

## Task 1: Add Export Format Constants

**Files:**
- Modify: `internal/export/tags.go`
- Modify: `cmd/jrnl-md/read.go`

- [ ] **Step 1: Add constants to tags.go**

Add to the top of `internal/export/tags.go`, after the imports:

```go
const (
	FormatJSON     = "json"
	FormatMarkdown = "md"
	FormatText     = "txt"
	FormatXML      = "xml"
	FormatYAML     = "yaml"
)
```

- [ ] **Step 2: Update read.go to use constants**

In `cmd/jrnl-md/read.go`, replace the export switch:

```go
	if f.export != "" {
		var output string
		var err error
		format := strings.ToLower(f.export)
		switch format {
		case export.FormatJSON:
			output, err = export.JSON(entries, cfg)
		case export.FormatMarkdown, "markdown":
			output, err = export.Markdown(entries, cfg)
		case export.FormatText, "text":
			output, err = export.Text(entries, cfg)
		case export.FormatXML:
			output, err = export.XML(entries, cfg)
		case export.FormatYAML:
			output, err = export.YAML(entries, cfg)
		default:
			return fmt.Errorf("unknown export format %q (supported: %s, %s, %s, %s, %s)",
				f.export, export.FormatJSON, export.FormatMarkdown,
				export.FormatText, export.FormatXML, export.FormatYAML)
		}
		if err != nil {
			return fmt.Errorf("exporting as %s: %w", format, err)
		}
		fmt.Print(output)
		return nil
	}
```

- [ ] **Step 3: Run tests**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go vet ./... && go test ./... -count=1`
Expected: ALL PASS.

- [ ] **Step 4: Commit**

```bash
cd /home/glw907/Projects/jrnl-md-rewrite
git add internal/export/tags.go cmd/jrnl-md/read.go
git commit -m "export: add format constants, use in read.go switch"
```

---

## Task 2: Add editor.Config Struct and Update Signatures

**Files:**
- Modify: `internal/editor/editor.go`
- Modify: `internal/editor/editor_test.go`

- [ ] **Step 1: Update tests to use Config struct**

Replace the three `prepareEncryptedContent` test calls in `internal/editor/editor_test.go`:

`TestPrepareEncryptedNew`:
```go
func TestPrepareEncryptedNew(t *testing.T) {
	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	cfg := Config{DateFmt: "2006-01-02", TimeFmt: "03:04 PM"}
	content, lineCount := prepareEncryptedContent("", date, cfg)

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
```

`TestPrepareEncryptedExisting`:
```go
func TestPrepareEncryptedExisting(t *testing.T) {
	existing := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMorning entry.\n"
	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	cfg := Config{DateFmt: "2006-01-02", TimeFmt: "03:04 PM"}
	content, _ := prepareEncryptedContent(existing, date, cfg)

	if !strings.Contains(content, "Morning entry.") {
		t.Error("lost existing content")
	}
	if !strings.Contains(content, "## [02:30 PM]") {
		t.Error("missing new entry heading")
	}
}
```

`TestPrepareEncryptedWithTemplate`:
```go
func TestPrepareEncryptedWithTemplate(t *testing.T) {
	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	cfg := Config{DateFmt: "2006-01-02", TimeFmt: "03:04 PM", Template: "## Mood\n"}
	content, _ := prepareEncryptedContent("", date, cfg)

	if !strings.Contains(content, "## Mood") {
		t.Error("missing template content")
	}
}
```

Update `TestPrepareDayFileNew`:
```go
func TestPrepareDayFileNew(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "29.md")

	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	cfg := Config{DateFmt: "2006-01-02", TimeFmt: "03:04 PM"}
	lineCount, err := PrepareDayFile(path, date, cfg)
	if err != nil {
		t.Fatalf("PrepareDayFile failed: %v", err)
	}

	if lineCount < 4 {
		t.Errorf("expected at least 4 lines, got %d", lineCount)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "# 2026-03-29 Sunday") {
		t.Errorf("missing day title, got: %q", content[:40])
	}
	if !strings.Contains(content, "## [02:30 PM]") {
		t.Errorf("missing entry heading, got: %q", content)
	}
}
```

Update `TestPrepareDayFileExisting`:
```go
func TestPrepareDayFileExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "29.md")

	existing := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMorning entry.\n"
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	cfg := Config{DateFmt: "2006-01-02", TimeFmt: "03:04 PM"}
	lineCount, err := PrepareDayFile(path, date, cfg)
	if err != nil {
		t.Fatalf("PrepareDayFile failed: %v", err)
	}

	if lineCount < 6 {
		t.Errorf("expected at least 6 lines, got %d", lineCount)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "Morning entry.") {
		t.Error("lost existing content")
	}
	if !strings.Contains(content, "## [02:30 PM]") {
		t.Error("missing new entry heading")
	}
	if strings.Count(content, "# 2026-03-29") != 1 {
		t.Error("should have exactly one day title")
	}
}
```

Update `TestPrepareDayFileWithTemplate`:
```go
func TestPrepareDayFileWithTemplate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "day.md")
	date := time.Date(2026, 3, 29, 17, 13, 0, 0, time.Local)

	tests := []struct {
		name     string
		template string
		check    func(t *testing.T, content string)
	}{
		{
			name:     "no template",
			template: "",
			check: func(t *testing.T, content string) {
				if strings.Count(content, "##") != 1 {
					t.Errorf("expected 1 entry heading, content:\n%s", content)
				}
			},
		},
		{
			name:     "simple template",
			template: "## Mood\n\n## Gratitude\n",
			check: func(t *testing.T, content string) {
				if !strings.Contains(content, "## Mood") {
					t.Error("missing template Mood heading")
				}
				if !strings.Contains(content, "## Gratitude") {
					t.Error("missing template Gratitude heading")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Remove(path)

			cfg := Config{DateFmt: "2006-01-02", TimeFmt: "03:04 PM", Template: tt.template}
			_, err := PrepareDayFile(path, date, cfg)
			if err != nil {
				t.Fatalf("PrepareDayFile() error: %v", err)
			}

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading file: %v", err)
			}
			tt.check(t, string(data))
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./internal/editor/ -v`
Expected: FAIL — `Config` undefined, wrong number of arguments.

- [ ] **Step 3: Add Config struct and update function signatures**

In `internal/editor/editor.go`, add the struct after the imports:

```go
// Config holds editor-related configuration.
type Config struct {
	Command    string
	DateFmt    string
	TimeFmt    string
	Passphrase string
	Template   string
}
```

Update `PrepareDayFile`:

```go
func PrepareDayFile(path string, date time.Time, cfg Config) (int, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("creating directory: %w", err)
	}

	var content string
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return 0, fmt.Errorf("reading existing day file: %w", err)
	}
	if err == nil {
		content = string(data)
	}

	if content == "" {
		content = journal.DayHeading(date, cfg.DateFmt) + "\n"
	}

	content += "\n" + journal.EntryHeading(date, cfg.TimeFmt, false) + "\n\n"

	if cfg.Template != "" {
		content += cfg.Template
		if !strings.HasSuffix(cfg.Template, "\n") {
			content += "\n"
		}
	}

	if err := atomicfile.WriteFile(path, []byte(content), 0644); err != nil {
		return 0, fmt.Errorf("writing day file: %w", err)
	}

	return countLines(content), nil
}
```

Update `prepareEncryptedContent`:

```go
func prepareEncryptedContent(existing string, date time.Time, cfg Config) (string, int) {
	if existing == "" {
		existing = journal.DayHeading(date, cfg.DateFmt) + "\n"
	}
	existing += "\n" + journal.EntryHeading(date, cfg.TimeFmt, false) + "\n\n"
	if cfg.Template != "" {
		existing += cfg.Template
		if !strings.HasSuffix(cfg.Template, "\n") {
			existing += "\n"
		}
	}
	return existing, countLines(existing)
}
```

Update `LaunchEncrypted`:

```go
func LaunchEncrypted(encPath string, date time.Time, cfg Config) error {
	var existing string
	data, err := os.ReadFile(encPath)
	if err == nil {
		plain, err := crypto.Decrypt(data, cfg.Passphrase)
		if err != nil {
			return fmt.Errorf("decrypting %s: %w", encPath, err)
		}
		existing = string(plain)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading %s: %w", encPath, err)
	}

	content, lineCount := prepareEncryptedContent(existing, date, cfg)

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

	if err := Launch(cfg.Command, tmpPath, lineCount); err != nil {
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

	enc, err := crypto.Encrypt(edited, cfg.Passphrase)
	if err != nil {
		return fmt.Errorf("encrypting: %w", err)
	}

	if err := atomicfile.WriteFile(encPath, enc, 0600); err != nil {
		return fmt.Errorf("writing %s: %w", encPath, err)
	}

	return nil
}
```

- [ ] **Step 4: Run editor tests**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./internal/editor/ -v`
Expected: ALL PASS.

- [ ] **Step 5: Update edit.go caller**

Replace `cmd/jrnl-md/edit.go` with:

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/editor"
	"github.com/glw907/jrnl-md/internal/journal"
)

func editEntry(fj *journal.FolderJournal, cfg config.Config, configPath string, passphrase string) error {
	if cfg.General.Editor == "" {
		return fmt.Errorf("no editor configured. Set editor in %s", configPath)
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

	ecfg := editor.Config{
		Command:    cfg.General.Editor,
		DateFmt:    cfg.Format.Date,
		TimeFmt:    cfg.Format.Time,
		Passphrase: passphrase,
		Template:   tmpl,
	}

	if fj.Encrypted() {
		return editor.LaunchEncrypted(fj.DayFilePath(now), now, ecfg)
	}

	path := fj.DayFilePath(now)
	lineCount, err := editor.PrepareDayFile(path, now, ecfg)
	if err != nil {
		return fmt.Errorf("preparing day file: %w", err)
	}

	return editor.Launch(cfg.General.Editor, path, lineCount)
}
```

- [ ] **Step 6: Update root.go caller**

In `cmd/jrnl-md/root.go`, replace:

```go
		return editEntry(fj, cfg, configPath, encrypted, passphrase)
```

With:

```go
		return editEntry(fj, cfg, configPath, passphrase)
```

- [ ] **Step 7: Verify — will fail until Task 3 adds Encrypted()**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go vet ./cmd/jrnl-md/`
Expected: FAIL — `fj.Encrypted` undefined. This is expected; Task 3 adds it.

Do NOT commit yet. Continue to Task 3.

---

## Task 3: Add Encrypted() and LoadedPaths() to FolderJournal

**Files:**
- Modify: `internal/journal/folder.go`
- Modify: `internal/journal/folder_test.go`

- [ ] **Step 1: Write failing tests**

Add to `internal/journal/folder_test.go`:

```go
func TestEncrypted(t *testing.T) {
	dir := t.TempDir()

	fj := NewFolderJournal(dir, testOpts)
	if fj.Encrypted() {
		t.Error("expected Encrypted() = false for plain journal")
	}

	encOpts := testOpts
	encOpts.Encrypt = true
	encOpts.Passphrase = "test"
	fj2 := NewFolderJournal(dir, encOpts)
	if !fj2.Encrypted() {
		t.Error("expected Encrypted() = true for encrypted journal")
	}
}

func TestLoadedPathsAfterLoad(t *testing.T) {
	dir := t.TempDir()

	content1 := "# 2026-03-28 Saturday\n\n## [09:00 AM]\n\nDay one.\n"
	content2 := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nDay two.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 28, 0, 0, 0, 0, time.Local), content1, "md")
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content2, "md")

	fj := NewFolderJournal(dir, testOpts)
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	paths := fj.LoadedPaths()
	if len(paths) != 2 {
		t.Fatalf("expected 2 loaded paths, got %d", len(paths))
	}
}

func TestLoadedPathsAfterLoadDay(t *testing.T) {
	dir := t.TempDir()

	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nToday.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	if err := fj.LoadDay(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("LoadDay failed: %v", err)
	}

	paths := fj.LoadedPaths()
	if len(paths) != 1 {
		t.Fatalf("expected 1 loaded path, got %d", len(paths))
	}
}

func TestLoadedPathsMissingDir(t *testing.T) {
	dir := t.TempDir()

	fj := NewFolderJournal(filepath.Join(dir, "nonexistent"), testOpts)
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	paths := fj.LoadedPaths()
	if len(paths) != 0 {
		t.Fatalf("expected 0 loaded paths, got %d", len(paths))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./internal/journal/ -run 'TestEncrypted|TestLoadedPaths' -v`
Expected: FAIL — `Encrypted` and `LoadedPaths` undefined.

- [ ] **Step 3: Add Encrypted(), loadedPaths field, LoadedPaths()**

In `internal/journal/folder.go`, update the struct:

```go
type FolderJournal struct {
	path        string
	opts        Options
	days        map[dateKey]*day
	tagParser   *TagParser
	loadedPaths []string
}
```

Add methods after `DayFilePath`:

```go
// Encrypted reports whether the journal uses encryption.
func (fj *FolderJournal) Encrypted() bool { return fj.opts.Encrypt }

// LoadedPaths returns the file paths read by Load or LoadDay.
func (fj *FolderJournal) LoadedPaths() []string { return fj.loadedPaths }
```

- [ ] **Step 4: Record paths in Load**

In `Load`, after the successful `parseDay` + tag parsing block (just before `fj.days[key] = &parsed`), add:

```go
		fj.loadedPaths = append(fj.loadedPaths, path)
```

- [ ] **Step 5: Record path in LoadDay**

In `LoadDay`, after `fj.days[key] = &parsed`, add:

```go
	fj.loadedPaths = append(fj.loadedPaths, path)
```

- [ ] **Step 6: Remove DayFiles()**

Delete the entire `DayFiles` method (lines 371-391 of folder.go).

- [ ] **Step 7: Update TestEncryptDecryptConversion**

In `internal/journal/folder_test.go`, replace the two `DayFiles()` calls:

Replace:
```go
	oldFiles, err := fj.DayFiles()
	if err != nil {
		t.Fatal(err)
	}
```

With:
```go
	oldFiles := fj.LoadedPaths()
```

Replace:
```go
	encFiles, _ := fj2.DayFiles()
```

With:
```go
	encFiles := fj2.LoadedPaths()
```

- [ ] **Step 8: Run all journal tests**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./internal/journal/ -v`
Expected: ALL PASS.

- [ ] **Step 9: Update encrypt.go caller**

In `cmd/jrnl-md/encrypt.go`, replace:

```go
	oldFiles, err := fj.DayFiles()
	if err != nil {
		return fmt.Errorf("listing day files: %w", err)
	}
```

With:

```go
	oldFiles := fj.LoadedPaths()
```

- [ ] **Step 10: Run full test suite**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go vet ./... && go test ./... -count=1`
Expected: ALL PASS.

- [ ] **Step 11: Commit Tasks 2 + 3 together**

```bash
cd /home/glw907/Projects/jrnl-md-rewrite
git add internal/editor/editor.go internal/editor/editor_test.go internal/journal/folder.go internal/journal/folder_test.go cmd/jrnl-md/edit.go cmd/jrnl-md/root.go cmd/jrnl-md/encrypt.go
git commit -m "editor: Config struct; journal: Encrypted(), LoadedPaths(); remove DayFiles"
```

---

## Task 4: Final Verification

**Files:** None (verification only).

- [ ] **Step 1: Run go vet**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go vet ./...`
Expected: no issues.

- [ ] **Step 2: Run full test suite**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./... -count=1`
Expected: ALL PASS.

- [ ] **Step 3: Run e2e tests**

Run: `cd /home/glw907/Projects/jrnl-md-rewrite && go test ./e2e/ -count=1 -v`
Expected: ALL PASS.

- [ ] **Step 4: Verify no raw format strings remain in read.go**

Confirm `read.go` has no bare `"json"`, `"md"`, `"txt"`, `"xml"`, `"yaml"` strings in the switch cases (only `export.Format*` constants and the aliases `"markdown"`, `"text"`).

- [ ] **Step 5: Verify DayFiles is fully removed**

Confirm `DayFiles` does not appear in any `.go` file (only in the plan/spec docs).

- [ ] **Step 6: Verify editor.go no longer has loose string params**

Confirm `PrepareDayFile` signature is `(string, time.Time, Config)` and `LaunchEncrypted` is `(string, time.Time, Config)`.
