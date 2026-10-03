---
title: Prototype protocol for vm-agent plans
summary: Define the prototype in the plan and get the user's agreement before building anything; shape may be one for the plan, one per plan file, several working together, or none.
tags: [prototypes, planning, sqlite, tooling]
updated: 2026-10-03
---

# Prototype protocol for vm-agent plans

What a prototype needs differs for each part of a plan. A plan's `index.md` therefore carries a prototype definition that the user agrees to before any prototype code exists. Building a prototype, a scaffold, or shared tooling ahead of that discussion is out of protocol.

## How to act on it

- Propose a shape with reasons: one prototype for the whole plan, one per plan file, several prototypes working together, or none because the change is scoped enough to judge from the plan. A plan may mix them.
- For each proposed prototype state its question, what it includes and leaves out, language, storage, what the user will judge, and what it cannot show.
- Build only what the user agreed to. Record the agreement and date in the plan index.
- Agreed prototypes are standalone folders under `prototypes/`, written in Go or Deno (Bash for simple glue), using SQLite unless the question is about the production store itself. They are kept as reference after the robust code exists.
- Propose extracting shared prototype tooling only after real prototypes show what repeats, and extract only with the user's agreement.
- A prototype demo is never acceptance evidence. Robust work starts after recorded signoff.

## Seen in

On 2026-10-03 an analytics prototype and a scaffold were built before the approach was discussed; the user had them removed and asked for this protocol. Commit `ab6b1c9` holds the removed code if it is ever useful as reference.

## Limits

This is a workflow contract, not a verified tooling fact. No prototype currently exists in the repository.

## Canonical contracts

- [Plans guide](../../plans/AGENTS.md).
- [Prototypes guide](../../../prototypes/AGENTS.md).
- [Plan 01 analytics prototype definition](../../plans/01-analytics/index.md).
