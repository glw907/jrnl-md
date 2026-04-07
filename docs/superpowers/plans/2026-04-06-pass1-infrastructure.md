# jrnl-md 2.0 Pass 1: Infrastructure — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Archive the 1.x codebase, clear main for a clean-sheet 2.0 build, rewrite CLAUDE.md, set up the new go.mod, and document the three-pass implementation plan.

**Architecture:** Sequential git and file operations — tag, branch, delete, write new docs. No application code is written in this pass.

**Tech Stack:** bash, git, gh

---

## File Map

**New files:**
- `docs/superpowers/plans/2026-04-06-pass2-core-build.md` — Pass 2 implementation plan
- `docs/superpowers/plans/2026-04-06-pass3-docs-polish.md` — Pass 3 implementation plan

**Modified files:**
- `CLAUDE.md` — rewrite for 2.0
- `go.mod` — new module file with 2.0 dependencies only
- `go.sum` — regenerated
- `Makefile` — updated version, simplified
- `.gitignore` — unchanged (already correct)
- `.golangci.yml` — unchanged (already correct)

**Deleted files:**
- `cmd/` — all 1.x CLI code
- `internal/` — all 1.x packages
- `e2e/` — all 1.x tests (including testdata/)
- `jrnl-md` — compiled binary
- `BACKLOG.md`
- `docs/jrnl-compat.md`
- `docs/superpowers/plans/2026-03-30-*.md` — all 1.x plans
- `docs/superpowers/plans/2026-03-31-*.md` — all 1.x plans
- `docs/superpowers/plans/2026-04-01-*.md` — all 1.x plans
- `docs/superpowers/specs/2026-03-30-*.md` — all 1.x specs
- `docs/superpowers/specs/2026-03-31-*.md` — all 1.x specs
- `docs/superpowers/specs/2026-04-01-*.md` — all 1.x specs

**Kept as-is:**
- `.gitignore`
- `.golangci.yml`
- `docs/superpowers/specs/2026-04-06-jrnl-md-2.0-design.md`

---

### Task 1: Tag and branch 1.x

Preserve the full 1.x history on a tag and branch before deleting anything.

**Files:** none (git metadata only)

- [ ] **Step 1: Tag current HEAD as v1.0.0**

```bash
git tag v1.0.0
```

- [ ] **Step 2: Create v1 branch from the tag**

```bash
git branch v1 v1.0.0
```

- [ ] **Step 3: Push tag and branch to origin**

```bash
git push origin v1.0.0
git push origin v1
```

- [ ] **Step 4: Verify**

```bash
git tag -l
git branch -a | grep v1
```

Expected: `v1.0.0` tag listed, `v1` branch listed locally and on origin.

---

### Task 2: Remove all 1.x source code

Delete all application code, tests, and the compiled binary from main.

**Files:**
- Delete: `cmd/`, `internal/`, `e2e/`, `jrnl-md`

- [ ] **Step 1: Remove source directories**

```bash
git rm -rf cmd/ internal/ e2e/
```

- [ ] **Step 2: Remove compiled binary**

```bash
rm -f jrnl-md
```

(Binary is gitignored, so `rm` not `git rm`.)

- [ ] **Step 3: Remove old go.mod and go.sum**

```bash
git rm go.mod go.sum
```

- [ ] **Step 4: Verify**

```bash
git status
```

Expected: staged deletions for cmd/, internal/, e2e/, go.mod, go.sum. Binary gone.

---

### Task 3: Remove 1.x docs

Delete docs that reference the jrnl-workalike framing, the backlog, and all 1.x specs and plans. Keep the 2.0 design spec.

**Files:**
- Delete: `BACKLOG.md`, `docs/jrnl-compat.md`
- Delete: `docs/superpowers/plans/2026-03-30-*.md`
- Delete: `docs/superpowers/plans/2026-03-31-*.md`
- Delete: `docs/superpowers/plans/2026-04-01-*.md`
- Delete: `docs/superpowers/specs/2026-03-30-*.md`
- Delete: `docs/superpowers/specs/2026-03-31-*.md`
- Delete: `docs/superpowers/specs/2026-04-01-*.md`

- [ ] **Step 1: Remove top-level docs**

```bash
git rm BACKLOG.md docs/jrnl-compat.md
```

- [ ] **Step 2: Remove 1.x plans**

```bash
git rm docs/superpowers/plans/2026-03-30-*.md
git rm docs/superpowers/plans/2026-03-31-*.md
git rm docs/superpowers/plans/2026-04-01-*.md
```

- [ ] **Step 3: Remove 1.x specs**

```bash
git rm docs/superpowers/specs/2026-03-30-*.md
git rm docs/superpowers/specs/2026-03-31-*.md
git rm docs/superpowers/specs/2026-04-01-*.md
```

- [ ] **Step 4: Verify only 2.0 spec remains**

```bash
ls docs/superpowers/specs/
ls docs/superpowers/plans/
```

Expected specs: `2026-04-06-jrnl-md-2.0-design.md` only.
Expected plans: `2026-04-06-pass1-infrastructure.md` only (this file).

---

### Task 4: Rewrite CLAUDE.md

Replace the jrnl-workalike CLAUDE.md with 2.0 project instructions.

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Write new CLAUDE.md**

Use the Write tool to replace `CLAUDE.md` with:

```markdown
# jrnl-md Project Instructions

## Goal

jrnl-md is a markdown journaling CLI inspired by [jrnl](https://jrnl.sh). It manages a
directory of markdown day files — one file per calendar day. The unix ecosystem handles
everything else.

## Design Principle: The Day Is the Atom

One markdown file per calendar day (`YYYY/MM/DD.md`). All operations work at the day level.
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
| Pass 2: Core Build | Pending | All packages, CLI, unit tests, e2e tests |
| Pass 3: Docs + Polish | Pending | README.md, docs/config.md, --help text |

## Key Design Decisions

- **No per-entry concept**: no Entry type, no starred, no per-entry tags, no delete
- **No encryption**: removed in 2.0
- **No export formats**: files are already markdown; unix tools handle the rest
- **No import**: removed in 2.0
- **Timestamps optional**: `timestamps` config key (default true) controls `## time` headings
- **Config**: TOML at `~/.config/jrnl-md/config.toml`

See `docs/superpowers/specs/2026-04-06-jrnl-md-2.0-design.md` for the full design spec.
```

- [ ] **Step 2: Verify**

```bash
head -5 CLAUDE.md
```

Expected: starts with `# jrnl-md Project Instructions`.

---

### Task 5: Create new go.mod and Makefile

Initialize a fresh go module with only the dependencies needed for 2.0. Update Makefile version.

**Files:**
- Create: `go.mod`
- Modify: `Makefile`

- [ ] **Step 1: Write go.mod**

Use the Write tool to create `go.mod` with:

```
module github.com/glw907/jrnl-md

go 1.25.0

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/fatih/color v1.19.0
	github.com/olebedev/when v1.1.0
	github.com/spf13/cobra v1.10.2
	golang.org/x/term v0.41.0
)
```

Note: `go get` won't work with no Go source files. Write go.mod directly.
The go.sum will be generated when Pass 2 adds source code and runs
`go mod tidy`. Do not create go.sum in this pass.

- [ ] **Step 2: Update Makefile version**

Use the Edit tool to change `VERSION := 0.1.0` to `VERSION := 2.0.0` in `Makefile`.

- [ ] **Step 3: Verify**

```bash
cat go.mod
grep VERSION Makefile
```

Expected: go.mod has the 5 direct dependencies listed. Makefile shows `VERSION := 2.0.0`.

---

### Task 6: Update project memory

Update Claude memory to reflect the 2.0 project state.

**Files:**
- Modify: `/home/glw907/.claude/projects/-home-glw907-Projects-jrnl-md/memory/MEMORY.md`
- Modify: `/home/glw907/.claude/projects/-home-glw907-Projects-jrnl-md/memory/project_pass_state.md`
- Delete: `/home/glw907/.claude/projects/-home-glw907-Projects-jrnl-md/memory/feedback_jrnl_compat.md`
- Delete: `/home/glw907/.claude/projects/-home-glw907-Projects-jrnl-md/memory/project_journal_backend.md`

- [ ] **Step 1: Update project_pass_state.md**

Use the Write tool:

```markdown
---
name: 2.0 implementation pass state
description: Current status of jrnl-md 2.0 implementation passes
type: project
---

jrnl-md 2.0 is a clean-sheet rebuild (not a refactor of 1.x). Three passes:

- Pass 1: Infrastructure — Done. 1.x archived as v1.0.0 tag + v1 branch. main cleared.
- Pass 2: Core Build — Pending. All packages, CLI subcommands, unit tests, e2e tests.
- Pass 3: Docs + Polish — Pending. README.md, docs/config.md, Cobra help text.

**Why:** 2.0 moves from "jrnl workalike" to "inspired by jrnl." Day-file-first, subcommand CLI, no starred/encryption/export/import.

**How to apply:** Design spec is at docs/superpowers/specs/2026-04-06-jrnl-md-2.0-design.md. Plans at docs/superpowers/plans/2026-04-06-pass*.md.
```

- [ ] **Step 2: Delete obsolete memory files**

```bash
rm /home/glw907/.claude/projects/-home-glw907-Projects-jrnl-md/memory/feedback_jrnl_compat.md
rm /home/glw907/.claude/projects/-home-glw907-Projects-jrnl-md/memory/project_journal_backend.md
```

These reference the jrnl-workalike framing which no longer applies.

- [ ] **Step 3: Rewrite MEMORY.md**

Use the Write tool:

```markdown
- [2.0 pass state](project_pass_state.md) — done/pending passes for the 2.0 clean-sheet rebuild
- [Skip worktree setup](feedback_worktree.md) — work directly on current branch; no worktrees needed on this solo project
```

- [ ] **Step 4: Verify**

```bash
ls /home/glw907/.claude/projects/-home-glw907-Projects-jrnl-md/memory/
cat /home/glw907/.claude/projects/-home-glw907-Projects-jrnl-md/memory/MEMORY.md
```

Expected: 3 files (MEMORY.md, project_pass_state.md, feedback_worktree.md). MEMORY.md shows 2 entries.

---

### Task 7: Write Pass 2 and Pass 3 plans

Write implementation plans for the remaining passes so subsequent sessions can pick them up. These are detailed plans following the same format as this file.

**Files:**
- Create: `docs/superpowers/plans/2026-04-06-pass2-core-build.md`
- Create: `docs/superpowers/plans/2026-04-06-pass3-docs-polish.md`

- [ ] **Step 1: Write Pass 2 plan**

This is the largest plan. It must cover all packages and subcommands from the design spec. Use the writing-plans skill to write it, referencing the design spec at `docs/superpowers/specs/2026-04-06-jrnl-md-2.0-design.md` for all types, methods, config keys, CLI flags, display formats, and test cases.

The plan should follow TDD: write failing tests first, then implement. Tasks should be ordered by dependency:

1. `internal/config` — Config struct, load/save, defaults, auto-create (no dependencies)
2. `internal/dateparse` — date parsing (no dependencies)
3. `internal/journal` — Day type, Store, Filter (depends on config for tag symbols)
4. `internal/display` — wrapping, truncation, highlighting, short format (depends on journal.Day)
5. `internal/editor` — launch editor, cursor positioning (depends on config for editor)
6. `cmd/jrnl-md` — main + all subcommands (depends on all internal packages)
7. E2E tests (depends on working binary)

- [ ] **Step 2: Write Pass 3 plan**

Covers documentation and polish:

1. README.md — what/install/quickstart/reference/philosophy
2. docs/config.md — full config reference
3. Cobra help text — Use, Short, Long, Example for each subcommand
4. Any issues found using the tool after Pass 2
5. Update CLAUDE.md pass table to mark Pass 2 as Done

---

### Task 8: Commit everything

- [ ] **Step 1: Stage all changes**

```bash
git add CLAUDE.md Makefile go.mod
git add docs/superpowers/plans/2026-04-06-pass1-infrastructure.md
git add docs/superpowers/plans/2026-04-06-pass2-core-build.md
git add docs/superpowers/plans/2026-04-06-pass3-docs-polish.md
git status
```

Review the staged changes. Expected:
- Deleted: cmd/, internal/, e2e/, go.mod (old), go.sum (old), BACKLOG.md, docs/jrnl-compat.md, all 1.x plans and specs
- Modified: CLAUDE.md, Makefile
- Added: go.mod (new), three plan files

- [ ] **Step 2: Commit**

```bash
git commit -m "$(cat <<'EOF'
Archive 1.x, clear main for 2.0 clean-sheet rebuild

- Tag v1.0.0 and branch v1 to preserve 1.x history
- Remove all 1.x source (cmd/, internal/, e2e/)
- Remove 1.x docs (jrnl-compat.md, BACKLOG.md, 1.x specs/plans)
- Rewrite CLAUDE.md for 2.0 design philosophy
- New go.mod with 2.0 dependencies only
- Add implementation plans for all three passes

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 3: Verify clean state**

```bash
git status
git log --oneline -3
```

Expected: clean working tree. Top commit is the archive/reset.

- [ ] **Step 4: Push to origin**

```bash
git push origin main
```
