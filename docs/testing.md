# Testing

Use [Dropbox's testing approach](https://dropbox.tech/infrastructure/-testing-our-new-sync-engine) as the design reference: make protocols, data models, and control flow testable before building their suites.
The [roadmap](roadmap.md) assigns concrete test and observation boundaries to each platform slice.

These are platform implementation conventions. The platform module, simulator, and test CLI do not exist yet; the separately copied `tools/kb` module has its own upstream tests and requires the `fts5` build tag.

## Choose the boundary first

For every behavior, state its invariant, independent oracle, observable events, and sources of nondeterminism.
Reject invalid domain states in the protocol or data model instead of making every module handle them.

Use this sequence when designing a suite:

```text
choose boundary
      |
generate related state and actions
      |
control I/O, time, and completion order
      |
run production transitions + independent oracle
      |
capture trace -> replay -> reduce -> retain regression
```

Expand an existing suite when a module enlarges its boundary. Create a focused suite for a distinct invariant, protocol, or native-system requirement.
Codec, pure-policy, redaction, and query-construction properties usually need direct fuzz targets rather than an event scheduler.
Stateful simulation is appropriate for sequencing, interleavings, expiry, cancellation, failures, and recovery.

## File conventions

For a package named `foo`, use these suffixes when the corresponding checks exist:

| File | Purpose |
| --- | --- |
| `foo_test.go` | Named specification behaviors. |
| `foo_mcdc_test.go` | Decision tables and condition witness pairs. |
| `foo_fuzz_test.go` | Native Go `FuzzXxx` discovery and corpus replay. |
| `foo_dst_test.go` | Fixed-seed or explicit-scenario simulation regressions. |
| `foo_helpers_test.go` | Shared generators, contract doubles, and independent oracle helpers. |
| `foo_integration_test.go` | Real storage, process, socket, or protocol adapters. |
| `foo_e2e_test.go` | Real assembled-system workflows and Linux boundaries. |

Each file starts with a comment naming its test boundary. Each invariant check names the owner document's invariant ID in a comment.
Keep canonical promises in the module's `invariants.md`, or a simulation package's `architecture.md` under `What it proves`; tests provide the mapping to those promises.

## Native fuzzing and deterministic scenarios

Prefer [Go native fuzzing](https://go.dev/doc/security/fuzz/). Use `f.Add` for meaningful initial cases and preserve saved failures under `testdata/fuzz/<Target>/`.

- Decode corpus bytes deterministically into bounded initial data, actions, faults, and scheduling seeds.
- Generate valid domain states by construction. Generate deliberately invalid requests explicitly, and test malformed wire formats separately.
- Validate each fixed corpus case's decoded data and action sequence before naming the behavior it exercises.
- Use one fresh system per input, stable ordering, virtual time, and controlled I/O completion where required.
- Derive named PRNG streams from the input for larger generated data. No uncontrolled randomness belongs inside the scenario.
- Reuse the invariant oracle and scenario runner between discovery targets and fixed replay tests.
- Compare final state and normalized business-event traces on repeated runs of identical inputs.
- Retain full input, seed, decoder/generator version, revision, fixtures, and schedule. A seed alone cannot reproduce an uncontrolled external system.
- Keep size, step, and logical-time limits explicit. Discovery budgets may use wall time; scenario behavior must not depend on it.
- Preserve the original failure bundle even if reduction cannot retain the same fault or invariant failure.
- Add a regression whenever discovery exposes a real failure. Use richer generation libraries only when they materially improve a suite.

## MC/DC decision evidence

Use table tests for compound decisions. Name each condition and the decision it affects, then identify feasible witness pairs where one condition changes and the outcome changes independently.
Cover equality boundaries, short-circuit behavior, and rejected inputs where relevant. Explain infeasible condition combinations.
Do not assume every decision requires exactly `N+1` rows; complex or coupled conditions can need additional cases or explicit limitations.
Go statement coverage supplements the witness manifest; it does not measure MC/DC.

## Separate suites for different claims

- Focused decision and transition suites check policy, budgets, lifecycle, provenance, and preservation or progress.
- Platform simulation connects production control state machines to controlled adapters and fault schedules.
- Native-adapter suites check real SQLite, DuckDB, filesystems, sockets, and processes against those contracts.
- Protocol suites check actual Pi, CLI, HTTP/TLS, and service behavior so fixtures cannot silently drift.
- Linux E2E uses Go `testing` and `os/exec` to drive disposable systemd/nspawn/QEMU environments; use race checks for exercised Go concurrency.
- Browser E2E uses controlled Playwright website fixtures alongside the real broker and Chromium boundary.
- Live-model evaluations use versioned cases, recorded trajectories, repeated trials, and explicit spending budgets. They do not claim deterministic inference.

Simulation covers its controlled boundary. It cannot establish real kernel isolation, crash durability, adapter correctness, or remote protocol behavior by itself.
Progress assertions must state when faults stop and what scheduling fairness they assume.

## Analytics and explanation

Run acceptance checks with analytics both enabled and disabled. Business results and normalized control traces must agree.
Analytics tests cover redaction, producer identity, schema evolution, delivery limits, drop reporting, flush barriers, queries, and collector failure.
Use temporary DuckDB files and one writer owner per file. Match collector-backed query and export behavior to the helper CLI.

Require helpful comments with ASCII diagrams for generators, scenario decoders, state transitions, fault schedules, oracles, shrinking, and analytics flows.
Explain what a failure teaches and how to replay it. Do not introduce logging that changes random draws or control scheduling.

## Verification and cost

Once the Go module exists, ordinary `go test ./...` replays specification tests, fixed scenarios, and committed fuzz corpus cases.
Run discovery deliberately, one native fuzz target per invocation, with bounded time or iterations.
Run native, VM E2E, race, and live-model checks according to their documented prerequisites and budgets.
Report which boundary ran and which remained unverified. Do not invent a Makefile or CLI command before it exists.
