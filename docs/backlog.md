# jrnl-md Backlog

Items worth implementing, roughly prioritized. Items explicitly excluded by design are in `docs/jrnl-compat.md`.

---

## `--format` display mode aliases

**Priority: High** — users migrating from jrnl will hit these immediately.

jrnl treats `--format` as a unified flag covering both display modes and structured exports. jrnl-md only handles structured exports (`json`, `md`, `txt`, `xml`, `yaml`). Passing a display mode alias currently returns an error.

| `--format` value | Expected behavior | Current behavior |
|---|---|---|
| `pretty` | default display | `unknown export format "pretty"` |
| `short` | one line per entry (same as `--short`) | `unknown export format "short"` |
| `tags` | tag frequency list (same as `--tags`) | `unknown export format "tags"` |

**Fix:** Add `pretty`, `short`, `tags` cases to the `switch` in `cmd/jrnl-md/read.go`. `pretty` and `short` route to the existing display logic; `tags` routes to `showTags`.

---

## `--format dates`

**Priority: Medium** — no workaround exists; jrnl users expect this.

`jrnl --format dates` lists the dates of matching entries with their entry count:

```
2025-01-15: 2 entries
2025-03-01: 1 entry
```

Not currently implemented. No flag equivalent.

**Fix:** Add a `dates` case to the export switch in `read.go`, with a small helper that groups entries by date and prints counts.

---

## `--config-override key=value`

**Priority: Low** — power-user escape hatch; not commonly needed.

jrnl supports `--config-override key=value` to override individual config keys at the command line without editing the config file. Useful for one-off scripted invocations.

---

## `--debug` flag

**Priority: Low** — diagnostic tooling only.

jrnl's `--debug` flag enables verbose output for troubleshooting. No functional gap for normal use.

---

## Docs Pass

**Priority: Low** — polish only.

- `README.md` — project overview, install, quickstart
- `docs/config.md` — full config key reference (stub currently referenced from `docs/jrnl-compat.md`)
