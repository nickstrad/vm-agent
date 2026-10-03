# Agent guide for docs/plans/

`docs/plans/` holds the implementation plans for roadmap work. One folder plans one [roadmap](../roadmap.md) section.
A plan is documentation. Writing one does not authorize implementation; wait for the user to say which plan file to execute.

Follow the root and [docs](../AGENTS.md) agent guides, [slice planning](../slice-planning.md), [testing](../testing.md), and [coding style](../coding-style.md). This guide adds the folder layout, the prototype stages, and the testing emphasis.

## Layout

```text
docs/plans/
  AGENTS.md                  this guide (CLAUDE.md is a relative symlink to it)
  NN-<roadmap-section>/      one plan; NN follows roadmap order
    index.md                 required: outcome, order, concurrency, status, decisions, budgets
    architecture.md          normal: modules and diagrams before/after, one change diagram per plan file
    testing.md               normal: the plan's testing approach, targets, tables, oracles, limits
    01-<cli-outcome>.md      1..n plan files, each one working CLI outcome
    02-<cli-outcome>.md
prototypes/NN-<roadmap-section>/   standalone prototype for that plan (outside docs/, kept as reference)
```

- A plan has between 1 and n plan files. Split when a section holds more than one CLI outcome or when parts can proceed concurrently. A one-file plan may fold `architecture.md` and `testing.md` into that file as sections with the same headings.
- Number plan files in their default execution order. Numbers are identifiers, not a promise of strict sequence; `index.md` owns the real ordering.
- Link every plan's `index.md` from [the docs index](../index.md). The plan index links its own files, so the docs index does not list them individually.

## index.md: ordering and concurrency

The index is the first file an agent reads and the only place that states execution order. It must contain:

1. **Outcome** of the whole plan and the roadmap items it covers.
2. **Plan files** in a table: file, CLI outcome, depends on, may run concurrently with, stage, and status.
3. **Order diagram** in text art showing dependencies, concurrent lanes, and signoff gates.
4. **Concurrency rules:** for each concurrent group, the packages or files each lane owns and the shared contracts that are frozen while lanes run. Two plan files may run concurrently only when they own disjoint source and share only contracts that are already signed off. A lane that needs a shared contract changed stops and updates `architecture.md` first.
5. **Signoff questions and decisions:** settled contracts with reasons, and open decisions the user must answer. Keep the two visibly separate.
6. **Budgets and acceptance** for the plan: fuzz discovery limits, VM cost, live-service spend, and the evidence that completes the plan.

Update the status column when a stage changes. Record signoff with its date and who gave it.

## architecture.md: before and after

Explain what exists before the plan, what exists after it, and what each plan file changes.

- **Before** and **after** diagrams of processes, packages, files on disk, and trust boundaries. Mark new, changed, and removed parts.
- **Module table:** each package or binary, its responsibility, the invariants it owns, and the plan file that creates or changes it.
- **Contracts:** wire formats, schemas, state transitions, errors, and limits shared between plan files. These are what concurrent lanes freeze.
- **Change diagram per plan file** when the plan has more than one file: the delta that file adds on top of the previous state. Each plan file links to its diagram instead of copying it.
- Label everything planned as planned. After implementation, correct the diagrams to match what was built.

## testing.md: lead with the testing approach

Testing design is the core of a plan, not an appendix. Write it before the implementation steps and make it specific enough to review without reading code.

- Name the evidence kinds the plan uses and the ones it deliberately does not use, with the reason.
- List every fuzz target with its input decoding, the property, and the independent oracle.
- List every compound decision with its condition IDs and the witness pairs MC/DC requires; note infeasible pairs.
- Name the invariant IDs each check covers and the owner document that will hold them.
- State what remains unverified and which later plan or Linux check covers it.
- Require explanatory comments and ASCII diagrams in test and analytics code, as [testing.md](../testing.md#analytics-and-explanation) describes.

### Which evidence a module needs

| Module kind | Required evidence | Not required |
| --- | --- | --- |
| Analytics tool and modules: event contract, emitter, collector, store, queries, helper CLI | Native Go fuzzing and MC/DC decision tables. | Deterministic simulation: no seeded scheduler, virtual time, replay-twice trace comparison, or `_dst_test.go` files. |
| Control-plane modules whose correctness depends on sequences, interleavings, faults, or recovery | Fuzzing, MC/DC, and deterministic simulation per the [roadmap](../roadmap.md). | — |
| Codecs, parsers, pure policy | Fuzzing, plus MC/DC where decisions are compound. | Simulation. |

Analytics is observational. It never authorizes an action or changes an outcome, and it may lose events as long as it counts them. That is why it is exempt from deterministic simulation.
The exemption is narrow:

- Each analytics fuzz target must still be deterministic for a given input, because Go fuzzing needs that to minimize and replay a failure.
- Stateful analytics behavior (queues, flush barriers, sessions) is fuzzed as operation sequences against a simple model oracle, asserting properties that hold under any interleaving.
- Other modules' suites still check that their business results and control traces are identical with analytics enabled and disabled. That check belongs to those modules, not to the analytics plan.

## Prototype, signoff, generalize, robust

Every plan file moves through four stages. The index records the current stage.

```text
prototype  ---->  signoff  ---->  generalize  ---->  robust
SQLite, Go or     user runs the   move reusable      production store and contracts,
Deno, one demo    demo, answers   tools and scripts  full fuzzing and MC/DC, invariants,
command           the signoff     into the           analytics events, acceptance
                  questions       prototype scaffold evidence
```

**Prototype.** Build the smallest runnable thing that lets the user judge the approach: the CLI outcome, the contract shapes, and the failure behavior. Follow [the prototypes guide](../../prototypes/AGENTS.md).

- Each prototype is a standalone folder under `prototypes/<plan-folder>/`, started from `prototypes/scaffold/`, with its own module, a `demo` command, and a README stating what it shows and what it does not.
- Write it in Go or Deno. Use Bash only for small glue.
- Use SQLite for prototype storage, whatever the production store will be. It is one file, needs no service, and the user can inspect it with `sqlite3`.
- Skip fuzz targets, MC/DC tables, and invariant documents. Do keep the trust boundaries honest: a prototype that fakes authentication or redaction proves nothing about the approach.
- One prototype may cover several plan files. A plan file that adds no new approach question, such as a deployment step, states "no prototype" with the reason.
- Run the demo and record the date and environment in the prototype README.

**Signoff.** The user runs the demo and answers the plan's signoff questions. Record each answer in `index.md`. Changes the user asks for go into the plan and `architecture.md` before any robust code.

**Generalize.** Review the signed-off prototype for tools, scripts, and patterns the next prototype would need, and move them into `prototypes/scaffold/` as the [generalize step](../../prototypes/AGENTS.md#generalize-step) describes. Record the result, including "nothing", in the prototype README.

**Robust.** Then implement the plan file to full standard under the production module.

- Write new code against the signed-off contracts. Do not copy prototype code into production; the prototype is a reference for behavior, not a starting point for source.
- Use the production store and adapters, and verify whatever the SQLite stand-in could not show. The plan names those gaps.
- Deliver the complete testing design, the owner `invariants.md` or `architecture.md` documents, observable events, and acceptance evidence.
- Keep the prototype. When the plan completes, mark its README superseded with a link to the robust code, so it stays a reference for that aspect of the platform.

Do not treat a prototype's demo as acceptance evidence, and do not start the robust stage on a plan file whose signoff is not recorded.

## Plan file contents

Each plan file is a bounded brief for one CLI outcome. Together with the plan's shared files it covers the [required brief](../slice-planning.md#required-brief):

| Brief item | Where it lives |
| --- | --- |
| Outcome | Each plan file; whole-plan outcome in `index.md`. |
| Scope and prerequisites, Mac and Linux split | Each plan file. |
| Contracts and ownership | `architecture.md`; each plan file lists what it adds. |
| Testing design, deterministic boundary | `testing.md`; each plan file lists its targets and tables. |
| Native and E2E evidence | Each plan file, with Mac checks and Linux checks named separately. |
| Observability | Each plan file. |
| Prototype and signoff | Each plan file; status in `index.md`. |
| Ordered implementation steps | Each plan file. |
| Acceptance and budgets | Each plan file; plan-wide budgets in `index.md`. |
| Decisions and status | `index.md`. |

## Maintenance

- Keep planned behavior distinct from verified behavior. After a plan file completes, add its actual evidence and note departures from the plan.
- When evidence changes scope, update the plan file, `architecture.md`, and the index together.
- Do not invent commands. Mark every command that does not exist yet as proposed.
- Save durable findings through the [knowledge workflow](../knowledge-base.md); link the plan rather than copying it.
