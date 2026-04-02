# Feature Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the 11 remaining jrnl features to achieve 100% feature parity (beyond documented by-design exceptions).

**Architecture:** Each feature is a self-contained task: add the flag/config, update the filter or export path, write compat tests. Features are grouped so that related changes land together (e.g., all date-component filters share the same Filter fields). Two new documented exceptions are added for `timeformat` and `--import --format TYPE`.

**Tech Stack:** Go 1.22+, cobra CLI, TOML config, `log/slog` for debug logging

---

### Task 1: Add Skipped Compat Test Stubs for All 11 Features + 2 Exceptions

**Files:**
- Modify: `e2e/jrnl_compat_test.go` (append after line 991)

- [ ] **Step 1: Add all 13 skipped test stubs**

Append to `e2e/jrnl_compat_test.go`:

```go
// --- Feature Parity: Pending ---

// TestCompat_Tagged: jrnl --tagged shows only entries that have at least one tag.
func TestCompat_Tagged(t *testing.T) {
	t.Skip("pending: --tagged filter not yet implemented")
}

// TestCompat_YearFilter: jrnl -year YYYY shows entries from a specific year.
func TestCompat_YearFilter(t *testing.T) {
	t.Skip("pending: -year filter not yet implemented")
}

// TestCompat_MonthFilter: jrnl -month M shows entries from a specific month across years.
func TestCompat_MonthFilter(t *testing.T) {
	t.Skip("pending: -month filter not yet implemented")
}

// TestCompat_DayFilter: jrnl -day D shows entries on a specific day-of-month.
func TestCompat_DayFilter(t *testing.T) {
	t.Skip("pending: -day filter not yet implemented")
}

// TestCompat_TodayInHistory: jrnl -today-in-history shows entries from today's date in prior years.
func TestCompat_TodayInHistory(t *testing.T) {
	t.Skip("pending: -today-in-history filter not yet implemented")
}

// TestCompat_Diagnostic: jrnl --diagnostic prints version/runtime info and exits.
func TestCompat_Diagnostic(t *testing.T) {
	t.Skip("pending: --diagnostic flag not yet implemented")
}

// TestCompat_Debug: jrnl --debug enables verbose logging.
func TestCompat_Debug(t *testing.T) {
	t.Skip("pending: --debug flag not yet implemented")
}

// TestCompat_DisplayFormat: display_format config sets default output format.
func TestCompat_DisplayFormat(t *testing.T) {
	t.Skip("pending: display_format config key not yet implemented")
}

// TestCompat_ConfigOverride: jrnl --config-override key value overrides config at runtime.
func TestCompat_ConfigOverride(t *testing.T) {
	t.Skip("pending: --config-override flag not yet implemented")
}

// TestCompat_FormatFancy: jrnl --format fancy outputs box-drawing cards.
func TestCompat_FormatFancy(t *testing.T) {
	t.Skip("pending: --format fancy/boxed not yet implemented")
}

// TestCompat_YAMLDirectoryExport: jrnl --format yaml --file dir/ writes one file per entry.
func TestCompat_YAMLDirectoryExport(t *testing.T) {
	t.Skip("pending: YAML directory export not yet implemented")
}

// TestCompat_TimeformatException: timeformat config is a documented exception (use format.date + format.time).
func TestCompat_TimeformatException(t *testing.T) {
	t.Skip("documented exception: timeformat replaced by format.date + format.time")
}

// TestCompat_ImportFormatException: --import --format TYPE is a documented exception (single format only).
func TestCompat_ImportFormatException(t *testing.T) {
	t.Skip("documented exception: --import --format TYPE not needed, single import format only")
}
```

- [ ] **Step 2: Run tests to verify all stubs skip**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run 'TestCompat_(Tagged|YearFilter|MonthFilter|DayFilter|TodayInHistory|Diagnostic|Debug|DisplayFormat|ConfigOverride|FormatFancy|YAMLDirectoryExport|TimeformatException|ImportFormatException)' -v`

Expected: All 13 tests SKIP with their pending messages.

- [ ] **Step 3: Commit**

```bash
git add e2e/jrnl_compat_test.go
git commit -m "test: add skipped compat stubs for 11 pending features + 2 exceptions"
```

---

### Task 2: `--tagged` Filter

**Files:**
- Modify: `internal/journal/filter.go`
- Modify: `cmd/jrnl-md/root.go`
- Modify: `cmd/jrnl-md/args.go`
- Modify: `e2e/jrnl_compat_test.go`

- [ ] **Step 1: Write the compat test**

Replace the `TestCompat_Tagged` stub in `e2e/jrnl_compat_test.go`:

```go
// TestCompat_Tagged: jrnl --tagged shows only entries that have at least one tag.
func TestCompat_Tagged(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	// 3 entries: 2 have tags (@work, @personal/@life), 1 has no tags (starred)
	stdout, stderr := runAll(t, env, "--tagged")

	assertEntriesFound(t, stderr, 2)
	if !strings.Contains(stdout, "First @work entry") {
		t.Errorf("expected @work entry in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "Mid-month @personal") {
		t.Errorf("expected @personal entry in output, got: %q", stdout)
	}
	if strings.Contains(stdout, "Starred afternoon entry") {
		t.Errorf("expected untagged entry NOT in output, got: %q", stdout)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_Tagged -v`

Expected: FAIL (unknown flag `--tagged`)

- [ ] **Step 3: Add `Tagged` to Filter and matching logic**

In `internal/journal/filter.go`, add `Tagged bool` field to `Filter` struct after `NotTagged`:

```go
type Filter struct {
	Tags       []string
	AndTags    bool
	NotTags    []string
	NotStarred bool
	NotTagged  bool
	Tagged     bool
	StartDate  *time.Time
	EndDate    *time.Time
	Starred    bool
	Contains   string
	N          int
}
```

In `isEmpty()`, add `!f.Tagged` to the chain:

```go
func (f Filter) isEmpty() bool {
	return len(f.Tags) == 0 &&
		len(f.NotTags) == 0 &&
		!f.NotStarred &&
		!f.NotTagged &&
		!f.Tagged &&
		f.StartDate == nil &&
		f.EndDate == nil &&
		!f.Starred &&
		f.Contains == ""
}
```

In `matches()`, add the Tagged check after the NotTagged check:

```go
	if f.NotTagged && len(e.Tags) > 0 {
		return false
	}
	if f.Tagged && len(e.Tags) == 0 {
		return false
	}
```

- [ ] **Step 4: Add `--tagged` CLI flag**

In `cmd/jrnl-md/root.go`, add to the `flags` struct after `notTagged`:

```go
	tagged     bool
```

In `newRootCmd()`, register the flag after the `--not-tagged` registration:

```go
	cmd.Flags().BoolVar(&f.tagged, "tagged", false, "Show only entries that have tags")
```

In `hasFilterFlags()`, add `f.tagged` to the return expression:

```go
	return f.n > 0 || f.short || f.starred || f.delete || f.encrypt || f.decrypt ||
		f.changeTime != "" || f.from != "" || f.to != "" || f.on != "" ||
		f.contains != "" || f.tags || f.export != "" ||
		f.notStarred || f.notTagged || f.tagged || len(f.not) > 0 || f.importFile != ""
```

In `cmd/jrnl-md/args.go`, in `buildFilter()`, add after `flt.NotTagged = f.notTagged`:

```go
	flt.Tagged = f.tagged
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_Tagged -v`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/journal/filter.go cmd/jrnl-md/root.go cmd/jrnl-md/args.go e2e/jrnl_compat_test.go
git commit -m "feat: add --tagged filter flag"
```

---

### Task 3: `-year`, `-month`, `-day` Date Component Filters

**Files:**
- Modify: `internal/journal/filter.go`
- Modify: `cmd/jrnl-md/root.go`
- Modify: `cmd/jrnl-md/args.go`
- Modify: `e2e/jrnl_compat_test.go`

- [ ] **Step 1: Write the compat tests**

Replace the `TestCompat_YearFilter`, `TestCompat_MonthFilter`, `TestCompat_DayFilter` stubs in `e2e/jrnl_compat_test.go`:

```go
// TestCompat_YearFilter: jrnl -year YYYY shows entries from a specific year.
func TestCompat_YearFilter(t *testing.T) {
	env := newTestEnv(t)
	// Seed entries in two different years
	day2025 := time.Date(2025, 6, 15, 0, 0, 0, 0, time.Local)
	day2026 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	writeDayFile(t, env.journalDir, day2025,
		"# 2025-06-15 Sunday\n\n## [09:00 AM]\n\nEntry from 2025.\n")
	writeDayFile(t, env.journalDir, day2026,
		"# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nEntry from 2026.\n")

	stdout, stderr := runAll(t, env, "-year", "2025")

	assertEntriesFound(t, stderr, 1)
	if !strings.Contains(stdout, "Entry from 2025") {
		t.Errorf("expected 2025 entry in output, got: %q", stdout)
	}
	if strings.Contains(stdout, "Entry from 2026") {
		t.Errorf("expected 2026 entry NOT in output, got: %q", stdout)
	}
}

// TestCompat_MonthFilter: jrnl -month M shows entries from a specific month across years.
func TestCompat_MonthFilter(t *testing.T) {
	env := newTestEnv(t)
	// March entries in 2025 and 2026, plus a June entry
	day1 := time.Date(2025, 3, 10, 0, 0, 0, 0, time.Local)
	day2 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	day3 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	writeDayFile(t, env.journalDir, day1,
		"# 2025-03-10 Monday\n\n## [09:00 AM]\n\nMarch 2025 entry.\n")
	writeDayFile(t, env.journalDir, day2,
		"# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nMarch 2026 entry.\n")
	writeDayFile(t, env.journalDir, day3,
		"# 2026-06-01 Monday\n\n## [09:00 AM]\n\nJune 2026 entry.\n")

	stdout, stderr := runAll(t, env, "-month", "3")

	assertEntriesFound(t, stderr, 2)
	if !strings.Contains(stdout, "March 2025 entry") {
		t.Errorf("expected March 2025 entry in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "March 2026 entry") {
		t.Errorf("expected March 2026 entry in output, got: %q", stdout)
	}
	if strings.Contains(stdout, "June 2026 entry") {
		t.Errorf("expected June entry NOT in output, got: %q", stdout)
	}
}

// TestCompat_DayFilter: jrnl -day D shows entries on a specific day-of-month.
func TestCompat_DayFilter(t *testing.T) {
	env := newTestEnv(t)
	// Day 1 entries in two months, plus a day-15 entry
	day1a := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	day1b := time.Date(2026, 4, 1, 0, 0, 0, 0, time.Local)
	day15 := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)
	writeDayFile(t, env.journalDir, day1a,
		"# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nMarch 1st entry.\n")
	writeDayFile(t, env.journalDir, day1b,
		"# 2026-04-01 Wednesday\n\n## [09:00 AM]\n\nApril 1st entry.\n")
	writeDayFile(t, env.journalDir, day15,
		"# 2026-03-15 Sunday\n\n## [09:00 AM]\n\nMarch 15th entry.\n")

	stdout, stderr := runAll(t, env, "-day", "1")

	assertEntriesFound(t, stderr, 2)
	if !strings.Contains(stdout, "March 1st entry") {
		t.Errorf("expected March 1st entry in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "April 1st entry") {
		t.Errorf("expected April 1st entry in output, got: %q", stdout)
	}
	if strings.Contains(stdout, "March 15th entry") {
		t.Errorf("expected March 15th entry NOT in output, got: %q", stdout)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run 'TestCompat_(YearFilter|MonthFilter|DayFilter)' -v`

Expected: FAIL (unknown flags)

- [ ] **Step 3: Add date component fields to Filter**

In `internal/journal/filter.go`, add three fields to the `Filter` struct after `Contains`:

```go
type Filter struct {
	Tags       []string
	AndTags    bool
	NotTags    []string
	NotStarred bool
	NotTagged  bool
	Tagged     bool
	StartDate  *time.Time
	EndDate    *time.Time
	Starred    bool
	Contains   string
	Year       int
	Month      int
	Day        int
	N          int
}
```

Update `isEmpty()` to include the new fields:

```go
func (f Filter) isEmpty() bool {
	return len(f.Tags) == 0 &&
		len(f.NotTags) == 0 &&
		!f.NotStarred &&
		!f.NotTagged &&
		!f.Tagged &&
		f.StartDate == nil &&
		f.EndDate == nil &&
		!f.Starred &&
		f.Contains == "" &&
		f.Year == 0 &&
		f.Month == 0 &&
		f.Day == 0
}
```

Add matching logic in `matches()` after the Contains check:

```go
	if f.Year != 0 && e.Date.Year() != f.Year {
		return false
	}
	if f.Month != 0 && int(e.Date.Month()) != f.Month {
		return false
	}
	if f.Day != 0 && e.Date.Day() != f.Day {
		return false
	}
```

- [ ] **Step 4: Add CLI flags and buildFilter mapping**

In `cmd/jrnl-md/root.go`, add to the `flags` struct:

```go
	year  string
	month string
	day   string
```

In `newRootCmd()`, register the flags (note: single-dash flags like `-year` are registered the same way in cobra):

```go
	cmd.Flags().StringVar(&f.year, "year", "", "Show entries from a specific year")
	cmd.Flags().StringVar(&f.month, "month", "", "Show entries from a specific month across years")
	cmd.Flags().StringVar(&f.day, "day", "", "Show entries on a specific day-of-month")
```

In `hasFilterFlags()`, add:

```go
		f.notStarred || f.notTagged || f.tagged || len(f.not) > 0 || f.importFile != "" ||
		f.year != "" || f.month != "" || f.day != ""
```

In `cmd/jrnl-md/args.go`, add a helper function and use it in `buildFilter()`:

```go
// parseDateComponent parses a date string and extracts a component (year, month, or day).
// Accepts plain integers or date strings parseable by dateparse.
func parseDateComponent(input string, component string) (int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, nil
	}

	// Try plain integer first
	var n int
	if _, err := fmt.Sscanf(input, "%d", &n); err == nil {
		switch component {
		case "year":
			if n >= 1000 && n <= 9999 {
				return n, nil
			}
		case "month":
			if n >= 1 && n <= 12 {
				return n, nil
			}
		case "day":
			if n >= 1 && n <= 31 {
				return n, nil
			}
		}
	}

	// Fall back to date parsing
	t, err := dateparse.Parse(input, 0, 0)
	if err != nil {
		return 0, fmt.Errorf("parsing -%s value %q: %w", component, input, err)
	}
	switch component {
	case "year":
		return t.Year(), nil
	case "month":
		return int(t.Month()), nil
	case "day":
		return t.Day(), nil
	}
	return 0, fmt.Errorf("unknown component %q", component)
}
```

In `buildFilter()`, add after the `--to` block:

```go
	if f.year != "" {
		y, err := parseDateComponent(f.year, "year")
		if err != nil {
			return flt, err
		}
		flt.Year = y
	}
	if f.month != "" {
		m, err := parseDateComponent(f.month, "month")
		if err != nil {
			return flt, err
		}
		flt.Month = m
	}
	if f.day != "" {
		d, err := parseDateComponent(f.day, "day")
		if err != nil {
			return flt, err
		}
		flt.Day = d
	}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run 'TestCompat_(YearFilter|MonthFilter|DayFilter)' -v`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/journal/filter.go cmd/jrnl-md/root.go cmd/jrnl-md/args.go e2e/jrnl_compat_test.go
git commit -m "feat: add -year, -month, -day date component filters"
```

---

### Task 4: `-today-in-history` Filter

**Files:**
- Modify: `cmd/jrnl-md/root.go`
- Modify: `cmd/jrnl-md/args.go`
- Modify: `e2e/jrnl_compat_test.go`

- [ ] **Step 1: Write the compat test**

Replace the `TestCompat_TodayInHistory` stub:

```go
// TestCompat_TodayInHistory: jrnl -today-in-history shows entries from today's date in prior years.
func TestCompat_TodayInHistory(t *testing.T) {
	env := newTestEnv(t)
	now := time.Now()
	// Create entries on today's month/day in two different years, plus a different date
	thisDay2024 := time.Date(2024, now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	thisDay2025 := time.Date(2025, now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	otherDay := time.Date(2025, now.Month(), 28, 0, 0, 0, 0, time.Local)
	if now.Day() == 28 {
		otherDay = time.Date(2025, now.Month(), 27, 0, 0, 0, 0, time.Local)
	}
	writeDayFile(t, env.journalDir, thisDay2024,
		fmt.Sprintf("# %s\n\n## [09:00 AM]\n\nToday-in-history 2024.\n", thisDay2024.Format("2006-01-02 Monday")))
	writeDayFile(t, env.journalDir, thisDay2025,
		fmt.Sprintf("# %s\n\n## [09:00 AM]\n\nToday-in-history 2025.\n", thisDay2025.Format("2006-01-02 Monday")))
	writeDayFile(t, env.journalDir, otherDay,
		fmt.Sprintf("# %s\n\n## [09:00 AM]\n\nDifferent day entry.\n", otherDay.Format("2006-01-02 Monday")))

	stdout, stderr := runAll(t, env, "-today-in-history")

	assertEntriesFound(t, stderr, 2)
	if !strings.Contains(stdout, "Today-in-history 2024") {
		t.Errorf("expected 2024 entry in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "Today-in-history 2025") {
		t.Errorf("expected 2025 entry in output, got: %q", stdout)
	}
	if strings.Contains(stdout, "Different day entry") {
		t.Errorf("expected different-day entry NOT in output, got: %q", stdout)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_TodayInHistory -v`

Expected: FAIL

- [ ] **Step 3: Add the flag and wire it up**

In `cmd/jrnl-md/root.go`, add to the `flags` struct:

```go
	todayInHistory bool
```

In `newRootCmd()`:

```go
	cmd.Flags().BoolVar(&f.todayInHistory, "today-in-history", false, "Show entries from today's date in prior years")
```

In `hasFilterFlags()`, add `f.todayInHistory` to the return expression.

In `cmd/jrnl-md/args.go`, in `buildFilter()`, add before the return:

```go
	if f.todayInHistory {
		now := time.Now()
		flt.Month = int(now.Month())
		flt.Day = now.Day()
	}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_TodayInHistory -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/jrnl-md/root.go cmd/jrnl-md/args.go e2e/jrnl_compat_test.go
git commit -m "feat: add -today-in-history filter flag"
```

---

### Task 5: `--diagnostic` Flag

**Files:**
- Modify: `cmd/jrnl-md/root.go`
- Modify: `e2e/jrnl_compat_test.go`

- [ ] **Step 1: Write the compat test**

Replace the `TestCompat_Diagnostic` stub:

```go
// TestCompat_Diagnostic: jrnl --diagnostic prints version/runtime info and exits.
func TestCompat_Diagnostic(t *testing.T) {
	// --diagnostic runs without needing a config file
	cmd := exec.Command(binary, "--diagnostic")
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		t.Fatalf("--diagnostic failed: %v\nstderr: %s", err, errBuf.String())
	}
	stdout := outBuf.String()
	if !strings.Contains(stdout, "jrnl-md:") {
		t.Errorf("expected 'jrnl-md:' in diagnostic output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "Go:") {
		t.Errorf("expected 'Go:' in diagnostic output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "OS:") {
		t.Errorf("expected 'OS:' in diagnostic output, got: %q", stdout)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_Diagnostic -v`

Expected: FAIL

- [ ] **Step 3: Implement --diagnostic**

In `cmd/jrnl-md/root.go`, add to the `flags` struct:

```go
	diagnostic bool
```

In `newRootCmd()`, register as hidden:

```go
	cmd.Flags().BoolVar(&f.diagnostic, "diagnostic", false, "Print diagnostic info")
	cmd.Flag("diagnostic").Hidden = true
```

In `runRoot()`, add immediately after the version check (after `if f.version { ... }`):

```go
	if f.diagnostic {
		fmt.Printf("jrnl-md: %s\n", version)
		fmt.Printf("Go: %s\n", runtime.Version())
		fmt.Printf("OS: %s %s\n", runtime.GOOS, runtime.GOARCH)
		return nil
	}
```

Add `"runtime"` to the imports in `root.go`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_Diagnostic -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/jrnl-md/root.go e2e/jrnl_compat_test.go
git commit -m "feat: add --diagnostic hidden flag"
```

---

### Task 6: `--debug` Flag

**Files:**
- Modify: `cmd/jrnl-md/root.go`
- Modify: `e2e/jrnl_compat_test.go`

- [ ] **Step 1: Write the compat test**

Replace the `TestCompat_Debug` stub:

```go
// TestCompat_Debug: jrnl --debug enables verbose logging.
func TestCompat_Debug(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	stdout, stderr := runAll(t, env, "--debug")

	// Debug mode should still show entries normally
	if !strings.Contains(stdout, "First @work entry") {
		t.Errorf("expected entries in output with --debug, got: %q", stdout)
	}
	// Debug mode should produce debug-level log output on stderr
	if !strings.Contains(stderr, "DEBUG") {
		t.Errorf("expected DEBUG messages in stderr, got: %q", stderr)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_Debug -v`

Expected: FAIL

- [ ] **Step 3: Implement --debug**

In `cmd/jrnl-md/root.go`, add to the `flags` struct:

```go
	debug bool
```

In `newRootCmd()`:

```go
	cmd.Flags().BoolVar(&f.debug, "debug", false, "Print debug information")
```

In `runRoot()`, add after the diagnostic check and before config loading:

```go
	if f.debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}
```

Add `"log/slog"` to the imports.

Then add debug log calls at key points in `runRoot()`:

After config loading succeeds:
```go
	slog.Debug("config loaded", "path", configPath)
```

After journal resolution (after `cfg = config.ResolvedJournalConfig(cfg, journalCfg)`):
```go
	slog.Debug("journal resolved", "name", journalName, "path", path, "encrypted", encrypted)
```

After filter building in the read path (in `readEntries` in `read.go`), add after `entries, err := fj.Entries(&flt)`:
```go
	slog.Debug("entries loaded", "count", len(entries), "filter", fmt.Sprintf("%+v", flt))
```

Add `"log/slog"` to imports in `read.go` as well.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_Debug -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/jrnl-md/root.go cmd/jrnl-md/read.go e2e/jrnl_compat_test.go
git commit -m "feat: add --debug flag for verbose logging"
```

---

### Task 7: `display_format` Config Key

**Files:**
- Modify: `internal/config/config.go`
- Modify: `cmd/jrnl-md/read.go`
- Modify: `e2e/jrnl_compat_test.go`

- [ ] **Step 1: Write the compat test**

Replace the `TestCompat_DisplayFormat` stub:

```go
// TestCompat_DisplayFormat: display_format config sets default output format.
func TestCompat_DisplayFormat(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	// Patch config to set display_format = "short"
	data, err := os.ReadFile(env.configPath)
	if err != nil {
		t.Fatal(err)
	}
	patched := strings.Replace(string(data),
		"indent_character = \"\"",
		"indent_character = \"\"\ndisplay_format = \"short\"",
		1)
	if err := os.WriteFile(env.configPath, []byte(patched), 0644); err != nil {
		t.Fatal(err)
	}

	// Run without --format flag — should use display_format from config
	stdout, stderr := runAll(t, env)

	assertEntriesFound(t, stderr, 3)
	// Short format: one line per entry, no full body
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	nonEmpty := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty++
		}
	}
	if nonEmpty < 3 {
		t.Errorf("expected at least 3 non-empty lines for short format, got %d", nonEmpty)
	}
	// Short listing should NOT contain markdown headings
	if strings.Contains(stdout, "## [") {
		t.Errorf("expected short format (no markdown headings), got: %q", stdout)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_DisplayFormat -v`

Expected: FAIL (display_format ignored, shows full pretty output)

- [ ] **Step 3: Add DisplayFormat to config**

In `internal/config/config.go`, add to `GeneralConfig`:

```go
	DisplayFormat string `toml:"display_format"`
```

No change to `Default()` needed — zero value `""` means "pretty" (current behavior).

- [ ] **Step 4: Use DisplayFormat in readEntries**

In `cmd/jrnl-md/read.go`, in `readEntries()`, add before `if f.export != ""`:

```go
	if f.export == "" && cfg.General.DisplayFormat != "" {
		f.export = cfg.General.DisplayFormat
	}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_DisplayFormat -v`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/config/config.go cmd/jrnl-md/read.go e2e/jrnl_compat_test.go
git commit -m "feat: add display_format config key for default output format"
```

---

### Task 8: `--config-override key value`

**Files:**
- Modify: `cmd/jrnl-md/root.go`
- Modify: `cmd/jrnl-md/args.go`
- Modify: `internal/config/config.go`
- Modify: `e2e/jrnl_compat_test.go`

- [ ] **Step 1: Write the compat test**

Replace the `TestCompat_ConfigOverride` stub:

```go
// TestCompat_ConfigOverride: jrnl --config-override key value overrides config at runtime.
func TestCompat_ConfigOverride(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	// Override linewrap to 40 from the command line
	stdout, _ := runAll(t, env, "--config-override", "linewrap", "40")

	// The long body text from seed should be wrapped
	// seedCompatJournal bodies are short, so test with a longer entry
	today := time.Now()
	longBody := "This is a deliberately long entry body that should definitely be wrapped when linewrap is set to forty."
	writeDayFile(t, env.journalDir, today,
		fmt.Sprintf("# %s\n\n## [09:00 AM]\n\n%s\n", today.Format("2006-01-02 Monday"), longBody))

	stdout, _ = runAll(t, env, "--config-override", "linewrap", "40")

	if strings.Contains(stdout, longBody) {
		t.Errorf("expected long body to be wrapped with --config-override linewrap 40, got single line")
	}
	if !strings.Contains(stdout, "deliberately long") {
		t.Errorf("expected body content present in output, got: %q", stdout)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_ConfigOverride -v`

Expected: FAIL

- [ ] **Step 3: Add config override infrastructure**

In `cmd/jrnl-md/root.go`, add to the `flags` struct:

```go
	configOverride []string
```

In `newRootCmd()`:

```go
	cmd.Flags().StringArrayVar(&f.configOverride, "config-override", nil, "Override config key value")
```

In `internal/config/config.go`, add the `ApplyOverrides` function:

```go
// ApplyOverrides applies command-line config overrides to a Config.
// Each override is a pair of strings: [key, value].
// Keys use dot notation for nested fields (e.g., "colors.tags").
func ApplyOverrides(cfg *Config, overrides []string) error {
	if len(overrides)%2 != 0 {
		return fmt.Errorf("config overrides must be key-value pairs")
	}
	for i := 0; i < len(overrides); i += 2 {
		key := overrides[i]
		value := overrides[i+1]
		if err := applyOverride(cfg, key, value); err != nil {
			return fmt.Errorf("applying override %q=%q: %w", key, value, err)
		}
	}
	return nil
}

func applyOverride(cfg *Config, key, value string) error {
	switch strings.ToLower(key) {
	// General
	case "editor":
		cfg.General.Editor = value
	case "encrypt":
		b, err := parseBool(value)
		if err != nil {
			return err
		}
		cfg.General.Encrypt = b
	case "default_hour":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer: %w", err)
		}
		cfg.General.DefaultHour = n
	case "default_minute":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer: %w", err)
		}
		cfg.General.DefaultMinute = n
	case "highlight":
		b, err := parseBool(value)
		if err != nil {
			return err
		}
		cfg.General.Highlight = b
	case "linewrap":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer: %w", err)
		}
		cfg.General.Linewrap = n
	case "indent_character":
		cfg.General.IndentCharacter = value
	case "template":
		cfg.General.Template = value
	case "display_format":
		cfg.General.DisplayFormat = value
	// Format
	case "format.time", "time":
		cfg.Format.Time = value
	case "format.date", "date":
		cfg.Format.Date = value
	case "format.tag_symbols", "tag_symbols":
		cfg.Format.TagSymbols = value
	case "format.file_extension", "file_extension":
		cfg.Format.FileExtension = value
	// Colors
	case "colors.date":
		cfg.Colors.Date = value
	case "colors.body":
		cfg.Colors.Body = value
	case "colors.tags":
		cfg.Colors.Tags = value
	default:
		return fmt.Errorf("unknown config key %q", key)
	}
	return nil
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean %q", s)
	}
}
```

Add `"strconv"` and `"strings"` to imports in `config.go` (strings is already imported).

- [ ] **Step 4: Wire up overrides in runRoot**

In `cmd/jrnl-md/root.go`, in `runRoot()`, add after config loading (after `cfg = config.Default()` fallback and the error-handling block), before the `if f.list` check:

```go
	if len(f.configOverride) > 0 {
		if err := config.ApplyOverrides(&cfg, f.configOverride); err != nil {
			return fmt.Errorf("applying config overrides: %w", err)
		}
	}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_ConfigOverride -v`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add cmd/jrnl-md/root.go cmd/jrnl-md/args.go internal/config/config.go e2e/jrnl_compat_test.go
git commit -m "feat: add --config-override for runtime config overrides"
```

---

### Task 9: `--format fancy` / `--format boxed` Export

**Files:**
- Create: `internal/export/fancy.go`
- Modify: `internal/export/format.go`
- Modify: `cmd/jrnl-md/read.go`
- Modify: `e2e/jrnl_compat_test.go`

- [ ] **Step 1: Write the compat test**

Replace the `TestCompat_FormatFancy` stub:

```go
// TestCompat_FormatFancy: jrnl --format fancy outputs box-drawing cards.
func TestCompat_FormatFancy(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	for _, format := range []string{"fancy", "boxed"} {
		t.Run(format, func(t *testing.T) {
			stdout, stderr := runAll(t, env, "--format", format)

			assertEntriesFound(t, stderr, 3)
			// Should contain box-drawing characters
			if !strings.Contains(stdout, "┎") {
				t.Errorf("expected top-left box character in output, got: %q", stdout)
			}
			if !strings.Contains(stdout, "┖") {
				t.Errorf("expected bottom-left box character in output, got: %q", stdout)
			}
			// Should contain entry content
			if !strings.Contains(stdout, "First @work entry") {
				t.Errorf("expected entry body in fancy output, got: %q", stdout)
			}
			// Should contain formatted date
			if !strings.Contains(stdout, "2026-03-01") {
				t.Errorf("expected date in fancy output, got: %q", stdout)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_FormatFancy -v`

Expected: FAIL (unknown format)

- [ ] **Step 3: Add format constants**

In `internal/export/format.go`:

```go
const (
	FormatJSON     = "json"
	FormatMarkdown = "md"
	FormatText     = "txt"
	FormatXML      = "xml"
	FormatYAML     = "yaml"
	FormatFancy    = "fancy"
	FormatBoxed    = "boxed"
)
```

- [ ] **Step 4: Implement the fancy exporter**

Create `internal/export/fancy.go`:

```go
package export

import (
	"fmt"
	"strings"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

// Fancy formats entries as Unicode box-drawing cards.
// "boxed" is an alias for the same format.
func Fancy(entries []journal.Entry, cfg config.Config) (string, error) {
	width := cfg.General.Linewrap
	if width <= 0 {
		width = 79
	}

	var b strings.Builder
	for i, e := range entries {
		if i > 0 {
			b.WriteString("\n")
		}
		dateStr := fmt.Sprintf("%s %s",
			e.Date.Format(cfg.Format.Date),
			e.Date.Format(cfg.Format.Time))
		if e.Starred {
			dateStr += " *"
		}
		writeCard(&b, trimBody(e.Body), dateStr, width)
	}
	return b.String(), nil
}

func writeCard(b *strings.Builder, body, dateStr string, width int) {
	innerWidth := width - 2 // left border + right border

	// Top border: ┎─── date ╮
	dateLen := len(dateStr)
	dashCount := innerWidth - dateLen - 1 // -1 for the space before date
	if dashCount < 1 {
		dashCount = 1
	}
	b.WriteString("┎")
	b.WriteString(strings.Repeat("─", dashCount))
	b.WriteString(" ")
	b.WriteString(dateStr)
	b.WriteString("╮\n")

	// Body lines
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		// Wrap long lines
		wrapped := wrapLine(line, innerWidth-2) // -2 for padding spaces
		for _, wl := range wrapped {
			padded := wl + strings.Repeat(" ", max(0, innerWidth-2-len(wl)))
			fmt.Fprintf(b, "┃ %s│\n", padded)
		}
	}

	// Bottom border: ┖───┘
	b.WriteString("┖")
	b.WriteString(strings.Repeat("─", innerWidth))
	b.WriteString("┘\n")
}

func wrapLine(line string, width int) []string {
	if width <= 0 || len(line) <= width {
		return []string{line}
	}
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
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
	return lines
}

```

- [ ] **Step 5: Wire up in readEntries**

In `cmd/jrnl-md/read.go`, add cases in the export switch (inside the inner `switch format` block):

```go
			case export.FormatFancy, export.FormatBoxed:
				output, err = export.Fancy(entries, cfg)
```

Update the error message to include `fancy`/`boxed`:

```go
				return fmt.Errorf("unknown export format %q (supported: pretty, short, tags, dates, %s, %s, %s, %s, %s, %s)",
					f.export, export.FormatJSON, export.FormatMarkdown,
					export.FormatText, export.FormatXML, export.FormatYAML, export.FormatFancy)
```

- [ ] **Step 6: Run test to verify it passes**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_FormatFancy -v`

Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/export/fancy.go internal/export/format.go cmd/jrnl-md/read.go e2e/jrnl_compat_test.go
git commit -m "feat: add --format fancy/boxed unicode box-drawing export"
```

---

### Task 10: YAML Directory Export

**Files:**
- Create: `internal/export/yaml_dir.go`
- Modify: `cmd/jrnl-md/read.go`
- Modify: `e2e/jrnl_compat_test.go`

- [ ] **Step 1: Write the compat test**

Replace the `TestCompat_YAMLDirectoryExport` stub:

```go
// TestCompat_YAMLDirectoryExport: jrnl --format yaml --file dir/ writes one file per entry.
func TestCompat_YAMLDirectoryExport(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	outDir := filepath.Join(env.dir, "yaml-export")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, stderr := runAll(t, env, "--format", "yaml", "--file", outDir+"/")

	assertEntriesFound(t, stderr, 3)

	// Check that files were created
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("failed to read export dir: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 YAML files, got %d", len(entries))
	}

	// Check content of one file — should have YAML front matter
	for _, entry := range entries {
		content, err := os.ReadFile(filepath.Join(outDir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		s := string(content)
		if !strings.HasPrefix(s, "---\n") {
			t.Errorf("expected YAML front matter in %s, got: %q", entry.Name(), s[:min(len(s), 50)])
		}
		if !strings.Contains(s, "date:") {
			t.Errorf("expected 'date:' in YAML front matter of %s", entry.Name())
		}
		if !strings.Contains(s, "tags:") {
			t.Errorf("expected 'tags:' in YAML front matter of %s", entry.Name())
		}
		break // check just the first one
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_YAMLDirectoryExport -v`

Expected: FAIL (YAML goes to single file, not directory)

- [ ] **Step 3: Create the YAML directory exporter**

Create `internal/export/yaml_dir.go`:

```go
package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

// YAMLDir writes one YAML-frontmatter markdown file per entry into the given directory.
func YAMLDir(entries []journal.Entry, cfg config.Config, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating export directory: %w", err)
	}

	for _, e := range entries {
		filename := yamlEntryFilename(e, cfg)
		path := filepath.Join(dir, filename)

		content := yamlEntryContent(e, cfg)
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}
	return nil
}

func yamlEntryFilename(e journal.Entry, cfg config.Config) string {
	timestamp := e.Date.Format("2006-01-02_150405")
	body := strings.TrimSpace(e.Body)

	// Use first line as slug, truncated
	slug := body
	if idx := strings.IndexByte(slug, '\n'); idx >= 0 {
		slug = slug[:idx]
	}
	slug = strings.TrimSpace(slug)
	if len(slug) > 40 {
		slug = slug[:40]
	}
	// Sanitize for filesystem
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, slug)
	slug = strings.Trim(slug, "_")
	if slug == "" {
		slug = "entry"
	}

	return timestamp + "_" + slug + ".md"
}

func yamlEntryContent(e journal.Entry, cfg config.Config) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "date: %q\n", e.Date.Format(cfg.Format.Date+" "+cfg.Format.Time))
	fmt.Fprintf(&b, "starred: %t\n", e.Starred)
	if len(e.Tags) == 0 {
		b.WriteString("tags: []\n")
	} else {
		b.WriteString("tags:\n")
		for _, tag := range e.Tags {
			fmt.Fprintf(&b, "  - %q\n", tag)
		}
	}
	b.WriteString("---\n\n")

	body := strings.TrimRight(e.Body, "\n ")
	if body != "" {
		b.WriteString(body)
		b.WriteString("\n")
	}

	return b.String()
}
```

- [ ] **Step 4: Wire up YAML directory export in readEntries**

In `cmd/jrnl-md/read.go`, modify the YAML case to detect directory paths. Replace the existing YAML case in the inner switch:

```go
			case export.FormatYAML:
				if f.file != "" && isDir(f.file) {
					if err := export.YAMLDir(entries, cfg, f.file); err != nil {
						return fmt.Errorf("exporting YAML directory: %w", err)
					}
					return nil
				}
				output, err = export.YAML(entries, cfg)
```

Add the `isDir` helper at the bottom of `read.go`:

```go
func isDir(path string) bool {
	if strings.HasSuffix(path, "/") || strings.HasSuffix(path, string(os.PathSeparator)) {
		return true
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat_YAMLDirectoryExport -v`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/export/yaml_dir.go cmd/jrnl-md/read.go e2e/jrnl_compat_test.go
git commit -m "feat: add YAML directory export (--format yaml --file dir/)"
```

---

### Task 11: Update Compat Doc and Document New Exceptions

**Files:**
- Modify: `docs/jrnl-compat.md`
- Modify: `e2e/jrnl_compat_test.go` (exception tests)

- [ ] **Step 1: Update the exception tests to be real assertions**

Replace the two exception stubs in `e2e/jrnl_compat_test.go`:

```go
// TestCompat_TimeformatException: timeformat config is a documented exception.
// jrnl-md uses format.date + format.time (Go layouts) instead of timeformat (Python strftime).
func TestCompat_TimeformatException(t *testing.T) {
	env := newTestEnv(t)
	// Verify that format.time and format.date work as the replacement
	data, err := os.ReadFile(env.configPath)
	if err != nil {
		t.Fatal(err)
	}
	patched := strings.Replace(string(data),
		`time = "03:04 PM"`,
		`time = "15:04"`,
		1)
	if err := os.WriteFile(env.configPath, []byte(patched), 0644); err != nil {
		t.Fatal(err)
	}

	today := time.Now()
	run(t, env, "Timeformat exception test entry.")
	stdout, _ := runAll(t, env)

	// Time should display in 24h format per the overridden format.time
	if !strings.Contains(stdout, today.Format("15:04")) {
		t.Errorf("expected 24h time format in output, got: %q", stdout)
	}
}

// TestCompat_ImportFormatException: --import --format TYPE is not needed.
// jrnl-md only has one import format (its own markdown), so the flag is unnecessary.
func TestCompat_ImportFormatException(t *testing.T) {
	env := newTestEnv(t)

	importContent := "# 2026-05-01 Friday\n\n## [09:00 AM]\n\nImport format exception test.\n"
	importPath := filepath.Join(env.dir, "import.md")
	if err := os.WriteFile(importPath, []byte(importContent), 0644); err != nil {
		t.Fatal(err)
	}

	// --import works without any --format flag
	_, stderr := run(t, env, "--import", importPath)
	if !strings.Contains(stderr, "Imported 1 entries") {
		t.Errorf("expected 'Imported 1 entries', got: %q", stderr)
	}
}
```

- [ ] **Step 2: Update docs/jrnl-compat.md**

Move the 11 implemented features from "Not Implemented" to the appropriate sections in the feature compatibility table, and update the two new exceptions. The "Not Implemented" section should only contain by-design exceptions after this change.

Read the current compat doc, then update the "Not Implemented" section to remove the 11 now-implemented features and add the two new exceptions. The remaining "Not Implemented" items should be:

```markdown
### Not Implemented

| Feature | Notes |
|---|---|
| DayOne backend | Folder-only; no plans to add |
| Single-file journal | Folder-only; no plans to add |
| `--export dayone` | By design — requires DayOne backend |
| GPG encryption | Uses age instead |
| `timeformat` config key | Uses `format.date` + `format.time` (Go time layouts) instead of Python strftime |
| `--import --format TYPE` | Single import format only; flag unnecessary |
```

And add the new features to the appropriate tables above (Filtering, Display, Config, etc.).

- [ ] **Step 3: Run the full compat suite**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./e2e/... -run TestCompat -v`

Expected: All tests PASS. Zero SKIP. Zero FAIL.

- [ ] **Step 4: Run all tests**

Run: `cd /home/glw907/Projects/jrnl-md && go test ./... -v`

Expected: All PASS.

- [ ] **Step 5: Commit**

```bash
git add docs/jrnl-compat.md e2e/jrnl_compat_test.go
git commit -m "docs: update compat doc for 100% feature parity"
```
