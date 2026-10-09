---
title: Slice planning and implementation authorization
summary: Agree architecture, then plan files and signoff; implement only the plan file the user instructs the agent to execute.
tags: [vm-agent, planning, authorization]
updated: 2026-10-08
---

# Slice planning and implementation authorization

Roadmap work uses two documentation passes before production implementation:

1. Discuss technical diagrams, module boundaries, and contracts in `architecture.md`, with a short `index.md`. A request to plan stops here for discussion.
2. After architecture agreement, write the numbered plan files, `testing.md`, and the full index. Review acceptance criteria, budgets, and open decisions; record user signoff with its date and who gave it.

Recorded signoff and an instruction to execute a specific plan file are both required before implementation. Plan agreement alone is insufficient. Implement production modules against the agreed contracts, and mark completion only when the required acceptance evidence exists.

## Canonical contracts and limits

- [Root workflow](../../../AGENTS.md#plan-slices-before-implementation).
- [Plan stages and authorization](../../plans/AGENTS.md#order-of-work).
- [Required slice brief](../../slice-planning.md#required-brief).

These are repository workflow contracts, not evidence of implemented platform capabilities. Documentation links, anchors, index coverage, relative symlinks, and whitespace were checked on macOS on 2026-10-08; no platform tests or Linux acceptance checks were run.
