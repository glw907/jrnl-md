# Git Sync, Search Index & Journal Manifest

## Overview

Built-in git sync for jrnl-md journals, with search index and manifest generation
designed to support a future web app for browsing and searching journal history
(including encrypted journals).

jrnl-md's day-file storage model (`YYYY/MM/DD.md`) makes git a natural sync
backend: two devices writing different days never conflict, and same-day conflicts
are rare and text-mergeable. This feature makes git sync a first-class operation
rather than leaving users to wire up shell scripts.

## Motivation

jrnl users have requested git sync since 2016 (jrnl-org/jrnl#412, #921, #1444).
The original jrnl can't deliver it cleanly because its single encrypted file
produces unresolvable binary merge conflicts. jrnl-md's architecture eliminates
this problem.

The sync layer also serves as foundation for a future GitHub-hosted web app that
lets users browse and search their journal history — including encrypted journals
via client-side decryption with age-js.

## Config

New per-journal keys in `config.toml`:

```toml
[journals.default]
path = "~/Documents/Journal/"
sync = true              # enable auto-sync (default: false)
sync_remote = "origin"   # git remote name (default: "origin")
sync_branch = "main"     # branch to sync (default: "main")
```

Sync is per-journal. Some journals may sync while others stay local.

## CLI Surface

### `jrnl-md sync`

Manual sync trigger. Commits any uncommitted changes, pulls, pushes.

Works even if `sync = false` in config (manual-only mode).

**`--init` flag** for first-time setup:

1. `git init` in the journal directory (if not already a repo)
2. Create `.gitignore` (exclude `*.tmp`, `*.lock`)
3. If no remote exists, prompt for remote URL and run `git remote add origin <url>`
4. Generate initial index and manifest
5. Initial commit of all existing files
6. Push to remote

### `jrnl-md index`

Manual index and manifest rebuild.

- Regenerates `search-index.json(.age)` and `journal.json`
- Useful after editing files outside jrnl-md or after migration
- Prompts for passphrase if journal is encrypted

### Auto-sync

When `sync = true`, fires automatically after write operations:

- Writing a new entry (inline or editor)
- `--edit`
- `--delete`
- `--change-time`
- `--import`
- `--encrypt` / `--decrypt`

Does NOT fire after read-only operations (`--short`, `--tags`, `--export`, etc.).

## Sync Operation

Both manual `jrnl-md sync` and auto-sync perform:

1. Regenerate `search-index.json(.age)` and `journal.json`
2. `git pull --rebase` from remote
3. Stage all changed files (day files, index, manifest)
4. Commit with descriptive message
5. `git push`

If pull hits a conflict (rare — only same-day edits from two devices), abort the
rebase and warn the user. Do not silently resolve.

## Commit Messages

One operation = one commit. The commit includes modified day file(s), updated
search index, and updated manifest.

| Operation | Message |
|---|---|
| New entry | `add 2026-03-30 09:04 AM` |
| Edit entry | `edit 2026-03-30` |
| Delete entry | `delete 2026-03-30 09:04 AM` |
| Change time | `change-time 2026-03-30 09:04 AM → 02:30 PM` |
| Import | `import 12 entries (2024-01-15 to 2024-03-20)` |
| Encrypt/decrypt | `encrypt journal` / `decrypt journal` |
| Index-only rebuild | `rebuild index` |

No merge commits (`pull --rebase` keeps history linear). No squashing. No signing
(users configure that via git if desired).

## Search Index

Generated alongside each sync. Contains everything a web app needs to render
search results without fetching individual day files.

**Unencrypted journals** produce `search-index.json`:

```json
{
  "version": 1,
  "generated": "2026-03-30T14:00:00Z",
  "entries": [
    {
      "date": "2026-03-30",
      "time": "09:04 AM",
      "starred": false,
      "tags": ["@project", "@idea"],
      "body": "Full entry text here for search matching."
    }
  ]
}
```

**Encrypted journals** produce `search-index.json.age`:

- Same JSON structure, encrypted with the journal's passphrase via age
- Generated at sync time when the passphrase is already in memory
- `jrnl-md index` prompts for the passphrase if run standalone

**Scope**: full rebuild of the entire journal each time. For a 10-year daily
journal (~3,650 entries), the JSON is roughly 5-10 MB — manageable. Incremental
indexing is a future optimization if needed.

## Journal Manifest

`journal.json` at the repo root. Always unencrypted, even for encrypted journals.
Contains structural metadata only, never entry content.

```json
{
  "version": 1,
  "generator": "jrnl-md",
  "generator_version": "0.2.0",
  "encrypted": true,
  "date_range": {
    "first": "2024-01-15",
    "last": "2026-03-30"
  },
  "entry_count": 847,
  "tag_counts": {
    "@work": 203,
    "@idea": 45,
    "@health": 112
  },
  "file_extension": "md",
  "time_format": "03:04 PM",
  "date_format": "2006-01-02",
  "index_file": "search-index.json.age"
}
```

**What this enables for a web app**:

- Show journal structure (date range, entry count, tag cloud) before decryption
- Know whether to prompt for a passphrase (`encrypted: true`)
- Locate the search index file
- Understand the file naming scheme to fetch individual day files

**Deliberately excluded**: no entry content, no individual entry timestamps, no
file path listing. The manifest is safe to expose publicly.

Regenerated alongside the search index on every sync.

## Error Handling

If sync fails (no network, auth failure, conflict), the local write still
succeeds. jrnl-md prints a warning to stderr (`sync failed: <reason>`) but exits
0. The user's entry is never lost because sync couldn't reach the remote. They can
run `jrnl-md sync` manually later.

## Compatibility

This feature has no jrnl equivalent. It is an addition to jrnl-md, not a
behavioral divergence. Document in `docs/jrnl-compat.md` under a new "jrnl-md
Extensions" section.

## Out of Scope

- The web app itself (separate project consuming index/manifest)
- Conflict resolution beyond aborting (no auto-merge, no three-way merge UI)
- SSH key management or credential helpers (git handles auth)
- Sync over anything other than git (no Syncthing, no cloud storage)
- Incremental indexing (full rebuild; optimize later if needed)
- Multi-remote or multi-branch sync (one remote, one branch per journal)
