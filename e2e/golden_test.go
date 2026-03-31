package e2e

import (
	"flag"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files from jrnl oracle output")

// goldenTest describes a single golden test case.
type goldenTest struct {
	slug       string
	mdArgs     []string                         // args for jrnl-md
	jrnlArgs   []string                         // args for jrnl (if nil, uses mdArgs)
	normalize  func(string) string              // if nil, only normalizeUniversal + stripANSI
	ansi       bool                             // if true, use golden-ansi/ dir, skip ANSI stripping
	seed       func(t *testing.T) (testEnv, jrnlOracle) // if nil, use seedGoldenJournal
	skipOracle bool                             // if true, don't run jrnl for capture
}

var goldenTests = []goldenTest{
	{
		slug:      "read-all",
		mdArgs:    []string{"-n", "99"},
		normalize: normalizeDefault,
	},
	{
		slug:      "read-last-1",
		mdArgs:    []string{"-n", "1"},
		normalize: normalizeDefault,
	},
	{
		slug:      "short",
		mdArgs:    []string{"--short", "-n", "99"},
		normalize: normalizeShort,
	},
	{
		slug:      "starred",
		mdArgs:    []string{"--starred", "-n", "99"},
		jrnlArgs:  []string{"-starred", "-n", "99"},
		normalize: normalizeDefault,
	},
	{
		slug:      "tags-list",
		mdArgs:    []string{"--tags"},
		normalize: normalizeTags,
	},
	{
		slug:      "list-journals",
		mdArgs:    []string{"--list"},
		normalize: normalizeList,
	},
	// --- Filters ---
	{slug: "filter-tag", mdArgs: []string{"@work", "-n", "99"}, normalize: normalizeDefault},
	{slug: "filter-tag-and", mdArgs: []string{"@personal", "--and", "@life", "-n", "99"}, jrnlArgs: []string{"@personal", "-and", "@life", "-n", "99"}, normalize: normalizeDefault},
	{slug: "filter-not-tag", mdArgs: []string{"--not", "@work", "-n", "99"}, jrnlArgs: []string{"-not", "@work", "-n", "99"}, normalize: normalizeDefault},
	{slug: "filter-not-starred", mdArgs: []string{"--not-starred", "-n", "99"}, jrnlArgs: []string{"-not", "starred", "-n", "99"}, normalize: normalizeDefault},
	{slug: "filter-not-tagged", mdArgs: []string{"--not-tagged", "-n", "99"}, jrnlArgs: []string{"-not", "tagged", "-n", "99"}, normalize: normalizeDefault},
	{slug: "filter-contains", mdArgs: []string{"--contains", "afternoon", "-n", "99"}, jrnlArgs: []string{"-contains", "afternoon", "-n", "99"}, normalize: normalizeDefault},
	{slug: "filter-from", mdArgs: []string{"--from", "2026-03-10", "-n", "99"}, jrnlArgs: []string{"-from", "2026-03-10", "-n", "99"}, normalize: normalizeDefault},
	{slug: "filter-to", mdArgs: []string{"--to", "2026-03-10", "-n", "99"}, jrnlArgs: []string{"-to", "2026-03-10", "-n", "99"}, normalize: normalizeDefault},
	{slug: "filter-on", mdArgs: []string{"--on", "2026-03-01", "-n", "99"}, jrnlArgs: []string{"-on", "2026-03-01", "-n", "99"}, normalize: normalizeDefault},
	// --- Combined Filters ---
	{slug: "filter-tag-starred", mdArgs: []string{"@work", "--starred", "-n", "99"}, jrnlArgs: []string{"@work", "-starred", "-n", "99"}, normalize: normalizeDefault},
	{slug: "filter-from-to", mdArgs: []string{"--from", "2026-03-05", "--to", "2026-03-10", "-n", "99"}, jrnlArgs: []string{"-from", "2026-03-05", "-to", "2026-03-10", "-n", "99"}, normalize: normalizeDefault},
	{slug: "filter-tag-from", mdArgs: []string{"@work", "--from", "2026-03-05", "-n", "99"}, jrnlArgs: []string{"@work", "-from", "2026-03-05", "-n", "99"}, normalize: normalizeDefault},
	{slug: "filter-contains-n1", mdArgs: []string{"--contains", "morning", "-n", "1"}, jrnlArgs: []string{"-contains", "morning", "-n", "1"}, normalize: normalizeDefault},
	{slug: "filter-not-tag-starred", mdArgs: []string{"--not", "@work", "--starred", "-n", "99"}, jrnlArgs: []string{"-not", "@work", "-starred", "-n", "99"}, normalize: normalizeDefault},
	// --- Export Formats ---
	{slug: "export-json", mdArgs: []string{"--format", "json", "-n", "99"}, normalize: normalizeJSON},
	{slug: "export-txt", mdArgs: []string{"--format", "txt", "-n", "99"}, normalize: normalizeTxt},
	{slug: "export-md", mdArgs: []string{"--format", "md", "-n", "99"}, normalize: normalizeMd},
	{slug: "export-xml", mdArgs: []string{"--format", "xml", "-n", "99"}, skipOracle: true},
	{slug: "export-yaml", mdArgs: []string{"--format", "yaml", "-n", "99"}, skipOracle: true},
	// --- Export + Filter Combos ---
	{slug: "export-json-tag", mdArgs: []string{"@work", "--format", "json", "-n", "99"}, normalize: normalizeJSON},
	{slug: "export-txt-from", mdArgs: []string{"--from", "2026-03-10", "--format", "txt", "-n", "99"}, jrnlArgs: []string{"-from", "2026-03-10", "--format", "txt", "-n", "99"}, normalize: normalizeTxt},
	{slug: "export-json-starred", mdArgs: []string{"--starred", "--format", "json", "-n", "99"}, jrnlArgs: []string{"-starred", "--format", "json", "-n", "99"}, normalize: normalizeJSON},
	// --- Config Variations ---
	{slug: "linewrap-40", mdArgs: []string{"-n", "99"}, normalize: normalizeDefault, seed: seedGoldenLinewrap40},
	{slug: "highlight-off", mdArgs: []string{"-n", "99"}, normalize: normalizeDefault},
	{slug: "default-hour-minute", mdArgs: []string{"-n", "99"}, normalize: normalizeDefault, seed: seedGoldenDefaultHourMinute},
	{slug: "tag-symbols-hash", mdArgs: []string{"--tags"}, normalize: normalizeTags, seed: seedGoldenHashTags},
	// --- ANSI Color Tests ---
	{slug: "color-tags-cyan", mdArgs: []string{"-n", "99"}, ansi: true, seed: seedGoldenANSI},
	{slug: "color-tags-list", mdArgs: []string{"--tags"}, ansi: true, seed: seedGoldenANSI},
	// --- Multiple Journals ---
	{slug: "multi-list", mdArgs: []string{"--list"}, normalize: normalizeList, seed: seedGoldenMulti},
	{slug: "multi-read", mdArgs: []string{"work:", "-n", "99"}, jrnlArgs: []string{"work:", "-n", "99"}, normalize: normalizeDefault, seed: seedGoldenMulti},
	{slug: "multi-tags", mdArgs: []string{"work:", "--tags"}, jrnlArgs: []string{"work:", "--tags"}, normalize: normalizeTags, seed: seedGoldenMulti},
	// --- Edge Cases ---
	{slug: "empty-journal", mdArgs: []string{"-n", "99"}, seed: seedGoldenEmpty},
	{slug: "single-entry", mdArgs: []string{"-n", "99"}, normalize: normalizeDefault, seed: seedGoldenSingle},
	{slug: "no-matches", mdArgs: []string{"--contains", "nonexistent", "-n", "99"}, jrnlArgs: []string{"-contains", "nonexistent", "-n", "99"}},
}

func TestGolden(t *testing.T) {
	for _, tt := range goldenTests {
		t.Run(tt.slug, func(t *testing.T) {
			// 1. Seed environments
			var env testEnv
			var oracle jrnlOracle
			if tt.seed != nil {
				env, oracle = tt.seed(t)
			} else {
				env, oracle = seedGoldenJournal(t)
			}

			// 2. Determine golden file directory
			dir := goldenDir(t)
			if tt.ansi {
				dir = goldenANSIDir(t)
			}
			filename := tt.slug + ".txt"

			// 3. Update or read golden file
			if *updateGolden {
				if tt.skipOracle {
					// Capture jrnl-md output as golden (for broken jrnl features)
					stdout, _ := run(t, env, tt.mdArgs...)
					golden := stdout
					if !tt.ansi {
						golden = stripANSI(golden)
					}
					golden = normalizeUniversal(golden)
					writeGolden(t, dir, filename, golden)
					t.Logf("captured jrnl-md output as golden (no oracle): %s/%s", dir, filename)
					return
				}
				jArgs := tt.jrnlArgs
				if jArgs == nil {
					jArgs = tt.mdArgs
				}
				stdout := oracle.run(t, jArgs...)
				golden := stdout
				if !tt.ansi {
					golden = stripANSI(golden)
				}
				golden = normalizeUniversal(golden)
				if tt.normalize != nil {
					golden = tt.normalize(golden)
				}
				writeGolden(t, dir, filename, golden)
				t.Logf("updated golden file: %s/%s", dir, filename)
				return
			}

			golden, ok := readGolden(t, dir, filename)
			if !ok {
				t.Skipf("golden file missing: %s/%s (run with -update-golden)", dir, filename)
				return
			}

			// 4. Run jrnl-md
			stdout, _ := run(t, env, tt.mdArgs...)

			// 5. Normalize
			actual := stdout
			if !tt.ansi {
				actual = stripANSI(actual)
			}
			actual = normalizeUniversal(actual)
			if tt.normalize != nil {
				actual = tt.normalize(actual)
			}
			golden = normalizeUniversal(golden)
			if tt.normalize != nil {
				golden = tt.normalize(golden)
			}

			// 6. Compare
			if actual != golden {
				t.Errorf("golden mismatch for %s\n\nGolden: %s/%s\n\nDiff:\n%s\n\nActual:\n%s",
					tt.slug, dir, filename, unifiedDiff(golden, actual), stdout)
			}
		})
	}
}
