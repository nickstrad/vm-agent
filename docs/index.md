# Docs index

Start here for project plans and engineering rules.
The platform is currently in the planning stage; roadmap commands and modules describe intended behavior.

- [Roadmap](roadmap.md) - architecture, ordered vertical slices, CLI outcomes, tests, analytics, and optional extensions.
- [Slice planning](slice-planning.md) - the planning workflow and required brief before implementing each slice.
- [Plans guide](plans/AGENTS.md) - plan-folder layout, ordering and concurrency index, before/after architecture, testing emphasis, and the prototype → signoff → generalize → robust stages.
- [Plan 01: analytics](plans/01-analytics/index.md) - collector, DuckDB store, and helper CLI in five plan files; planned, with a SQLite prototype awaiting signoff.
- [Prototypes guide](../prototypes/AGENTS.md) - standalone reference prototypes, the copyable scaffold, and the generalize step.
- [Testing](testing.md) - deterministic suite design, native fuzzing, MC/DC evidence, test file conventions, and verification boundaries.
- [Coding style](coding-style.md) - custom Go code, Linux reuse, explicit contracts, and explanatory test and analytics comments.
- [Docs agent guide](AGENTS.md) - instructions for maintaining this folder and the roadmap.
- [Knowledge base](knowledge-base.md) - the copied kb source, pinned revision, included skill, and configuration reference.
- [Knowledge-store contract](knowledge/AGENTS.md) - repo-local storage, CLI access, durable findings, and verification rules.

Reusable findings live in `docs/knowledge/data/`; use the copied `kb` CLI to discover them. The local database is ignored by Git, and no corpus Markdown index is maintained.
