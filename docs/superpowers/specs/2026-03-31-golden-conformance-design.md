# Golden Conformance Test Suite — Design Spec

## Goal

A one-time, comprehensive test suite that verifies jrnl-md's output matches canonical
jrnl exactly — not just behavioral correctness but visual/output fidelity. The suite
captures jrnl's output as golden files and compares jrnl-md's output against them, with
documented normalizations for known exceptions.

## Strategy: Snapshot Oracle with Golden Files

1. A **capture step** (`-update-golden` flag) runs canonical jrnl against a fixed seed
   dataset and saves stdout to `e2e/testdata/golden/` as plain-text fixtures.
2. Normal test runs compare jrnl-md output against these saved golden files.
3. No jrnl dependency at test time — only needed when re-capturing.
4. CI runs against golden files only; re-capture is a manual, intentional audit step.

## Scope

### What is tested
- Every read/display command
- Every filter flag and common filter combinations
- Every export format
- Config keys that affect visible output: `linewrap`, `highlight`, `colors.tags`,
  `default_hour`/`default_minute`, `tagsymbols`
- ANSI color output for tag highlighting
- Multiple journals
- Edge cases: empty journal, single entry, no matches

### What is NOT tested
- Interactive behavior (editor, templates, prompts)
- Encryption (changes storage, not display)
- Write operations (already covered by existing e2e tests)
- `--edit`, `--delete`, `--change-time` (mutating operations with interactive prompts)

## Files

| File | Purpose |
|------|---------|
| `e2e/golden_test.go` | Test table, runner loop, comparison logic |
| `e2e/golden_helpers_test.go` | `jrnlOracle` struct, seeding, normalization, ANSI stripping, diff |
| `e2e/testdata/golden/*.txt` | Plain-text golden files (ANSI stripped) |
| `e2e/testdata/golden-ansi/*.txt` | Raw ANSI golden files (color tests) |

No changes to existing test files.

## Seed Data

6 entries across 4 days, shared by both jrnl and jrnl-md environments:

| Date | Time | Body | Tags | Star |
|------|------|------|------|------|
| 2026-03-01 | 09:00 | First @work entry with a longer body that can test linewrap behavior when the configuration sets a narrow column width. | @work | no |
| 2026-03-01 | 14:00 | Starred afternoon entry. | — | yes |
| 2026-03-05 | 10:30 | A @personal reflection about @life and the importance of journaling regularly. | @personal @life | no |
| 2026-03-10 | 08:00 | Morning @work standup notes covering the sprint review. | @work | no |
| 2026-03-10 | 20:00 | Evening thoughts. | — | yes |
| 2026-03-15 | 11:00 | Mid-month @personal entry. | @personal | no |

The `tag-symbols-hash` config variation uses a parallel seed with `#` instead of `@`.

## Test Matrix (~40 cases)

### Read/Display (default config)

| Slug | Args |
|------|------|
| `read-all` | `-n 99` |
| `read-last-1` | `-n 1` |
| `short` | `--short` |
| `starred` | `--starred` |
| `tags-list` | `--tags` |
| `list-journals` | `--list` |

### Filters

| Slug | Args |
|------|------|
| `filter-tag` | `@work` |
| `filter-tag-and` | `@personal --and @life` |
| `filter-not-tag` | `--not @work` |
| `filter-not-starred` | `--not-starred` |
| `filter-not-tagged` | `--not-tagged` |
| `filter-contains` | `--contains afternoon` |
| `filter-from` | `--from 2026-03-10` |
| `filter-to` | `--to 2026-03-10` |
| `filter-on` | `--on 2026-03-01` |

### Combined Filters

| Slug | Args |
|------|------|
| `filter-tag-starred` | `@work --starred` |
| `filter-from-to` | `--from 2026-03-05 --to 2026-03-10` |
| `filter-tag-from` | `@work --from 2026-03-05` |
| `filter-contains-n1` | `--contains morning -n 1` |
| `filter-not-tag-starred` | `--not @work --starred` |

### Export Formats

| Slug | Args |
|------|------|
| `export-json` | `--format json` |
| `export-txt` | `--format txt` |
| `export-md` | `--format md` |
| `export-xml` | `--format xml` |
| `export-yaml` | `--format yaml` |

### Export + Filter Combos

| Slug | Args |
|------|------|
| `export-json-tag` | `@work --format json` |
| `export-txt-from` | `--from 2026-03-10 --format txt` |
| `export-json-starred` | `--starred --format json` |

### Export to File

| Slug | Args | Notes |
|------|------|-------|
| `export-file-json` | `--format json --file <tmpdir>/out.json` | Verify file contents match `export-json` golden file |

### Config Variations (separate seed per config)

| Slug | Config Override | Notes |
|------|----------------|-------|
| `linewrap-40` | `linewrap: 40` | Uses the long-body entry |
| `highlight-off` | `highlight: false` | Verify no ANSI codes |
| `default-hour-minute` | `default_hour: 14, default_minute: 30` | Seed includes entry at 14:30 (the default time); verify display timestamp |
| `tag-symbols-hash` | `tagsymbols: "#"` | Parallel seed with `#tags` |

### ANSI Color Tests (stored in `golden-ansi/`)

| Slug | Config | Args |
|------|--------|------|
| `color-tags-cyan` | `highlight: true, colors.tags: cyan` | `-n 99` |
| `color-tags-list` | same | `--tags` |

### Multiple Journals

| Slug | Args | Notes |
|------|------|-------|
| `multi-list` | `--list` | Two journals configured |
| `multi-read` | `work: -n 99` | Read from named journal |
| `multi-tags` | `work: --tags` | Tags from named journal |

### Edge Cases

| Slug | Seed | Args |
|------|------|------|
| `empty-journal` | no entries | `-n 99` |
| `single-entry` | 1 entry | `-n 99` |
| `no-matches` | standard seed | `--contains nonexistent` |

## Normalization

### Layer 1: Universal (all tests)

- Strip trailing whitespace per line
- Normalize line endings to `\n`
- Trim trailing blank lines
- Stderr is never compared — only stdout

### Layer 2: Format-specific

- **JSON export:** Remove `title` key from each entry object, sort remaining keys.
  jrnl includes `title`; jrnl-md does not.
- **ANSI tests:** No ANSI stripping — raw bytes compared. All other tests strip ANSI
  escape sequences before comparison.
- **`--short`:** Replace text portion after timestamp with placeholder. jrnl shows
  title (first sentence); jrnl-md shows truncated body. Date/time column format and
  alignment are verified.

### Layer 3: Per-test overrides

Each table row can declare an optional `normalize func(string) string`. Each use must
have a comment explaining the known difference and referencing `docs/jrnl-compat.md`.

## Test Runner

```go
var updateGolden = flag.Bool("update-golden", false, "re-capture golden files from jrnl")

func TestGolden(t *testing.T) {
    for _, tt := range goldenTests {
        t.Run(tt.slug, func(t *testing.T) {
            if *updateGolden {
                // run jrnl oracle → write golden file
            } else {
                // read golden file (skip if missing)
            }
            // run jrnl-md → capture stdout
            // apply normalizations
            // compare → unified diff on failure
        })
    }
}
```

## Comparison Failure Output

On mismatch, the test prints:
- The golden file path
- A unified diff between expected (golden) and actual (jrnl-md)
- The raw actual output for inspection

## jrnlOracle Struct

```go
type jrnlOracle struct {
    configPath string
    journalPath string
}

func (o jrnlOracle) run(args ...string) (string, error) {
    // calls: jrnl --config-file <configPath> <args...>
    // returns stdout
}
```

Seeding functions:
- `seedGoldenJournal(t) (testEnv, jrnlOracle)` — standard 6-entry seed for both
- `seedWithConfig(t, configOverrides) (testEnv, jrnlOracle)` — config variation tests
- `seedMultiJournal(t) (testEnv, jrnlOracle)` — multi-journal tests

## Dependencies

- `jrnl` (pip/pipx) required only for `-update-golden`
- Normal test runs: zero external dependencies
- Golden files committed to git — diffs are reviewable in PRs
