---
title: Prototype workflow and tooling for vm-agent plans
summary: Plans go prototype → signoff → generalize → robust; prototypes are standalone Go/Deno folders with SQLite under prototypes/, run with `go -C prototypes/<name> run . demo`, and are kept as reference.
tags: [prototypes, planning, sqlite, tooling, macos]
updated: 2026-10-03
verified: 2026-10-03 — macOS darwin/arm64, Go 1.26.4; `go vet .` and `go run . demo` in prototypes/scaffold and prototypes/01-analytics.
---

# Prototype workflow and tooling for vm-agent plans

Every plan file under `docs/plans/` moves through four stages: a SQLite prototype the user can run, recorded signoff of the plan's questions, a generalize step that moves reusable tooling into `prototypes/scaffold/`, and only then the robust implementation with full fuzzing and MC/DC evidence. A passing demo is never acceptance evidence.

## How to act on it

- Start a prototype by copying the scaffold; each prototype has its own `go.mod` and imports nothing from other prototypes or the platform. Share code by copying, so folders stay standalone and can be kept as reference after the robust code exists.
- Write prototypes and demo drivers in Go or Deno; use Bash only for small glue. Use SQLite for storage whatever the production store will be, and state in the README what the stand-in cannot show.
- Run a demo from the repository root:

```bash
go -C prototypes/scaffold run . demo
go -C prototypes/01-analytics run . demo
KEEP=1 go -C prototypes/01-analytics run . demo   # keep the temporary working directory
```

## Tooling facts

- `demokit.go` re-executes the prototype's own binary via `os.Executable()`, so `go run . demo` needs no separate build step.
- Demo working directories are created directly under `/tmp`, not under `$TMPDIR`: macOS limits Unix socket paths to about 104 bytes and `$TMPDIR` paths under `/var/folders/...` can exceed that once a socket name is appended.
- SQLite prototypes use `github.com/mattn/go-sqlite3`, which needs cgo and a C compiler. It is the same driver version `tools/kb` already pulls into the module cache.
- Deno 2.7.14 is installed on the Mac at `~/.deno/bin/deno`; no Deno prototype or shared Deno entry point exists yet.

## Limits

Verified on macOS only. No prototype has been run on the Linux VM. A SQLite prototype says nothing about DuckDB's single-process ownership, appender, or read-only statement detection; plan 01 verifies those in its robust stage.

## Canonical contracts

- [Plans guide](../../plans/AGENTS.md).
- [Prototypes guide](../../../prototypes/AGENTS.md).
- [Plan 01 analytics](../../plans/01-analytics/index.md).
