package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// seedCompatJournal creates a 3-entry journal across 2 days for compat testing.
//
//	day1 (2026-03-01): @work entry (09:00 AM), starred entry with no tags (02:00 PM) *
//	day2 (2026-03-15): @personal and @life entry (10:00 AM)
func seedCompatJournal(t *testing.T, env testEnv) {
	t.Helper()
	day1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	day2 := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)
	writeDayFile(t, env.journalDir, day1,
		"# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nFirst @work entry.\n\n## [02:00 PM] *\n\nStarred afternoon entry.\n")
	writeDayFile(t, env.journalDir, day2,
		"# 2026-03-15 Sunday\n\n## [10:00 AM]\n\nMid-month @personal and @life entry.\n")
}

// runAll runs the binary with --num 99 appended, returning all entries.
func runAll(t *testing.T, env testEnv, args ...string) (stdout, stderr string) {
	t.Helper()
	return run(t, env, append(args, "--num", "99")...)
}

// assertEntriesFound checks that stderr reports exactly n entries found.
func assertEntriesFound(t *testing.T, stderr string, n int) {
	t.Helper()
	want := fmt.Sprintf("%d entries found", n)
	if !strings.Contains(stderr, want) {
		t.Errorf("expected %q in stderr, got: %q", want, stderr)
	}
}

// --- Write ---

// TestCompat_WriteInlineEntry: jrnl "Entry text" creates an entry at current time.
func TestCompat_WriteInlineEntry(t *testing.T) {
	env := newTestEnv(t)
	today := time.Now()

	_, stderr := run(t, env, "Inline compat entry.")

	if !strings.Contains(stderr, "Entry added") {
		t.Errorf("expected 'Entry added' in stderr, got: %q", stderr)
	}
	if !dayFileExists(t, env.journalDir, today) {
		t.Fatal("expected day file for today")
	}
	content := dayFileContent(t, env.journalDir, today)
	if !strings.Contains(content, "Inline compat entry.") {
		t.Errorf("expected entry body in day file, got:\n%s", content)
	}
}

func TestCompat_WriteFromStdin(t *testing.T) {
	env := newTestEnv(t)
	today := time.Now()

	_, stderr := runWithStdin(t, env, "Compat stdin entry.\n")

	if !strings.Contains(stderr, "Entry added") {
		t.Errorf("expected 'Entry added' in stderr, got: %q", stderr)
	}
	if !dayFileExists(t, env.journalDir, today) {
		t.Fatal("expected day file for today after stdin write")
	}
	content := dayFileContent(t, env.journalDir, today)
	if !strings.Contains(content, "Compat stdin entry.") {
		t.Errorf("expected stdin body in day file, got:\n%s", content)
	}
}

func TestCompat_DatePrefixedEntry(t *testing.T) {
	env := newTestEnv(t)
	yesterday := time.Now().AddDate(0, 0, -1)

	_, stderr := run(t, env, "yesterday: Compat date-prefix entry.")

	if !strings.Contains(stderr, "Entry added") {
		t.Errorf("expected 'Entry added' in stderr, got: %q", stderr)
	}
	if !dayFileExists(t, env.journalDir, yesterday) {
		t.Fatal("expected day file for yesterday after date-prefixed entry")
	}
	content := dayFileContent(t, env.journalDir, yesterday)
	if !strings.Contains(content, "Compat date-prefix entry.") {
		t.Errorf("expected entry body in yesterday's day file, got:\n%s", content)
	}
}

// --- Reading / Listing ---

// TestCompat_LastNEntries: jrnl -n N / --num N shows last N entries.
func TestCompat_LastNEntries(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	stdout, stderr := run(t, env, "--num", "2")

	assertEntriesFound(t, stderr, 2)
	if strings.Contains(stdout, "First @work entry") {
		t.Errorf("expected oldest entry NOT in output (only last 2), got: %q", stdout)
	}
	if !strings.Contains(stdout, "Mid-month @personal") {
		t.Errorf("expected newest entry in output, got: %q", stdout)
	}
}

// TestCompat_ShortListing: jrnl --short / -s shows one-line summary per entry.
func TestCompat_ShortListing(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	stdout, stderr := runAll(t, env, "--short")

	assertEntriesFound(t, stderr, 3)
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	nonEmpty := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		nonEmpty++
		if !strings.Contains(line, "2026-03") {
			t.Errorf("expected --short line to contain date prefix, got: %q", line)
		}
	}
	if nonEmpty < 3 {
		t.Errorf("expected at least 3 non-empty lines for --short with 3 entries, got %d:\n%s", nonEmpty, stdout)
	}
}

// --- Filters ---

// TestCompat_StarredFilter: jrnl --starred shows only starred entries.
func TestCompat_StarredFilter(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	stdout, stderr := runAll(t, env, "--starred")

	assertEntriesFound(t, stderr, 1)
	if !strings.Contains(stdout, "Starred afternoon entry") {
		t.Errorf("expected starred entry in output, got: %q", stdout)
	}
	if strings.Contains(stdout, "First @work entry") {
		t.Errorf("expected non-starred entry NOT in output, got: %q", stdout)
	}
}

// TestCompat_TextSearch: jrnl --contains text shows only matching entries.
func TestCompat_TextSearch(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	stdout, stderr := runAll(t, env, "--contains", "Starred")

	assertEntriesFound(t, stderr, 1)
	if !strings.Contains(stdout, "Starred afternoon entry") {
		t.Errorf("expected matching entry in output, got: %q", stdout)
	}
	if strings.Contains(stdout, "First @work entry") {
		t.Errorf("expected non-matching entry NOT in output, got: %q", stdout)
	}
}

// TestCompat_DateRangeFrom: jrnl --from DATE shows entries on or after DATE.
func TestCompat_DateRangeFrom(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	stdout, stderr := runAll(t, env, "--from", "2026-03-15")

	assertEntriesFound(t, stderr, 1)
	if !strings.Contains(stdout, "Mid-month @personal") {
		t.Errorf("expected mid-month entry in output, got: %q", stdout)
	}
	if strings.Contains(stdout, "First @work entry") {
		t.Errorf("expected earlier entry NOT in output, got: %q", stdout)
	}
}

// TestCompat_DateRangeTo: jrnl --to DATE shows entries on or before DATE.
func TestCompat_DateRangeTo(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	stdout, stderr := runAll(t, env, "--to", "2026-03-01")

	assertEntriesFound(t, stderr, 2)
	if !strings.Contains(stdout, "First @work entry") {
		t.Errorf("expected March 1 entry in output, got: %q", stdout)
	}
	if strings.Contains(stdout, "Mid-month @personal") {
		t.Errorf("expected later entry NOT in output, got: %q", stdout)
	}
}

// TestCompat_DateRangeOn: jrnl --on DATE shows entries on that exact day.
func TestCompat_DateRangeOn(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	stdout, stderr := runAll(t, env, "--on", "2026-03-01")

	assertEntriesFound(t, stderr, 2)
	if !strings.Contains(stdout, "First @work entry") {
		t.Errorf("expected March 1 entry in output, got: %q", stdout)
	}
	if strings.Contains(stdout, "Mid-month @personal") {
		t.Errorf("expected March 15 entry NOT in output, got: %q", stdout)
	}
}

// TestCompat_TagFilter: jrnl @tag shows only entries containing that tag.
func TestCompat_TagFilter(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	stdout, stderr := runAll(t, env, "@work")

	assertEntriesFound(t, stderr, 1)
	if !strings.Contains(stdout, "First @work entry") {
		t.Errorf("expected @work entry in output, got: %q", stdout)
	}
	if strings.Contains(stdout, "Mid-month @personal") {
		t.Errorf("expected non-@work entry NOT in output, got: %q", stdout)
	}
}

// TestCompat_AndTagFilter: jrnl --and @tag1 @tag2 shows entries with ALL listed tags.
func TestCompat_AndTagFilter(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	// Only mid-month entry has both @personal and @life
	stdout, stderr := runAll(t, env, "--and", "@personal", "@life")

	assertEntriesFound(t, stderr, 1)
	if !strings.Contains(stdout, "Mid-month @personal") {
		t.Errorf("expected AND-matching entry in output, got: %q", stdout)
	}
	if strings.Contains(stdout, "First @work entry") {
		t.Errorf("expected non-matching entry NOT in output, got: %q", stdout)
	}
}

// TestCompat_ExcludeTag: jrnl --not @tag excludes entries containing that tag.
func TestCompat_ExcludeTag(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	// 3 entries total; 1 has @work → 2 remain
	stdout, stderr := runAll(t, env, "--not", "@work")

	assertEntriesFound(t, stderr, 2)
	if strings.Contains(stdout, "First @work entry") {
		t.Errorf("expected @work entry NOT in output, got: %q", stdout)
	}
}

// TestCompat_ExcludeStarred: jrnl --not-starred shows only unstarred entries.
func TestCompat_ExcludeStarred(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	// 3 entries; 1 starred → 2 unstarred
	stdout, stderr := runAll(t, env, "--not-starred")

	assertEntriesFound(t, stderr, 2)
	if strings.Contains(stdout, "Starred afternoon entry") {
		t.Errorf("expected starred entry NOT in output, got: %q", stdout)
	}
}

// TestCompat_ExcludeTagged: jrnl --not-tagged shows only untagged entries.
func TestCompat_ExcludeTagged(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	// Only "Starred afternoon entry" has no tags → 1 entry
	stdout, stderr := runAll(t, env, "--not-tagged")

	assertEntriesFound(t, stderr, 1)
	if !strings.Contains(stdout, "Starred afternoon entry") {
		t.Errorf("expected untagged entry in output, got: %q", stdout)
	}
	if strings.Contains(stdout, "@work") || strings.Contains(stdout, "@personal") {
		t.Errorf("expected tagged entries NOT in output, got: %q", stdout)
	}
}

// TestCompat_ListTags: jrnl --tags lists all tags in "tag: N" format.
func TestCompat_ListTags(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	stdout, _ := runAll(t, env, "--tags")

	for _, tag := range []string{"@work", "@personal", "@life"} {
		if !strings.Contains(stdout, tag) {
			t.Errorf("expected %q in --tags output, got: %q", tag, stdout)
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if line == "" {
			continue
		}
		if !strings.Contains(line, ": ") {
			t.Errorf("expected --tags line in 'tag: N' format, got: %q", line)
		}
	}
}

// TestCompat_ListTagsFrequencySorted: tags with higher frequency appear first.
// Note: TestTagsFrequencySort in tags_test.go covers the same sort logic with a single-file
// fixture. This test uses a multi-file fixture to confirm frequency sort works across days.
func TestCompat_ListTagsFrequencySorted(t *testing.T) {
	env := newTestEnv(t)
	// @zebra appears 3 times, @alpha appears once — @zebra must come first
	day1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	day2 := time.Date(2026, 3, 2, 0, 0, 0, 0, time.Local)
	day3 := time.Date(2026, 3, 3, 0, 0, 0, 0, time.Local)
	day4 := time.Date(2026, 3, 4, 0, 0, 0, 0, time.Local)
	writeDayFile(t, env.journalDir, day1, "# 2026-03-01 Sunday\n\n## [09:00 AM]\n\n@zebra entry one.\n")
	writeDayFile(t, env.journalDir, day2, "# 2026-03-02 Monday\n\n## [09:00 AM]\n\n@zebra entry two.\n")
	writeDayFile(t, env.journalDir, day3, "# 2026-03-03 Tuesday\n\n## [09:00 AM]\n\n@zebra entry three.\n")
	writeDayFile(t, env.journalDir, day4, "# 2026-03-04 Wednesday\n\n## [09:00 AM]\n\n@alpha entry once.\n")

	stdout, _ := runAll(t, env, "--tags")

	zebraIdx := strings.Index(stdout, "@zebra")
	alphaIdx := strings.Index(stdout, "@alpha")
	if zebraIdx < 0 || alphaIdx < 0 {
		t.Fatalf("expected both @zebra and @alpha in --tags output, got: %q", stdout)
	}
	if zebraIdx > alphaIdx {
		t.Errorf("expected @zebra (3x) to appear before @alpha (1x) in --tags output, got:\n%s", stdout)
	}
}

// --- Edit ---

// TestCompat_EditNoFilter: jrnl --edit (no filter) opens all entries via editFiltered.
func TestCompat_EditNoFilter(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	editorPath := writeMockEditor(t, env.dir, "Starred afternoon entry", "Edited starred entry")
	patchConfigEditor(t, env.configPath, editorPath)

	_, stderr := run(t, env, "--edit", "--num", "99")

	if strings.Contains(stderr, "error") || strings.Contains(stderr, "Error") {
		t.Fatalf("unexpected error: %q", stderr)
	}

	march1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	content := dayFileContent(t, env.journalDir, march1)
	if !strings.Contains(content, "Edited starred entry") {
		t.Errorf("expected edited body in day file, got:\n%s", content)
	}
	if strings.Contains(content, "Starred afternoon entry") {
		t.Errorf("expected original text to be replaced, got:\n%s", content)
	}
}

// TestCompat_EditWithFilter: jrnl --edit @tag edits only matching entries.
func TestCompat_EditWithFilter(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	editorPath := writeMockEditor(t, env.dir, "First @work entry", "Edited via filter")
	patchConfigEditor(t, env.configPath, editorPath)

	_, stderr := run(t, env, "@work", "--edit")

	if strings.Contains(stderr, "error") || strings.Contains(stderr, "Error") {
		t.Fatalf("unexpected error: %q", stderr)
	}

	march1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	march15 := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)

	content1 := dayFileContent(t, env.journalDir, march1)
	if !strings.Contains(content1, "Edited via filter") {
		t.Errorf("expected @work entry to be edited in march1, got:\n%s", content1)
	}

	content15 := dayFileContent(t, env.journalDir, march15)
	if !strings.Contains(content15, "Mid-month @personal") {
		t.Errorf("expected unfiltered entry to be unchanged in march15, got:\n%s", content15)
	}
}

// --- Delete ---

// TestCompat_DeleteEntries: jrnl --delete removes matched entries after confirmation.
func TestCompat_DeleteEntries(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	// Delete only the @work entry (March 1, 09:00)
	_, stderr := runWithStdin(t, env, "y\n", "--on", "2026-03-01", "--contains", "First", "--delete", "--num", "99")

	if !strings.Contains(stderr, "1 entry deleted") {
		t.Errorf("expected '1 entry deleted' in stderr, got: %q", stderr)
	}
	day1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	content := dayFileContent(t, env.journalDir, day1)
	if strings.Contains(content, "First @work entry") {
		t.Errorf("expected deleted entry to be gone from day file, got:\n%s", content)
	}
	// Starred entry on same day must survive
	if !strings.Contains(content, "Starred afternoon entry") {
		t.Errorf("expected non-deleted entry to remain, got:\n%s", content)
	}
}

// --- Change time ---

// TestCompat_ChangeTime: jrnl --change-time moves an entry to a new timestamp.
func TestCompat_ChangeTime(t *testing.T) {
	env := newTestEnv(t)
	day1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	writeDayFile(t, env.journalDir, day1,
		"# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nEntry to reschedule.\n")

	_, stderr := runWithStdin(t, env, "y\n", "--on", "2026-03-01", "--change-time=2026-03-05", "--num", "99")

	if !strings.Contains(stderr, "1 entry modified") {
		t.Errorf("expected '1 entry modified' in stderr, got: %q", stderr)
	}
	target := time.Date(2026, 3, 5, 0, 0, 0, 0, time.Local)
	if !dayFileExists(t, env.journalDir, target) {
		t.Fatal("expected day file for 2026-03-05 after --change-time")
	}
	content := dayFileContent(t, env.journalDir, target)
	if !strings.Contains(content, "Entry to reschedule") {
		t.Errorf("expected rescheduled entry in new day file, got:\n%s", content)
	}
}

// --- Journals ---

// TestCompat_ListJournals: jrnl --list shows configured journal names and paths.
func TestCompat_ListJournals(t *testing.T) {
	env := newTestEnv(t)

	stdout, _ := run(t, env, "--list")

	if !strings.Contains(stdout, "default") {
		t.Errorf("expected 'default' journal in --list output, got: %q", stdout)
	}
}

// TestCompat_MultipleJournals: jrnl work: text writes to the named journal.
func TestCompat_MultipleJournals(t *testing.T) {
	workDir := t.TempDir()
	env := newMultiTestEnv(t, map[string]string{
		"default": t.TempDir(),
		"work":    workDir,
	})
	today := time.Now()

	_, stderr := run(t, env, "work:", "Work journal entry.")

	if !strings.Contains(stderr, "Entry added") {
		t.Errorf("expected 'Entry added' when writing to named journal, got: %q", stderr)
	}
	if !dayFileExists(t, workDir, today) {
		t.Fatal("expected day file in work journal")
	}
	content := dayFileContent(t, workDir, today)
	if !strings.Contains(content, "Work journal entry.") {
		t.Errorf("expected entry body in work journal day file, got:\n%s", content)
	}
}

// --- Export ---

// TestCompat_ExportJSON: jrnl --format json outputs valid JSON with expected keys.
func TestCompat_ExportJSON(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	stdout, _ := runAll(t, env, "--format", "json")

	var result struct {
		Entries []map[string]any `json:"entries"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("--format json output is not valid JSON: %v\noutput: %q", err, stdout)
	}
	if len(result.Entries) != 3 {
		t.Errorf("expected 3 entries in JSON, got %d", len(result.Entries))
	}
	for i, entry := range result.Entries {
		for _, key := range []string{"date", "time", "body", "starred"} {
			if _, ok := entry[key]; !ok {
				t.Errorf("entry %d missing key %q in JSON output", i, key)
			}
		}
	}
}

// TestCompat_ExportMarkdown: jrnl --format md outputs markdown.
func TestCompat_ExportMarkdown(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	stdout, _ := runAll(t, env, "--format", "md")

	if !strings.Contains(stdout, "## ") {
		t.Errorf("expected markdown entry headings (## ) in --format md output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "First @work entry") {
		t.Errorf("expected entry body in --format md output, got: %q", stdout)
	}
}

// TestCompat_ExportText: jrnl --format txt outputs plain text.
func TestCompat_ExportText(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	stdout, _ := runAll(t, env, "--format", "txt")

	if !strings.Contains(stdout, "First @work entry") {
		t.Errorf("expected entry body in --format txt output, got: %q", stdout)
	}
	if strings.Contains(stdout, "{") {
		t.Errorf("expected plain text output (no JSON/markdown), got: %q", stdout)
	}
}

// TestCompat_ExportToFile: jrnl --file path writes output to file instead of stdout.
func TestCompat_ExportToFile(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	outPath := filepath.Join(env.dir, "export.json")
	runAll(t, env, "--format", "json", "--file", outPath)

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("expected output file at %s: %v", outPath, err)
	}
	if !strings.Contains(string(data), "entries") {
		t.Errorf("expected JSON 'entries' key in exported file, got: %q", string(data))
	}
}

func TestCompat_Import(t *testing.T) {
	env := newTestEnv(t)

	importContent := "# 2026-01-10 Saturday\n\n## [09:00 AM]\n\nImported first entry.\n\n# 2026-01-11 Sunday\n\n## [03:00 PM]\n\nImported second entry.\n"
	importPath := filepath.Join(env.dir, "import.md")
	if err := os.WriteFile(importPath, []byte(importContent), 0644); err != nil {
		t.Fatalf("failed to write import file: %v", err)
	}

	_, stderr := run(t, env, "--import", importPath)
	if !strings.Contains(stderr, "Imported 2 entries") {
		t.Errorf("expected 'Imported 2 entries' in stderr, got: %q", stderr)
	}
	if !strings.Contains(stderr, "Skipped 0 duplicates") {
		t.Errorf("expected 'Skipped 0 duplicates' in stderr, got: %q", stderr)
	}

	day1 := time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)
	day2 := time.Date(2026, 1, 11, 0, 0, 0, 0, time.Local)
	if !dayFileExists(t, env.journalDir, day1) {
		t.Fatal("expected day file for 2026-01-10 after import")
	}
	if !dayFileExists(t, env.journalDir, day2) {
		t.Fatal("expected day file for 2026-01-11 after import")
	}
	content1 := dayFileContent(t, env.journalDir, day1)
	if !strings.Contains(content1, "Imported first entry.") {
		t.Errorf("expected 'Imported first entry.' in day1 file, got:\n%s", content1)
	}
	content2 := dayFileContent(t, env.journalDir, day2)
	if !strings.Contains(content2, "Imported second entry.") {
		t.Errorf("expected 'Imported second entry.' in day2 file, got:\n%s", content2)
	}

	// Re-import — all entries should be skipped as duplicates
	_, stderr2 := run(t, env, "--import", importPath)
	if !strings.Contains(stderr2, "Imported 0 entries") {
		t.Errorf("expected 'Imported 0 entries' on re-import, got: %q", stderr2)
	}
	if !strings.Contains(stderr2, "Skipped 2 duplicates") {
		t.Errorf("expected 'Skipped 2 duplicates' on re-import, got: %q", stderr2)
	}
}

// --- Config ---

func TestCompat_ConfigFileFlag(t *testing.T) {
	env := newTestEnv(t)
	today := time.Now()

	// Invoke binary directly with --config-file (bypassing newCmd which uses --config)
	cmd := exec.Command(binary, "--config-file", env.configPath, "Config-file-flag compat entry.")
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		t.Fatalf("binary failed with --config-file: %v\nstderr: %s", err, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "Entry added") {
		t.Errorf("expected 'Entry added', got: %q", errBuf.String())
	}
	if !dayFileExists(t, env.journalDir, today) {
		t.Fatal("expected day file for today after --config-file write")
	}
}

func TestCompat_DefaultHourMinute(t *testing.T) {
	env := newTestEnv(t)

	// Patch default_hour and default_minute into the existing config
	data, err := os.ReadFile(env.configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	patched := strings.Replace(string(data),
		"[general]\neditor = \"\"\nhighlight = false\nlinewrap = 0\nindent_character = \"\"",
		"[general]\neditor = \"\"\nhighlight = false\nlinewrap = 0\nindent_character = \"\"\ndefault_hour = 14\ndefault_minute = 30",
		1)
	if !strings.Contains(patched, "default_hour = 14") {
		t.Fatal("config patch for default_hour did not apply — check testConfigHeader")
	}
	if err := os.WriteFile(env.configPath, []byte(patched), 0644); err != nil {
		t.Fatalf("failed to write patched config: %v", err)
	}

	target := time.Date(2025, 6, 1, 0, 0, 0, 0, time.Local)
	run(t, env, "2025-06-01: Default hour test entry.")

	if !dayFileExists(t, env.journalDir, target) {
		t.Fatal("expected day file for 2025-06-01")
	}
	content := dayFileContent(t, env.journalDir, target)
	// The entry heading should reflect 02:30 PM (14:30 in 12h format)
	if !strings.Contains(content, "02:30 PM") {
		t.Errorf("expected time heading '02:30 PM' in day file content, got:\n%s", content)
	}
}

// --- Per-journal config ---

func TestCompat_PerJournalConfig(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	// Set a per-journal editor (global editor stays empty)
	editorPath := writeMockEditor(t, env.dir, "First @work entry", "PerJournal-edited entry")

	data, err := os.ReadFile(env.configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	patched := strings.Replace(string(data),
		fmt.Sprintf("[journals.default]\npath = %q", env.journalDir),
		fmt.Sprintf("[journals.default]\npath = %q\neditor = %q", env.journalDir, editorPath),
		1)
	if !strings.Contains(patched, editorPath) {
		t.Fatalf("per-journal editor patch did not apply")
	}
	if err := os.WriteFile(env.configPath, []byte(patched), 0644); err != nil {
		t.Fatalf("failed to write patched config: %v", err)
	}

	_, stderr := run(t, env, "--edit", "--num", "99")
	if strings.Contains(stderr, "error") || strings.Contains(stderr, "Error") {
		t.Fatalf("unexpected error in stderr: %q", stderr)
	}

	march1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	content := dayFileContent(t, env.journalDir, march1)
	if !strings.Contains(content, "PerJournal-edited entry") {
		t.Errorf("expected per-journal editor to be invoked, got:\n%s", content)
	}
}

func TestCompat_Templates(t *testing.T) {
	t.Skip("pending: template rendering (editor pre-fill) not yet implemented")
}

// --- Display ---

// TestCompat_TagHighlighting: with highlight=true, tags appear in output without errors.
// Note: ANSI codes are suppressed when stdout is not a TTY (e.g. in tests), so this test
// confirms the feature runs without error and produces output — not that it emits ANSI codes.
func TestCompat_TagHighlighting(t *testing.T) {
	env := newTestEnv(t)
	data, err := os.ReadFile(env.configPath)
	if err != nil {
		t.Fatal(err)
	}
	highlighted := strings.Replace(string(data), "highlight = false", "highlight = true", 1)
	highlighted = strings.Replace(highlighted, `tags = "none"`, `tags = "cyan"`, 1)
	if !strings.Contains(highlighted, "highlight = true") {
		t.Fatal("config patch for highlight=true did not apply — check testConfigHeader")
	}
	if err := os.WriteFile(env.configPath, []byte(highlighted), 0644); err != nil {
		t.Fatal(err)
	}

	seedCompatJournal(t, env)
	stdout, stderr := runAll(t, env)

	if strings.Contains(stderr, "error") || strings.Contains(stderr, "Error") {
		t.Errorf("unexpected error with highlight=true: %q", stderr)
	}
	if !strings.Contains(stdout, "@work") {
		t.Errorf("expected @work tag to appear in highlighted output, got: %q", stdout)
	}
}
