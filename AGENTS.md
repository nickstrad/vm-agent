# Agent guide

## Project facts

- This is an AI-built personal agent platform inspired by Meta's Muse architecture.
- A CLI sends messages to one Linux VM, where an isolated agent performs work under host-side policy and returns observable results.
- The user learns to direct AI, design systems, evaluate tests, and inspect analytics. Do not simplify module correctness or complexity on the assumption that the user will hand-code it.
- Solo constraints concern cost and operational burden. Prefer one VM and existing infrastructure over additional hosted services.
- Use standard Linux software for service management, isolation, networking, and virtualization. Use Go for custom platform code.
- Reuse Pi through JSONL RPC with a Go adapter. Pi keeps its supported runtime. Route model traffic through trusted inference services to OpenRouter. fx is an alternative that must pass the same adapter contracts.
- Browser support, subagents, and eBPF tracking remain extensions below the roadmap's core boundary.
- The agent platform is currently planning and documentation. Its roadmap commands and suites are proposed; `tools/kb` is a copied upstream implementation, not an installed local tool.

## Read first

1. Read [docs/index.md](docs/index.md).
2. Read [docs/roadmap.md](docs/roadmap.md) for scope and ordering.
3. Before code or tests, read [docs/coding-style.md](docs/coding-style.md) and [docs/testing.md](docs/testing.md).
4. Follow scoped `AGENTS.md` files in the directories being changed.
5. Read the project [update-knowledge-store skill](.agents/skills/update-knowledge-store/SKILL.md) at session start and follow the knowledge workflow below throughout the task.
6. Before starting a roadmap slice, follow [docs/slice-planning.md](docs/slice-planning.md). A roadmap bullet is direction, not an implementation brief.

## Plan slices before implementation

- Write a bounded slice plan under `docs/` and add it to the index before implementing the slice. Follow the planning template and distinguish open decisions from settled contracts.
- Plan one working CLI outcome at a time, including necessary prerequisites, owned invariants, testing boundaries, observable events, acceptance evidence, and resource budgets.
- Treat a request to plan as documentation work. Do not begin implementation merely because a plan exists or the user asks whether the project is ready; wait for an instruction to implement that slice.
- Implement against the agreed slice plan once authorized. Update it when evidence changes scope, and report departures rather than silently expanding into later roadmap work.

## Knowledge workflow

- The complete skill and templates live in `.agents/skills/update-knowledge-store`; `.claude/skills/update-knowledge-store` is a relative symlink to the same package. Keep both agents on these shared instructions.
- This repo's store is `docs/knowledge`. Read [its contract](docs/knowledge/AGENTS.md) first and explicitly set `KB_ROOT` to its absolute path for every CLI invocation; never use the tool's computer-wide default. Follow [the invocation recipe](docs/knowledge-base.md#run-the-repo-scoped-cli).
- Search for relevant prior findings before substantive implementation, investigation, or environment/tooling advice; skip redundant searches when the current task already has relevant results.
- Use `kb search` with `--caller codex` or `--caller claude`, then `kb show` to read relevant entries. Use the CLI for all corpus discovery and maintenance, rather than filesystem searches or direct edits.
- Judge every search using its explicit search ID: mark useful or unhelpful hits, or record `--none` when nothing helped, including zero-hit searches. Avoid ambiguous `--last` feedback across concurrent sessions.
- Before finishing, save newly established durable, non-obvious findings about this project's environment, tooling, and engineering contracts. Mac/Linux execution boundaries, replay limitations, and repository-specific setup belong here; keep personal preferences and transient task progress out. Search before adding; revise existing entries where appropriate. Draft outside the corpus, import or edit through `kb`, and confirm the result with `kb show`.
- Record evidence and verification limits, never secrets or transient task progress. Keep architecture plans and implementation status in `docs/`.
- Run the copied CLI from `tools/kb`, using a temporary binary or `go run` with `fts5`; do not install it globally or change shell profiles. This store starts in full-text mode (`KB_EMBEDDER=none`); do not switch providers or models without a request.
- If unavailable, report the limitation once per session and continue independent work; do not browse the corpus or silently skip required knowledge maintenance.
- Follow the skill for diagnosis, index repair, and optional exports. Do not generate a corpus Markdown index unless asked; Git tracks entries and contracts, while `docs/knowledge/.kb/` stays local.

## Architecture rules

- Keep the agent runtime separate from trusted policy, credentials, state, and network services.
- Sentinel authorizes actions and outbound requests. Runtime code cannot grant itself permissions or access real credentials.
- Use explicit protocols, constrained state transitions, and authenticated caller identity. Reject invalid states at their boundary.
- Keep SQLite task state and security audit authoritative. Optional analytics does not authorize actions or affect correctness.
- Implement roadmap work as vertical slices with CLI outcomes, test evidence, and observable events. Preserve existing authorized scope.

## Testing and observability

- Treat testability and observability as design inputs for every change, following the Dropbox approach linked in the roadmap.
- Prefer Go native fuzzing. Corpus bytes deterministically generate scenario data, actions, faults, and scheduling choices.
- Expand an existing suite when a new module enlarges its boundary. Add a focused suite for a distinct invariant or contract; do not force every concern into a universal simulator.
- Use independent invariant oracles, explicit MC/DC decision witnesses, saved fuzz inputs, and reproducible scenario replay.
- State which sources of nondeterminism a suite controls. Validate native adapters, protocols, real Linux boundaries, and live model behavior with the appropriate separate checks.
- With debug or analytics enabled, every module emits correlated, redacted events to one collector owning `analytics.duckdb`. Never have service processes independently open that file for writing.
- Treat the analytics helper CLI as a first-class platform interface.
- Give test and analytics code generous explanatory comments and small ASCII diagrams covering state, timing, decisions, generators, oracles, and replay. Keep the explanations technically accurate.

## Invariant ownership

- A production module owns its promises in `invariants.md` beside its source.
- A shared simulation package owns its promises and limits in the `What it proves` section of its `architecture.md`; do not duplicate them in a second invariant file.
- Use stable IDs, one sentence per invariant, and a concrete fenced text-art example.
- State limits under `Not promised`. Link related simulation or native-adapter boundaries.
- Tests identify the invariant IDs they check in comments. Invariant documents do not maintain lists of test names.
- Update the owner document and checks together when behavior changes. Documentation may state planned promises before implementation, but must label that status explicitly.

## Documentation and changes

- Keep project docs in `docs/` and its index current.
- Follow the knowledge workflow and project skill consistently; tool installation and store initialization remain separate from exposing the skill.
- Preserve the distinction between proposed architecture and verified implementation.
- Keep the README environment-variable tables current when configuration changes. Distinguish supported settings from proposed platform variables.
- Keep the copied tool's source unchanged unless a tool change is explicitly in scope; its revision is recorded in [docs/knowledge-base.md](docs/knowledge-base.md).
- Keep `AGENTS.md` canonical. `CLAUDE.md` is a relative symlink to it; update the target rather than creating divergent instructions.
- Do not commit unless the user asks.

## Branches and worktrees

- Keep the root checkout on `main`. Do branch work in a Git worktree under `worktrees/`, one folder per branch, named after the branch.
- Start a branch with `git worktree add worktrees/<branch> -b <branch> main` from the repository root. Check out an existing branch with `git worktree add worktrees/<branch> <branch>`.
- Run the agent, builds, and checks for that branch inside its worktree folder. Do not switch the root checkout to a feature branch.
- `worktrees/` is tracked only through its `.gitignore`; everything else in it is ignored. Never commit a worktree's contents from the root checkout.
- Ignored local state is per checkout. In a worktree, set `KB_ROOT` to that worktree's `docs/knowledge` and run `kb reindex --all` before the first search.
- After a branch merges, remove its worktree with `git worktree remove worktrees/<branch>` and delete the branch. Use `git worktree list` to see what exists and `git worktree prune` to clear stale records.
- Commit directly on `main` only when the user asks for that, such as a small repository-protocol change.

## Verification

- Run checks appropriate to the changed boundary and report what was actually exercised.
- For documentation-only work, check relative links, indexes, symlinks, and whitespace.
- Once a Go module and code exist, use its documented checks, including relevant tests, `go vet`, formatting, corpus replay, and race checks where applicable.
- Run expensive fuzz discovery, VM E2E, and live-model evaluations within explicit resource and spending budgets.
- Do not report nonexistent commands or unavailable environments as passing checks.
