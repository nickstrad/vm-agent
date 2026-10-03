# Agent guide for prototypes/

`prototypes/` holds small runnable experiments that let the user judge an approach before robust work starts. It is empty until a plan's prototype definition is agreed.
Prototypes are not platform code. Production code never imports them, and a passing demo is never acceptance evidence for a plan.

Follow the root agent guide and [the plans guide](../docs/plans/AGENTS.md#prototype-pass), which says when a prototype is discussed and how it is defined, agreed, and signed off.

## Discuss before building

Do not create a prototype, a scaffold, or shared prototype tooling on your own initiative. What a prototype needs differs for each part of a plan, so the plan defines it first and the user agrees to it.

1. After the architecture and plan files are agreed, the plan's `index.md` proposes a prototype definition: the shape, and for each prototype its question, scope, language, storage, and what the user will judge.
2. The user agrees, changes, or rejects the definition. Record the answer and date in the index.
3. Only then build what was agreed, and nothing beyond it.

## Rules for an agreed prototype

- **Standalone.** Each prototype is its own folder, `prototypes/<plan-folder>-<topic>/`, that builds and runs alone. It imports nothing from another prototype or from the platform module. Prototypes that work together do so by running side by side, as their definition describes.
- **Kept.** Do not delete a prototype when its plan completes. Mark its README superseded, with a link to the robust code, so it stays a reference for that aspect of the platform.
- **Go or Deno.** Write prototype programs and their drivers in Go or Deno TypeScript. Use Bash where it is the simplest glue.
- **SQLite by default.** Use SQLite for storage unless the question being judged is about the production store itself. State in the README what a stand-in cannot show.
- **Sized to the question.** Build the least that answers the agreed question. Leave out tests, invariant documents, and polish, but keep any trust boundary under judgment real.
- **Runnable and recorded.** The README gives the commands to run, what each shows, what it does not show, and the date and environment where it was run. Do not claim an environment that was not run.
- **No secrets, databases, or build output in Git.**

## Generalizing shared tooling

Shared scaffolding or common scripts are worth having only once real prototypes show what repeats. There is none yet.

- After a prototype is signed off, note in its README anything the next prototype would likely need again.
- Propose extracting it to the user, naming what would move and where. Extract only after the user agrees.
- Do not generalize from a single prototype unless the user asks.

## Prototypes

None yet.
