# jrnl-md Backlog

Items worth implementing, roughly prioritized. Items explicitly excluded by design are in `docs/jrnl-compat.md`.

---

~~## `--format` display mode aliases~~
~~## `--format dates`~~

**Done** — `--format pretty/short/tags/dates` all implemented.

---

## Compat test suite audit

**Priority: High** — we know the suite has gaps (e.g. `--format` display aliases were implemented but untested until caught manually).

Do a thorough review of canonical jrnl functionality against `e2e/jrnl_compat_test.go`:

1. Cross-reference every feature in `docs/jrnl-compat.md` against the test suite — confirm each has a real assertion, not just a pass-through.
2. Check jrnl's full feature set (flags, config keys, output formats, edge cases) against what we test — the compat table itself may have omissions.
3. Add `TestCompat_*` tests for anything missing. Update `docs/jrnl-compat.md` if new rows are needed.

Known gaps to start with: `--format pretty/short/tags/dates`, `--export xml`, `--export yaml`, `linewrap`, `indent_character`, starred-entry write syntax (`jrnl * text`), `-N` shorthand, encrypt/decrypt.

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
