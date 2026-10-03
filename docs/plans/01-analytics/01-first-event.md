# 01 First event

Part of [plan 01](index.md). Roadmap items 1 and 2. **Stage:** prototype definition proposed. Everything below is planned.

## Outcome

One event travels from an emitter through the collector into `analytics.duckdb`, and the helper CLI shows it.

```text
$ vm-agent-collector --state-dir ./state &                      # proposed
$ vm-agent-analytics --state-dir ./state emit-test --analytics  # proposed
$ vm-agent-analytics --state-dir ./state events
ingest_seq  producer       source_trust  module         name            outcome  attrs
1           analytics-cli  trusted       analytics-cli  analytics.test  ok       {"note":"first event"}
```

## Scope

Included:

- The root Go module and the DuckDB driver build on the Mac.
- `event`: envelope v1, JSONL codec, validation, redaction with a per-name attribute registry.
- `wire`: frame reader and writer with `hello`, `emit`, `flush`, and the `events` query kind.
- `store`: open the database as sole owner, schema v1, append, read recent events.
- `collector`: accept sessions, authenticate, admit, write each event as it arrives.
- `emit`: flag gate, redaction, synchronous send. No queue yet.
- `vm-agent-collector`, and `vm-agent-analytics` with `emit-test` and `events`.

Deferred: bounded queues, drop accounting, and `health` ([02](02-emitter-delivery.md)); other queries ([03](03-helper-cli-queries.md)); retention and degraded state ([04](04-retention-limits-failure.md)); Linux ([05](05-linux-deployment.md)).

Prerequisites: [signoff](index.md#signoff-questions) of S1–S4, S7, and S8. Go 1.26 or newer and a C toolchain on the Mac.

**Mac:** all of this file. **Linux:** nothing here; 05 builds and runs the same code on the VM.

## Change

See the [01 change diagram](architecture.md#01-first-event). This file creates every contract in [architecture.md](architecture.md#contracts); they freeze when it completes.

## Testing

From [testing.md](testing.md). This file delivers:

| Kind | Items | Invariants |
| --- | --- | --- |
| Fuzz | `FuzzEventDecode`, `FuzzEventValidate`, `FuzzRedact`, `FuzzFrameReader`, `FuzzStoreRoundTrip` | AN-EVT-1..4, AN-WIRE-1, AN-STORE-1 |
| MC/DC | D1 emission gate, D2 session admission, D3 event admission (without `serving`), D7 attribute retention | AN-EMIT-1, AN-COL-1, AN-COL-2, AN-EVT-3 |

No deterministic simulation. The redaction target matters most here: its canary oracle searches the encoded bytes, so it cannot be satisfied by a redactor that merely agrees with itself.

Owner documents created: `invariants.md` in `event`, `wire`, `store`, `collector`, and `emit`, each holding the promises this file implements and marking later ones as planned.

## Prototype and signoff

The [prototype definition](index.md#prototype-definition) proposes prototype A, shared with file 02, for S1–S4 and S7. It is not agreed or built.

Whatever is agreed, a SQLite stand-in cannot show the following, so this file verifies them against DuckDB:

- A second process opening the same database file is refused while the collector holds it.
- The driver builds with cgo on the Mac, and which build tags it needs.
- Batch append and read-back of every envelope type, including unsigned 64-bit values and timestamps.

## Observability

- The collector emits its own events through the same admission path: `collector.started`, `collector.session.opened`, `collector.session.rejected`. Tokens and raw frames never appear in them.
- Critical collector errors go to standard error, which becomes journald under systemd in 05.
- With both flags off, `emit-test` sends nothing and says so.

## Steps

1. Create the root module and add the DuckDB driver. Record the build command, tags, and toolchain in the README and as a knowledge finding. Check: an empty program opens and closes a temporary database.
2. Write `event` with D7 and the three event fuzz targets. Check: tables pass; each target runs 60 seconds without a finding.
3. Write `wire` with `FuzzFrameReader`.
4. Write `store` with schema v1 and `FuzzStoreRoundTrip`. Decide JSON text or `MAP` for `attrs` and record it in `architecture.md`.
5. Write `collector` session handling and admission with D2 and D3.
6. Write `emit` with the gate and D1.
7. Write the two binaries and the `events` and `emit-test` commands.
8. Write the owner `invariants.md` files. Run the acceptance walkthrough and paste its transcript below.

## Acceptance

- The outcome transcript above is reproduced with the real binaries on the Mac, including a non-registered attribute that does not appear in the output or in the database file.
- With the collector running, a second process that tries to open `analytics.duckdb` directly fails. This is the one check for the cross-process half of AN-STORE-2.
- `go test`, `go test -race`, `go vet`, and `gofmt -l` are clean for `./internal/analytics/...`.
- Each listed fuzz target has committed seeds and has run its discovery budget without an unresolved finding.
- Each listed MC/DC table contains every row in [testing.md](testing.md#mcdc-decision-tables).

## Evidence

None yet. Record commands run, environment, and results here after implementation.
