# 02 Emitter and delivery

Part of [plan 01](index.md). Roadmap item 3. **Stage:** prototype definition proposed. Everything below is planned.
Depends on [01](01-first-event.md). May run concurrently with [03](03-helper-cli-queries.md) and 05a as **lane A**; see [the concurrency rules](index.md#concurrency-rules).

## Outcome

Emission is gated by flags, never blocks its caller, and accounts for every event. `health` shows where each one went.

```text
$ vm-agent-analytics emit-test --analytics --count 5000     # proposed
$ vm-agent-analytics health
producer       accepted  written  dropped_queue_full  dropped_local  rejected_invalid  rejected_auth
analytics-cli  4310      4310     0                   690            0                 0
collector: serving   ingest queue 0/4096   lag p50 3ms
```

## Scope

Included:

- `emit`: bounded producer queue, sender goroutine, local drop counter, `Flush`, reconnect on the next send after a failure.
- `collector`: bounded ingest queue, batch writer, per-producer counters in `producer_stats`, flush barrier, `health` request.
- `vm-agent-analytics health`, and `--count` on `emit-test`.

Deferred: reconnect backoff and the degraded state ([04](04-retention-limits-failure.md)); Linux peer credentials ([05](05-linux-deployment.md)).

Owns: `internal/analytics/emit/` and the collector's ingest files. Must not edit `query/` or the helper CLI beyond `health`.

**Mac:** all of this file. **Linux:** nothing here.

## Change

See the [02 change diagram](architecture.md#02-emitter-and-delivery). The envelope, store schema for `events`, and frame format do not change. This file adds the `health` request kind and starts filling `producer_stats`.

The two queues are plain single-threaded state machines; goroutines only move items in and out of them. That shape is what lets them be fuzzed as operation sequences.

```text
caller --> gate --> redact --> [ producer queue ] --> sender --> socket --> admit --> [ ingest queue ] --> batch writer --> store
                                 full: drop newest,                                    full: drop newest,
                                 count dropped_local                                   count dropped_queue_full
```

## Testing

From [testing.md](testing.md). This file delivers:

| Kind | Items | Invariants |
| --- | --- | --- |
| Fuzz | `FuzzEmitterOps`, `FuzzIngestQueueOps`, `FuzzCollectorSession` | AN-EMIT-1..3, AN-COL-1..4 |
| MC/DC | D4 queue offer, D5 flush acknowledgement | AN-EMIT-2, AN-COL-3 |

No deterministic simulation. The roadmap's earlier "delivery simulation" for buffering faults is replaced by the two operation-sequence targets, whose model oracle is a slice and four counters. `FuzzCollectorSession` runs real goroutines and therefore asserts only conservation after a flush barrier, which must hold under every interleaving.

## Prototype and signoff

The [prototype definition](index.md#prototype-definition) proposes prototype A, shared with file 01, for S4 and S5. It is not agreed or built.

As proposed, prototype A leaves these out, so this file verifies them directly:

- The producer-side queue.
- Unacknowledged `emit` frames.
- Batch writes.

## Observability

- Per producer: accepted, written, dropped at the ingest queue, dropped locally, rejected as invalid, rejected for authorization, attributes removed.
- Collector: queue depth and capacity, batch size, ingestion lag from `emitted_at_ns` to `received_at`, schema versions seen.
- Local drops reach the collector in the next `flush` frame, so a producer that never reconnects leaves them uncounted. `source_seq` gaps still reveal the loss.
- Counters contain no event content.

## Steps

1. Write the producer queue state machine with D4 and `FuzzEmitterOps`.
2. Add the sender goroutine and `Flush` to `emit`.
3. Write the ingest queue and barrier state machine with D5 and `FuzzIngestQueueOps`.
4. Add the batch writer and `producer_stats` updates in the same transaction as each batch.
5. Add `health` to the collector and the helper CLI.
6. Write `FuzzCollectorSession` against the assembled collector.
7. Update `invariants.md` in `emit` and `collector`. Run the walkthrough and paste the transcript.

## Acceptance

- The outcome transcript is reproduced: a burst larger than the queues completes without blocking, and `accepted + dropped_queue_full + dropped_local + rejected == offered` after a flush.
- With both flags off, the emitter opens no socket, verified by the D1 table and by the walkthrough against a collector that logs sessions.
- Tests, race run, vet, and format are clean; each fuzz target met its discovery budget; each table has every listed row.

## Evidence

None yet.
