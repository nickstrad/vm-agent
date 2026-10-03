// PROTOTYPE for plan docs/plans/01-analytics. Reference code, not platform code.
//
// It exists so the user can run the approach and sign it off before the
// robust implementation starts. It uses SQLite as a stand-in for DuckDB, has
// no fuzz targets, no MC/DC tables, and no invariant documents. Production
// code must never import it. See README.md for what it does and does not show.
//
// What the prototype demonstrates:
//
//	producer process                      collector process (sole DB owner)
//	----------------                      ---------------------------------
//	flag gate (--analytics | --debug)
//	      | off -> nothing leaves the process
//	      v on
//	redact attrs (allowlist)
//	      |
//	      |  JSONL over a Unix socket     authenticate token -> producer
//	      +-----------------------------> check producer may speak for module
//	                                      validate envelope, redact again
//	                                             |
//	                                      bounded queue (full -> drop + count)
//	                                             |
//	                                      single writer goroutine -> SQLite
//	                                             ^
//	helper CLI: events | task | query | health   |
//	      +-----------------------------> read-only handle, row cap
//
// The helper CLI never opens the database file. Every read goes through the
// collector, which is the property the DuckDB version depends on.
package main

import (
	"bufio"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"text/tabwriter"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	schemaVersion = 1
	maxLineBytes  = 16 << 10 // one wire frame; larger frames end the session
	maxFieldBytes = 128
	maxAttrBytes  = 256
	maxRows       = 1000
)

// Event is the proposed v1 envelope. The collector, not the producer, assigns
// producer identity and source trust, so those two fields are absent here.
type Event struct {
	SchemaVersion  int               `json:"schema_version"`
	EventID        string            `json:"event_id"`
	Module         string            `json:"module"`
	Name           string            `json:"name"`
	TaskID         string            `json:"task_id,omitempty"`
	ActionID       string            `json:"action_id,omitempty"`
	TraceID        string            `json:"trace_id,omitempty"`
	ParentEventID  string            `json:"parent_event_id,omitempty"`
	SourceInstance string            `json:"source_instance"`
	SourceSeq      uint64            `json:"source_seq"`
	LogicalTime    uint64            `json:"logical_time"`
	DurationNS     int64             `json:"duration_ns"`
	Outcome        string            `json:"outcome"`
	TestSeed       string            `json:"test_seed,omitempty"`
	Attrs          map[string]string `json:"attrs,omitempty"`
}

type Request struct {
	Op     string `json:"op"` // emit | flush | events | task | query | health
	Token  string `json:"token"`
	Event  *Event `json:"event,omitempty"`
	TaskID string `json:"task_id,omitempty"`
	SQL    string `json:"sql,omitempty"`
}

type Response struct {
	OK      bool             `json:"ok"`
	Error   string           `json:"error,omitempty"`
	Dropped bool             `json:"dropped,omitempty"`
	Columns []string         `json:"columns,omitempty"`
	Rows    [][]any          `json:"rows,omitempty"`
	Health  map[string]int64 `json:"health,omitempty"`
}

// attrAllowlist is the prototype's redaction rule: only these attribute keys
// survive. The robust version registers allowed keys per event name.
var attrAllowlist = map[string]bool{"bytes": true, "status": true, "reason": true, "note": true}

var outcomes = map[string]bool{"ok": true, "error": true, "denied": true, "timeout": true}

func redact(e *Event) (removed int) {
	for k, v := range e.Attrs {
		if !attrAllowlist[k] {
			delete(e.Attrs, k)
			removed++
		} else if len(v) > maxAttrBytes {
			e.Attrs[k] = v[:maxAttrBytes]
		}
	}
	return removed
}

func validate(e *Event) error {
	if e.SchemaVersion != schemaVersion {
		return fmt.Errorf("unsupported schema_version %d", e.SchemaVersion)
	}
	for name, v := range map[string]string{"event_id": e.EventID, "module": e.Module, "name": e.Name, "source_instance": e.SourceInstance} {
		if v == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	for name, v := range map[string]string{"event_id": e.EventID, "module": e.Module, "name": e.Name, "task_id": e.TaskID, "action_id": e.ActionID,
		"trace_id": e.TraceID, "parent_event_id": e.ParentEventID, "source_instance": e.SourceInstance, "test_seed": e.TestSeed} {
		if len(v) > maxFieldBytes {
			return fmt.Errorf("%s exceeds %d bytes", name, maxFieldBytes)
		}
	}
	if !outcomes[e.Outcome] {
		return fmt.Errorf("unknown outcome %q", e.Outcome)
	}
	if e.DurationNS < 0 {
		return errors.New("duration_ns is negative")
	}
	return nil
}

// ---------------------------------------------------------------- collector

type producer struct {
	Modules []string `json:"modules"`
	Trust   string   `json:"trust"`
	Query   bool     `json:"query"`
	token   string
}

type queued struct {
	ev       *Event
	producer string
	trust    string
	barrier  chan struct{} // non-nil: a flush marker, closed once the writer reaches it
}

type collector struct {
	write      *sql.DB
	read       *sql.DB // opened read-only, so `query` cannot modify the store
	ingest     chan queued
	producers  map[string]*producer
	writeDelay time.Duration

	accepted, written, droppedQueueFull, rejectedAuth, rejectedInvalid, redactedAttrs, storeErrors atomic.Int64
}

const schemaSQL = `CREATE TABLE IF NOT EXISTS events (
  ingest_seq      INTEGER PRIMARY KEY AUTOINCREMENT,
  schema_version  INTEGER NOT NULL,
  event_id        TEXT NOT NULL UNIQUE,
  producer        TEXT NOT NULL,
  source_trust    TEXT NOT NULL,
  module          TEXT NOT NULL,
  name            TEXT NOT NULL,
  task_id         TEXT, action_id TEXT, trace_id TEXT, parent_event_id TEXT,
  source_instance TEXT NOT NULL,
  source_seq      INTEGER NOT NULL,
  logical_time    INTEGER NOT NULL,
  duration_ns     INTEGER NOT NULL,
  outcome         TEXT NOT NULL,
  test_seed       TEXT,
  attrs           TEXT NOT NULL,
  received_at     TEXT NOT NULL
)`

func runCollector(args []string) error {
	fs := flag.NewFlagSet("collector", flag.ExitOnError)
	dir := fs.String("dir", "", "state directory holding producers.json, tokens/, the socket, and the database")
	queue := fs.Int("queue", 64, "ingest queue capacity; a full queue drops and counts")
	delay := fs.Duration("write-delay", 0, "artificial per-insert delay, to make queue drops visible")
	fs.Parse(args)

	c := &collector{ingest: make(chan queued, *queue), producers: map[string]*producer{}, writeDelay: *delay}
	raw, err := os.ReadFile(filepath.Join(*dir, "producers.json"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &c.producers); err != nil {
		return err
	}
	for name, p := range c.producers {
		tok, err := os.ReadFile(filepath.Join(*dir, "tokens", name+".token"))
		if err != nil {
			return err
		}
		p.token = strings.TrimSpace(string(tok))
	}

	dbPath := filepath.Join(*dir, "analytics.sqlite")
	if c.write, err = sql.Open("sqlite3", "file:"+dbPath+"?_journal_mode=WAL&_busy_timeout=5000"); err != nil {
		return err
	}
	c.write.SetMaxOpenConns(1)
	if _, err := c.write.Exec(schemaSQL); err != nil {
		return err
	}
	if c.read, err = sql.Open("sqlite3", "file:"+dbPath+"?mode=ro&_busy_timeout=5000"); err != nil {
		return err
	}

	sock := filepath.Join(*dir, "collector.sock")
	os.Remove(sock)
	ln, err := net.Listen("unix", sock)
	if err != nil {
		return err
	}
	os.Chmod(sock, 0o600)
	go c.writer()
	fmt.Fprintf(os.Stderr, "collector: listening on %s (queue=%d)\n", sock, *queue)
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go c.serve(conn)
	}
}

// writer is the only goroutine that modifies the store. Items arrive in queue
// order, so reaching a barrier means everything queued before it is resolved.
func (c *collector) writer() {
	for q := range c.ingest {
		if q.barrier != nil {
			close(q.barrier)
			continue
		}
		time.Sleep(c.writeDelay)
		e := q.ev
		if e.Attrs == nil {
			e.Attrs = map[string]string{}
		}
		attrs, _ := json.Marshal(e.Attrs)
		_, err := c.write.Exec(`INSERT INTO events (schema_version,event_id,producer,source_trust,module,name,task_id,action_id,trace_id,
			parent_event_id,source_instance,source_seq,logical_time,duration_ns,outcome,test_seed,attrs,received_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(event_id) DO NOTHING`,
			e.SchemaVersion, e.EventID, q.producer, q.trust, e.Module, e.Name, e.TaskID, e.ActionID, e.TraceID, e.ParentEventID,
			e.SourceInstance, e.SourceSeq, e.LogicalTime, e.DurationNS, e.Outcome, e.TestSeed, string(attrs), time.Now().UTC().Format(time.RFC3339Nano))
		if err != nil {
			c.storeErrors.Add(1)
			continue
		}
		c.written.Add(1)
	}
}

func (c *collector) authenticate(token string) (string, *producer) {
	// Sorted so the lookup order is stable; constant-time compare per candidate.
	names := make([]string, 0, len(c.producers))
	for n := range c.producers {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if p := c.producers[n]; p.token != "" && subtle.ConstantTimeCompare([]byte(p.token), []byte(token)) == 1 {
			return n, p
		}
	}
	return "", nil
}

func (c *collector) serve(conn net.Conn) {
	defer conn.Close()
	sc := bufio.NewScanner(conn)
	sc.Buffer(make([]byte, 0, 4096), maxLineBytes)
	enc := json.NewEncoder(conn)
	for sc.Scan() {
		var req Request
		if err := json.Unmarshal(sc.Bytes(), &req); err != nil {
			c.rejectedInvalid.Add(1)
			enc.Encode(Response{Error: "malformed request"})
			continue
		}
		enc.Encode(c.handle(req))
	}
}

func (c *collector) handle(req Request) Response {
	name, p := c.authenticate(req.Token)
	if p == nil {
		c.rejectedAuth.Add(1)
		return Response{Error: "unauthenticated"}
	}
	switch req.Op {
	case "emit":
		return c.emit(name, p, req.Event)
	case "flush":
		b := make(chan struct{})
		select {
		case c.ingest <- queued{barrier: b}:
			<-b
			return Response{OK: true}
		case <-time.After(5 * time.Second):
			return Response{Error: "flush timed out: queue stayed full"}
		}
	case "health":
		return Response{OK: true, Health: map[string]int64{
			"accepted": c.accepted.Load(), "written": c.written.Load(), "dropped_queue_full": c.droppedQueueFull.Load(),
			"rejected_auth": c.rejectedAuth.Load(), "rejected_invalid": c.rejectedInvalid.Load(),
			"redacted_attrs": c.redactedAttrs.Load(), "store_errors": c.storeErrors.Load(),
			"queue_len": int64(len(c.ingest)), "queue_cap": int64(cap(c.ingest)),
		}}
	}
	if !p.Query {
		c.rejectedAuth.Add(1)
		return Response{Error: fmt.Sprintf("producer %q may not query", name)}
	}
	const cols = "ingest_seq, producer, source_trust, module, name, task_id, outcome, source_seq, logical_time, attrs"
	switch req.Op {
	case "events":
		return c.query(fmt.Sprintf("SELECT %s FROM events ORDER BY ingest_seq DESC LIMIT 20", cols))
	case "task":
		return c.query(fmt.Sprintf("SELECT %s FROM events WHERE task_id = ? ORDER BY logical_time, ingest_seq LIMIT %d", cols, maxRows), req.TaskID)
	case "query":
		sqlText := strings.TrimRight(strings.TrimSpace(req.SQL), ";")
		return c.query(fmt.Sprintf("SELECT * FROM (%s) LIMIT %d", sqlText, maxRows))
	}
	return Response{Error: "unknown op"}
}

func (c *collector) emit(name string, p *producer, e *Event) Response {
	if e == nil {
		c.rejectedInvalid.Add(1)
		return Response{Error: "emit without event"}
	}
	allowed := false
	for _, m := range p.Modules {
		allowed = allowed || m == e.Module
	}
	if !allowed {
		c.rejectedAuth.Add(1)
		return Response{Error: fmt.Sprintf("producer %q may not emit for module %q", name, e.Module)}
	}
	if err := validate(e); err != nil {
		c.rejectedInvalid.Add(1)
		return Response{Error: err.Error()}
	}
	c.redactedAttrs.Add(int64(redact(e))) // defense in depth: producers redact first
	select {
	case c.ingest <- queued{ev: e, producer: name, trust: p.Trust}:
		c.accepted.Add(1)
		return Response{OK: true}
	default: // never block a producer: drop the newest event and count it
		c.droppedQueueFull.Add(1)
		return Response{OK: true, Dropped: true}
	}
}

func (c *collector) query(sqlText string, args ...any) Response {
	rows, err := c.read.Query(sqlText, args...)
	if err != nil {
		return Response{Error: err.Error()}
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	resp := Response{OK: true, Columns: cols}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return Response{Error: err.Error()}
		}
		for i, v := range vals {
			if b, ok := v.([]byte); ok {
				vals[i] = string(b)
			}
		}
		resp.Rows = append(resp.Rows, vals)
	}
	if err := rows.Err(); err != nil {
		return Response{Error: err.Error()}
	}
	return resp
}

// ------------------------------------------------------------------ clients

type client struct {
	conn  net.Conn
	rd    *bufio.Reader
	token string
}

func dial(dir, as string) (*client, error) {
	tok, err := os.ReadFile(filepath.Join(dir, "tokens", as+".token"))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout("unix", filepath.Join(dir, "collector.sock"), time.Second)
	if err != nil {
		return nil, err
	}
	return &client{conn: conn, rd: bufio.NewReader(conn), token: strings.TrimSpace(string(tok))}, nil
}

func (c *client) call(req Request) (Response, error) {
	req.Token = c.token
	var resp Response
	if err := json.NewEncoder(c.conn).Encode(req); err != nil {
		return resp, err
	}
	line, err := c.rd.ReadBytes('\n')
	if err != nil {
		return resp, err
	}
	return resp, json.Unmarshal(line, &resp)
}

func randomID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type attrFlags map[string]string

func (a attrFlags) String() string { return "" }
func (a attrFlags) Set(s string) error {
	k, v, ok := strings.Cut(s, "=")
	if !ok {
		return errors.New("want key=value")
	}
	a[k] = v
	return nil
}

func runEmit(args []string) error {
	fs := flag.NewFlagSet("emit", flag.ExitOnError)
	dir := fs.String("dir", "", "state directory")
	as := fs.String("as", "", "producer identity (reads tokens/<as>.token)")
	analytics := fs.Bool("analytics", false, "enable event emission")
	debug := fs.Bool("debug", false, "enable event emission (debug implies analytics)")
	module := fs.String("module", "", "module the event is about")
	name := fs.String("name", "", "event name")
	task := fs.String("task", "", "task ID")
	outcome := fs.String("outcome", "ok", "ok | error | denied | timeout")
	seed := fs.String("seed", "", "test seed or corpus ID")
	count := fs.Int("count", 1, "number of events to send on one connection")
	attrs := attrFlags{}
	fs.Var(attrs, "attr", "key=value attribute (repeatable)")
	fs.Parse(args)

	// The emission gate. With both flags off nothing is built, redacted, or sent.
	if !*analytics && !*debug {
		fmt.Println("emit: analytics disabled; nothing sent")
		return nil
	}
	c, err := dial(*dir, *as)
	if err != nil {
		// Analytics must never fail the caller: count the loss locally and move on.
		fmt.Printf("emit: collector unavailable (%v); %d event(s) dropped locally; caller continues\n", err, *count)
		return nil
	}
	defer c.conn.Close()
	instance := randomID()
	var sent, dropped, rejected int
	for i := 1; i <= *count; i++ {
		e := &Event{SchemaVersion: schemaVersion, EventID: randomID(), Module: *module, Name: *name, TaskID: *task,
			TraceID: instance, SourceInstance: instance, SourceSeq: uint64(i), LogicalTime: uint64(i),
			Outcome: *outcome, TestSeed: *seed, Attrs: map[string]string{}}
		for k, v := range attrs {
			e.Attrs[k] = v
		}
		removed := redact(e)
		resp, err := c.call(Request{Op: "emit", Event: e})
		switch {
		case err != nil:
			return err
		case !resp.OK:
			rejected++
			fmt.Printf("emit: rejected: %s\n", resp.Error)
		case resp.Dropped:
			dropped++
		default:
			sent++
		}
		if i == 1 && removed > 0 {
			fmt.Printf("emit: redacted %d attribute(s) before sending\n", removed)
		}
	}
	fmt.Printf("emit: queued=%d dropped_by_collector=%d rejected=%d\n", sent, dropped, rejected)
	return nil
}

func runRead(op string, args []string) error {
	fs := flag.NewFlagSet(op, flag.ExitOnError)
	dir := fs.String("dir", "", "state directory")
	as := fs.String("as", "", "caller identity (reads tokens/<as>.token)")
	fs.Parse(args)
	req := Request{Op: op}
	switch op {
	case "task":
		req.TaskID = fs.Arg(0)
	case "query":
		req.SQL = fs.Arg(0)
	}
	c, err := dial(*dir, *as)
	if err != nil {
		return err
	}
	defer c.conn.Close()
	resp, err := c.call(req)
	if err != nil {
		return err
	}
	if !resp.OK {
		return errors.New(resp.Error)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	defer tw.Flush()
	if resp.Health != nil {
		keys := make([]string, 0, len(resp.Health))
		for k := range resp.Health {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(tw, "%s\t%d\n", k, resp.Health[k])
		}
		return nil
	}
	if op == "flush" {
		fmt.Fprintln(tw, "flush: every event queued before this barrier is written or counted")
		return nil
	}
	fmt.Fprintln(tw, strings.Join(resp.Columns, "\t"))
	for _, row := range resp.Rows {
		cells := make([]string, len(row))
		for i, v := range row {
			cells[i] = fmt.Sprint(v)
		}
		fmt.Fprintln(tw, strings.Join(cells, "\t"))
	}
	fmt.Fprintf(tw, "(%d rows)\n", len(resp.Rows))
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: proto demo | collector|emit|flush|health|events|task|query [flags] [arg]")
		os.Exit(2)
	}
	var err error
	switch cmd, args := os.Args[1], os.Args[2:]; cmd {
	case "demo":
		runDemo()
	case "collector":
		err = runCollector(args)
	case "emit":
		err = runEmit(args)
	case "flush", "health", "events", "task", "query":
		err = runRead(cmd, args)
	default:
		err = fmt.Errorf("unknown command %q", cmd)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
