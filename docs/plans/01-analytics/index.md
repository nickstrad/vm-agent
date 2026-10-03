# Plan 01: analytics — inspect the first event

Covers [roadmap](../../roadmap.md#analytics-collector-analytics-duckdb-and-analytics-helper-cli--inspect-the-first-event) items 1–5: the analytics collector, `analytics.duckdb`, and the analytics helper CLI.

**Status:** planned. No platform code exists. A [SQLite prototype](../../../prototypes/01-analytics/README.md) exists and awaits signoff. Every command and package below is proposed.

## Outcome

`vm-agent-analytics events`, `task <id>`, and `query` explain recorded events before any agent feature exists. A collector process is the only owner of `analytics.duckdb`; modules emit redacted events to it only when `--analytics` or `--debug` is set; the platform keeps working when analytics is off, slow, full, or down.

## Read in this order

1. This index: order, concurrency, signoff questions.
2. [architecture.md](architecture.md): before and after, contracts, one change diagram per plan file.
3. [testing.md](testing.md): fuzzing and MC/DC design. Analytics has no deterministic simulation.
4. The plan file you were asked to execute.

## Plan files

| File | CLI outcome | Roadmap items | Depends on | May run concurrently with | Stage |
| --- | --- | --- | --- | --- | --- |
| [01-first-event.md](01-first-event.md) | `events` shows one event that travelled emitter → collector → DuckDB. | 1, 2 | Signoff | — | Prototype built, awaiting signoff |
| [02-emitter-delivery.md](02-emitter-delivery.md) | `health` shows accepted, dropped, and rejected counts; flags gate emission. | 3 | 01 | 03, 05a | Prototype built, awaiting signoff |
| [03-helper-cli-queries.md](03-helper-cli-queries.md) | `task`, `query`, `modules`, `seed`, and `export` answer questions through the collector. | 4 | 01 | 02, 05a | Prototype built for `task` and `query`, awaiting signoff |
| [04-retention-limits-failure.md](04-retention-limits-failure.md) | `health` reports retention and degraded state; callers survive collector failure and a full disk. | 5 | 02, 03 | 05a | Prototype built for collector-down only, awaiting signoff |
| [05-linux-deployment.md](05-linux-deployment.md) | The collector runs under systemd on the VM; the Mac helper CLI reads it through SSH. | 1 (deploy) | 01 for 05a; 04 for 05b | 02, 03, 04 (05a only) | No prototype: deployment adds no new protocol or storage approach |

## Order

```text
                       +--> 02 emitter + delivery ----+
                       |    (lane A)                  |
prototype --signoff--> 01 first event                 +--> 04 retention, limits, --> 05b Linux
(SQLite)   gate        |                              |    failure isolation         failure checks
                       +--> 03 helper CLI queries ----+
                       |    (lane B)
                       |
                       +--> 05a Linux build + systemd deploy (lane C, when the VM is ready)
```

- **Gate:** no robust work starts before the user signs off the prototype and the [generalize step](../../../prototypes/AGENTS.md#generalize-step) has run. Signoff may cover all files at once or one file at a time.
- **01 is serial.** It creates the Go module, the event contract, the store schema, and the wire protocol that every other file builds on.
- **02, 03, and 05a may run concurrently** after 01 completes.
- **04 needs both 02 and 03**: it changes the ingest path and the query path.
- **05b** repeats the Linux service checks after 04, because 04 changes failure behavior.

## Concurrency rules

Lanes share only the contracts in [architecture.md](architecture.md#contracts), which freeze when 01 completes.

| Lane | Plan file | Owns (proposed paths) | Must not edit |
| --- | --- | --- | --- |
| A | 02 | `internal/analytics/emit/`, ingest half of `internal/analytics/collector/` (`ingest*.go`, `auth*.go`, `queue*.go`) | `query/`, `cmd/vm-agent-analytics/` except the `health` command |
| B | 03 | `internal/analytics/query/`, query half of `collector/` (`query*.go`), `cmd/vm-agent-analytics/` | `emit/`, ingest files |
| C | 05a | `deploy/systemd/`, Linux-only files (`*_linux.go`) | Portable packages |

- `internal/analytics/event/`, `internal/analytics/store/` schema, and the wire protocol are frozen while lanes run. A lane that needs one changed stops, updates `architecture.md`, and tells the user, because the other lane depends on it.
- Lanes A and B both add a request kind to the collector's dispatch table. Each adds its own file and one registration line; the later lane to merge rebases that line.
- Run lanes in separate worktrees. Merge one lane at a time and rerun the other lane's ordinary tests after each merge.

## Signoff questions

Run `go -C prototypes/01-analytics run . demo` from the repository root, then answer these. Each answer is recorded here with its date.

| ID | Question | Proposed answer | Shown in demo step | Status |
| --- | --- | --- | --- | --- |
| S1 | Is the v1 envelope field set right? | The fields in [architecture.md](architecture.md#event-envelope-v1). | 3, 4 | Open |
| S2 | Who assigns `producer` and `source_trust`? | The collector, from the authenticated identity. A producer cannot claim trust. | 4, 5 | Open |
| S3 | What transport do producers and the helper CLI use? | One Unix socket, JSONL frames, one request and one response per line. | all | Open |
| S4 | How are producers authenticated? | A per-identity token file with `0600` permissions now; Linux peer credentials are added as a second check in 05. | 5 | Open |
| S5 | What happens when a queue is full? | Drop the newest event, count it per producer, never block the caller. | 6 | Open |
| S6 | How open is `query`? | One read-only `SELECT` per request with a row cap and timeout; `export` produces a snapshot for unrestricted offline SQL. | 4, 5 | Open |
| S7 | How is redaction decided? | An allowlist of attribute keys per event name, applied by the producer and again by the collector; unknown keys are removed and counted. | 3, 8 | Open |
| S8 | What is the Go module layout? | One root module `github.com/nickstrad/vm-agent`; packages under `internal/analytics/`; binaries under `cmd/`. | — | Open |

## Decisions

Settled:

- **Analytics testing is native fuzzing plus MC/DC tables, with no deterministic simulation.** User decision, 2026-10-03. Analytics is observational and may lose events as long as it counts them. See [testing.md](testing.md).
- **Prototype in SQLite, sign off, generalize reusable tooling, then build the robust DuckDB version.** User decision, 2026-10-03. The prototype stays in the repository as reference. See [the plans guide](../AGENTS.md#prototype-signoff-generalize-robust).
- **The collector is the only process that opens `analytics.duckdb`.** Roadmap contract. The helper CLI reads through the collector.
- **Mac first.** Roadmap callout. Linux adds the native build, systemd deployment, and service-failure checks in file 05.
- **DuckDB driver:** [duckdb-go](https://github.com/duckdb/duckdb-go), as the roadmap names. It needs cgo and a C toolchain on each target.

Open, beyond the signoff questions:

- **O1: read-only statement detection in DuckDB.** The plan assumes the driver exposes a prepared statement's type so `query` can refuse anything but `SELECT`. Unverified. File 03 starts with a spike; the fallback is named queries plus `export` only.
- **O2: store-level size cap.** File 04 assumes the collector can measure database size from the file and DuckDB's own metadata. The exact measure is chosen during 04.
- **O3: whether the Linux VM exists yet.** 05 waits for it; nothing else does.

## Budgets and acceptance

| Budget | Limit |
| --- | --- |
| Fuzz discovery | 60 seconds per target by default; at most 10 minutes per target per session; one target per invocation. |
| Ordinary tests | Replay committed corpus and MC/DC tables; target under 60 seconds for `./internal/analytics/...` on the Mac. |
| VM cost | None until 05. 05 uses the existing lab VM; no additional hosted service. |
| Live-service spend | None. |

The plan is complete when every plan file's acceptance section is satisfied with recorded evidence, the owner documents hold the invariants in [testing.md](testing.md#invariants), the README lists implemented analytics configuration, and the prototype's README is marked superseded with a link to the robust code.

## Limits of this plan

- Other modules do not exist yet. The only producers are the helper CLI's proposed `emit-test` command and the collector's own health events.
- The check that analytics on and off give identical business results belongs to each later module's suite. This plan supplies the emitter contract they will use.
- Mac evidence does not establish Linux service behavior. See [05](05-linux-deployment.md).
