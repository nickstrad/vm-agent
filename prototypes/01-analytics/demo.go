package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// runDemo walks through the approach under review. Each step maps to a signoff
// question in docs/plans/01-analytics/index.md; README.md has the table.
func runDemo() {
	d := NewDemo("proto")
	defer d.Close()
	dir := d.Work

	// Producer registry: who may speak for which module, with what trust, and
	// who may query. "stranger" holds a token the collector has never seen.
	d.WriteFile(filepath.Join(dir, "producers.json"), `{
  "cli":      {"modules": ["cli"],           "trust": "trusted",   "query": false},
  "runtime":  {"modules": ["agent"],         "trust": "untrusted", "query": false},
  "operator": {"modules": ["analytics-cli"], "trust": "trusted",   "query": true}
}`, 0o600)
	for _, name := range []string{"cli", "runtime", "operator", "stranger"} {
		d.WriteSecret(filepath.Join(dir, "tokens", name+".token"))
	}

	d.Step("Start the collector: sole owner of the database (queue=8, slow writer so drops show up later)")
	collector := d.Start("collector", "--dir", dir, "--queue", "8", "--write-delay", "5ms")
	d.WaitFor(filepath.Join(dir, "collector.sock"))

	d.Step("Flag gate: without --analytics or --debug nothing is sent")
	d.Run("emit", "--dir", dir, "--as", "cli", "--module", "cli", "--name", "message.sent", "--task", "t-demo-1")
	d.Run("flush", "--dir", dir, "--as", "operator")
	d.Run("events", "--dir", dir, "--as", "operator")

	d.Step("The first event: sent with --analytics; the secret attribute is removed before it leaves the producer")
	d.Run("emit", "--dir", dir, "--as", "cli", "--analytics", "--module", "cli", "--name", "message.sent", "--task", "t-demo-1",
		"--attr", "bytes=5", "--attr", "api_key=sk-live-SECRET-CANARY")
	d.Run("flush", "--dir", dir, "--as", "operator")
	d.Run("events", "--dir", dir, "--as", "operator")

	d.Step("Task timeline and an ad hoc read-only query, both served by the collector")
	d.Run("emit", "--dir", dir, "--as", "runtime", "--debug", "--module", "agent", "--name", "tool.finished", "--task", "t-demo-1",
		"--outcome", "error", "--attr", "reason=exit-1")
	d.Run("flush", "--dir", dir, "--as", "operator")
	d.Run("task", "--dir", dir, "--as", "operator", "t-demo-1")
	d.Run("query", "--dir", dir, "--as", "operator", "SELECT producer, source_trust, module, count(*) AS n FROM events GROUP BY 1, 2, 3")

	d.Step("Rejections: unknown token, a producer speaking for another module, a producer querying, a write through query")
	d.Run("emit", "--dir", dir, "--as", "stranger", "--analytics", "--module", "cli", "--name", "message.sent")
	d.Run("emit", "--dir", dir, "--as", "runtime", "--analytics", "--module", "sentinel", "--name", "decision.allow")
	d.Run("events", "--dir", dir, "--as", "runtime")
	d.Run("query", "--dir", dir, "--as", "operator", "DELETE FROM events RETURNING *")

	d.Step("Burst of 200 events into a queue of 8: the producer is never blocked; drops are counted")
	d.Run("emit", "--dir", dir, "--as", "cli", "--analytics", "--module", "cli", "--name", "burst", "--task", "t-burst", "--count", "200")
	d.Run("flush", "--dir", dir, "--as", "operator")
	d.Run("health", "--dir", dir, "--as", "operator")
	d.Run("query", "--dir", dir, "--as", "operator",
		"SELECT count(*) AS stored, min(source_seq) AS first_seq, max(source_seq) AS last_seq FROM events WHERE task_id = 't-burst'")

	d.Step("Collector failure: the caller reports a local drop and still exits 0")
	d.Stop(collector)
	code := d.Run("emit", "--dir", dir, "--as", "cli", "--analytics", "--module", "cli", "--name", "message.sent", "--task", "t-demo-2")
	fmt.Printf("exit status: %d\n", code)

	d.Step("Direct look at the file now that its owner has stopped: the canary must be absent")
	db := filepath.Join(dir, "analytics.sqlite")
	d.Tool("sqlite3", "-header", "-column", db, "SELECT ingest_seq, producer, name, attrs FROM events ORDER BY ingest_seq LIMIT 3")
	files, _ := filepath.Glob(db + "*")
	for _, f := range files {
		if raw, _ := os.ReadFile(f); bytes.Contains(raw, []byte("SECRET-CANARY")) {
			fmt.Printf("FAIL: canary found in %s\n", filepath.Base(f))
			os.Exit(1)
		}
	}
	fmt.Printf("ok: canary not present in %d database file(s)\n", len(files))
}
