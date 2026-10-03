# 05 Linux deployment

Part of [plan 01](index.md). Roadmap item 1's deployment half and the slice's Linux checks. **Stage:** planned; no prototype, because deployment adds no new protocol or storage approach. Everything below is planned.

Two checkpoints:

- **05a** needs [01](01-first-event.md) and a ready Linux VM. It may run concurrently with 02, 03, and 04 as **lane C**.
- **05b** needs [04](04-retention-limits-failure.md) and repeats the failure checks.

## Outcome

The collector runs as a systemd service on the VM, and the helper CLI on the Mac reads it through SSH.

```text
vm$  systemctl status vm-agent-collector            # active (running)        (all proposed)
mac$ ssh -N -L "$HOME/.vm-agent/collector.sock:/run/vm-agent-collector/collector.sock" vm &
mac$ vm-agent-analytics --socket "$HOME/.vm-agent/collector.sock" events
```

## Scope

Included:

- Native Linux build of both binaries with cgo for the VM's architecture.
- `deploy/systemd/vm-agent-collector.service`: a dedicated system user, `StateDirectory` for the database, `RuntimeDirectory` for the socket, restart policy, and standard hardening options.
- A Linux-only peer-credential check on accepted connections, added to token authentication, in `*_linux.go` files.
- Install and removal notes. No installer script unless the user asks.

Deferred: other platform services and their identities, which arrive with the roadmap's CLI and VM API section.

**Mac:** cross-checking that portable packages still build and pass. **Linux VM agent:** everything else in this file. Sync source by commit; databases and tokens stay on their own machine.

## Change

See the [05 change diagram](architecture.md#05-linux-deployment). Portable packages do not change. The peer-credential check is an additional condition on session admission, so D2 in [testing.md](testing.md#d2-session-admission-collector-plan-file-01) gains a `peer_allowed` condition and one more witness row, evaluated as always true on other systems.

## Testing

Portable evidence is unchanged: the same fuzz corpus and MC/DC tables are replayed on Linux with `go test` and `go test -race`. The extended D2 table is the only new table.

The remaining checks are real-system observations, not suites. Each has a pass criterion and is recorded as a transcript:

| Check | Pass when | Checkpoint |
| --- | --- | --- |
| Build | Both binaries build natively on the VM and report the same schema version as the Mac build. | 05a |
| Corpus replay | `go test ./internal/analytics/...` and the race run pass on the VM. | 05a |
| Service start | The unit starts, creates the database under its state directory with owner-only permissions, and `events` answers locally. | 05a |
| Sole owner | Another user cannot read the database file or connect to the socket; a second collector instance fails to open the file. | 05a |
| Peer credentials | A connection from a user not in the registry is refused even with a valid token. | 05a |
| Remote read | The Mac helper CLI lists events through the SSH-forwarded socket. | 05a |
| Kill during ingest | After `SIGKILL` during a burst, systemd restarts the service, the database opens, and events stored before the kill are present. Lost events are within the unflushed window. | 05b |
| Full disk | With the state directory on a small dedicated filesystem filled to capacity, the collector goes degraded, `emit-test` still exits 0, and the service recovers after space is freed. | 05b |
| Stopped service | With the unit stopped, `emit-test --analytics` exits 0 and reports a local drop. | 05b |
| journald | Critical errors from the cases above appear in the unit's journal without tokens or event content. | 05b |

Simulation is not involved, and none of these is claimed from the Mac.

## Observability

No new events. Confirm that `health` reports the Linux build's version and that ingestion lag is sensible over the SSH-forwarded path.

## Steps

1. 05a: build on the VM; record toolchain requirements as a knowledge finding.
2. 05a: write the unit file and install notes; add the peer-credential check and the extended D2 table.
3. 05a: run the six 05a checks; paste transcripts.
4. 05b: run the four 05b checks after 04 merges; paste transcripts.
5. Update `architecture.md` and the knowledge entry on Mac and Linux boundaries with what was verified.

## Acceptance

Every row in the table passes on the VM with a recorded transcript, date, kernel, and distribution. Any row that cannot be run is listed as unverified with the reason; the plan is not complete while a row is unverified.

## Budgets

The existing lab VM only. The full-disk check uses a small loopback or tmpfs filesystem, never the VM's root filesystem.

## Evidence

None yet.
