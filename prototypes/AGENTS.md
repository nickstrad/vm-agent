# Agent guide for prototypes/

`prototypes/` holds small runnable experiments that let the user judge an approach before robust work starts. Each prototype is a standalone folder that stays in the repository as reference for one aspect of the platform.
Prototypes are not platform code. Production code never imports them, and a passing demo is never acceptance evidence for a plan.

Follow the root agent guide and [the plans guide](../docs/plans/AGENTS.md#prototype-signoff-generalize-robust), which defines when a prototype is built and signed off.

## Layout

```text
prototypes/
  AGENTS.md            this guide (CLAUDE.md is a relative symlink to it)
  scaffold/            copyable starting point; holds the canonical shared files
  NN-<topic>/          one standalone prototype; NN matches its plan folder when it has one
    README.md          what it shows, what it does not, how to run, verification date
    go.mod             its own module
    main.go, demo.go   subcommands and the `demo` walkthrough
    demokit.go         copied from scaffold/
```

## Rules

- **Standalone.** A prototype folder builds and runs alone, with its own `go.mod` or Deno entry point. It imports nothing from another prototype, from `scaffold/`, or from the platform module. Share code by copying it.
- **Kept.** Do not delete a prototype when its plan completes. Mark it superseded in its README, with a link to the robust code, so it remains a readable reference for that approach.
- **Go or Deno for scripting.** Write prototype programs and their demo drivers in Go, or in Deno TypeScript when that is the shorter path. Use Bash only for glue that would be awkward otherwise, and keep it small.
- **SQLite for storage.** Use SQLite whatever the production store will be: one file, no service, and the user can open it with `sqlite3`. State in the README what the stand-in cannot show.
- **One command to judge it.** `go -C prototypes/NN-<topic> run . demo` (or `deno run` with explicit permissions) walks through numbered steps, prints each command before its output, shows failure paths as well as the happy path, and cleans up after itself. `KEEP=1` keeps the working directory.
- **Honest boundaries.** Keep authentication, redaction, and ownership real enough to judge. Leave out tests, invariant documents, and polish.
- **No secrets, no databases, no build output in Git.** Demos write to a temporary directory.
- **Record verification.** After running the demo, add the date and environment to the README. Do not claim an environment that was not run.

## Start a prototype

```bash
cp -R prototypes/scaffold prototypes/NN-<topic>     # from the repository root
```

Then rename the module in its `go.mod`, replace `main.go`, and write the README from the template below.

## Generalize step

After the user signs off a prototype, and before robust work starts, review it for anything worth reusing and move that into `scaffold/`:

1. List the helpers, scripts, and patterns the prototype needed that the next one would also need.
2. Generalize each one: remove topic-specific names, document it, and put the canonical copy in `scaffold/`.
3. Recopy an updated shared file into existing prototypes only when they benefit; rerun their demos if so.
4. Record what was generalized in the prototype's README and in the table below.
5. Save a durable finding through the [knowledge workflow](../docs/knowledge-base.md) when the prototype taught something about the environment or tooling.

If nothing is worth generalizing, say so in the README. The step is a review, not a quota.

| Shared file | Purpose | Came from |
| --- | --- | --- |
| `scaffold/demokit.go` | Demo runner: numbered steps, foreground and background subcommands, temporary state, secret files, optional external tools. | `01-analytics` |
| `scaffold/main.go` | Subcommand skeleton with a `demo` entry. | `01-analytics` |

A shared Deno entry point is not in the scaffold yet; add one the first time a prototype uses Deno.

## README template

```markdown
# <Topic> prototype

Status: awaiting signoff | signed off <date> | superseded by <link>
Plan: <link to docs/plans/NN-.../index.md>

## Run it
## What each step shows      (table: step, approach under review, plan file)
## What it does not show
## Generalized               (what moved to scaffold/, or "nothing")
## Verified                  (date, OS/arch, toolchain, what ran)
```

## Prototypes

| Folder | Aspect of the platform | Status |
| --- | --- | --- |
| [01-analytics](01-analytics/README.md) | Event envelope, collector as sole store owner, flag gate, redaction, bounded queue with counted drops, flush barrier, queries through the collector. | Awaiting signoff |
