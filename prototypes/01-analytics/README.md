# Analytics prototype

Status: awaiting signoff
Plan: [docs/plans/01-analytics](../../docs/plans/01-analytics/index.md)

A SQLite stand-in for the analytics collector, store, and helper CLI. Run it, then answer the plan's [signoff questions](../../docs/plans/01-analytics/index.md#signoff-questions).
This folder is standalone and stays in the repository as reference; see [the prototypes guide](../AGENTS.md).

## Run it

```bash
go -C prototypes/01-analytics run . demo          # from the repository root
KEEP=1 go -C prototypes/01-analytics run . demo   # keep the temporary working directory
```

It needs Go 1.26 or newer and a C compiler for the SQLite driver. `sqlite3` is optional and only used for the last step.
The demo works in a temporary directory under `/tmp`, starts a collector, and removes everything on exit.

Individual subcommands (`collector`, `emit`, `flush`, `health`, `events`, `task`, `query`) can be run by hand against a kept directory.

## What each step shows

| Step | Approach under review | Signoff question | Plan file |
| --- | --- | --- | --- |
| 1 | One collector process owns the database file and a Unix socket. | S3 | [01](../../docs/plans/01-analytics/01-first-event.md) |
| 2 | Flags off: nothing is built, redacted, or sent. | — | [02](../../docs/plans/01-analytics/02-emitter-delivery.md) |
| 3 | The first event travels producer → collector → store → helper command; a non-allowlisted attribute is removed before sending. | S1, S7 | [01](../../docs/plans/01-analytics/01-first-event.md) |
| 4 | Task timeline and ad hoc query go through the collector. The collector, not the producer, assigns `producer` and `source_trust`. | S2, S6 | [03](../../docs/plans/01-analytics/03-helper-cli-queries.md) |
| 5 | Unknown tokens, a producer speaking for another module, a non-query identity, and a write through `query` are rejected. | S4, S6 | [02](../../docs/plans/01-analytics/02-emitter-delivery.md), [03](../../docs/plans/01-analytics/03-helper-cli-queries.md) |
| 6 | A full queue drops the newest event and counts it; the producer is never blocked; a flush barrier makes counts stable. | S5 | [02](../../docs/plans/01-analytics/02-emitter-delivery.md) |
| 7 | With the collector stopped, the caller records a local drop and exits 0. | — | [04](../../docs/plans/01-analytics/04-retention-limits-failure.md) |
| 8 | The redacted value is absent from every database file. | S7 | [01](../../docs/plans/01-analytics/01-first-event.md) |

## What it does not show

- **DuckDB behavior.** SQLite stands in for the store. DuckDB's single-process ownership, appender, checkpoints, and read-only statement detection are verified in the robust stage.
- **Correctness under test.** There are no fuzz targets, MC/DC tables, or invariant documents. Passing the demo is not acceptance evidence for the plan.
- **The real emitter.** `emit` is a one-shot command, so the producer-side bounded queue is not exercised; only the collector's ingest queue is.
- **Unacknowledged emits.** Here the collector replies to every `emit`. The planned protocol sends no reply, so a slow collector cannot slow a producer.
- **Retention, export, module summary, seed lookup, schema evolution, and systemd deployment.**
- **Linux.** It has run on macOS only.

The `query` command is guarded twice here: the collector wraps the text as a subquery and runs it on a read-only SQLite handle. Step 5's write is stopped by the wrapper. The robust design replaces both with DuckDB-specific checks.

## Generalized

`demokit.go` and the subcommand-plus-`demo` layout moved to [`../scaffold`](../scaffold/README.md) on 2026-10-03. Repeat the [generalize step](../AGENTS.md#generalize-step) after signoff in case review changes the prototype.

## Verified

2026-10-03, macOS (darwin/arm64), Go 1.26.4: `go vet .` and `go run . demo` ran to completion with the canary check passing.
