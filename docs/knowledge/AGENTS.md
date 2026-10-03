# Knowledge-store contract

This directory is the knowledge root for `vm-agent` only. Follow the root and docs agent guides plus the project [update-knowledge-store skill](../../.agents/skills/update-knowledge-store/SKILL.md).

## Storage and access

- Set `KB_ROOT` to the absolute path of this directory for every invocation. Never rely on `~/knowledge` or a different repository's store.
- Use the copied `tools/kb` CLI via the [temporary-build recipe](../knowledge-base.md#run-the-repo-scoped-cli); do not install it computer-wide or change shell profiles.
- `kb doctor` may warn about missing `/usr/local/bin/kb`; that global-installation check is expected to warn for this repo-scoped setup and does not justify running the installer.
- Use `KB_EMBEDDER=none` for this initial full-text store. Do not change providers/models without a request; intentional changes need a complete reindex.
- `data/` holds tracked knowledge entries and companions. `.kb/` holds ignored local SQLite state, including search history and feedback.
- Read and maintain corpus entries through `kb`, never filesystem browsing, direct edits, or a Markdown entry index. Draft outside `data/` and import through `kb add`.
- Search before adding, read relevant hits, prefer revision over duplication, and judge every search with its explicit ID. Confirm changes using `kb show`.

## What to retain

- Save durable, non-obvious facts about repo tooling, native environment boundaries, integration pitfalls, simulation limits, replay, and analytics behavior.
- Save newly established engineering contracts when they prevent future mistakes; link their canonical docs rather than duplicating whole plans.
- In particular, record where Mac implementation/tests suffice and which Linux checks remain necessary. This is a reusable execution boundary, not merely task progress.
- Keep personal preferences, errands, temporary progress, secrets, databases, and generated traces out of entries.
- Require `title`, a useful one-line `summary`, `tags`, and `updated` front matter. Mark proposed contracts and untested runtime claims explicitly; use `verified` only with actual date, environment, and checks.
- Root and docs instruction files remain authoritative for workflow; knowledge findings provide supporting facts and evidence.

## Version control

- Track entries, companions, and this contract; `CLAUDE.md` is a relative symlink to this guide.
- Do not commit `.kb/` or generate `index.md` unless specifically requested. Rebuild a checkout's index with `kb reindex --all` while preserving existing local history.
- Commit only when the user asks.
