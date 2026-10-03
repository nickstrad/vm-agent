---
title: Repo-local knowledge CLI and storage
summary: Run the copied tools/kb CLI through scripts/kb, which pins KB_ROOT to this checkout's docs/knowledge and KB_EMBEDDER=none, with no global installation; ignore its local database.
tags: [knowledge, tooling, configuration, repository]
updated: 2026-10-03
verified: 2026-10-03 — macOS; scripts/kb build and doctor, stale-index reindex, temporary fts5 build, empty-store reindex, FTS search, and explicit-ID search feedback. The wrapper's missing-index auto-reindex branch was not exercised.
---

# Repo-local knowledge CLI and storage

This project's knowledge store is `docs/knowledge`, while its unchanged pinned tool source is `tools/kb`. These paths serve different purposes. The CLI must always receive this repo's absolute `KB_ROOT`; the upstream `~/knowledge` fallback is not this project's store.

## Storage layout

```text
vm-agent/
  tools/kb/               copied CLI source, not a global installation
  .agents/skills/         shared skill plus templates
  .claude/skills/         symlink to the shared skill
  docs/knowledge/
    AGENTS.md             store contract
    CLAUDE.md -> AGENTS.md
    data/                 tracked entries and companions
    .kb/                  ignored local SQLite index/history/feedback
```

Start with `KB_EMBEDDER=none`: full-text indexing/search needs no embedding service or paid inference. Do not switch an existing store's provider/model implicitly. Search and feedback history remain local; a new checkout reconstructs the index using `kb reindex --all`.

## Invocation

Preferred: `scripts/kb <kb args>` from any checkout or worktree. It derives `KB_ROOT` from the script location, pins `KB_EMBEDDER=none`, builds `tools/kb` with `fts5` into the ignored `docs/knowledge/.kb/bin/` (rebuilt only when sources or the Go version change), and runs `reindex --all` when `.kb/kb.sqlite` is missing. A merely stale index is not detected: after pulling or merging entry changes, or when `kb doctor` reports `db-vs-files` stale/orphan failures, run `scripts/kb reindex --all` (it preserves search history).

Manual equivalent, from the repository root:

```bash
export KB_ROOT="$PWD/docs/knowledge"
export KB_EMBEDDER=none
vm_agent_kb_bin="$(mktemp -d)/kb"
go -C tools/kb build -tags fts5 -o "$vm_agent_kb_bin" ./cmd/kb
"$vm_agent_kb_bin" reindex --all
"$vm_agent_kb_bin" search "Mac Linux development boundaries" --caller codex
# Judge the explicit search ID with the feedback commands printed by search.
```

These variables apply to the terminal session; do not change shell profiles or install the binary into a computer-wide path. Building requires Go 1.26 or newer and a C compiler. Use `--caller claude` for Claude Code.

Discover/read/maintain entries through the CLI. Draft outside the corpus, import with `kb add`, and confirm with `kb show`; never directly browse/edit corpus files or maintain a second Markdown entry index.

## Canonical guidance and limits

- [Repo-scoped invocation and provenance](../../knowledge-base.md).
- [Store contract](../AGENTS.md).
- [Agent knowledge workflow](../../../AGENTS.md#knowledge-workflow).

The initial build, index creation, empty FTS search, and explicit-ID feedback succeeded on Mac. This does not claim the imported tool's complete upstream test suite or live embedding integrations have passed.
