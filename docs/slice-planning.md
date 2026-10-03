# Plan a slice

The [roadmap](roadmap.md) orders the platform's capabilities. Before implementing a slice, write a concrete brief here in `docs/` and link it from [index.md](index.md).
Planning and implementation are separate tasks. A request to plan produces a reviewable document; implementation begins when the user instructs an agent to implement that slice.

Start with the first analytics slice on Mac, aiming for one emitted event that the analytics helper CLI can retrieve. Identify its local build prerequisites and later Linux deployment checks in the brief. Broader query features can follow in later slices if needed to keep this outcome bounded.

## Required brief

1. **Outcome:** State the CLI action and observable result the completed slice provides, with a concrete acceptance example.
2. **Scope and prerequisites:** Name what is included, what is deferred, which earlier capabilities it needs, what the Mac agent can implement/test, when to use the Linux VM agent, and how source and acceptance evidence cross that boundary.
3. **Contracts and ownership:** Define module boundaries, input/output schemas, state transitions, errors, trust, persistence, and failure behavior; identify the owner of each invariant.
4. **Testing design:** Name the invariant IDs and independent oracles; specify MC/DC witness decisions, focused fuzz targets, and the existing suite to extend or distinct suite to create.
5. **Deterministic boundary:** Define generated state/actions/faults, controlled I/O/time/scheduling, decoder version, scenario limits, replay artifacts, and what simulation cannot establish; omit scheduling machinery when focused fuzzing suffices.
6. **Native and E2E evidence:** Specify real adapters and system boundaries that need separate checks, the software used, prerequisites, and explicit pass/fail criteria.
7. **Observability:** Define redacted events, correlation, metrics, helper-CLI inspection, enabled/disabled behavior, and failure/drop handling; specify explanatory comments and text art for tests and analytics.
8. **Ordered implementation steps:** Give a short dependency-ordered list of concrete changes and their checks, keeping each step within this slice.
9. **Acceptance and budgets:** State the evidence required for completion, ordinary-test commands once available, discovery limits, VM costs, and any permitted live-service spend.
10. **Decisions and status:** Resolve routine design choices with reasons; clearly list remaining blockers, assumptions, and planned versus verified behavior.

## After implementation

Update the plan with actual verification evidence and meaningful scope changes. Update owned invariants, relevant docs, and the README configuration reference together with behavior.
Follow the project knowledge skill to save durable environment/tooling findings and engineering boundaries in the repo-local store, linking canonical plans and recording verification limits. Mark the slice complete only when its acceptance criteria are satisfied, and plan the next slice using what the completed work established.
