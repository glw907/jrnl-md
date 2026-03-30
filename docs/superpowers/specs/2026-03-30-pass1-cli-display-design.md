---
name: Pass 1 — CLI & Display Polish
description: --format alias, --file export, tags sorted by frequency, tag highlighting, --and AND logic, --not exclusion
type: project
---

# Pass 1: CLI & Display Polish

## Goal

Close the gap between jrnl-md and jrnl for CLI flags and display output.

## Scope

1. `--format` as alias for `--export` (jrnl uses `--format`)
2. `--file` flag to write export output to a file instead of stdout
3. `--tags` output sorted by frequency descending
4. Tag highlighting in body output (respects `colors.tags` and `highlight` config)
5. `--and` flag: require ALL specified tags (default is OR)
6. `--not` flag: exclude entries containing specified tags
7. `--not-starred`: exclude starred entries
8. `--not-tagged`: exclude entries that have any tags

---

## 1. `--format` alias and `--file` flag

**`cmd/jrnl-md/root.go`**

Add two fields to the `flags` struct:

```go
format string
file   string
```

In `newRootCmd`, register:

```go
cmd.Flags().StringVar(&f.format, "format", "", "Export format (alias for --export)")
cmd.Flags().StringVar(&f.file, "file", "", "Write export output to file")
```

At the top of `runRoot`, before calling `readEntries`, coalesce:

```go
if f.format != "" && f.export == "" {
    f.export = f.format
}
```

**`cmd/jrnl-md/read.go`**

In `exportEntries`, change the final write from `fmt.Fprint(os.Stdout, output)` to:

```go
if f.file != "" {
    if err := atomicfile.WriteFile(f.file, []byte(output)); err != nil {
        return fmt.Errorf("writing export to %s: %w", f.file, err)
    }
    return nil
}
fmt.Fprint(os.Stdout, output)
return nil
```

---

## 2. `--tags` sorted by frequency

**`cmd/jrnl-md/read.go`** — `listTags` function

Current: iterates entries, builds `map[string]int`, prints in map order.

Change: after building the map, collect into a `[]struct{ tag string; count int }` slice, sort by count descending (secondary: tag name ascending for stability), then print.

```go
type tagCount struct {
    tag   string
    count int
}
counts := make([]tagCount, 0, len(freq))
for tag, n := range freq {
    counts = append(counts, tagCount{tag, n})
}
sort.Slice(counts, func(i, j int) bool {
    if counts[i].count != counts[j].count {
        return counts[i].count > counts[j].count
    }
    return counts[i].tag < counts[j].tag
})
for _, tc := range counts {
    fmt.Printf("@%s (%d)\n", tc.tag, tc.count)
}
```

---

## 3. Tag highlighting

**`internal/display/display.go`**

Add a new exported function:

```go
// HighlightTags replaces tag occurrences in body with colorFn-wrapped versions.
// tagSymbols is the set of tag prefix characters (e.g. "@"). colorFn wraps a
// string with ANSI color. If colorFn is nil, body is returned unchanged.
func HighlightTags(body, tagSymbols string, colorFn func(a ...any) string) string
```

Implementation:
- If `colorFn == nil` return body unchanged.
- Build a regex from `tagSymbols`: match `[<symbols>]\w+` (whole-word boundary not needed; tags are delimited by whitespace or punctuation per jrnl convention).
- Use `regexp.ReplaceAllStringFunc` to wrap each match with `colorFn(match)`.

**`cmd/jrnl-md/read.go`** — `printEntries` (or wherever body is rendered)

After formatting body text, apply highlighting:

```go
colorFn := display.ColorFunc(cfg.Colors.Tags) // nil if "none"
if cfg.General.Highlight && colorFn == nil {
    colorFn = display.ColorFunc("cyan") // default when highlight=true but tags="none"
}
body = display.HighlightTags(body, cfg.Format.TagSymbols, colorFn)
```

**`internal/display/display.go`** — `ColorFunc`

Current behavior: returns `nil` for `"none"`. No change needed to ColorFunc itself.

---

## 4. `--and` AND tag logic

**`internal/journal/filter.go`**

Add field to `Filter`:

```go
AndTags bool // if true, entry must match ALL Tags (default: any)
```

In `matches()`, the existing tag loop does OR logic. Add a branch:

```go
if len(f.Tags) > 0 {
    if f.AndTags {
        for _, tag := range f.Tags {
            found := false
            for _, et := range entry.Tags {
                if et == tag { found = true; break }
            }
            if !found { return false }
        }
    } else {
        // existing OR logic
    }
}
```

**`cmd/jrnl-md/root.go`**

Add to `flags`:

```go
and bool
```

Register:

```go
cmd.Flags().BoolVar(&f.and, "and", false, "Require all tags (AND logic)")
```

**`cmd/jrnl-md/read.go`** — `buildFilter`

Set `AndTags: f.and`.

---

## 5. `--not`, `--not-starred`, `--not-tagged`

**`internal/journal/filter.go`**

Add fields:

```go
NotTags      []string // exclude entries containing any of these tags
NotStarred   bool     // exclude starred entries
NotTagged    bool     // exclude entries that have any tags
```

In `matches()`:

```go
// NotTags exclusion
for _, tag := range f.NotTags {
    for _, et := range entry.Tags {
        if et == tag { return false }
    }
}
// NotStarred exclusion
if f.NotStarred && entry.Starred { return false }
// NotTagged exclusion
if f.NotTagged && len(entry.Tags) > 0 { return false }
```

**`cmd/jrnl-md/root.go`**

Add to `flags`:

```go
not        []string
notStarred bool
notTagged  bool
```

Register:

```go
cmd.Flags().StringArrayVar(&f.not, "not", nil, "Exclude entries with tag")
cmd.Flags().BoolVar(&f.notStarred, "not-starred", false, "Exclude starred entries")
cmd.Flags().BoolVar(&f.notTagged, "not-tagged", false, "Exclude tagged entries")
```

Add `f.notStarred` and `f.notTagged` to `hasFilterFlags`.

**`cmd/jrnl-md/read.go`** — `buildFilter`

Set `NotTags: f.not`, `NotStarred: f.notStarred`, `NotTagged: f.notTagged`.

---

## Testing

- `filter_test.go`: table tests for AndTags, NotTags, NotStarred, NotTagged
- `display_test.go`: HighlightTags with multi-symbol tagSymbols, nil colorFn, no-op when no tags
- `read_test.go` (or e2e): --tags output order; --file writes to disk
- `args_test.go`: no changes needed

## Files touched

| File | Change |
|------|--------|
| `cmd/jrnl-md/root.go` | Add format, file, and, not, not-starred, not-tagged flags; coalesce format→export |
| `cmd/jrnl-md/read.go` | --file export target; tag-sorted listTags; apply HighlightTags; buildFilter additions |
| `internal/journal/filter.go` | AndTags, NotTags, NotStarred, NotTagged fields + matches logic |
| `internal/display/display.go` | HighlightTags function |
