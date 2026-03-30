# jrnl-md Backlog

Items worth implementing, roughly prioritized. Items explicitly excluded by design are in `docs/jrnl-compat.md`.

---

~~## `--format` display mode aliases~~
~~## `--format dates`~~

**Done** — `--format pretty/short/tags/dates` all implemented.

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
