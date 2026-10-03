# Plan 01 testing approach: analytics

All targets, tables, and commands here are planned. This file is the testing contract for every plan file in this folder; each plan file names the rows it delivers.

## Position

Analytics modules are tested with **native Go fuzzing** and **MC/DC decision tables**. They are not tested with deterministic simulation.

| Used | Not used |
| --- | --- |
| `FuzzXxx` targets with committed seeds and saved failures | A seeded scheduler or controlled interleavings |
| Operation-sequence fuzzing against a model oracle for stateful parts | Virtual time |
| MC/DC tables for every compound decision | Replay-twice comparison of normalized traces |
| Real temporary DuckDB files inside the store and query fuzz targets | `_dst_test.go` files, scenario decoders, replay bundles |
| The race detector over corpus replay | A separate integration or simulation suite for analytics |

**Why.** Analytics observes. It never authorizes an action, never alters an outcome, and is allowed to lose events provided it counts them. A scheduler-controlled simulation earns its cost where a missed interleaving can corrupt authoritative state or grant authority; analytics has neither. The user settled this on 2026-10-03 and the rule is recorded in [the plans guide](../AGENTS.md#which-evidence-a-module-needs).

**What still holds.**

- Every fuzz target is deterministic for a given input: no wall clock, no unseeded randomness, no dependence on goroutine timing in an assertion. Go needs this to minimize and replay a failure.
- Oracles are independent of the code under test: a second, simpler statement of the rule, never a call into the production function.
- Other modules' suites will check that their results are identical with analytics on and off. That evidence is theirs; this plan supplies the emitter they will use and its guarantee that it does nothing when gated off.

## How stateful behavior is fuzzed without a simulator

Queues, flush barriers, and sessions have state, so they get operation-sequence fuzzing. The input bytes decode into a bounded list of operations. The target applies each one to the production state machine and to a small model, then compares.

```text
fuzz bytes --> decode: [offer, offer, take, fail-write, flush, offer, ...]   (bounded length)
                              |
              +---------------+----------------+
              v                                v
   production queue / session          model: a slice and four counters
              |                                |
              +----------- compare ------------+
   after every op:  offered == queued + written + dropped + rejected
   after a flush:   queued == 0, and every earlier event is written or counted
```

Two rules keep this honest without controlling a scheduler:

1. **Make the core single-threaded.** The queue and admission logic are plain state machines with no goroutines inside. Goroutines live in a thin shell that only moves items between the socket, the state machine, and the store.
2. **Where real goroutines run, assert only what every interleaving must satisfy.** The session target drives a real in-process collector over an in-memory connection and checks conservation after a flush barrier: the counts must add up whatever order the goroutines ran in. It never asserts an ordering that depends on timing.

The race detector runs over the same corpus. It can find a data race on an interleaving that happened to occur; it does not prove the absence of others. That gap is accepted for analytics and stated under [limits](#what-this-does-not-establish).

## Fuzz targets

| Target | Package | Plan file | Input decoding | Property and independent oracle | Invariants |
| --- | --- | --- | --- | --- | --- |
| `FuzzEventDecode` | `event` | 01 | Raw bytes as one frame. | Never panics; rejects frames over the limit; a decoded event re-encodes and decodes to an equal value. Oracle: structural equality, written in the test. | AN-EVT-1, AN-EVT-2 |
| `FuzzEventValidate` | `event` | 01 | Bytes fill a structured envelope, including out-of-range lengths and enums. | Accepts exactly when a separately written checklist of field rules passes. | AN-EVT-1 |
| `FuzzRedact` | `event` | 01 | Bytes choose an event name, attribute keys, and values; a marker canary is planted in every attribute the registry does not allow. | No canary appears anywhere in the encoded output; output passes validation; redacting twice equals redacting once. Oracle: byte search of the encoded event. | AN-EVT-3, AN-EVT-4 |
| `FuzzFrameReader` | `wire` | 01 | Bytes are a stream; a second stream chooses how reads are split. | Same frames regardless of read splitting; memory bounded by the frame limit; an oversize or malformed frame ends the stream with the right error. | AN-WIRE-1 |
| `FuzzStoreRoundTrip` | `store` | 01 | Bytes generate a batch of valid events. | Append to a real temporary DuckDB, read back, compare to the input batch field by field. | AN-STORE-1 |
| `FuzzEmitterOps` | `emit` | 02 | Operation sequence: emit, sender-takes, connection-fails, reconnect, flush, toggle gate. | Model of a slice and counters: `offered == sent + queued + dropped_local`; gate off performs no I/O; emit never reports an error to the caller. | AN-EMIT-1, AN-EMIT-2, AN-EMIT-3 |
| `FuzzIngestQueueOps` | `collector` | 02 | Operation sequence: admit, writer-takes-batch, write-succeeds, write-fails, barrier. | Model: every admitted event ends written exactly once or counted exactly once; a barrier resolves only after everything before it. | AN-COL-3 |
| `FuzzCollectorSession` | `collector` | 02 | Bytes become a bounded sequence of frames for several sessions with different identities, some hostile. | Real in-process collector and real temporary DuckDB. After a flush: stored rows carry only the session's identity and trust; no row exists for a module the session may not speak for; counts are conserved; no token appears in any reply or row. | AN-COL-1, AN-COL-2, AN-COL-3, AN-COL-4 |
| `FuzzNamedQueryParams` | `query` | 03 | Bytes become parameters for each named query, including quotes, semicolons, and SQL fragments. | Result equals a filter over the same events computed in Go; store row count and schema are unchanged afterwards. | AN-QRY-1, AN-QRY-3 |
| `FuzzAdHocAdmit` | `query` | 03 | Bytes become SQL text, seeded with reads, writes, DDL, multi-statement, `COPY`, `ATTACH`, `PRAGMA`, and `INSTALL` forms. | For every admitted statement, executing it leaves a fingerprint of all tables and the schema unchanged and touches no file outside the store. Oracle: the fingerprint, not the admission function. | AN-QRY-1 |
| `FuzzResultLimit` | `query` | 03 | Bytes choose a requested limit and a row count. | Returned rows never exceed the cap; `truncated` is true exactly when rows were cut. | AN-QRY-2 |
| `FuzzRetentionPlan` | `store` | 04 | Bytes generate event ages and sizes plus a retention window and size cap. | The delete set equals the one a brute-force oracle computes; nothing inside the window is deleted while the store is under its cap. | AN-STORE-3 |
| `FuzzDegradedOps` | `collector` | 04 | Operation sequence: admit, write-fails with a chosen error class, probe-succeeds, query, health. | Model of the two-state machine: in degraded state nothing is admitted, everything is counted, health still answers. | AN-COL-5 |

Conventions for every target:

- Seed with `f.Add` cases that each name the behavior they show. These seeds double as the readable specification examples; analytics does not need a separate example suite.
- Keep saved failures under `testdata/fuzz/<Target>/` and add a comment in the test naming the invariant each one broke.
- Bound every decoded size and operation count explicitly, so one input cannot run long.
- Targets that use DuckDB open a fresh temporary database per input. If that proves too slow to fuzz usefully, switch to one database per worker with tables emptied between inputs, and record the change here.

## MC/DC decision tables

Each compound decision lives in one pure function with named conditions. Its `_mcdc_test.go` table lists every row below; each condition has a pair of rows that differ only in that condition and in the outcome.

### D1 emission gate (`emit`, plan file 01)

`send = (analytics OR debug) AND configured AND name_registered`

| Row | analytics | debug | configured | name_registered | send | Witness for |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | T | F | T | T | T | base |
| 2 | F | F | T | T | F | analytics (1, 2); debug (2, 3) |
| 3 | F | T | T | T | T | debug (2, 3) |
| 4 | T | F | F | T | F | configured (1, 4) |
| 5 | T | F | T | F | F | name_registered (1, 5) |

### D2 session admission (`collector`, plan file 01)

`welcome = version_supported AND token_matches AND identity_enabled`

| Row | version_supported | token_matches | identity_enabled | welcome | Witness for |
| --- | --- | --- | --- | --- | --- |
| 1 | T | T | T | T | base |
| 2 | F | T | T | F | version_supported |
| 3 | T | F | T | F | token_matches |
| 4 | T | T | F | F | identity_enabled |

### D3 event admission (`collector`, plan file 01; `serving` added by 04)

`admit = may_speak_for_module AND version_is_1 AND envelope_valid AND serving`

Five rows: the all-true base and one row flipping each condition to false. Each false row must also increment exactly one counter, and the table asserts which.

### D4 queue offer (`emit` and `collector`, plan file 02)

`enqueue = NOT closed AND length < capacity`

Rows: open with room (enqueue); closed with room (drop); open at capacity (drop). Add the equality boundary `length == capacity - 1` (enqueue) and `length == capacity` (drop).

### D5 flush acknowledgement (`collector`, plan file 02)

`ack = barrier_reached AND (batch_committed OR batch_failure_counted)`

| Row | barrier_reached | batch_committed | batch_failure_counted | ack | Witness for |
| --- | --- | --- | --- | --- | --- |
| 1 | T | T | F | T | base |
| 2 | F | T | F | F | barrier_reached |
| 3 | T | F | F | F | batch_committed (1, 3); batch_failure_counted (3, 4) |
| 4 | T | F | T | T | batch_failure_counted |

`batch_committed` and `batch_failure_counted` both true is infeasible: a batch has one outcome. The table states this instead of inventing a row.

### D6 ad hoc query admission (`query`, plan file 03; `serving` added by 04)

`run = may_query AND single_statement AND read_only AND within_limit AND serving`

Six rows: the base and one flipping each condition. Each rejection maps to one error code, asserted in the table.

### D7 attribute retention during redaction (`event`, plan file 01)

`keep = key_registered_for_name AND value_within_limit AND NOT matches_secret_pattern`

Four rows: the base and one flipping each condition. A removed attribute is counted, never truncated into place.

### D8 retention delete (`store`, plan file 04)

`delete = older_than_window OR (store_over_cap AND among_oldest)`

| Row | older_than_window | store_over_cap | among_oldest | delete | Witness for |
| --- | --- | --- | --- | --- | --- |
| 1 | F | F | F | F | base |
| 2 | T | F | F | T | older_than_window |
| 3 | F | T | T | T | store_over_cap (3, 4); among_oldest (3, 5) |
| 4 | F | F | T | F | store_over_cap |
| 5 | F | T | F | F | among_oldest |

Add the boundary where an event's age equals the window exactly.

### D9 degraded transition (`collector`, plan file 04)

`degrade = write_failed AND (disk_full OR consecutive_failures >= threshold)`

Rows: write succeeded; failed once with a transient error (stay serving); failed with disk full (degrade); failed at the threshold (degrade); the boundary one below the threshold (stay serving).

Go statement coverage is reported alongside these tables. It supplements the witness rows and does not measure MC/DC.

## Invariants

Planned promises. Each moves into the owning package's `invariants.md` with a text-art example when that package is written; tests cite these IDs in comments.

| ID | Promise | Owner |
| --- | --- | --- |
| AN-EVT-1 | Any byte string decodes to a valid envelope or an error, without a panic and within the frame limit. | `event` |
| AN-EVT-2 | A valid envelope survives encode then decode unchanged. | `event` |
| AN-EVT-3 | A redacted event contains only attributes registered for its name, and no removed value appears anywhere in its encoding. | `event` |
| AN-EVT-4 | Redaction is idempotent and its output is valid. | `event` |
| AN-WIRE-1 | An oversize or malformed frame ends the session using bounded memory. | `wire` |
| AN-EMIT-1 | With both flags off the emitter builds nothing and performs no I/O. | `emit` |
| AN-EMIT-2 | Emitting never blocks the caller and never returns an error the caller must handle. | `emit` |
| AN-EMIT-3 | Every offered event is sent, still queued, or counted as dropped locally. | `emit` |
| AN-COL-1 | A stored event's producer and trust come only from its authenticated session. | `collector` |
| AN-COL-2 | No session stores an event for a module outside its registry entry. | `collector` |
| AN-COL-3 | After a flush reply, every earlier admitted event is stored exactly once or counted exactly once. | `collector` |
| AN-COL-4 | No token appears in a stored event, a reply, or a log line. | `collector` |
| AN-COL-5 | While degraded the collector admits nothing, counts everything, and still answers health. | `collector` |
| AN-STORE-1 | A stored event reads back equal to the admitted event. | `store` |
| AN-STORE-2 | Only the `store` package opens the database, and one process holds it at a time. | `store` |
| AN-STORE-3 | Retention deletes nothing inside the window while the store is under its cap, and deletes oldest first when over it. | `store` |
| AN-QRY-1 | No query request changes stored data, schema, or files outside the store. | `query` |
| AN-QRY-2 | A result never exceeds the row cap and reports when it was cut. | `query` |
| AN-QRY-3 | Named queries pass every caller-supplied value as a bound parameter. | `query` |

AN-STORE-2's cross-process half rests on DuckDB's own file lock. It is checked once by the acceptance walkthrough in [01](01-first-event.md#acceptance), not by a fuzz target.

## Files and comments

For a package `foo`, analytics uses `foo_fuzz_test.go`, `foo_mcdc_test.go`, and `foo_helpers_test.go` from [the testing conventions](../../testing.md#file-conventions). It has no `foo_dst_test.go`.

Every test file opens with a comment naming its boundary. Each target and table carries:

- an ASCII diagram of the decoder or state machine it exercises;
- the oracle and why it is independent of the production code;
- the invariant IDs it checks;
- the exact command that replays a saved failure;
- what the target does not cover.

## Commands (proposed, valid once the module exists)

```bash
go test ./internal/analytics/...                 # MC/DC tables + committed fuzz seeds and saved failures
go test -race ./internal/analytics/...           # same, with the race detector
go test -run '^$' -fuzz '^FuzzRedact$' -fuzztime 60s ./internal/analytics/event   # discovery, one target
go vet ./... && gofmt -l .
```

DuckDB needs cgo. The build tags and toolchain notes are settled in [01](01-first-event.md).

## Acceptance walkthrough

Each plan file ends with a short scripted run of the real binaries on the Mac, and its transcript is pasted into the plan file as evidence. It is a demonstration of the CLI outcome, not a test suite, and it takes over from the prototype's `demo` command as the thing to run once the robust code exists.

## What this does not establish

- **Unexplored interleavings.** No scheduler is controlled. A race the detector did not happen to observe can remain. Accepted: the worst result is a lost or miscounted analytics event.
- **Timing.** The 200 ms batch timer and query timeouts are not driven by virtual time. Their decision logic takes "deadline passed" as an input condition and is covered by tables; the real timer is exercised only by the walkthrough.
- **DuckDB crash durability.** What survives a kill during a write is checked on Linux in [05](05-linux-deployment.md), by observation.
- **Linux service behavior, peer credentials, and disk exhaustion on the target filesystem.** See [05](05-linux-deployment.md).
- **Analytics on/off equivalence for other modules.** Owned by those modules' suites.
