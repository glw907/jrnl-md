# Pass 7: Edit Validation & Day-File Edit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add post-edit validation with actionable errors and re-open loop to all edit paths, abort on empty buffer, and redirect single-day full-match filtered edits to open the day file directly.

**Architecture:** `parseDay`/`ParseMultiDay` gain structured `ParseError` returns with line numbers. A new `cleanup.go` file handles light whitespace normalization. The editor package gains a validation loop with re-open prompt. `editFiltered` in `cmd/jrnl-md/edit.go` detects single-day full-match and redirects to the direct edit path.

**Tech Stack:** Go 1.25, existing `internal/prompt`, `internal/editor`, `internal/journal` packages

**Spec:** `docs/superpowers/specs/2026-03-30-edit-validation-design.md`

---

### Task 1: ParseError structured type

**Files:**
- Modify: `internal/journal/day.go`
- Test: `internal/journal/day_test.go`

- [ ] **Step 1: Write the failing test for ParseError**

Add to `internal/journal/day_test.go`:

```go
func TestParseDayErrorMissingTitle(t *testing.T) {
	_, err := parseDay("no heading here\n\n## [09:00 AM]\n\nBody.\n", "2006-01-02", "03:04 PM")
	if err == nil {
		t.Fatal("expected error for missing day heading")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	if pe.Line != 1 {
		t.Errorf("line = %d, want 1", pe.Line)
	}
	if !strings.Contains(pe.Error(), "expected") {
		t.Errorf("error should contain expected format: %s", pe.Error())
	}
}

func TestParseDayErrorBadTime(t *testing.T) {
	_, err := parseDay("# 2026-03-29 Sunday\n\n## [3:59pm]\n\nBody.\n", "2006-01-02", "03:04 PM")
	if err == nil {
		t.Fatal("expected error for bad time")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	if pe.Line == 0 {
		t.Error("expected non-zero line number")
	}
	if pe.Value != "3:59pm" {
		t.Errorf("value = %q, want %q", pe.Value, "3:59pm")
	}
}

func TestParseDayErrorBadDate(t *testing.T) {
	_, err := parseDay("# not-a-date Sunday\n\n## [09:00 AM]\n\nBody.\n", "2006-01-02", "03:04 PM")
	if err == nil {
		t.Fatal("expected error for bad date")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	if pe.Line != 1 {
		t.Errorf("line = %d, want 1", pe.Line)
	}
}

func TestParseMultiDayErrorIncludesLineOffset(t *testing.T) {
	// Second day section has a bad time — error line should be relative to full text
	text := "# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nGood entry.\n\n# 2026-03-15 Sunday\n\n## [bad-time]\n\nBad entry.\n"
	_, err := ParseMultiDay(text, "2006-01-02", "03:04 PM")
	if err == nil {
		t.Fatal("expected error")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	// The bad time is on line 9 of the full text (1-based)
	if pe.Line < 7 {
		t.Errorf("line = %d, expected >= 7 (offset into full text)", pe.Line)
	}
}
```

Add `"errors"` and `"strings"` to the test file imports if not already present.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./internal/journal/ -run "TestParseDay(Error|.*Error)" -v`
Expected: FAIL — `ParseError` undefined

- [ ] **Step 3: Implement ParseError and update parseDay**

In `internal/journal/day.go`, add the `ParseError` type and update `parseDay` to return structured errors. Add `"errors"` to imports if needed.

Add `ParseError` type before `parseDay`:

```go
// ParseError reports a parse failure with location and context.
type ParseError struct {
	File     string // file path or description (set by caller)
	Line     int    // 1-based line number in the source text
	Value    string // the bad value found
	Expected string // what was expected
}

func (e *ParseError) Error() string {
	loc := e.File
	if loc == "" {
		loc = "input"
	}
	if e.Value != "" {
		return fmt.Sprintf("%s: line %d: can't parse %q (expected %s)", loc, e.Line, e.Value, e.Expected)
	}
	return fmt.Sprintf("%s: line %d: %s (expected %s)", loc, e.Line, "missing", e.Expected)
}
```

Replace `parseDay` with a version that returns `*ParseError`:

```go
func parseDay(text, dateFmt, timeFmt string) (day, error) {
	var d day

	lines := strings.Split(text, "\n")

	titleMatch := titleRe.FindStringSubmatch(text)
	if titleMatch == nil {
		return d, &ParseError{
			Line:     1,
			Expected: fmt.Sprintf("day heading like \"# %s %s\"", time.Now().Format(dateFmt), time.Now().Format("Monday")),
		}
	}

	dayDate, err := time.ParseInLocation(dateFmt, titleMatch[1], time.Local)
	if err != nil {
		return d, &ParseError{
			Line:     1,
			Value:    titleMatch[1],
			Expected: fmt.Sprintf("date in format %q", dateFmt),
		}
	}
	d.date = dayDate

	matches := entryRe.FindAllStringSubmatchIndex(text, -1)

	for i, match := range matches {
		timeStr := text[match[2]:match[3]]

		// Compute 1-based line number for this match
		lineNum := 1 + strings.Count(text[:match[0]], "\n")
		_ = lines // used indirectly via text

		entryTime, err := time.ParseInLocation(timeFmt, timeStr, time.Local)
		if err != nil {
			return d, &ParseError{
				Line:     lineNum,
				Value:    timeStr,
				Expected: fmt.Sprintf("time in format %q, e.g. \"## [%s]\"", timeFmt, time.Now().Format(timeFmt)),
			}
		}

		starred := match[4] != -1

		entryDate := time.Date(
			dayDate.Year(), dayDate.Month(), dayDate.Day(),
			entryTime.Hour(), entryTime.Minute(), entryTime.Second(),
			0, time.Local,
		)

		bodyStart := match[1]
		var bodyEnd int
		if i+1 < len(matches) {
			bodyEnd = matches[i+1][0]
		} else {
			bodyEnd = len(text)
		}

		body := strings.TrimSpace(text[bodyStart:bodyEnd])

		d.entries = append(d.entries, Entry{
			Date:    entryDate,
			Body:    body,
			Starred: starred,
		})
	}

	return d, nil
}
```

- [ ] **Step 4: Update ParseMultiDay to propagate line offsets**

Replace `ParseMultiDay` in `internal/journal/day.go`:

```go
func ParseMultiDay(text, dateFmt, timeFmt string) ([]Entry, error) {
	lines := strings.Split(text, "\n")
	var sectionStarts []int
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") && len(line) > 2 && line[2] >= '0' && line[2] <= '9' {
			sectionStarts = append(sectionStarts, i)
		}
	}

	if len(sectionStarts) == 0 {
		return nil, nil
	}

	var entries []Entry
	for i, start := range sectionStarts {
		end := len(lines)
		if i+1 < len(sectionStarts) {
			end = sectionStarts[i+1]
		}
		section := strings.Join(lines[start:end], "\n")
		d, err := parseDay(section, dateFmt, timeFmt)
		if err != nil {
			var pe *ParseError
			if errors.As(err, &pe) {
				// Offset line number to be relative to the full text
				pe.Line += start
				return nil, pe
			}
			return nil, fmt.Errorf("parsing day section at line %d: %w", start+1, err)
		}
		entries = append(entries, d.entries...)
	}

	return entries, nil
}
```

Add `"errors"` to the `day.go` import block.

- [ ] **Step 5: Run all tests**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./internal/journal/ -v`
Expected: All PASS (new ParseError tests + existing parseDay/ParseMultiDay tests)

- [ ] **Step 6: Run e2e tests for regression**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -v -count=1`
Expected: All PASS/SKIP

- [ ] **Step 7: Commit**

```bash
git add internal/journal/day.go internal/journal/day_test.go
git commit -m "feat: add ParseError structured type with line numbers to parseDay/ParseMultiDay"
```

---

### Task 2: Light cleanup functions

**Files:**
- Create: `internal/journal/cleanup.go`
- Test: `internal/journal/cleanup_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/journal/cleanup_test.go`:

```go
package journal

import (
	"testing"
)

func TestCleanupDayContent(t *testing.T) {
	t.Run("strips trailing empty entry heading", func(t *testing.T) {
		input := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nReal entry.\n\n## [02:30 PM]\n\n"
		got := CleanupDayContent(input, "03:04 PM")
		want := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nReal entry.\n"
		if got != want {
			t.Errorf("got:\n%q\nwant:\n%q", got, want)
		}
	})

	t.Run("preserves non-empty trailing entry", func(t *testing.T) {
		input := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nReal entry.\n\n## [02:30 PM]\n\nAlso real.\n"
		got := CleanupDayContent(input, "03:04 PM")
		if got != input {
			t.Errorf("should not change content with no empty headings:\ngot: %q\nwant: %q", got, input)
		}
	})

	t.Run("normalizes blank lines before ## headings", func(t *testing.T) {
		input := "# 2026-03-29 Sunday\n\n\n\n## [09:00 AM]\n\nEntry.\n\n\n\n\n## [02:30 PM]\n\nSecond.\n"
		got := CleanupDayContent(input, "03:04 PM")
		want := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nEntry.\n\n## [02:30 PM]\n\nSecond.\n"
		if got != want {
			t.Errorf("got:\n%q\nwant:\n%q", got, want)
		}
	})

	t.Run("trims trailing whitespace from lines", func(t *testing.T) {
		input := "# 2026-03-29 Sunday  \n\n## [09:00 AM]  \n\nEntry with spaces.  \n"
		got := CleanupDayContent(input, "03:04 PM")
		want := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nEntry with spaces.\n"
		if got != want {
			t.Errorf("got:\n%q\nwant:\n%q", got, want)
		}
	})

	t.Run("ensures single trailing newline", func(t *testing.T) {
		input := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nEntry.\n\n\n\n"
		got := CleanupDayContent(input, "03:04 PM")
		want := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nEntry.\n"
		if got != want {
			t.Errorf("got:\n%q\nwant:\n%q", got, want)
		}
	})

	t.Run("no change returns identical content", func(t *testing.T) {
		input := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nClean content.\n"
		got := CleanupDayContent(input, "03:04 PM")
		if got != input {
			t.Errorf("expected no change, got:\n%q", got)
		}
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./internal/journal/ -run TestCleanupDayContent -v`
Expected: FAIL — `CleanupDayContent` undefined

- [ ] **Step 3: Implement CleanupDayContent**

Create `internal/journal/cleanup.go`:

```go
package journal

import (
	"regexp"
	"strings"
)

// CleanupDayContent applies light normalization to a day file's content:
//   - strips trailing empty entry headings (## [time] with no body)
//   - normalizes to exactly one blank line before ## headings
//   - trims trailing whitespace from each line
//   - ensures a single trailing newline
func CleanupDayContent(text, timeFmt string) string {
	lines := strings.Split(text, "\n")

	// Trim trailing whitespace from each line
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}

	// Strip trailing empty entry headings:
	// A ## [time] heading is "empty" if everything after it is blank.
	for {
		// Find the last ## heading
		lastHeading := -1
		for i := len(lines) - 1; i >= 0; i-- {
			if entryRe.MatchString(lines[i]) {
				lastHeading = i
				break
			}
		}
		if lastHeading == -1 {
			break
		}
		// Check if everything after it is blank
		allBlank := true
		for i := lastHeading + 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) != "" {
				allBlank = false
				break
			}
		}
		if !allBlank {
			break
		}
		// Remove the heading and everything after it
		lines = lines[:lastHeading]
		// Also remove any trailing blank lines before where the heading was
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}
	}

	// Normalize blank lines before ## headings: exactly one blank line
	blankBeforeHeading := regexp.MustCompile(`(?m)\n{3,}(## \[)`)
	result := strings.Join(lines, "\n")
	result = blankBeforeHeading.ReplaceAllString(result, "\n\n$1")

	// Ensure single trailing newline
	result = strings.TrimRight(result, "\n") + "\n"

	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./internal/journal/ -run TestCleanupDayContent -v`
Expected: All PASS

- [ ] **Step 5: Run all journal tests**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./internal/journal/ -v`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add internal/journal/cleanup.go internal/journal/cleanup_test.go
git commit -m "feat: add CleanupDayContent for light post-edit normalization"
```

---

### Task 3: Editor validation loop

Add a validation function to the editor package that handles the parse → error → re-open loop, and an empty-content check.

**Files:**
- Modify: `internal/editor/editor.go`
- Test: `internal/editor/editor_test.go`

- [ ] **Step 1: Write the failing tests**

Read `internal/editor/editor_test.go` first to understand existing test patterns. Then add:

```go
func TestIsEmptyContent(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", true},
		{"   \n\n  ", true},
		{"\t\n", true},
		{"# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nEntry.\n", false},
		{"some text", false},
	}
	for _, tt := range tests {
		got := IsEmptyContent(tt.input)
		if got != tt.want {
			t.Errorf("IsEmptyContent(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./internal/editor/ -run TestIsEmptyContent -v`
Expected: FAIL — `IsEmptyContent` undefined

- [ ] **Step 3: Implement IsEmptyContent**

Add to `internal/editor/editor.go`:

```go
// IsEmptyContent returns true if the text is empty or whitespace-only.
func IsEmptyContent(text string) bool {
	return strings.TrimSpace(text) == ""
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./internal/editor/ -run TestIsEmptyContent -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/editor/editor.go internal/editor/editor_test.go
git commit -m "feat: add IsEmptyContent check for post-edit validation"
```

---

### Task 4: Direct edit with validation

Rewrite the direct edit flow (`editEntry`) to include backup, validation, cleanup, empty check, and re-open loop. Also add the encrypted variant.

**Files:**
- Modify: `cmd/jrnl-md/edit.go`
- Modify: `internal/editor/editor.go`
- Test: `e2e/edit_test.go`

- [ ] **Step 1: Add e2e test for empty buffer abort**

Add to `e2e/edit_test.go`:

```go
func TestEditEmptyBufferAborts(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	// Mock editor that empties the file
	script := "#!/bin/bash\nFILE=\"${@: -1}\"\necho -n '' > \"$FILE\"\n"
	editorPath := filepath.Join(env.dir, "empty-editor.sh")
	if err := os.WriteFile(editorPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	patchConfigEditor(t, env.configPath, editorPath)

	// Edit with a filter that matches entries on March 1
	_, stderr, _ := runErr(t, env, "--edit", "--on", "2026-03-01")

	if !strings.Contains(stderr, "no changes made") && !strings.Contains(stderr, "No entries found") {
		t.Errorf("expected abort message, got: %q", stderr)
	}

	// Verify entries are still intact
	march1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	content := dayFileContent(t, env.journalDir, march1)
	if !strings.Contains(content, "First @work entry") {
		t.Errorf("entries should be preserved after empty buffer abort, got:\n%s", content)
	}
}
```

- [ ] **Step 2: Add e2e test for direct edit with cleanup**

Add to `e2e/edit_test.go`:

```go
func TestEditDirectCleansUpEmptyHeading(t *testing.T) {
	env := newTestEnv(t)

	// Write a day file with one entry
	today := time.Now()
	dayContent := fmt.Sprintf("# %s %s\n\n## [09:00 AM]\n\nExisting entry.\n",
		today.Format("2006-01-02"), today.Format("Monday"))
	writeDayFile(t, env.journalDir, today, dayContent)

	// Mock editor that does nothing (leaves the appended empty heading as-is)
	script := "#!/bin/bash\n# no-op editor\n"
	editorPath := filepath.Join(env.dir, "noop-editor.sh")
	if err := os.WriteFile(editorPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	patchConfigEditor(t, env.configPath, editorPath)

	// Bare edit — should append a heading, then cleanup should strip it
	run(t, env)

	content := dayFileContent(t, env.journalDir, today)
	if !strings.Contains(content, "Existing entry.") {
		t.Errorf("existing entry should be preserved, got:\n%s", content)
	}

	// The appended empty ## heading should have been cleaned up
	headingCount := strings.Count(content, "## [")
	if headingCount != 1 {
		t.Errorf("expected 1 entry heading after cleanup, got %d in:\n%s", headingCount, content)
	}
}
```

- [ ] **Step 3: Run e2e tests to verify they fail**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/ -run "TestEdit(EmptyBuffer|DirectCleans)" -v -count=1`
Expected: FAIL

- [ ] **Step 4: Update editEntry with validation flow**

Replace `editEntry` in `cmd/jrnl-md/edit.go`:

```go
func editEntry(fj *journal.FolderJournal, cfg config.Config, configPath string, passphrase string) error {
	return editDayFile(fj, cfg, configPath, passphrase, time.Now(), true)
}

// editDayFile opens a day file in the editor with validation, cleanup, and
// re-open loop. If appendHeading is true, a new entry heading is appended
// before opening (bare edit). If false, the file is opened as-is (redirected
// filtered edit).
func editDayFile(fj *journal.FolderJournal, cfg config.Config, configPath string, passphrase string, date time.Time, appendHeading bool) error {
	if cfg.General.Editor == "" {
		return fmt.Errorf("no editor configured. Set editor in %s", configPath)
	}

	path := fj.DayFilePath(date)

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
		return editDayFileEncrypted(path, date, ecfg, appendHeading)
	}

	return editDayFilePlain(path, date, ecfg, appendHeading)
}

func editDayFilePlain(path string, date time.Time, ecfg editor.Config, appendHeading bool) error {
	// Read backup
	backup, _ := os.ReadFile(path)

	// Prepare file
	var startLine int
	if appendHeading {
		var err error
		startLine, err = editor.PrepareDayFile(path, date, ecfg)
		if err != nil {
			return fmt.Errorf("preparing day file: %w", err)
		}
	} else {
		startLine = 1
	}

	for {
		if err := editor.Launch(ecfg.Command, path, startLine); err != nil {
			return fmt.Errorf("launching editor: %w", err)
		}

		edited, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading edited file: %w", err)
		}

		content := string(edited)

		// Empty check
		if editor.IsEmptyContent(content) {
			fmt.Fprintln(os.Stderr, "No entries found after editing. Were you trying to delete all entries? Aborting — no changes made.")
			if backup != nil {
				atomicfile.WriteFile(path, backup, 0644)
			}
			return nil
		}

		// Validate
		_, parseErr := journal.ParseDayContent(content, ecfg.DateFmt, ecfg.TimeFmt)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Error in edited content:\n  %s\n", parseErr)
			if !prompt.YesNo(os.Stdin, os.Stderr, "Re-open editor?") {
				fmt.Fprintf(os.Stderr, "Warning: %s may contain invalid entries\n", path)
				return nil
			}
			startLine = 1
			continue
		}

		// Cleanup
		cleaned := journal.CleanupDayContent(content, ecfg.TimeFmt)
		if cleaned != content {
			if err := atomicfile.WriteFile(path, []byte(cleaned), 0644); err != nil {
				return fmt.Errorf("writing cleaned file: %w", err)
			}
		}

		return nil
	}
}

func editDayFileEncrypted(encPath string, date time.Time, ecfg editor.Config, appendHeading bool) error {
	// Read and decrypt existing content
	var existing string
	data, err := os.ReadFile(encPath)
	if err == nil {
		plain, err := crypto.Decrypt(data, ecfg.Passphrase)
		if err != nil {
			return fmt.Errorf("decrypting %s: %w", encPath, err)
		}
		existing = string(plain)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading %s: %w", encPath, err)
	}

	content := existing
	var startLine int
	if appendHeading {
		content, startLine = editor.PrepareEncryptedContent(content, date, ecfg)
	} else {
		startLine = 1
	}

	for {
		editedBytes, err := editor.WriteTempAndEdit(ecfg.Command, content, startLine)
		if err != nil {
			return err
		}

		content = string(editedBytes)

		// Empty check
		if editor.IsEmptyContent(content) {
			fmt.Fprintln(os.Stderr, "No entries found after editing. Were you trying to delete all entries? Aborting — no changes made.")
			return nil
		}

		// Validate
		_, parseErr := journal.ParseDayContent(content, ecfg.DateFmt, ecfg.TimeFmt)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Error in edited content:\n  %s\n", parseErr)
			if !prompt.YesNo(os.Stdin, os.Stderr, "Re-open editor?") {
				// Discard — original encrypted file untouched
				return nil
			}
			startLine = 1
			continue
		}

		// Cleanup
		content = journal.CleanupDayContent(content, ecfg.TimeFmt)

		// Re-encrypt and write
		dir := filepath.Dir(encPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}

		enc, err := crypto.Encrypt([]byte(content), ecfg.Passphrase)
		if err != nil {
			return fmt.Errorf("encrypting: %w", err)
		}

		if err := atomicfile.WriteFile(encPath, enc, 0600); err != nil {
			return fmt.Errorf("writing %s: %w", encPath, err)
		}

		return nil
	}
}
```

Note: This code references `journal.ParseDayContent` — a thin exported wrapper around `parseDay` that callers outside the `journal` package can use. Add it now.

- [ ] **Step 5: Add ParseDayContent to journal package**

Add to `internal/journal/day.go`:

```go
// ParseDayContent validates day file content by parsing it. Returns the
// parsed day and any parse error. Exported for use by the editor validation loop.
func ParseDayContent(text, dateFmt, timeFmt string) (day, error) {
	return parseDay(text, dateFmt, timeFmt)
}
```

- [ ] **Step 6: Export prepareEncryptedContent**

In `internal/editor/editor.go`, rename `prepareEncryptedContent` to `PrepareEncryptedContent` (capitalize the P). Update the call in `LaunchEncrypted` at line 156 accordingly.

- [ ] **Step 7: Update imports in edit.go**

Add these imports to `cmd/jrnl-md/edit.go`:

```go
import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glw907/jrnl-md/internal/atomicfile"
	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/crypto"
	"github.com/glw907/jrnl-md/internal/editor"
	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/glw907/jrnl-md/internal/prompt"
)
```

- [ ] **Step 8: Verify compilation**

Run: `cd /home/glw907/Projects/jrnl-md && go build ./cmd/jrnl-md/`
Expected: Compiles. Fix any unused import issues.

- [ ] **Step 9: Run e2e tests**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/ -run "TestEdit" -v -count=1`
Expected: All PASS including the new tests.

- [ ] **Step 10: Commit**

```bash
git add cmd/jrnl-md/edit.go internal/editor/editor.go internal/journal/day.go
git commit -m "feat: add post-edit validation loop with backup, cleanup, and empty abort"
```

---

### Task 5: Filtered edit with validation and single-day redirect

Update `editFiltered` to: (a) detect single-day full-match and redirect to `editDayFile`, (b) add empty check and validation loop for the temp-file path.

**Files:**
- Modify: `cmd/jrnl-md/edit.go`
- Test: `e2e/edit_test.go`

- [ ] **Step 1: Add e2e test for single-day redirect**

Add to `e2e/edit_test.go`:

```go
func TestEditFilteredSingleDayRedirect(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	// Edit --on 2026-03-01 matches all 2 entries on that day → should redirect to direct edit
	editorPath := writeMockEditor(t, env.dir, "First @work entry", "Redirected edit")
	patchConfigEditor(t, env.configPath, editorPath)

	_, stderr := run(t, env, "--edit", "--on", "2026-03-01")

	if strings.Contains(stderr, "error") || strings.Contains(stderr, "Error") {
		t.Fatalf("unexpected error: %q", stderr)
	}

	march1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	content := dayFileContent(t, env.journalDir, march1)
	if !strings.Contains(content, "Redirected edit") {
		t.Errorf("expected redirected edit content, got:\n%s", content)
	}
}
```

- [ ] **Step 2: Add e2e test for filtered edit empty buffer abort**

Add to `e2e/edit_test.go`:

```go
func TestEditFilteredEmptyAborts(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	// Mock editor that empties the temp file
	script := "#!/bin/bash\nFILE=\"${@: -1}\"\necho -n '' > \"$FILE\"\n"
	editorPath := filepath.Join(env.dir, "empty-editor.sh")
	if err := os.WriteFile(editorPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	patchConfigEditor(t, env.configPath, editorPath)

	// Filter that won't redirect to direct (partial day match — only @work, not starred)
	_, stderr, _ := runErr(t, env, "--edit", "@work")

	if !strings.Contains(stderr, "no changes made") && !strings.Contains(stderr, "No entries found") {
		t.Errorf("expected abort message, got: %q", stderr)
	}

	// Verify @work entry still exists
	march1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	content := dayFileContent(t, env.journalDir, march1)
	if !strings.Contains(content, "First @work entry") {
		t.Errorf("@work entry should be preserved after abort, got:\n%s", content)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/ -run "TestEdit(FilteredSingleDay|FilteredEmpty)" -v -count=1`
Expected: FAIL

- [ ] **Step 4: Update editFiltered with single-day redirect and validation**

Replace `editFiltered` in `cmd/jrnl-md/edit.go`:

```go
func editFiltered(fj *journal.FolderJournal, cfg config.Config, configPath string, passphrase string, entries []journal.Entry) error {
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No entries to edit.")
		return nil
	}

	if cfg.General.Editor == "" {
		return fmt.Errorf("no editor configured. Set editor in %s", configPath)
	}

	// Single-day redirect: if all entries are from one day and that day has no
	// other entries, open the day file directly instead of the temp-file round-trip.
	if isSingleDayFullMatch(fj, entries) {
		date := entries[0].Date
		return editDayFile(fj, cfg, configPath, passphrase, date, false)
	}

	// Multi-day or partial-day: temp file round-trip
	content := journal.FormatEntries(entries, cfg.Format.Date, cfg.Format.Time)

	for {
		edited, err := editor.WriteTempAndEdit(cfg.General.Editor, content, 1)
		if err != nil {
			return err
		}

		editedStr := string(edited)

		// Empty check
		if editor.IsEmptyContent(editedStr) {
			fmt.Fprintln(os.Stderr, "No entries found after editing. Were you trying to delete all entries? Aborting — no changes made.")
			return nil
		}

		// Validate
		newEntries, parseErr := journal.ParseMultiDay(editedStr, cfg.Format.Date, cfg.Format.Time)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Error in edited content:\n  %s\n", parseErr)
			if !prompt.YesNo(os.Stdin, os.Stderr, "Re-open editor?") {
				fmt.Fprintln(os.Stderr, "Edits discarded. Journal unchanged.")
				return nil
			}
			content = editedStr
			continue
		}

		if err := fj.DeleteEntries(entries); err != nil {
			return fmt.Errorf("removing old entries: %w", err)
		}

		if err := fj.AddEntries(newEntries); err != nil {
			return fmt.Errorf("adding edited entries: %w", err)
		}

		n := len(newEntries)
		switch {
		case n == 0:
			fmt.Fprintf(os.Stderr, "%d entries deleted.\n", len(entries))
		case n == 1:
			fmt.Fprintf(os.Stderr, "1 entry edited.\n")
		default:
			fmt.Fprintf(os.Stderr, "%d entries edited.\n", n)
		}

		return nil
	}
}

// isSingleDayFullMatch returns true if all entries are from the same calendar
// day and that day file has no other entries (full match).
func isSingleDayFullMatch(fj *journal.FolderJournal, entries []journal.Entry) bool {
	if len(entries) == 0 {
		return false
	}
	firstDay := entries[0].Date
	for _, e := range entries[1:] {
		if e.Date.Year() != firstDay.Year() || e.Date.Month() != firstDay.Month() || e.Date.Day() != firstDay.Day() {
			return false
		}
	}
	dayEntries, err := fj.DayEntries(firstDay)
	if err != nil {
		return false
	}
	return len(dayEntries) == len(entries)
}
```

- [ ] **Step 5: Verify compilation**

Run: `cd /home/glw907/Projects/jrnl-md && go build ./cmd/jrnl-md/`
Expected: Compiles.

- [ ] **Step 6: Run all e2e edit tests**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/ -run TestEdit -v -count=1`
Expected: All PASS

- [ ] **Step 7: Run full compat suite**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat -v -count=1`
Expected: All PASS

- [ ] **Step 8: Commit**

```bash
git add cmd/jrnl-md/edit.go
git commit -m "feat: add single-day redirect and validation loop to editFiltered"
```

---

### Task 6: Update CLAUDE.md and compat tests

**Files:**
- Modify: `CLAUDE.md`
- Modify: `BACKLOG.md`
- Modify: `e2e/jrnl_compat_test.go`

- [ ] **Step 1: Update CLAUDE.md pass table**

Add Pass 7 row and mark it Done:

```
| Pass 7: Edit Validation | Done | Post-edit validation, empty abort, single-day direct edit |
```

- [ ] **Step 2: Close backlog items**

In `BACKLOG.md`, move #5, #6, #7 from High to Done with a note referencing Pass 7.

- [ ] **Step 3: Run full test suite**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./... -v -count=1`
Expected: All PASS

- [ ] **Step 4: Run go vet**

Run: `cd /home/glw907/Projects/jrnl-md && go vet ./...`
Expected: No issues

- [ ] **Step 5: Commit**

```bash
git add CLAUDE.md BACKLOG.md e2e/jrnl_compat_test.go
git commit -m "Pass 7: edit validation complete"
```
