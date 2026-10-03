# 04 Retention, limits, and failure isolation

Part of [plan 01](index.md). Roadmap item 5. **Stage:** draft, awaiting architecture agreement. Everything below is planned.
Depends on [02](02-emitter-delivery.md) and [03](03-helper-cli-queries.md). 05a may run concurrently.

## Outcome

Analytics stays bounded, and its failure never becomes the platform's failure.

```text
$ vm-agent-analytics health                                   # proposed
collector: degraded (store_unavailable: disk full) since 12:04:10
retention: 14d window, 256 MiB cap, store 255.8 MiB, oldest event 13d 22h, last run deleted 18,204
analytics-cli  accepted 0  dropped_degraded 412  dropped_local 0

$ vm-agent-analytics emit-test --analytics ; echo $?          # collector stopped
0
```

## Scope

Included:

- `store`: retention by age, then by size, oldest first, run on a schedule and on demand.
- `collector`: the `serving` and `degraded` states; query timeout; cap on concurrent queries.
- `emit`: reconnect with capped backoff; local drops reported at the next flush.
- `health`: retention figures, store size, state, and reason.
- Configuration for window, size cap, and limits, documented in the README's environment-variable tables as implemented settings.

Deferred: disk exhaustion on the real target filesystem and service restarts ([05](05-linux-deployment.md)).

**Mac:** all of this file, with write failures injected at the store boundary. **Linux:** 05b repeats the failure cases against a real full filesystem and systemd.

## Change

See the [04 change diagram](architecture.md#04-retention-limits-and-failure-isolation).

```text
                 write fails with disk full,
                 or N consecutive failures
     serving  ------------------------------->  degraded
        ^         admit, write, query              | admit nothing, count dropped_degraded
        |                                          | queries refused with store_unavailable
        +------------------------------------------+ health still answers
                 periodic probe write succeeds
```

The collector owns this state machine. Producers never learn the state: their sends succeed or drop exactly as before, which is what keeps them independent of it.

## Testing

From [testing.md](testing.md). This file delivers:

| Kind | Items | Invariants |
| --- | --- | --- |
| Fuzz | `FuzzRetentionPlan`, `FuzzDegradedOps`; extend `FuzzEmitterOps` with reconnect operations | AN-STORE-3, AN-COL-5, AN-EMIT-2 |
| MC/DC | D8 retention delete, D9 degraded transition; add the `serving` condition to D3 and D6 | AN-STORE-3, AN-COL-5 |

No deterministic simulation and no virtual time. Timers are kept out of the decisions: "the window has passed" and "the deadline has passed" enter the pure decision functions as inputs, so the tables and fuzz targets cover the logic, and the real timer is exercised only by the walkthrough.

## Prototype and signoff

The [prototype definition](index.md#prototype-definition) proposes none for this file: its decisions are small and stated in full as tables D8 and D9. If the user wants to see the degraded state or retention running before signoff, add that to the definition first.

## Observability

- `collector.degraded` and `collector.recovered` with the error class, never the raw error text if it could contain a path under a user directory.
- `collector.retention` with rows deleted, bytes before and after, and duration.
- While degraded, these events cannot be stored; the collector writes them to standard error and counts them, then records one summary event on recovery.

## Steps

1. Write the retention decision with D8 and `FuzzRetentionPlan`; add the delete to `store`.
2. Write the two-state machine with D9 and `FuzzDegradedOps`; add `serving` to D3 and D6.
3. Add query timeout and the concurrent-query cap.
4. Add reconnect with backoff to `emit` and extend `FuzzEmitterOps`.
5. Extend `health`. Add configuration and update the README tables.
6. Update `invariants.md` in `store`, `collector`, and `emit`. Run the walkthrough and paste the transcript.

## Acceptance

- With the collector stopped, killed mid-burst, or degraded, `emit-test` exits 0 within its normal time and reports its local drops.
- With analytics disabled, no socket is opened and no file is created.
- An injected disk-full error moves the collector to degraded, `health` still answers, and a successful probe returns it to serving with a summary event.
- A store filled past its cap shrinks below it, deleting oldest first and nothing inside the window that the cap did not force.
- A query past its time limit is cancelled and reported with `query_limit`.
- Tests, race run, vet, and format are clean; each fuzz target met its discovery budget; each table has every listed row.

## Evidence

None yet.
