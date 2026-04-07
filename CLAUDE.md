# jrnl-md Project Instructions

## Goal

jrnl-md is a markdown journaling CLI inspired by [jrnl](https://jrnl.sh). It manages a
directory of markdown day files — one file per calendar day. The unix ecosystem handles
everything else.

## Design Principle: The Day Is the Atom

One markdown file per calendar day (`YYYY/MM/YYYY-MM-DD.md`). All operations work at the day level.
There is no per-entry structure within a day file.

- Load only the day files an operation needs. Never load the entire journal for a
  single-day operation.
- The directory structure (`YYYY/MM/`) is the primary index. Use it to skip irrelevant
  files rather than loading everything and filtering in memory.
- `edit` opens exactly one day file. No multi-day serialization.
- `write` always appends to today's file.

## CLI Structure

Subcommand-based CLI using Cobra:

- `jrnl-md write <text>` — append to today
- `jrnl-md edit [--on <date>]` — open one day file in editor
- `jrnl-md list [flags] [@tag...]` — display matching days
- `jrnl-md tags [flags]` — tag frequencies
- `jrnl-md completion` — shell completion

No bare-invocation magic. No subcommand = help.

## Implementation Passes

| Pass | Status | Focus |
|------|--------|-------|
| Pass 1: Infrastructure | Done | Archive 1.x, clear main, CLAUDE.md, go.mod, plans |
| Pass 2: Core Build | Done | All packages, CLI, unit tests, e2e tests |
| Pass 3: Docs + Polish | Done | README.md, docs/config.md, --help text |

## Key Design Decisions

- **No per-entry concept**: no Entry type, no starred, no per-entry tags, no delete
- **No encryption**: removed in 2.0
- **No export formats**: files are already markdown; unix tools handle the rest
- **No import**: removed in 2.0
- **Timestamps optional**: `timestamps` config key (default true) controls `## time` headings
- **Config**: TOML at `~/.config/jrnl-md/config.toml`

See `docs/superpowers/specs/2026-04-06-jrnl-md-2.0-design.md` for the full design spec.
