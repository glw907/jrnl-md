# jrnl Compatibility

jrnl-md is a markdown-native reimplementation of [jrnl](https://jrnl.sh). This document describes what is compatible, what differs, and what is not implemented.

## Backend

jrnl supports multiple storage backends: DayOne, encrypted single-file, and folder-of-files. **jrnl-md implements only the folder-based markdown backend.** There is no `type` config key and no plan to add one.

Journal files are stored as one Markdown file per day at `YYYY/MM/DD.md` within the configured journal directory. If you are migrating from a jrnl DayOne or single-file journal, export to plain text first, then import into jrnl-md with `--import`.

## File Format

jrnl stores entries in plain text with one entry per line:

```
2025-01-15 09:00 Entry body here.
2025-01-15 14:30 Another entry.
```

jrnl-md stores entries as Markdown with day-level and entry-level headings:

```markdown
# 2025-01-15 Wednesday

## [09:00 AM]

Entry body here.

## [02:30 PM]

Another entry.
```

## Encryption

jrnl encrypts using AES (legacy) or GPG. jrnl-md uses [age](https://age-encryption.org) (scrypt + XChaCha20-Poly1305). Encrypted jrnl journals cannot be read by jrnl-md directly.

## Config File

| | jrnl | jrnl-md |
|---|---|---|
| Format | YAML | TOML |
| Location | `~/.config/jrnl/jrnl.yaml` | `~/.config/jrnl-md/config.toml` |
| Journal path key | `journals.<name>.journal` | `journals.<name>.path` |
| Journal type key | `journals.<name>.type` | *(not supported)* |
| Config file flag | `--config-file` | `--config-file` (alias: `--config`) *(Pass 2)* |

## Entry Titles

jrnl treats the first sentence of an entry body (text up to the first `.`, `!`, or `?`) as a distinct **title** field. This title appears in `--short` output and in export formats (e.g. the `title` key in JSON export).

jrnl-md has no title concept. There is no structural distinction between the first sentence and the rest of the body — the entire body is stored and displayed as-is. `--short` output shows the date/time followed by up to 60 characters of body text (truncated with `...`).

If you rely on jrnl's title field in exports or integrations, note that jrnl-md's JSON export has no `title` key.

## Feature Compatibility

### Implemented

| Feature | jrnl | jrnl-md |
|---|---|---|
| Write inline entry | `jrnl Entry text` | same |
| Write from stdin | `echo "text" \| jrnl` | same *(Pass 2)* |
| Date-prefixed entry | `jrnl yesterday: text` | same *(Pass 2)* |
| Last N entries | `--n N` / `-N` | same |
| Short listing | `--short` / `-s` | same |
| Starred entries | `--starred` | same |
| Text search | `--contains text` | same |
| Date range | `--from`, `--to`, `--on` | same |
| Tag filter | `jrnl @tag` | same |
| AND tag filter | `--and` | same |
| Exclude tag | `--not @tag` | same |
| Exclude starred | `--not-starred` | same |
| Exclude tagged | `--not-tagged` | same |
| List tags | `--tags` | same (frequency-sorted) |
| Edit (no filter) | `--edit` | opens all/last-N entries *(Pass 3)* |
| Edit with filters | `--edit @tag` | same *(Pass 3)* |
| Delete entries | `--delete` | same |
| Change entry time | `--change-time` | same |
| Encrypt journal | `--encrypt` / `--decrypt` | same (age, not GPG) |
| List journals | `--list` | same |
| Export formats | `--export`/`--format` json/md/txt/xml/yaml | same |
| Export to file | `--file path` | same |
| Import | `--import file` | same *(Pass 4)* |
| Multiple journals | `jrnl work: text` | same |
| Editor from env | `$VISUAL` / `$EDITOR` fallback | same |
| Per-journal config | editor, template, tag_symbols | same *(Pass 5)* |
| `default_hour` / `default_minute` | config keys | same *(Pass 5)* |
| Tag highlighting | `highlight`, `colors.tags` | same |
| Line wrapping | `linewrap` | same |
| Templates | `template` | same |
| Shell completion | `--completion` | `jrnl-md completion {bash,zsh,fish,powershell}` (subcommand) |

### Not Implemented

| Feature | Notes |
|---|---|
| DayOne backend | Folder-only; no plans to add |
| Single-file journal | Folder-only; no plans to add |
| `--export dayone` | By design — requires DayOne backend |
| GPG encryption | Uses age instead |
| `--config-override key=value` | Not implemented |
| `--debug` flag | Not implemented |
| `-today-in-history` filter | Shows entries from today's date in prior years; not implemented |
| `-month DATE` filter | Shows entries in a given month across years; not implemented |
| `-day DATE` filter | Shows entries on a day-of-month across months; not implemented |
| `-year DATE` filter | Shows entries in a specific year; not implemented |
| `-tagged` positive filter | Shows only entries with any tag; not implemented (use `--not-tagged` to exclude) |
| `--format fancy` / `boxed` | Unicode box-drawing display format; not implemented |
| `--diagnostic` flag | Prints debug info for bug reports; not implemented |
| `display_format` config key | Sets default output format; not implemented (always `pretty`) |
| `timeformat` config key | Custom timestamp storage format; not implemented (uses fixed Go layouts) |
| `--import --format TYPE` | Specifies import source format; jrnl-md always expects its own markdown format |
| YAML directory export | `--file dir/` with `--format yaml` exports one file per entry; not implemented |

## Config Key Reference

See [docs/config.md](config.md) for the full jrnl-md config reference.

## Compatibility Test Suite

Every row in the Feature Compatibility table above has a corresponding end-to-end test in
`e2e/jrnl_compat_test.go`. Tests for implemented features assert correct behavior; tests for
pending features call `t.Skip` with the pass number.

Run the suite:

```bash
go test ./e2e/... -run TestCompat -v
```

Each implementation pass must update this file to convert pending tests to real assertions.
