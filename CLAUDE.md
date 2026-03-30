# jrnl-md Project Instructions

## Goal

jrnl-md is a **100% workalike reimplementation of [jrnl](https://jrnl.sh)**, using markdown
as the storage format. Every jrnl CLI behavior must work identically **except** for the
documented exceptions in `docs/jrnl-compat.md`.

When in doubt about intended behavior, consult `jrnl --help` or the jrnl documentation at
https://jrnl.sh/en/stable/. The behavior there is the correct behavior for jrnl-md.

## Compatibility Rule

Before adding, changing, or removing any CLI flag, output format, or behavior:

1. Check `docs/jrnl-compat.md`. If jrnl has an opinion, match it.
2. If you're deviating, the deviation must be listed in `docs/jrnl-compat.md` as a documented
   exception with a reason.
3. Do not add flags, output formats, or behaviors that jrnl does not have, unless they are
   specifically required by the jrnl-md markdown backend (e.g. `--config` alias).

## Compat Suite Rule

`e2e/jrnl_compat_test.go` contains one `TestCompat_*` test per documented jrnl feature.

**Every implementation pass must include a task to update this file:**
- Convert `t.Skip("pending Pass N: ...")` to a real test for each feature the pass implements.
- Add new `TestCompat_*` tests if the pass adds new features that are jrnl-compatible.

After each pass, `go test ./e2e/... -run TestCompat -v` must show:
- All implemented features: PASS
- All pending features: SKIP (with the correct pending pass note)
- No FAIL results

## Implementation Passes

| Pass | Status | Focus |
|------|--------|-------|
| Compat Suite | Done | Established e2e/jrnl_compat_test.go and this CLAUDE.md |
| Pass 1: CLI & Display Polish | Done | Filter flags, tag highlighting, --format/--file, --tags sort |
| Pass 2: Date & Time | Pending | Date-prefixed entries, default_hour/default_minute, stdin write, --config-file |
| Pass 3: --edit with Filters | Pending | --edit always via editFiltered, filtered and unfiltered |
| Pass 4: Import | Pending | --import FILE |
| Pass 5: Per-journal Config | Pending | Per-journal config overrides, templates |
| Docs Pass | Pending | README.md, docs/config.md polish |

## Documented Exceptions

See `docs/jrnl-compat.md` for the full list. Key exceptions:

- **Backend**: folder-based markdown only (no DayOne, no single-file)
- **Encryption**: age (not GPG)
- **Config format**: TOML (not YAML)
- **Entry titles**: no title concept — full body is stored/displayed as-is
- **DayOne export**: not supported by design
