---
title: Deterministic simulation and observational analytics contract
summary: Fuzz corpus bytes drive bounded production-state simulations; analytics must preserve decisions and randomness and is itself tested by fuzzing plus MC/DC only; native checks prove boundaries simulation cannot.
tags: [testing, fuzzing, simulation, analytics, invariants]
updated: 2026-10-03
---

# Deterministic simulation and observational analytics contract

The platform's planned suites use native Go fuzz input as a deterministic scenario description, not as an invitation to create uncontrolled randomness inside the system under test. Analytics observes that same execution without changing its decisions or random draws.

```text
corpus bytes + decoder version + revision + fixtures
                         |
              bounded state/actions/faults/schedule
                         |
             production transitions -> independent oracle
                         |
              redacted events -> one collector -> DuckDB
```

## Replay boundary

Control time, IDs, randomness, I/O completions, and scheduling only where the named invariant requires them. Derive named random streams and sort map-derived choices. Save complete inputs and fixture/generator versions: a seed alone does not reproduce uncontrolled concurrency or remote services.

Replay identical inputs twice and compare final state and normalized business traces. Use independent safety/preservation/progress oracles, stating fairness assumptions and when faults stop. Focused parser, redaction, and pure-policy fuzzing need no universal scheduler.

Extend a suite when a module enlarges its boundary; create a distinct suite for a new invariant or contract. MC/DC tables supply condition witnesses; Go statement coverage alone does not establish MC/DC. Native storage, kernel, harness protocol, and live-model claims need separate checks.

## Analytics boundary

Optional debug/analytics emission never authorizes actions or changes correctness. One collector owns the platform's `analytics.duckdb`; other services emit bounded, correlated, redacted events rather than opening it for concurrent writing. Authoritative task state and security audit remain in SQLite, with critical errors in journald.

Analytics modules and tools themselves are exempt from deterministic simulation (user decision, 2026-10-03): test them with native Go fuzzing and MC/DC tables only, with no seeded scheduler, virtual time, replay-twice comparison, or `_dst_test.go` files. The exemption is narrow. Their fuzz targets must still be deterministic per input; stateful parts (queues, flush barriers, sessions) are fuzzed as operation sequences against a simple model oracle, asserting only properties that hold under any interleaving; and every other module still proves analytics on/off equivalence in its own suite.

Acceptance of other modules compares analytics-enabled and disabled business outcomes. Analytics fuzz targets and tables cover delivery limits, drop reporting, flush barriers, collector failure, producer identity, and redaction. Comments should explain generators, independent oracles, faults, replay, and event flows using concrete text art.

## Canonical contracts and limits

- [Testing conventions](../../testing.md).
- [Plans guide: which evidence a module needs](../../plans/AGENTS.md).
- [Plan 01 analytics testing approach](../../plans/01-analytics/testing.md).
- [Roadmap design contract](../../roadmap.md).
- [Coding and comment conventions](../../coding-style.md).

These are planned design contracts. Platform implementation, deterministic replay, and Linux/model acceptance have not yet been verified. The copied knowledge tool's in-process DuckDB statistics are separate from platform analytics.
