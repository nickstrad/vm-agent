# Agent guide for docs/

`docs/` holds project documentation for people and agents. It contains no application code.

## Read and route

- Read [index.md](index.md) before editing documentation.
- Follow the root agent guide.
- Follow its knowledge workflow and the project [update-knowledge-store skill](../.agents/skills/update-knowledge-store/SKILL.md): consult relevant configured findings before substantive design changes and save durable findings through `kb` with search feedback and verification evidence.
- Keep the roadmap in [roadmap.md](roadmap.md). It describes ordered vertical slices and the extension boundary.
- Keep implementation and verification conventions in [testing.md](testing.md) and [coding-style.md](coding-style.md).
- Follow [slice-planning.md](slice-planning.md) to turn each roadmap slice into a bounded implementation brief before code. Add each slice plan to the index and keep unresolved decisions visible.
- Keep the knowledge-tool choice, copy provenance, repo-scoped invocation, and shared project skill paths in [knowledge-base.md](knowledge-base.md). The store lives in `knowledge/` and follows its own agent contract.

## Maintain structure

- Add a descriptive link to `index.md` for every new canonical Markdown document outside the knowledge corpus. Link the knowledge contract here; discover corpus entries through `kb`, without maintaining a second entry index.
- Save durable environment/tooling findings and engineering boundaries through the project skill when documentation work establishes them, including where Mac checks stop and Linux acceptance begins. Keep canonical plans in docs and link them from findings rather than copying whole plans.
- When moving or renaming a document, update incoming links and the index.
- Index canonical instruction files; their `CLAUDE.md` symlink aliases do not need duplicate entries.
- Do not copy unrelated findings, environment setup, agent roles, or unavailable skill requirements from reference repositories.
- Keep planned decisions distinct from completed behavior and verified evidence.

## Roadmap format

- Keep the architecture visual at the top and give every diagram component a matching section heading.
- Use ordered, one-line implementation steps grouped into working vertical slices.
- End each slice with its testing boundary, observable events, and CLI outcome.
- Keep an explicit Mac-first, split, or Linux-first callout in each slice, naming portable work and the Linux checks needed for acceptance.
- Keep browser, subagent, and eBPF work below the extension boundary.
- Carry testing and analytics requirements through each slice rather than postponing them to a final phase.

## Writing

- Write in active voice with short, concrete sentences and useful links.
- Explain why a boundary, test, or metric matters; avoid generic padding.
- Use tables for comparisons and text art for concrete states, decisions, and schedules.
- Keep generous learning commentary in testing and analytics code, while keeping the roadmap easy to skim.
- Update this `AGENTS.md` through its canonical path; `CLAUDE.md` points here.
