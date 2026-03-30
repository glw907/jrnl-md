# Pass 1: CLI & Display Polish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `--format`/`--file` export flags, frequency-sorted `--tags`, tag highlighting, and `--and`/`--not`/`--not-starred`/`--not-tagged` filter flags to match jrnl behavior.

**Architecture:** Filter logic is extended in `internal/journal/filter.go`; display highlighting is added to `internal/display/display.go`; CLI flags are wired in `cmd/jrnl-md/root.go` and plumbed through `cmd/jrnl-md/args.go` (buildFilter) and `cmd/jrnl-md/read.go`.

**Tech Stack:** Go stdlib (`regexp`, `sort`), `github.com/fatih/color`, `github.com/glw907/jrnl-md/internal/atomicfile`

---

## e2e Test Helpers (reference — do not modify)

```go
// Create a test environment with config and empty journal dir
env := newTestEnv(t)

// Run the binary with args; returns (stdout, stderr string)
stdout, stderr := run(t, env, "arg1", "arg2", ...)

// Run the binary expecting a possible error; returns (stdout, stderr string, err error)
stdout, stderr, err := runErr(t, env, "arg1", "arg2", ...)

// Write a day file directly (used to seed specific-date entries)
writeDayFile(t, env.journalDir, time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local),
    "# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nEntry body.\n")

// Seed helpers already defined in filter_test.go:
//   seedFilterJournal(t, env) — writes 3 entries with @work, @project, @personal tags
//   seedJournal(t, env)       — writes 3 entries (see compat_test.go)
```

The e2e test config sets `highlight = false` and all colors to `"none"`, so ANSI codes will not appear in e2e output. Tag highlighting is a visual feature tested by unit tests only.

---

## File Map

| File | Change |
|------|--------|
| `internal/journal/filter.go` | Add AndTags, NotTags, NotStarred, NotTagged fields + matches logic |
| `internal/journal/filter_test.go` | New unit tests for AndTags, NotTags, NotStarred, NotTagged |
| `internal/display/display.go` | Fix ColorFunc to return nil for "none"; add HighlightTags |
| `internal/display/display_test.go` | New tests for HighlightTags; test ColorFunc nil return |
| `cmd/jrnl-md/root.go` | Add format, file, and, not, not-starred, not-tagged flags; coalesce format→export |
| `cmd/jrnl-md/args.go` | buildFilter: set AndTags, NotTags, NotStarred, NotTagged |
| `cmd/jrnl-md/read.go` | showTags: frequency sort; export: --file target; body: apply HighlightTags; date/bodyColor nil guard |
| `e2e/filter_test.go` | Append tests for --and, --not, --not-starred, --not-tagged |
| `e2e/tags_test.go` | Append test for frequency sort order |
| `e2e/export_test.go` | Append tests for --file and --format alias |

---

### Task 1: Filter — AndTags, NotTags, NotStarred, NotTagged

**Files:**
- Modify: `internal/journal/filter.go`
- Modify: `internal/journal/filter_test.go`

- [ ] **Step 1: Write failing unit tests**

Append to `internal/journal/filter_test.go` after the existing `TestFilterEmpty` test:

```go
func TestFilterAndTags(t *testing.T) {
	entries := makeTestEntries()
	// "Friday afternoon @personal @mood" has both tags; others don't
	f := Filter{Tags: []string{"@personal", "@mood"}, AndTags: true}
	result := f.Apply(entries)

	if len(result) != 1 {
		t.Fatalf("expected 1 entry with both tags, got %d", len(result))
	}
	if result[0].Body != "Friday afternoon @personal @mood" {
		t.Errorf("wrong entry: %q", result[0].Body)
	}
}

func TestFilterAndTagsNoMatch(t *testing.T) {
	entries := makeTestEntries()
	// No entry has both @work and @personal
	f := Filter{Tags: []string{"@work", "@personal"}, AndTags: true}
	result := f.Apply(entries)

	if len(result) != 0 {
		t.Fatalf("expected 0 entries for AND with no match, got %d", len(result))
	}
}

func TestFilterNotTags(t *testing.T) {
	entries := makeTestEntries()
	// Exclude entries containing @work
	f := Filter{NotTags: []string{"@work"}}
	result := f.Apply(entries)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries (without @work), got %d", len(result))
	}
	for _, e := range result {
		for _, tag := range e.Tags {
			if tag == "@work" {
				t.Errorf("entry with @work should be excluded: %q", e.Body)
			}
		}
	}
}

func TestFilterNotTagsMultiple(t *testing.T) {
	entries := makeTestEntries()
	// Exclude both @work and @personal — only "Saturday morning" remains
	f := Filter{NotTags: []string{"@work", "@personal"}}
	result := f.Apply(entries)

	if len(result) != 1 {
		t.Fatalf("expected 1 entry after excluding @work and @personal, got %d", len(result))
	}
	if result[0].Body != "Saturday morning" {
		t.Errorf("wrong remaining entry: %q", result[0].Body)
	}
}

func TestFilterNotStarred(t *testing.T) {
	entries := makeTestEntries()
	// Only one entry is starred; exclude it
	f := Filter{NotStarred: true}
	result := f.Apply(entries)

	if len(result) != 2 {
		t.Fatalf("expected 2 non-starred entries, got %d", len(result))
	}
	for _, e := range result {
		if e.Starred {
			t.Errorf("starred entry should be excluded: %q", e.Body)
		}
	}
}

func TestFilterNotTagged(t *testing.T) {
	entries := makeTestEntries()
	// Only "Saturday morning" has no tags
	f := Filter{NotTagged: true}
	result := f.Apply(entries)

	if len(result) != 1 {
		t.Fatalf("expected 1 untagged entry, got %d", len(result))
	}
	if result[0].Body != "Saturday morning" {
		t.Errorf("wrong entry: %q", result[0].Body)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/glw907/Projects/jrnl-md
go test ./internal/journal/ -run 'TestFilterAndTags|TestFilterNotTags|TestFilterNotStarred|TestFilterNotTagged' -v
```

Expected: FAIL — `Filter` struct has no `AndTags`, `NotTags`, `NotStarred`, `NotTagged` fields.

- [ ] **Step 3: Rewrite filter.go**

Replace `internal/journal/filter.go` entirely:

```go
package journal

import (
	"strings"
	"time"
)

// Filter specifies criteria for selecting entries.
type Filter struct {
	Tags       []string
	AndTags    bool // if true, entry must match ALL Tags (default: any)
	NotTags    []string
	NotStarred bool
	NotTagged  bool
	StartDate  *time.Time
	EndDate    *time.Time
	Starred    bool
	Contains   string
	N          int
}

// Apply returns entries matching all filter criteria. If N is set,
// only the last N matching entries are returned.
func (f Filter) Apply(entries []Entry) []Entry {
	if f.isEmpty() {
		if f.N > 0 && f.N < len(entries) {
			return entries[len(entries)-f.N:]
		}
		return entries
	}

	tagSet := make(map[string]bool, len(f.Tags))
	for _, t := range f.Tags {
		tagSet[strings.ToLower(t)] = true
	}

	notTagSet := make(map[string]bool, len(f.NotTags))
	for _, t := range f.NotTags {
		notTagSet[strings.ToLower(t)] = true
	}

	containsLower := strings.ToLower(f.Contains)

	var result []Entry
	for _, e := range entries {
		if f.matches(e, tagSet, notTagSet, containsLower) {
			result = append(result, e)
		}
	}

	if f.N > 0 && f.N < len(result) {
		result = result[len(result)-f.N:]
	}

	return result
}

func (f Filter) isEmpty() bool {
	return len(f.Tags) == 0 &&
		len(f.NotTags) == 0 &&
		!f.NotStarred &&
		!f.NotTagged &&
		f.StartDate == nil &&
		f.EndDate == nil &&
		!f.Starred &&
		f.Contains == ""
}

func (f Filter) matches(e Entry, tagSet, notTagSet map[string]bool, containsLower string) bool {
	if f.Starred && !e.Starred {
		return false
	}
	if f.NotStarred && e.Starred {
		return false
	}
	if f.NotTagged && len(e.Tags) > 0 {
		return false
	}
	if f.StartDate != nil && e.Date.Before(*f.StartDate) {
		return false
	}
	if f.EndDate != nil && e.Date.After(*f.EndDate) {
		return false
	}
	if len(tagSet) > 0 {
		if f.AndTags {
			for tag := range tagSet {
				found := false
				for _, et := range e.Tags {
					if strings.ToLower(et) == tag {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		} else {
			found := false
			for _, t := range e.Tags {
				if tagSet[t] {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	for tag := range notTagSet {
		for _, et := range e.Tags {
			if strings.ToLower(et) == tag {
				return false
			}
		}
	}
	if containsLower != "" {
		if !strings.Contains(strings.ToLower(e.Body), containsLower) {
			return false
		}
	}
	return true
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/journal/ -v
```

Expected: all journal tests pass including the six new filter tests.

- [ ] **Step 5: Commit**

```bash
git add internal/journal/filter.go internal/journal/filter_test.go
git commit -m "filter: add AndTags, NotTags, NotStarred, NotTagged"
```

---

### Task 2: ColorFunc nil return + HighlightTags

**Files:**
- Modify: `internal/display/display.go`
- Modify: `internal/display/display_test.go`

**Context:** `ColorFunc` currently returns `fmt.Sprint` for `"none"` and unknown names. `HighlightTags` needs `nil` as the sentinel meaning "no highlight". We change `ColorFunc` to return `nil` for `"none"` and all unrecognized names, and update the two callers in `read.go` that use `ColorFunc` for date/body to guard against nil.

- [ ] **Step 1: Write failing unit tests**

The existing `internal/display/display_test.go` imports only `"strings"` and `"testing"`. Replace the import block with:

```go
import (
	"fmt"
	"strings"
	"testing"
)
```

Append to `internal/display/display_test.go`:

```go
func TestColorFuncNoneReturnsNil(t *testing.T) {
	if ColorFunc("none") != nil {
		t.Error(`ColorFunc("none") should return nil`)
	}
}

func TestColorFuncUnknownReturnsNil(t *testing.T) {
	if ColorFunc("not-a-color") != nil {
		t.Error(`ColorFunc("not-a-color") should return nil`)
	}
}

func TestHighlightTagsNilColorFn(t *testing.T) {
	body := "Entry with @work tag"
	result := HighlightTags(body, "@", nil)
	if result != body {
		t.Errorf("nil colorFn should return body unchanged, got %q", result)
	}
}

func TestHighlightTagsNoTagsInBody(t *testing.T) {
	body := "Entry with no tags here"
	colorFn := func(a ...any) string { return "X" }
	result := HighlightTags(body, "@", colorFn)
	if result != body {
		t.Errorf("body with no tags should be returned unchanged, got %q", result)
	}
}

func TestHighlightTagsSingleSymbol(t *testing.T) {
	body := "Entry with @work and @home tags"
	colorFn := func(a ...any) string {
		s := fmt.Sprint(a...)
		return "[" + s + "]"
	}
	result := HighlightTags(body, "@", colorFn)
	want := "Entry with [@work] and [@home] tags"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestHighlightTagsMultipleSymbols(t *testing.T) {
	body := "Entry with @work and #project tags"
	colorFn := func(a ...any) string {
		s := fmt.Sprint(a...)
		return "[" + s + "]"
	}
	result := HighlightTags(body, "@#", colorFn)
	want := "Entry with [@work] and [#project] tags"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestHighlightTagsEmptySymbols(t *testing.T) {
	body := "Entry with @work tag"
	colorFn := func(a ...any) string { return "X" }
	result := HighlightTags(body, "", colorFn)
	if result != body {
		t.Errorf("empty tagSymbols should return body unchanged, got %q", result)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/display/ -run 'TestColorFunc|TestHighlight' -v
```

Expected: FAIL — `HighlightTags` is undefined; `ColorFunc("none")` currently returns non-nil.

- [ ] **Step 3: Rewrite display.go**

Replace `internal/display/display.go` entirely:

```go
package display

import (
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// WrapText wraps text to the given column width, preserving newlines
// between paragraphs.
func WrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		if len(paragraph) <= width {
			lines = append(lines, paragraph)
			continue
		}

		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		current := words[0]
		for _, word := range words[1:] {
			if len(current)+1+len(word) > width {
				lines = append(lines, current)
				current = word
			} else {
				current += " " + word
			}
		}
		lines = append(lines, current)
	}

	return strings.Join(lines, "\n")
}

// IndentBody prepends indent to each non-empty line of body.
func IndentBody(body, indent string) string {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// TerminalWidth returns the width of stdout, defaulting to 79.
func TerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 79
	}
	return w
}

// ColorFunc returns a function that wraps text in the named ANSI color.
// Returns nil for "none" or unrecognized names.
func ColorFunc(name string) func(a ...any) string {
	switch strings.ToLower(name) {
	case "black":
		return color.New(color.FgBlack).SprintFunc()
	case "red":
		return color.New(color.FgRed).SprintFunc()
	case "green":
		return color.New(color.FgGreen).SprintFunc()
	case "yellow":
		return color.New(color.FgYellow).SprintFunc()
	case "blue":
		return color.New(color.FgBlue).SprintFunc()
	case "magenta":
		return color.New(color.FgMagenta).SprintFunc()
	case "cyan":
		return color.New(color.FgCyan).SprintFunc()
	case "white":
		return color.New(color.FgWhite).SprintFunc()
	default:
		return nil
	}
}

// HighlightTags replaces tag occurrences in body with colorFn-wrapped
// versions. tagSymbols is the set of tag prefix characters (e.g. "@").
// If colorFn is nil or tagSymbols is empty, body is returned unchanged.
func HighlightTags(body, tagSymbols string, colorFn func(a ...any) string) string {
	if colorFn == nil || tagSymbols == "" {
		return body
	}
	escaped := regexp.QuoteMeta(tagSymbols)
	re := regexp.MustCompile(`[` + escaped + `]\w+`)
	return re.ReplaceAllStringFunc(body, func(match string) string {
		return colorFn(match)
	})
}
```

- [ ] **Step 4: Run display tests to verify they pass**

```bash
go test ./internal/display/ -v
```

Expected: all display tests pass.

- [ ] **Step 5: Fix the two nil-unsafe ColorFunc callers in read.go**

In `cmd/jrnl-md/read.go`, find the lines:

```go
	dateColor := display.ColorFunc(cfg.Colors.Date)
	bodyColor := display.ColorFunc(cfg.Colors.Body)
```

Replace with:

```go
	dateColorFn := display.ColorFunc(cfg.Colors.Date)
	dateColor := func(a ...any) string {
		if dateColorFn != nil {
			return dateColorFn(a...)
		}
		s := ""
		for _, v := range a {
			s += fmt.Sprint(v)
		}
		return s
	}
	bodyColorFn := display.ColorFunc(cfg.Colors.Body)
	bodyColor := func(a ...any) string {
		if bodyColorFn != nil {
			return bodyColorFn(a...)
		}
		s := ""
		for _, v := range a {
			s += fmt.Sprint(v)
		}
		return s
	}
```

Also add `"fmt"` to the imports of `read.go` if not already present (check the import block).

- [ ] **Step 6: Run full test suite**

```bash
go test ./... -v 2>&1 | grep -E '(FAIL|PASS|ok)'
```

Expected: all packages pass.

- [ ] **Step 7: Commit**

```bash
git add internal/display/display.go internal/display/display_test.go cmd/jrnl-md/read.go
git commit -m "display: add HighlightTags; ColorFunc returns nil for unrecognized colors"
```

---

### Task 3: --format alias, --file export target

**Files:**
- Modify: `cmd/jrnl-md/root.go`
- Modify: `cmd/jrnl-md/read.go`
- Modify: `e2e/export_test.go`

**Context:** `atomicfile.WriteFile` signature is `(path string, data []byte, perm os.FileMode)`.

- [ ] **Step 1: Write failing e2e tests**

Append to `e2e/export_test.go`:

```go
func TestExportToFile(t *testing.T) {
	env := newTestEnv(t)
	seedJournal(t, env)

	outFile := filepath.Join(env.dir, "out.json")
	run(t, env, "--export", "json", "--file", outFile, "--num", "99")

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if !strings.Contains(string(data), "Morning entry") {
		t.Errorf("exported file missing entry content, got: %s", data)
	}
}

func TestFormatAliasMatchesExport(t *testing.T) {
	env1 := newTestEnv(t)
	seedJournal(t, env1)
	env2 := newTestEnv(t)
	seedJournal(t, env2)

	out1, _ := run(t, env1, "--export", "json", "--num", "99")
	out2, _ := run(t, env2, "--format", "json", "--num", "99")

	if out1 != out2 {
		t.Errorf("--format and --export produced different output:\n--export: %s\n--format: %s", out1, out2)
	}
}
```

Also add the missing imports to `e2e/export_test.go` — check the current import block and add `"os"` and `"path/filepath"` if not present.

- [ ] **Step 2: Run to verify failure**

```bash
go test ./e2e/ -run 'TestExportToFile|TestFormatAliasMatchesExport' -v
```

Expected: FAIL — `--file` and `--format` flags don't exist.

- [ ] **Step 3: Add flags to root.go**

In `cmd/jrnl-md/root.go`, in the `flags` struct, add after `export string`:

```go
	format string
	file   string
```

In `newRootCmd`, after the existing `--export` flag registration line:

```go
	cmd.Flags().StringVar(&f.format, "format", "", "Export format (alias for --export)")
	cmd.Flags().StringVar(&f.file, "file", "", "Write export output to file instead of stdout")
```

In `runRoot`, after `if f.list { return listJournals(cfg) }` and before `journalName, text, tagArgs := parseArgs(...)`, add:

```go
	if f.format != "" && f.export == "" {
		f.export = f.format
	}
```

- [ ] **Step 4: Add --file write in read.go**

In `cmd/jrnl-md/read.go`, add `"github.com/glw907/jrnl-md/internal/atomicfile"` to the import block.

In the export section, find `fmt.Print(output)` and `return nil` (these are the last two lines of the export switch block). Replace:

```go
		fmt.Print(output)
		return nil
```

with:

```go
		if f.file != "" {
			if err := atomicfile.WriteFile(f.file, []byte(output), 0o600); err != nil {
				return fmt.Errorf("writing export to %s: %w", f.file, err)
			}
			return nil
		}
		fmt.Print(output)
		return nil
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./e2e/ -run 'TestExportToFile|TestFormatAliasMatchesExport' -v
go test ./...
```

Expected: both e2e tests pass; no regressions.

- [ ] **Step 6: Commit**

```bash
git add cmd/jrnl-md/root.go cmd/jrnl-md/read.go e2e/export_test.go
git commit -m "cli: add --format alias and --file export target"
```

---

### Task 4: --and, --not, --not-starred, --not-tagged flags

**Files:**
- Modify: `cmd/jrnl-md/root.go`
- Modify: `cmd/jrnl-md/args.go`
- Modify: `e2e/filter_test.go`

- [ ] **Step 1: Write failing e2e tests**

Append to `e2e/filter_test.go`:

```go
func TestAndFlag(t *testing.T) {
	env := newTestEnv(t)
	// Entry 1: has @work and @project; Entry 2: has only @work
	writeDayFile(t, env.journalDir, time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local),
		"# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nEntry with @work and @project tags.\n\n## [10:00 AM]\n\nEntry with only @work tag.\n")

	stdout, stderr := run(t, env, "--and", "@work", "@project", "--num", "99")

	if !strings.Contains(stderr, "1 entries found") {
		t.Errorf("--and should match only entry with both tags, stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "both @work and @project") {
		t.Errorf("--and should show entry with both tags, stdout: %q", stdout)
	}
	if strings.Contains(stdout, "only @work") {
		t.Errorf("--and should not show entry with only one tag, stdout: %q", stdout)
	}
}

func TestNotFlag(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	// seedFilterJournal has @work, @project, @personal entries
	// Excluding @work and @project should leave only @personal
	stdout, stderr := run(t, env, "--not", "@work", "--not", "@project", "--num", "99")

	if !strings.Contains(stderr, "1 entries found") {
		t.Errorf("expected 1 entry after --not exclusions, stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "personal") {
		t.Errorf("expected @personal entry in output, stdout: %q", stdout)
	}
	if strings.Contains(stdout, "@work") {
		t.Errorf("--not @work should exclude @work entries, stdout: %q", stdout)
	}
}

func TestNotStarredFlag(t *testing.T) {
	env := newTestEnv(t)
	writeDayFile(t, env.journalDir, time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local),
		"# 2026-03-01 Sunday\n\n## [09:00 AM] *\n\nStarred entry content.\n\n## [10:00 AM]\n\nUnstarred entry content.\n")

	stdout, stderr := run(t, env, "--not-starred", "--num", "99")

	if !strings.Contains(stderr, "1 entries found") {
		t.Errorf("--not-starred should find 1 entry, stderr: %q", stderr)
	}
	if strings.Contains(stdout, "Starred entry") {
		t.Errorf("--not-starred should exclude starred entries, stdout: %q", stdout)
	}
	if !strings.Contains(stdout, "Unstarred entry") {
		t.Errorf("--not-starred should include unstarred entries, stdout: %q", stdout)
	}
}

func TestNotTaggedFlag(t *testing.T) {
	env := newTestEnv(t)
	writeDayFile(t, env.journalDir, time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local),
		"# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nEntry with @tag.\n\n## [10:00 AM]\n\nEntry without tags.\n")

	stdout, stderr := run(t, env, "--not-tagged", "--num", "99")

	if !strings.Contains(stderr, "1 entries found") {
		t.Errorf("--not-tagged should find 1 entry, stderr: %q", stderr)
	}
	if strings.Contains(stdout, "@tag") {
		t.Errorf("--not-tagged should exclude tagged entries, stdout: %q", stdout)
	}
	if !strings.Contains(stdout, "without tags") {
		t.Errorf("--not-tagged should include untagged entries, stdout: %q", stdout)
	}
}
```

- [ ] **Step 2: Run to verify failure**

```bash
go test ./e2e/ -run 'TestAndFlag|TestNotFlag|TestNotStarredFlag|TestNotTaggedFlag' -v
```

Expected: FAIL — flags not registered.

- [ ] **Step 3: Add flags to root.go**

In `cmd/jrnl-md/root.go`, add to the `flags` struct (after `starred bool`):

```go
	and        bool
	not        []string
	notStarred bool
	notTagged  bool
```

In `newRootCmd`, after `cmd.Flags().BoolVar(&f.starred, ...)`:

```go
	cmd.Flags().BoolVar(&f.and, "and", false, "Require all specified tags (AND logic)")
	cmd.Flags().StringArrayVar(&f.not, "not", nil, "Exclude entries containing tag")
	cmd.Flags().BoolVar(&f.notStarred, "not-starred", false, "Exclude starred entries")
	cmd.Flags().BoolVar(&f.notTagged, "not-tagged", false, "Exclude entries that have any tags")
```

Replace `hasFilterFlags` with:

```go
func hasFilterFlags(f *flags) bool {
	return f.n > 0 || f.short || f.starred || f.delete || f.encrypt || f.decrypt ||
		f.changeTime != "" || f.from != "" || f.to != "" || f.on != "" ||
		f.contains != "" || f.tags || f.export != "" ||
		f.notStarred || f.notTagged || len(f.not) > 0
}
```

- [ ] **Step 4: Wire into buildFilter in args.go**

In `cmd/jrnl-md/args.go`, in `buildFilter`, after `flt.Starred = f.starred` add:

```go
	flt.AndTags = f.and
	flt.NotTags = f.not
	flt.NotStarred = f.notStarred
	flt.NotTagged = f.notTagged
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./e2e/ -run 'TestAndFlag|TestNotFlag|TestNotStarredFlag|TestNotTaggedFlag' -v
go test ./...
```

Expected: all four e2e tests pass; no regressions.

- [ ] **Step 6: Commit**

```bash
git add cmd/jrnl-md/root.go cmd/jrnl-md/args.go e2e/filter_test.go
git commit -m "cli: add --and, --not, --not-starred, --not-tagged filter flags"
```

---

### Task 5: --tags sorted by frequency descending

**Files:**
- Modify: `cmd/jrnl-md/read.go`
- Modify: `e2e/tags_test.go`

**Context:** `showTags` in `read.go` currently sorts alphabetically. `export.TagCounts` returns `map[string]int`.

- [ ] **Step 1: Write failing e2e test**

Append to `e2e/tags_test.go`:

```go
func TestTagsFrequencySort(t *testing.T) {
	env := newTestEnv(t)
	// @common appears 3 times, @medium 2 times, @rare 1 time
	writeDayFile(t, env.journalDir, time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local),
		"# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nEntry with @common and @rare.\n\n## [10:00 AM]\n\nEntry with @common and @medium.\n\n## [11:00 AM]\n\nEntry with @common only.\n\n## [12:00 PM]\n\nEntry with @medium only.\n")

	stdout, _ := run(t, env, "--tags", "--num", "99")

	posCommon := strings.Index(stdout, "@common")
	posMedium := strings.Index(stdout, "@medium")
	posRare := strings.Index(stdout, "@rare")

	if posCommon == -1 || posMedium == -1 || posRare == -1 {
		t.Fatalf("expected all three tags in output, got: %s", stdout)
	}
	if posCommon > posMedium {
		t.Errorf("@common (3 occurrences) should appear before @medium (2 occurrences)")
	}
	if posMedium > posRare {
		t.Errorf("@medium (2 occurrences) should appear before @rare (1 occurrence)")
	}
}
```

- [ ] **Step 2: Run to verify failure**

```bash
go test ./e2e/ -run 'TestTagsFrequencySort' -v
```

Expected: FAIL — tags appear in alphabetical order, not frequency order.

- [ ] **Step 3: Rewrite showTags in read.go**

Replace the `showTags` function in `cmd/jrnl-md/read.go`:

```go
func showTags(entries []journal.Entry) error {
	counts := export.TagCounts(entries)

	if len(counts) == 0 {
		fmt.Println("No tags found.")
		return nil
	}

	type tagCount struct {
		tag   string
		count int
	}
	tc := make([]tagCount, 0, len(counts))
	for tag, n := range counts {
		tc = append(tc, tagCount{tag, n})
	}
	sort.Slice(tc, func(i, j int) bool {
		if tc[i].count != tc[j].count {
			return tc[i].count > tc[j].count
		}
		return tc[i].tag < tc[j].tag
	})

	for _, item := range tc {
		fmt.Printf("%-20s : %d\n", item.tag, item.count)
	}
	return nil
}
```

The `sort` import is already present in `read.go`.

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./e2e/ -run 'TestTagsFrequencySort' -v
go test ./...
```

Expected: test passes; no regressions.

- [ ] **Step 5: Commit**

```bash
git add cmd/jrnl-md/read.go e2e/tags_test.go
git commit -m "display: sort --tags output by frequency descending"
```

---

### Task 6: Tag highlighting in body output

**Files:**
- Modify: `cmd/jrnl-md/read.go`

**Context:** `HighlightTags` and the updated `ColorFunc` (returning nil for "none") are in place from Task 2. Wire tag highlighting into the body rendering in `readEntries`. Logic: when `cfg.General.Highlight` is true and `cfg.Colors.Tags` is not "none" (ColorFunc returns non-nil), use that color. When `cfg.General.Highlight` is true but `cfg.Colors.Tags` is "none" (ColorFunc returns nil), default to cyan. When `cfg.General.Highlight` is false, no highlighting.

Unit testing of the coloring logic itself is covered by display_test.go (Task 2). No new tests needed here; the existing e2e suite (which uses `highlight = false` in test config) will catch regressions.

- [ ] **Step 1: Add tag highlighting to the body rendering block in read.go**

In `cmd/jrnl-md/read.go`, find the body rendering block inside `for _, e := range entries`:

```go
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
```

Replace with:

```go
		body := e.Body
		if linewrap > 0 && indent != "" {
			body = display.WrapText(body, linewrap-len(indent))
		} else if linewrap > 0 {
			body = display.WrapText(body, linewrap)
		}
		if indent != "" {
			body = display.IndentBody(body, indent)
		}
		if cfg.General.Highlight {
			tagColorFn := display.ColorFunc(cfg.Colors.Tags)
			if tagColorFn == nil {
				tagColorFn = display.ColorFunc("cyan")
			}
			body = display.HighlightTags(body, cfg.Format.TagSymbols, tagColorFn)
		}
		fmt.Println(bodyColor(body))
```

- [ ] **Step 2: Run full test suite**

```bash
go test ./... -v 2>&1 | grep -E '(FAIL|ok )'
```

Expected: all packages pass.

- [ ] **Step 3: Commit**

```bash
git add cmd/jrnl-md/read.go
git commit -m "display: apply tag highlighting in body output"
```

---

### Task 7: Final verification

- [ ] **Step 1: Run full test suite**

```bash
go test ./...
```

Expected: all tests pass.

- [ ] **Step 2: Run vet**

```bash
go vet ./...
```

Expected: no issues.

- [ ] **Step 3: Build binary and smoke test**

```bash
make build
echo "Entry with @sometag present." | ./jrnl-md
./jrnl-md --tags
./jrnl-md --not-tagged
./jrnl-md --format json
./jrnl-md --export json --file /tmp/test-export.json && cat /tmp/test-export.json
```

Expected: binary builds; entry written; --tags shows frequency-sorted output; --not-tagged shows entries without tags; --format json and --export json produce identical output; --file writes to the specified path.
