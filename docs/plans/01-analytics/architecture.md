# Plan 01 architecture: analytics

Everything here is planned. Paths, package names, commands, and schemas are proposals until [signoff](index.md#signoff-questions); after implementation this file is corrected to match what was built.

## Before

```text
repository
  docs/            roadmap, conventions, knowledge store
  tools/kb/        copied knowledge CLI (separate Go module)

no platform Go module, no binaries, no analytics store, no events
```

## After

```text
 Mac or Linux VM
 +----------------------------------------------------------------------------+
 |                                                                            |
 |  any producer process                 vm-agent-collector (one process)     |
 |  +-----------------------------+      +----------------------------------+ |
 |  | emit.Emitter                |      | session: hello -> identity       | |
 |  |  gate: --analytics/--debug  |      |   | emit            | query      | |
 |  |  redact (allowlist)         |      |   v                 v            | |
 |  |  bounded queue, drop+count  |      | admit + redact   admit + limit   | |
 |  |  sender goroutine           |      |   |                 |            | |
 |  +--------------+--------------+      | bounded queue       |            | |
 |                 | JSONL frames        |   | drop+count      |            | |
 |                 +-------------------->| batch writer --> store <---------+ |
 |                   Unix socket         |              (sole owner)        | |
 |  vm-agent-analytics (helper CLI)      +---------------+------------------+ |
 |  events | task | query | modules |                    |                    |
 |  seed | export | health  ------------> same socket    v                    |
 |                                             <state dir>/analytics.duckdb   |
 +----------------------------------------------------------------------------+

 Linux only (file 05): systemd unit, dedicated user, StateDirectory, RuntimeDirectory,
 peer-credential check, journald for critical errors, SSH-forwarded socket for the Mac CLI.
```

Trust boundaries:

- The collector trusts no field that states who sent an event. It derives `producer` and `source_trust` from the authenticated session.
- Events from the future runtime cell arrive as `untrusted`. Analytics never feeds a decision, so untrusted events can mislead a reader but cannot authorize anything.
- The helper CLI is a client. It never opens the database file.

## Modules

| Package or binary (proposed) | Responsibility | Owns invariants | Created or changed by |
| --- | --- | --- | --- |
| `internal/analytics/event` | Envelope type, JSONL codec, validation, redaction, attribute registry. Pure functions. | `AN-EVT-*` | 01 |
| `internal/analytics/wire` | Frame reader and writer, request and response types, size limits. Pure functions over `io.Reader` and `io.Writer`. | `AN-WIRE-*` | 01; 02 and 03 add request kinds |
| `internal/analytics/store` | The only code that opens DuckDB. Schema, migrations, batch append, parameterized reads, retention deletes. | `AN-STORE-*` | 01; 04 adds retention |
| `internal/analytics/collector` | Socket server, sessions, authentication, admission, ingest queue, batch writer, query dispatch, health. | `AN-COL-*` | 01 minimal; 02 ingest; 03 query; 04 degraded state |
| `internal/analytics/emit` | Producer-side library: flag gate, redaction, bounded queue, sender, local drop counts, flush. | `AN-EMIT-*` | 01 minimal synchronous sender; 02 full |
| `internal/analytics/query` | Named queries, ad hoc query admission, limits, result shaping, export. | `AN-QRY-*` | 03; 04 adds limits |
| `cmd/vm-agent-collector` | Daemon entry point and configuration. | — | 01 |
| `cmd/vm-agent-analytics` | Helper CLI. | — | 01 `events`, `emit-test`; 02 `health`; 03 the rest |
| `deploy/systemd/` | Unit file and install notes. | — | 05 |

Each production package gets an `invariants.md` beside its source, per the root guide. This plan lists the planned promises in [testing.md](testing.md#invariants); the owner documents become canonical when the code exists.

## Contracts

These freeze when plan file 01 completes. Concurrent lanes may add to them only as their plan file states.

### Event envelope v1

| Field | Type and limit | Set by | Purpose |
| --- | --- | --- | --- |
| `schema_version` | integer, `1` | producer | Lets the collector reject or migrate other versions. |
| `event_id` | string ≤ 128 bytes, required | producer | Identity for parent links. |
| `module` | string ≤ 64, required | producer | Platform concern the event describes. Must be one the session may speak for. |
| `name` | string ≤ 128, required, registered | producer | Event kind, such as `message.sent`. Selects the attribute allowlist. |
| `task_id`, `action_id`, `trace_id` | string ≤ 128, optional | producer | Correlation. |
| `parent_event_id` | string ≤ 128, optional | producer | Causal link. |
| `source_instance` | string ≤ 128, required | producer | One process lifetime of a producer. |
| `source_seq` | unsigned 64-bit, starts at 1, increases by 1 per event built | producer | Gaps reveal events lost before storage. |
| `logical_time` | unsigned 64-bit | producer | Ordering inside a controlled run; `0` when unused. |
| `emitted_at_ns` | signed 64-bit Unix nanoseconds | producer | Ingestion lag only. Never read by a decision. |
| `duration_ns` | signed 64-bit ≥ 0 | producer | Length of the operation described. |
| `outcome` | enum: `ok`, `error`, `denied`, `timeout`, `cancelled` | producer | Result. |
| `test_seed` | string ≤ 128, optional | producer | Seed or corpus ID of the test run that produced it. |
| `attrs` | ≤ 32 entries; key ≤ 64 bytes; value string ≤ 1024 bytes | producer | Redacted details. Only keys registered for `name` survive. |
| `producer` | string | **collector** | Authenticated identity. |
| `source_trust` | enum: `trusted`, `untrusted`, `test` | **collector** | From the producer registry. |
| `received_at` | timestamp | **collector** | Collector clock at admission. |
| `ingest_seq` | unsigned 64-bit | **collector** | Total order of stored events. |

`emitted_at_ns` is an addition to the roadmap's field list; it is the only way to observe ingestion lag. One encoded event must fit in a 16 KiB frame.

### Wire protocol v1

One Unix socket. Each frame is one JSON object on one line, at most 16 KiB. A longer or malformed frame ends the session and is counted.

```text
client                               collector
  | hello {v:1, token}                 |
  |----------------------------------->|  look up token -> producer, trust, permissions
  |<-----------------------------------|  welcome {producer} | error unauthenticated (closes)
  | emit {event}            (no reply) |  admit -> queue | drop+count | reject+count
  | emit {event}            (no reply) |
  | flush {dropped_local: n}           |
  |----------------------------------->|  barrier enters the queue behind earlier events
  |<-----------------------------------|  flushed {accepted, written, dropped, rejected}
  | query {kind, params} | health {}   |
  |<-----------------------------------|  result {columns, rows, truncated} | error {code}
```

- `emit` has no reply, so a slow collector cannot slow a producer.
- `flush` is the barrier: its reply means every event this session sent before it is stored, dropped and counted, or rejected and counted.
- Delivery is at most once. The emitter never retries, so the store needs no deduplication.
- Error codes are a closed set: `unauthenticated`, `not_permitted`, `invalid_event`, `frame_too_large`, `unsupported_version`, `query_not_read_only`, `query_limit`, `store_unavailable`. Go callers get matching sentinel errors.

### Producer registry

A collector-owned file maps each identity to the modules it may speak for, its trust, and whether it may query. Each identity's token lives in its own file readable only by that identity. Tokens never appear in events, logs, or error text.

### Store schema v1

```text
events(ingest_seq UBIGINT, schema_version, event_id, producer, source_trust, module, name,
       task_id, action_id, trace_id, parent_event_id, source_instance, source_seq UBIGINT,
       logical_time UBIGINT, emitted_at TIMESTAMP, received_at TIMESTAMP,
       duration_ns BIGINT, outcome, test_seed, attrs VARCHAR /* canonical JSON, sorted keys */)

producer_stats(producer, accepted, written, dropped_queue_full, dropped_local,
               rejected_invalid, rejected_auth, redacted_attrs, updated_at)

schema_meta(version, applied_at)
```

Only `store` runs SQL that writes. Whether `attrs` stays JSON text or becomes a DuckDB `MAP` is decided in file 01 after measuring the driver; the envelope does not change either way.

### Limits

| Limit | Default | Introduced by |
| --- | --- | --- |
| Frame size | 16 KiB | 01 |
| Producer queue | 1,024 events | 02 |
| Collector ingest queue | 4,096 events | 02 |
| Write batch | 512 events or 200 ms, whichever comes first | 02 |
| Query rows | 100 default, 10,000 maximum | 03 |
| Query time | 5 seconds | 04 |
| Retention | 14 days | 04 |
| Store size | 256 MiB | 04 |

## Change diagram per plan file

### 01 first event

```text
before: nothing
after:  + go.mod (root module)
        + event (envelope, codec, validate, redact)
        + wire (frames: hello, emit, flush, query kind "events")
        + store (open DuckDB, schema v1, append, read recent)
        + collector (accept, authenticate, admit, write one event at a time)
        + emit (gate + redact + synchronous send; no queue yet)
        + vm-agent-collector, vm-agent-analytics {emit-test, events}

 emit-test --> emit.Emitter --> socket --> collector --> store --> analytics.duckdb
 events    -------------------> socket --> collector --> store (read)
```

### 02 emitter and delivery

```text
 emit.Emitter:   synchronous send     ==>  gate -> redact -> [bounded queue] -> sender goroutine
                                                      full: drop newest, count locally
 collector:      write one at a time  ==>  admit -> [bounded queue] -> batch writer
                                                      full: drop newest, count per producer
 + flush barrier end to end           + producer_stats updated with each batch
 + health request and `vm-agent-analytics health`
 unchanged: event, store schema, query path
```

### 03 helper CLI queries

```text
 collector query dispatch:  kind "events" only  ==>  events | task | modules | seed | sql | export
 + query (named queries with bound parameters, ad hoc admission, result shaping)
 + vm-agent-analytics {task, query, modules, seed, export}
 + export writes a snapshot directory the user can open with the duckdb CLI
 unchanged: event, emit, ingest path, store schema
```

### 04 retention, limits, and failure isolation

```text
 store:      + retention delete by age, then by size
 collector:  serving  <==>  degraded (store unavailable: admit nothing, count everything)
             + query timeout and concurrent-query cap
 emit:       + reconnect with backoff; local drops reported at next flush
 health:     + oldest event, store size, retention runs, degraded state and reason
 unchanged: envelope, wire framing, named queries
```

### 05 Linux deployment

```text
 + deploy/systemd/vm-agent-collector.service (dedicated user, StateDirectory, RuntimeDirectory, hardening)
 + peer-credential check on Linux, in addition to tokens
 + critical errors to journald
 + Mac helper CLI --ssh-forwarded socket--> VM collector
 unchanged: all portable packages
```
