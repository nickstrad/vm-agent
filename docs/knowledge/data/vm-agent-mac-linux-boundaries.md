---
title: Mac development and Linux acceptance boundaries
summary: Portable platform logic and deterministic suites can start on Mac; real Linux isolation, services, networking, KVM, and eBPF need VM acceptance checks.
tags: [development, macos, linux, testing, architecture]
updated: 2026-10-03
---

# Mac development and Linux acceptance boundaries

The platform does not require all coding to happen in the Linux VM. Its planned portable Go control logic, analytics collector/helper CLI, database adapters, generators, oracles, and deterministic simulations can be developed on Mac. Linux-specific adapters and real-system acceptance need the KVM-capable Linux VM.

## How to choose the environment

| Concern | Mac work | Linux VM evidence |
| --- | --- | --- |
| Analytics and testing scaffolding | Collector, helper CLI, temporary DuckDB files, redaction fuzzing, generators, replay, independent oracles. | Native build, service deployment, systemd failure behavior; suite replay on the target. |
| CLI/API, SQLite, policy, supervision | Portable protocol/state logic, migrations, MC/DC tables, controlled faults and fixtures. | SSH delivery, target-filesystem crash behavior, real restart/cancellation and socket credentials. |
| Runtime and networking | Portable contracts, simulated process/network completion, proxy/TLS fixtures. | nspawn, namespaces, seccomp, cgroups, nftables, proxy-only routing and escape checks. |
| Harness/connectors/subagents | Go adapters, scripted Pi/model responses, API fixtures, permissions and budget simulation. | Actual harness/worker integration and containment within the runtime boundary. |
| Browser/eBPF extensions | Broker/provenance decisions, website fixtures, lost-event simulation. | QEMU/KVM guest integration, BPF programs/loaders, verifier, cgroup hooks, kernel attribution. |

The coding agent can author source on either machine. Choose the VM agent when implementing Linux-specific integration so it can exercise that boundary immediately. Separate portable packages from OS adapters and use build constraints where needed; native dependencies must be built for the target OS and architecture.

## Acceptance consequence

Passing a deterministic simulation or Mac native test does not establish Linux kernel isolation, filesystem crash durability, peer authentication, or external protocol correctness. Every slice plan must name its Mac checks, Linux checks, and remaining verification limits separately.

Plan the first analytics slice around a locally emitted event retrieved by the helper CLI; include later Linux deployment checks explicitly. Sync source commits between environments and keep databases and credentials local.

## Canonical contracts and limits

- [Roadmap and per-slice environment callouts](../../roadmap.md).
- [Slice planning brief](../../slice-planning.md).
- [Testing boundaries](../../testing.md).

These are planned execution and acceptance boundaries, not evidence that platform modules already exist or that Linux E2E has passed.
