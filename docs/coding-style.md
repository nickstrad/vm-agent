# Coding style

AI writes the implementation. These rules keep the platform correct, reviewable, and observable without rebuilding existing infrastructure or an agent harness.

## Reuse existing systems

- Use Linux tools for Linux responsibilities: systemd, OpenSSH, namespaces, cgroups, nspawn, seccomp, nftables, and QEMU/KVM.
- Use Go for custom CLI, API, policy, supervision, proxy, analytics, and connector code.
- Keep Pi in its supported runtime. Limit harness-native code to the integration hooks the platform requires.
- Choose dependencies and service boundaries for clear contracts and affordable operation.

## Make contracts visible

- Prefer explicit data types, state transitions, error handling, and ownership over hidden behavior.
- Keep trusted policy and credentials out of the agent runtime.
- Give each control state machine one owner and represent asynchronous I/O through requests and completion events.
- Supply injectable boundaries for nondeterminism needed by the relevant suite. Do not add a simulated scheduler to a stateless function.
- Keep production and simulation implementations behind the same contract. Doubles must obey that contract except where an explicit hostile-input test checks rejection.
- Return typed errors or sentinels when callers need to classify failures.
- Make cancellation, retries, resource limits, and ambiguous side effects explicit.

## Explain testing and analytics generously

- Write test names and bodies as executable specifications.
- Identify invariant IDs in test comments and keep an independent oracle separate from production decisions.
- Explain the generator, domain bounds, fault model, decision witness, and replay command.
- Use concrete ASCII state machines, timelines, truth tables, and event flows in package and helper comments.
- Explain why an assertion or telemetry field reveals a failure; do not merely restate the next line.
- Comment on what a simulation controls and what native or E2E checks still need to prove.
- Keep analytics event schemas versioned, redacted, correlated, and bounded.
- Make instrumentation observational: it must not consume simulation randomness or change business decisions.

## Finish changes coherently

- Update owned invariants, relevant tests, analytics contracts, and docs in the same change as behavior.
- Format Go code and run the checks appropriate to its boundary.
- Maintain reusable findings with evidence through the [knowledge CLI](knowledge-base.md), without browsing or editing corpus files directly.
