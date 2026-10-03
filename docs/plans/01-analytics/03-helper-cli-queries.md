# 03 Helper CLI queries

Part of [plan 01](index.md). Roadmap item 4. **Stage:** prototype definition proposed. Everything below is planned.
Depends on [01](01-first-event.md). May run concurrently with [02](02-emitter-delivery.md) and 05a as **lane B**; see [the concurrency rules](index.md#concurrency-rules).

## Outcome

The helper CLI answers questions about recorded events through the collector, never by opening the database.

```text
$ vm-agent-analytics task t-42            # timeline of one task              (all proposed)
$ vm-agent-analytics modules              # per-module counts, outcomes, drops
$ vm-agent-analytics seed 8f3a...         # events from one test seed or corpus ID
$ vm-agent-analytics query "SELECT name, count(*) FROM events GROUP BY 1"
$ vm-agent-analytics export ./snapshot    # snapshot to open with the duckdb CLI
```

## Scope

Included:

- `query`: named queries `events`, `task`, `modules`, `seed` with bound parameters; ad hoc `SELECT` admission; row cap and `truncated` flag; export.
- Collector query dispatch for those kinds, on a connection separate from the batch writer's.
- The five helper commands, with table and `--json` output.

Deferred: query timeout, concurrent-query cap, and behavior while degraded ([04](04-retention-limits-failure.md)).

Owns: `internal/analytics/query/`, the collector's query files, and `cmd/vm-agent-analytics/`. Must not edit `emit/` or ingest files.

**Mac:** all of this file. **Linux:** nothing here; 05 runs the commands over an SSH-forwarded socket.

## Change

See the [03 change diagram](architecture.md#03-helper-cli-queries).

```text
helper CLI --> query {kind, params} --> collector
                                          | named: fixed SQL + bound parameters
                                          | sql:   admit (single, read-only) -> wrap with row cap
                                          | export: write snapshot under a caller-named directory
                                          v
                                        store (read connection inside the owning process)
```

DuckDB allows many connections inside the one process that owns the file, which is how reads proceed while the writer appends. Other processes get data only through the collector or through an export, respecting [DuckDB's concurrency model](https://duckdb.org/docs/current/connect/concurrency).

## Open decision first

**O1** in the [index](index.md#decisions): ad hoc admission assumes the driver reports a prepared statement's type. Step 1 is a spike that checks this against statements that write, attach, copy, install extensions, or chain a second statement. If it cannot be made reliable, ad hoc `query` is removed from this file and users run unrestricted SQL against an `export` snapshot instead. Signoff question S6 already offers that choice.

## Testing

From [testing.md](testing.md). This file delivers:

| Kind | Items | Invariants |
| --- | --- | --- |
| Fuzz | `FuzzNamedQueryParams`, `FuzzAdHocAdmit`, `FuzzResultLimit` | AN-QRY-1..3 |
| MC/DC | D6 ad hoc query admission (without `serving`) | AN-QRY-1 |

No deterministic simulation. `FuzzAdHocAdmit` is the security-relevant target: its oracle fingerprints every table and the schema before and after an admitted statement runs against real DuckDB, so it judges the effect rather than trusting the admission function's own classification.

## Prototype and signoff

The [prototype definition](index.md#prototype-definition) proposes prototype B, a DuckDB spike that answers O1 and S6. It is not agreed or built. If it is agreed, it replaces step 1 below and its result is recorded in the index before the rest of this file starts.

Named queries, result limits, and export need no prototype; they are judged from this plan.

## Observability

- Each query emits `collector.query` with kind, row count, truncation, duration, and outcome. Ad hoc SQL text is not recorded; its length and a hash are.
- `export` records the destination's base name and size, not its full path.

## Steps

1. Spike O1 and record the result in the index.
2. Write named queries with `FuzzNamedQueryParams`.
3. Write result shaping and limits with `FuzzResultLimit`.
4. Write ad hoc admission with D6 and `FuzzAdHocAdmit`, or record its removal.
5. Write export.
6. Add collector dispatch and the helper commands.
7. Write `query/invariants.md`. Run the walkthrough and paste the transcript.

## Acceptance

- Each command in the outcome returns correct results for a store populated by `emit-test`, with identical content in table and `--json` form.
- A write, a DDL statement, a multi-statement string, and `COPY ... TO` are each refused through `query` with `query_not_read_only`, and the store is unchanged.
- A result larger than the cap is cut and reported as truncated.
- An export opens in the `duckdb` CLI while the collector keeps running.
- Tests, race run, vet, and format are clean; each fuzz target met its discovery budget; the D6 table has every listed row.

## Evidence

None yet.
