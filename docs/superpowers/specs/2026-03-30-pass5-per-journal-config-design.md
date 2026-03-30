---
name: Pass 5 — Per-Journal Config
description: Per-journal editor, template, and tag_symbols overrides in config
type: project
---

# Pass 5: Per-Journal Config

## Goal

Allow per-journal overrides for `editor`, `template`, and `tag_symbols`. When a journal config specifies these, they take precedence over the global `[general]` values. Matches jrnl's per-journal config model.

---

## Config changes

**`internal/config/config.go`** — `JournalConfig`:

```go
type JournalConfig struct {
    Path       string  `toml:"path"`
    Encrypt    *bool   `toml:"encrypt,omitempty"`
    Editor     string  `toml:"editor,omitempty"`
    Template   string  `toml:"template,omitempty"`
    TagSymbols string  `toml:"tag_symbols,omitempty"`
}
```

All three fields are optional (empty string = use global default).

Example TOML:

```toml
[journals.work]
path = "~/Documents/Work/"
editor = "vim"
tag_symbols = "@#"
```

---

## Resolution logic

Add a helper in `config.go`:

```go
// ResolvedJournalConfig returns cfg merged with journal-specific overrides.
func ResolvedJournalConfig(global Config, j JournalConfig) Config
```

Returns a copy of `global` with journal-specific fields applied where non-empty:
- `j.Editor != ""` → `result.General.Editor = j.Editor`
- `j.Template != ""` → `result.General.Template = j.Template`
- `j.TagSymbols != ""` → `result.Format.TagSymbols = j.TagSymbols`

This returns a fully resolved `Config` so the rest of the code (which already uses `cfg.General.Editor`, `cfg.General.Template`, `cfg.Format.TagSymbols`) requires no further changes.

---

## `cmd/jrnl-md/root.go`

After resolving `journalCfg`, apply the merge:

```go
cfg = config.ResolvedJournalConfig(cfg, journalCfg)
```

This single line is all that's needed in `runRoot`. All downstream calls already use `cfg`.

---

## Testing

- `config_test.go`: ResolvedJournalConfig — override all three fields; override some fields; override none (global used); empty journal config leaves global unchanged
- Integration/e2e: journal with `editor` override uses that editor when `--edit` is invoked

---

## Files touched

| File | Change |
|------|--------|
| `internal/config/config.go` | Add Editor, Template, TagSymbols to JournalConfig; add ResolvedJournalConfig |
| `cmd/jrnl-md/root.go` | Call ResolvedJournalConfig after loading journal config |
