# Knowledge base

The project selects [nickstrad/kb](https://github.com/nickstrad/kb) as its knowledge-base tool.
A copy of its tracked source lives under `tools/kb`, including the Go CLI, tests, and [knowledge-store skill](../tools/kb/skills/update-knowledge-store/SKILL.md).

Copied revision: `ad5c811609efa0381bf50c43b9bbaf500c5e6e9d`, on 2026-10-03.
The copied files retain upstream content and modes; the upstream Git directory is omitted.

## Configuration and scope

The tool is repo-scoped, with no global executable installation or shell-profile changes.
The complete knowledge-store skill and templates are copied into [`.agents/skills/update-knowledge-store`](../.agents/skills/update-knowledge-store/SKILL.md) for project use; `.claude/skills/update-knowledge-store` is a relative symlink to that package.
The upstream package under `tools/kb` remains unchanged. When refreshing it, deliberately synchronize the project skill and templates too.
The [upstream README](../tools/kb/README.md) describes its prerequisites and commands.
The [project README](../README.md#environment-variables) lists its supported environment variables and will grow with implemented platform settings.

Always set `KB_ROOT` to the absolute path of this repo's `docs/knowledge`; never use the upstream `~/knowledge` fallback.
Entries live under `docs/knowledge/data/`; `docs/knowledge/.kb/kb.sqlite` holds the derived search index, search history, and feedback. Git tracks source entries and ignores `.kb/`.
The store starts with `KB_EMBEDDER=none` for full-text search without a hosted service or embedding cost. Changing this later requires an explicit decision and reindexing.

## Run the repo-scoped CLI

From the repository root, build a temporary executable and scope configuration to the current terminal session:

```bash
export KB_ROOT="$PWD/docs/knowledge"
export KB_EMBEDDER=none
vm_agent_kb_bin="$(mktemp -d)/kb"
go -C tools/kb build -tags fts5 -o "$vm_agent_kb_bin" ./cmd/kb
"$vm_agent_kb_bin" doctor
"$vm_agent_kb_bin" search "Mac Linux development boundaries" --caller codex
# Judge the returned search ID using the feedback commands printed by search.
```

Use `--caller claude` in Claude Code. The same temporary executable supports all skill commands; it is not added to `PATH` or installed globally.
Upstream `kb doctor` checks `/usr/local/bin/kb` and warns when it is absent. That installation warning is expected here; do not follow its global-installer suggestion. Evaluate its actual driver, store, entry, and index health checks separately.
On a fresh checkout, run `"$vm_agent_kb_bin" reindex --all` to rebuild the ignored local index from tracked entries before searching. Existing local search history is preserved by reindexing but is not shared through Git.

## Agent use

Read the project skill at session start and follow the root [knowledge workflow](../AGENTS.md#knowledge-workflow) consistently.
Consult relevant findings before substantive implementation, investigation, or environment/tooling advice; judge every search and consider saving durable findings before finishing.
Follow the skill for `kb search`, `list`, `show`, `add`, `edit`, and search feedback.
Read [the store contract](knowledge/AGENTS.md) first.
Discover and maintain knowledge through the CLI rather than browsing its corpus files directly.
If the CLI is unavailable, report that limitation once per session and continue independent work rather than silently bypassing it.

Use `kb` for corpus discovery; do not generate or maintain a Markdown entry index unless requested.
Knowledge-tool setup is separate from the platform's planned analytics collector: `kb stats` analyzes knowledge search history with in-process DuckDB, while the platform roadmap calls for its own `analytics.duckdb` file.

## Refresh the copy

Copy a deliberate upstream revision, update this provenance, and verify the tracked files.
Any build or test run needs Go 1.26 or newer, a C compiler, and the `fts5` build tag.
Live Ollama tests require explicit opt-in and a configured service.
