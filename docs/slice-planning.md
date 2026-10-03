# Plan a slice

The [roadmap](roadmap.md) orders the platform's capabilities. Before implementing a slice, write a plan folder under [`docs/plans/`](plans/AGENTS.md) and link its index from [index.md](index.md).
A plan is a folder of one or more plan files. Its index states their order and which may run concurrently; its `architecture.md` and `testing.md` carry the shared contracts and testing approach. The [plans guide](plans/AGENTS.md#plan-file-contents) maps the brief below onto those files.
Planning and implementation are separate tasks. Plan in [passes](plans/AGENTS.md#order-of-work): the architecture and its technical diagrams first, discussed back and forth with the user; then the plan files; then a [prototype pass](plans/AGENTS.md#prototype-pass) to agree whether a prototype is needed and what kind. A request to plan starts the architecture pass and stops for discussion. It does not produce prototype code. Robust implementation begins after the user signs off and instructs an agent to implement a plan file.

Start with the first analytics slice on Mac, aiming for one emitted event that the analytics helper CLI can retrieve. Identify its local build prerequisites and later Linux deployment checks in the brief. Broader query features can follow in later slices if needed to keep this outcome bounded.

## Required brief

1. **Outcome:** State the CLI action and observable result the completed slice provides, with a concrete acceptance example.
2. **Scope and prerequisites:** Name what is included, what is deferred, which earlier capabilities it needs, what the Mac agent can implement/test, when to use the Linux VM agent, and how source and acceptance evidence cross that boundary.
3. **Contracts and ownership:** Define module boundaries, input/output schemas, state transitions, errors, trust, persistence, and failure behavior; identify the owner of each invariant.
4. **Testing design:** Name the invariant IDs and independent oracles; specify MC/DC witness decisions, focused fuzz targets, and the existing suite to extend or distinct suite to create.
5. **Deterministic boundary:** Define generated state/actions/faults, controlled I/O/time/scheduling, decoder version, scenario limits, replay artifacts, and what simulation cannot establish; omit scheduling machinery when focused fuzzing suffices. Analytics modules use fuzzing and MC/DC only and state that here.
6. **Native and E2E evidence:** Specify real adapters and system boundaries that need separate checks, the software used, prerequisites, and explicit pass/fail criteria.
7. **Observability:** Define redacted events, correlation, metrics, helper-CLI inspection, enabled/disabled behavior, and failure/drop handling; specify explanatory comments and text art for tests and analytics.
8. **Ordered implementation steps:** Give a short dependency-ordered list of concrete changes and their checks, keeping each step within this slice.
9. **Acceptance and budgets:** State the evidence required for completion, ordinary-test commands once available, discovery limits, VM costs, and any permitted live-service spend.
10. **Decisions and status:** Resolve routine design choices with reasons; clearly list remaining blockers, assumptions, and planned versus verified behavior.
11. **Prototype and signoff:** Propose whether the work needs one prototype, one per plan file, several working together, or none, and define each; list the questions the user must answer, what a stand-in cannot show, and the recorded agreement and signoff.

## After implementation

Update the plan with actual verification evidence and meaningful scope changes. Update owned invariants, relevant docs, and the README configuration reference together with behavior.
Follow the project knowledge skill to save durable environment/tooling findings and engineering boundaries in the repo-local store, linking canonical plans and recording verification limits. Mark any prototype's README superseded rather than deleting it. Mark the slice complete only when its acceptance criteria are satisfied, and plan the next slice using what the completed work established.
