# Prototype scaffold

Starting point for a new prototype. Copy it; do not import it.

```bash
cp -R prototypes/scaffold prototypes/NN-<topic>      # from the repository root
# then edit the module line in prototypes/NN-<topic>/go.mod
go -C prototypes/scaffold run . demo                 # see what the scaffold does
```

| File | Purpose |
| --- | --- |
| `main.go` | Subcommand skeleton with a `demo` walkthrough. Replace its contents. |
| `demokit.go` | Shared demo runner: steps, foreground and background commands, temporary state, secrets. Canonical copy. |
| `go.mod` | Standalone module with no dependencies. Add SQLite with `go get github.com/mattn/go-sqlite3` when the prototype stores data. |

Replace this README with the template in [the prototypes guide](../AGENTS.md#readme-template).

## Verified

2026-10-03, macOS (darwin/arm64), Go 1.26.4: `go vet` and `go run . demo` completed.
