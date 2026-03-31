package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// jrnlOracle — calls the real jrnl binary
// ---------------------------------------------------------------------------

// jrnlOracle wraps a jrnl config file for running the reference implementation.
type jrnlOracle struct {
	configPath string
}

// run executes `jrnl --config-file <configPath> <args>` and returns stdout.
// stderr (the "N entries found" box) is discarded.
func (o jrnlOracle) run(t *testing.T, args ...string) string {
	t.Helper()
	full := append([]string{"--config-file", o.configPath}, args...)
	cmd := exec.Command("jrnl", full...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		t.Logf("jrnl stderr: %s", errBuf.String())
		t.Fatalf("jrnl exited with error: %v", err)
	}
	return outBuf.String()
}

// ---------------------------------------------------------------------------
// Golden file directories and I/O
// ---------------------------------------------------------------------------

// goldenDir returns (and creates) the testdata/golden/ directory.
func goldenDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join("testdata", "golden")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("goldenDir: %v", err)
	}
	return dir
}

// goldenANSIDir returns (and creates) the testdata/golden-ansi/ directory.
func goldenANSIDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join("testdata", "golden-ansi")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("goldenANSIDir: %v", err)
	}
	return dir
}

// writeGolden writes content to dir/name.
func writeGolden(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeGolden %s: %v", name, err)
	}
}

// readGolden reads and returns dir/name, or returns ("", false) if not found.
func readGolden(t *testing.T, dir, name string) (string, bool) {
	t.Helper()
	path := filepath.Join(dir, name)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", false
	}
	if err != nil {
		t.Fatalf("readGolden %s: %v", name, err)
	}
	return string(data), true
}

// ---------------------------------------------------------------------------
// Config generators
// ---------------------------------------------------------------------------

// goldenJrnlConfig returns a jrnl YAML config string pointing at journalPath.
// journalPath should be a plain file (jrnl uses a single .txt file).
func goldenJrnlConfig(journalPath string) string {
	return fmt.Sprintf(`colors:
  body: none
  date: none
  tags: none
  title: none
default_hour: 9
default_minute: 0
editor: ''
encrypt: false
highlight: false
indent_character: ''
journals:
  default: %s
linewrap: 0
tagsymbols: '@'
template: false
timeformat: '%%Y-%%m-%%d %%H:%%M'
version: v4.3
`, journalPath)
}

// goldenJrnlMdConfigHeader is the TOML config preamble for jrnl-md golden tests.
// It uses 24h time format and disables colors/highlighting.
const goldenJrnlMdConfigHeader = `[general]
editor = ""
highlight = false
linewrap = 0
indent_character = ""

[format]
time = "15:04"
date = "2006-01-02"
tag_symbols = "@"
file_extension = "md"

[colors]
date = "none"
body = "none"
tags = "none"

`

// ---------------------------------------------------------------------------
// Seed entries
// ---------------------------------------------------------------------------

// goldenEntry is a single journal entry used to seed both jrnl and jrnl-md.
type goldenEntry struct {
	Date   string // YYYY-MM-DD
	Time   string // HH:MM (24h)
	Body   string
	Starred bool
}

// goldenEntries is the canonical set of 6 seed entries used in golden tests.
var goldenEntries = []goldenEntry{
	{
		Date:    "2026-03-01",
		Time:    "09:00",
		Body:    "First @work entry with a longer body that can test linewrap behavior when the configuration sets a narrow column width.",
		Starred: false,
	},
	{
		Date:    "2026-03-01",
		Time:    "14:00",
		Body:    "Starred afternoon entry.",
		Starred: true,
	},
	{
		Date:    "2026-03-05",
		Time:    "10:30",
		Body:    "A @personal reflection about @life and the importance of journaling regularly.",
		Starred: false,
	},
	{
		Date:    "2026-03-10",
		Time:    "08:00",
		Body:    "Morning @work standup notes covering the sprint review.",
		Starred: false,
	},
	{
		Date:    "2026-03-10",
		Time:    "20:00",
		Body:    "Evening thoughts.",
		Starred: true,
	},
	{
		Date:    "2026-03-15",
		Time:    "11:00",
		Body:    "Mid-month @personal entry.",
		Starred: false,
	},
}

// ---------------------------------------------------------------------------
// Seeding helpers
// ---------------------------------------------------------------------------

// seedGoldenJournal creates both a jrnl-md testEnv and a jrnlOracle seeded
// with the canonical goldenEntries. It uses the default config functions.
func seedGoldenJournal(t *testing.T) (testEnv, jrnlOracle) {
	t.Helper()
	return seedGoldenWithEntries(t, goldenEntries, goldenJrnlMdConfigHeader, goldenJrnlConfig)
}

// seedGoldenWithEntries seeds both jrnl-md and jrnl with the provided entries,
// using the supplied config header for jrnl-md and config generator for jrnl.
//
// jrnlConfigFn receives the path to the jrnl journal .txt file and returns
// the full YAML config string to write.
func seedGoldenWithEntries(
	t *testing.T,
	entries []goldenEntry,
	mdConfigHeader string,
	jrnlConfigFn func(journalPath string) string,
) (testEnv, jrnlOracle) {
	t.Helper()

	// --- jrnl-md environment ---
	dir := t.TempDir()
	journalDir := filepath.Join(dir, "journal")
	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("seedGoldenWithEntries: mkdir journal: %v", err)
	}

	configPath := filepath.Join(dir, "config.toml")
	config := mdConfigHeader + fmt.Sprintf("[journals.default]\npath = %q\n", journalDir)
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatalf("seedGoldenWithEntries: write config: %v", err)
	}

	env := testEnv{
		dir:        dir,
		configPath: configPath,
		journalDir: journalDir,
	}

	// Group entries by date to build day files.
	type dayGroup struct {
		date    time.Time
		entries []goldenEntry
	}
	seen := map[string]*dayGroup{}
	var order []string
	for _, e := range entries {
		if _, ok := seen[e.Date]; !ok {
			d, err := time.Parse("2006-01-02", e.Date)
			if err != nil {
				t.Fatalf("seedGoldenWithEntries: parse date %q: %v", e.Date, err)
			}
			seen[e.Date] = &dayGroup{date: d}
			order = append(order, e.Date)
		}
		seen[e.Date].entries = append(seen[e.Date].entries, e)
	}

	for _, dateStr := range order {
		g := seen[dateStr]
		content := buildMdDayFile(g.date, g.entries)
		writeDayFile(t, journalDir, g.date, content)
	}

	// --- jrnl environment ---
	jrnlDir := t.TempDir()
	jrnlJournalPath := filepath.Join(jrnlDir, "journal.txt")
	jrnlConfigPath := filepath.Join(jrnlDir, "jrnl.yaml")

	jrnlContent := buildJrnlFile(entries)
	if err := os.WriteFile(jrnlJournalPath, []byte(jrnlContent), 0644); err != nil {
		t.Fatalf("seedGoldenWithEntries: write jrnl journal: %v", err)
	}

	jrnlConfig := jrnlConfigFn(jrnlJournalPath)
	if err := os.WriteFile(jrnlConfigPath, []byte(jrnlConfig), 0644); err != nil {
		t.Fatalf("seedGoldenWithEntries: write jrnl config: %v", err)
	}

	oracle := jrnlOracle{configPath: jrnlConfigPath}
	return env, oracle
}

// buildMdDayFile constructs the jrnl-md markdown content for a single day.
func buildMdDayFile(date time.Time, entries []goldenEntry) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n", date.Format("2006-01-02 Monday")))
	for _, e := range entries {
		sb.WriteString("\n")
		if e.Starred {
			sb.WriteString(fmt.Sprintf("## [%s] *\n", e.Time))
		} else {
			sb.WriteString(fmt.Sprintf("## [%s]\n", e.Time))
		}
		sb.WriteString("\n")
		sb.WriteString(e.Body)
		sb.WriteString("\n")
	}
	return sb.String()
}

// buildJrnlFile constructs the jrnl single-file journal content.
func buildJrnlFile(entries []goldenEntry) string {
	var sb strings.Builder
	for _, e := range entries {
		line := fmt.Sprintf("[%s %s] %s", e.Date, e.Time, e.Body)
		if e.Starred {
			line += " *"
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}

// ---------------------------------------------------------------------------
// Normalization functions
// ---------------------------------------------------------------------------

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes all ANSI color/style escape sequences from s.
func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

// jrnl's "N entries found" box (unicode box-drawing output on stderr) and
// similar decorative lines we want to drop universally.
var jrnlBoxRe = regexp.MustCompile(`(?m)^[┏┗┃].+\n?`)

// normalizeUniversal applies transformations that apply to every output mode:
//   - Strip ANSI codes
//   - Remove jrnl's "N entries found" box-drawing lines
//   - Trim trailing whitespace on each line
//   - Normalize line endings to \n
//   - Trim leading/trailing blank lines
func normalizeUniversal(s string) string {
	s = stripANSI(s)
	s = jrnlBoxRe.ReplaceAllString(s, "")
	// Remove jrnl-md's plain "N entries found\n" header line.
	s = regexp.MustCompile(`(?m)^\d+ entr(y|ies) found\n?`).ReplaceAllString(s, "")
	// Trim trailing whitespace per line.
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t\r")
	}
	s = strings.Join(lines, "\n")
	// Collapse runs of blank lines to a single blank line.
	s = regexp.MustCompile(`\n{3,}`).ReplaceAllString(s, "\n\n")
	return strings.Trim(s, "\n")
}

// normalizeJSON normalizes JSON output for comparison:
//   - Removes the "title" key from each entry (jrnl has it, jrnl-md doesn't)
//   - Sorts entry tags for stable comparison
//   - Re-marshals with canonical indentation
func normalizeJSON(s string) string {
	s = normalizeUniversal(s)
	var data map[string]any
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		// If it won't parse, return as-is so the test failure is obvious.
		return s
	}

	// Normalize entries slice.
	if entries, ok := data["entries"].([]any); ok {
		for _, raw := range entries {
			entry, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			// jrnl splits text into title + body; jrnl-md has only body.
			// Merge title into body before removing the key.
			if title, ok := entry["title"].(string); ok && title != "" {
				if body, _ := entry["body"].(string); body != "" {
					entry["body"] = title + "\n" + body
				} else {
					entry["body"] = title
				}
			}
			delete(entry, "title")
			// Sort tags slice for stable comparison.
			if tags, ok := entry["tags"].([]any); ok {
				strs := make([]string, len(tags))
				for i, tg := range tags {
					strs[i] = fmt.Sprint(tg)
				}
				sort.Strings(strs)
				sorted := make([]any, len(strs))
				for i, s := range strs {
					sorted[i] = s
				}
				entry["tags"] = sorted
			}
		}
	}

	// Sort top-level tags map is already a map (stable in JSON marshal order
	// since Go 1.12 sorts map keys alphabetically when marshaling).

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return s
	}
	return string(out)
}

// normalizeShort normalizes --short output.
// jrnl: `DATE TIME Full title text (no truncation)`
// jrnl-md: `DATE TIME First 60 chars of body...`
// Strategy: keep only the `DATE TIME` prefix from each line; mask the rest.
func normalizeShort(s string) string {
	s = normalizeUniversal(s)
	// Match lines starting with a timestamp.
	tsRe := regexp.MustCompile(`(?m)^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}).*`)
	return tsRe.ReplaceAllString(s, "$1")
}

// normalizeList normalizes --list output.
// jrnl: `Journals defined in config (/path/to/config)\n * name -> /path/to/journal`
// jrnl-md: `Journals:\n  name -> /path/to/journal`
// Strategy: extract "name" from lines containing " -> ", one per line, sorted.
func normalizeList(s string) string {
	s = normalizeUniversal(s)
	arrowRe := regexp.MustCompile(`(\w+)\s*->`)
	var names []string
	for _, line := range strings.Split(s, "\n") {
		if m := arrowRe.FindStringSubmatch(line); m != nil {
			names = append(names, m[1])
		}
	}
	sort.Strings(names)
	return strings.Join(names, "\n")
}

// normalizeDefault normalizes default (full-body) display output.
//
// jrnl format:
//
//	DATE TIME Title text of entry.
//	[blank]
//	Body continuation (if any)
//
// jrnl-md format:
//
//	DATE TIME
//	[blank]
//	Full body text (possibly wrapped)
//
// Strategy: extract (date, time, body) tuples from both; compare them.
// Star markers differ too: jrnl appends " *" in the title, jrnl-md puts it
// in the heading — we strip stars for body comparison.
func normalizeDefault(s string) string {
	s = normalizeUniversal(s)
	type entry struct {
		datetime string
		body     string
	}

	// Regex to detect a timestamp line (may be followed by a title or alone).
	tsLineRe := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2})(.*)$`)

	var entries []entry
	lines := strings.Split(s, "\n")
	var cur *entry
	for _, line := range lines {
		if m := tsLineRe.FindStringSubmatch(line); m != nil {
			if cur != nil {
				cur.body = strings.TrimSpace(cur.body)
				entries = append(entries, *cur)
			}
			// m[2] may contain the title inline (jrnl) or be empty (jrnl-md).
			title := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(m[2]), " *"))
			cur = &entry{datetime: m[1], body: title}
		} else if cur != nil && line != "" {
			if cur.body == "" {
				cur.body = line
			} else {
				cur.body += " " + line
			}
		}
	}
	if cur != nil {
		cur.body = strings.TrimSpace(cur.body)
		entries = append(entries, *cur)
	}

	// Rebuild as canonical lines: "DATETIME | BODY"
	var out []string
	for _, e := range entries {
		// Normalize star from body if present.
		body := strings.TrimSuffix(strings.TrimSpace(e.body), " *")
		body = strings.TrimSpace(body)
		out = append(out, e.datetime+" | "+body)
	}
	return strings.Join(out, "\n")
}

// normalizeTxt normalizes --format txt / --export txt output.
// jrnl:    `[DATE TIME] Body. *\n\n`
// jrnl-md: `[DATE TIME] Body.\n\n`
// Strategy: universal normalize, then strip trailing " *" before closing `]`-adjacent
// positions — actually stars are outside the bracket in txt. Strip " *" at end of
// content lines.
func normalizeTxt(s string) string {
	s = normalizeUniversal(s)
	// Strip trailing " *" from content lines.
	s = regexp.MustCompile(`(?m) \*$`).ReplaceAllString(s, "")
	return s
}

// normalizeMd normalizes --format md / --export md output.
// Both produce `### DATE TIME [Title]` headings but jrnl appends the title
// inline whereas jrnl-md does not. Strip text after the timestamp in headings.
func normalizeMd(s string) string {
	s = normalizeUniversal(s)
	// Heading lines: ### YYYY-MM-DD HH:MM [rest...]  — keep only timestamp.
	headingRe := regexp.MustCompile(`(?m)^(#{1,6} \d{4}-\d{2}-\d{2} \d{2}:\d{2}).*$`)
	s = headingRe.ReplaceAllString(s, "$1")
	// Strip trailing " *" anywhere.
	s = regexp.MustCompile(`(?m) \*$`).ReplaceAllString(s, "")
	// Remove jrnl's trailing ` ` lines (single space) left after body.
	s = regexp.MustCompile(`(?m)^ $`).ReplaceAllString(s, "")
	return s
}

// normalizeTags normalizes --tags output.
// Both tools use `%-20s : %d` format. Sort lines for stable comparison.
func normalizeTags(s string) string {
	s = normalizeUniversal(s)
	lines := strings.Split(s, "\n")
	var tagLines []string
	for _, l := range lines {
		if strings.Contains(l, ":") {
			tagLines = append(tagLines, strings.TrimSpace(l))
		}
	}
	sort.Strings(tagLines)
	return strings.Join(tagLines, "\n")
}

// ---------------------------------------------------------------------------
// Config variation seed helpers
// ---------------------------------------------------------------------------

// seedGoldenLinewrap40 seeds with linewrap: 40 in both configs.
func seedGoldenLinewrap40(t *testing.T) (testEnv, jrnlOracle) {
	t.Helper()
	mdConfig := strings.Replace(goldenJrnlMdConfigHeader, "linewrap = 0", "linewrap = 40", 1)
	jrnlCfgFn := func(journalPath string) string {
		return strings.Replace(goldenJrnlConfig(journalPath), "linewrap: 0", "linewrap: 40", 1)
	}
	return seedGoldenWithEntries(t, goldenEntries, mdConfig, jrnlCfgFn)
}

// seedGoldenDefaultHourMinute seeds with default_hour 14 and default_minute 30,
// with one entry already at 14:30.
func seedGoldenDefaultHourMinute(t *testing.T) (testEnv, jrnlOracle) {
	t.Helper()
	entries := []goldenEntry{
		{"2026-03-01", "14:30", "Entry at default time.", false},
	}
	jrnlCfgFn := func(journalPath string) string {
		cfg := goldenJrnlConfig(journalPath)
		cfg = strings.Replace(cfg, "default_hour: 9", "default_hour: 14", 1)
		cfg = strings.Replace(cfg, "default_minute: 0", "default_minute: 30", 1)
		return cfg
	}
	return seedGoldenWithEntries(t, entries, goldenJrnlMdConfigHeader, jrnlCfgFn)
}

// seedGoldenHashTags seeds with '#' as the tag symbol in both configs.
func seedGoldenHashTags(t *testing.T) (testEnv, jrnlOracle) {
	t.Helper()
	entries := []goldenEntry{
		{"2026-03-01", "09:00", "First #work entry.", false},
		{"2026-03-01", "14:00", "Starred #personal moment.", true},
		{"2026-03-05", "10:30", "A #personal reflection about #life.", false},
	}
	mdConfig := strings.Replace(goldenJrnlMdConfigHeader, `tag_symbols = "@"`, `tag_symbols = "#"`, 1)
	jrnlCfgFn := func(journalPath string) string {
		return strings.Replace(goldenJrnlConfig(journalPath), "tagsymbols: '@'", "tagsymbols: '#'", 1)
	}
	return seedGoldenWithEntries(t, entries, mdConfig, jrnlCfgFn)
}

// seedGoldenANSI seeds with highlight: true and colors.tags: cyan.
func seedGoldenANSI(t *testing.T) (testEnv, jrnlOracle) {
	t.Helper()
	mdConfig := strings.Replace(
		strings.Replace(goldenJrnlMdConfigHeader, "highlight = false", "highlight = true", 1),
		`tags = "none"`, `tags = "cyan"`, 1)
	jrnlCfgFn := func(journalPath string) string {
		cfg := goldenJrnlConfig(journalPath)
		cfg = strings.Replace(cfg, "highlight: false", "highlight: true", 1)
		return strings.Replace(cfg, "tags: none", "tags: cyan", 1)
	}
	return seedGoldenWithEntries(t, goldenEntries, mdConfig, jrnlCfgFn)
}

// seedGoldenMulti seeds two journals (default + work) in both jrnl-md and jrnl.
func seedGoldenMulti(t *testing.T) (testEnv, jrnlOracle) {
	t.Helper()
	dir := t.TempDir()

	// jrnl-md: two journal directories
	defaultDir := filepath.Join(dir, "md-default")
	workDir := filepath.Join(dir, "md-work")
	for _, d := range []string{defaultDir, workDir} {
		if err := os.MkdirAll(filepath.Join(d, "2026", "03"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(defaultDir, "2026", "03", "01.md"),
		[]byte("# 2026-03-01 Sunday\n\n## [09:00]\n\nDefault journal entry.\n"), 0644)
	os.WriteFile(filepath.Join(workDir, "2026", "03", "01.md"),
		[]byte("# 2026-03-01 Sunday\n\n## [09:00]\n\nWork @project entry.\n\n## [14:00]\n\nWork @meeting notes.\n"), 0644)

	mdConfigPath := filepath.Join(dir, "config.toml")
	mdConfig := goldenJrnlMdConfigHeader +
		fmt.Sprintf("[journals.default]\npath = %q\n\n[journals.work]\npath = %q\n", defaultDir, workDir)
	os.WriteFile(mdConfigPath, []byte(mdConfig), 0644)
	mdEnv := testEnv{dir: dir, configPath: mdConfigPath, journalDir: defaultDir}

	// jrnl: two journal files
	jrnlDir := t.TempDir()
	defaultPath := filepath.Join(jrnlDir, "default.txt")
	workPath := filepath.Join(jrnlDir, "work.txt")
	os.WriteFile(defaultPath, []byte("[2026-03-01 09:00] Default journal entry.\n"), 0644)
	os.WriteFile(workPath, []byte("[2026-03-01 09:00] Work @project entry.\n[2026-03-01 14:00] Work @meeting notes.\n"), 0644)

	jrnlConfigPath := filepath.Join(jrnlDir, "jrnl.yaml")
	jrnlConfig := fmt.Sprintf(`colors:
  body: none
  date: none
  tags: none
  title: none
default_hour: 9
default_minute: 0
editor: ''
encrypt: false
highlight: false
indent_character: ''
journals:
  default: %s
  work: %s
linewrap: 0
tagsymbols: '@'
template: false
timeformat: '%%Y-%%m-%%d %%H:%%M'
version: v4.3
`, defaultPath, workPath)
	os.WriteFile(jrnlConfigPath, []byte(jrnlConfig), 0644)

	oracle := jrnlOracle{configPath: jrnlConfigPath}
	return mdEnv, oracle
}

// seedGoldenEmpty seeds an empty journal.
func seedGoldenEmpty(t *testing.T) (testEnv, jrnlOracle) {
	t.Helper()
	return seedGoldenWithEntries(t, nil, goldenJrnlMdConfigHeader, goldenJrnlConfig)
}

// seedGoldenSingle seeds a journal with a single entry.
func seedGoldenSingle(t *testing.T) (testEnv, jrnlOracle) {
	t.Helper()
	entries := []goldenEntry{
		{"2026-03-01", "09:00", "Only entry in the journal.", false},
	}
	return seedGoldenWithEntries(t, entries, goldenJrnlMdConfigHeader, goldenJrnlConfig)
}

// ---------------------------------------------------------------------------
// Diff helper
// ---------------------------------------------------------------------------

// unifiedDiff returns a simple unified-style diff of want vs got.
// It is intentionally simple — sufficient for test failure messages.
func unifiedDiff(want, got string) string {
	if want == got {
		return ""
	}
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")

	var sb strings.Builder
	sb.WriteString("--- want\n+++ got\n")

	max := len(wantLines)
	if len(gotLines) > max {
		max = len(gotLines)
	}
	for i := 0; i < max; i++ {
		var w, g string
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if w == g {
			sb.WriteString("  " + w + "\n")
		} else {
			if i < len(wantLines) {
				sb.WriteString("- " + w + "\n")
			}
			if i < len(gotLines) {
				sb.WriteString("+ " + g + "\n")
			}
		}
	}
	return sb.String()
}
