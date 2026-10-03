// demokit: the shared demo runner for prototypes.
//
// The canonical copy lives in prototypes/scaffold/demokit.go. Each prototype
// carries its own copy so its folder stays standalone; improve the canonical
// copy during a prototype's "generalize" step, then recopy where useful.
//
// A demo is a numbered walkthrough that re-executes this same binary with
// different subcommands, printing each command before its output:
//
//	d := NewDemo("proto")
//	defer d.Close()
//	d.Step("Start the server")
//	srv := d.Start("serve", "--dir", d.Work)
//	d.Run("status", "--dir", d.Work)
//	d.Stop(srv)
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Demo struct {
	Work  string // temporary working directory, removed by Close unless KEEP=1
	name  string
	self  string
	steps int
	procs []*exec.Cmd
}

// NewDemo creates a short temporary directory. It sits directly under /tmp
// because Unix socket paths are limited to about 104 bytes on macOS.
func NewDemo(name string) *Demo {
	work, err := os.MkdirTemp("/tmp", "vma-"+name+".")
	must(err)
	self, err := os.Executable()
	must(err)
	return &Demo{Work: work, name: name, self: self}
}

func (d *Demo) Step(title string) {
	d.steps++
	fmt.Printf("\n=== %d. %s\n", d.steps, title)
}

// Run executes a subcommand of this binary in the foreground and returns its
// exit code. A non-zero exit is printed, not fatal: demos show rejections too.
func (d *Demo) Run(args ...string) int {
	fmt.Printf("$ %s %s\n", d.name, d.display(args))
	code := d.exec(exec.Command(d.self, args...))
	if code != 0 {
		fmt.Printf("(exit %d)\n", code)
	}
	return code
}

// Tool runs an external program if it is installed and reports whether it ran.
func (d *Demo) Tool(program string, args ...string) bool {
	path, err := exec.LookPath(program)
	if err != nil {
		fmt.Printf("(%s not installed; skipped)\n", program)
		return false
	}
	fmt.Printf("$ %s %s\n", program, d.display(args))
	d.exec(exec.Command(path, args...))
	return true
}

// Start launches a subcommand in the background, such as a server.
func (d *Demo) Start(args ...string) *exec.Cmd {
	fmt.Printf("$ %s %s &\n", d.name, d.display(args))
	cmd := exec.Command(d.self, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	must(cmd.Start())
	d.procs = append(d.procs, cmd)
	return cmd
}

func (d *Demo) Stop(cmd *exec.Cmd) {
	if cmd.ProcessState == nil {
		cmd.Process.Kill()
		cmd.Wait()
	}
}

// Close stops background processes and removes the working directory.
// Set KEEP=1 to keep it for inspection.
func (d *Demo) Close() {
	for _, p := range d.procs {
		d.Stop(p)
	}
	if os.Getenv("KEEP") == "1" {
		fmt.Printf("\nkept: %s\n", d.Work)
		return
	}
	os.RemoveAll(d.Work)
}

// WaitFor blocks until path exists, for example a server's socket.
func (d *Demo) WaitFor(path string) {
	for i := 0; i < 100; i++ {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	must(fmt.Errorf("timed out waiting for %s", path))
}

// WriteSecret writes a fresh random token readable only by the current user.
func (d *Demo) WriteSecret(path string) {
	b := make([]byte, 16)
	rand.Read(b)
	d.WriteFile(path, hex.EncodeToString(b)+"\n", 0o600)
}

func (d *Demo) WriteFile(path, content string, mode os.FileMode) {
	must(os.MkdirAll(filepath.Dir(path), 0o700))
	must(os.WriteFile(path, []byte(content), mode))
}

func (d *Demo) exec(cmd *exec.Cmd) int {
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		must(err)
	}
	return 0
}

// display shortens the working directory to $WORK and quotes spaced arguments.
func (d *Demo) display(args []string) string {
	out := make([]string, len(args))
	for i, a := range args {
		a = strings.ReplaceAll(a, d.Work, "$WORK")
		if strings.ContainsAny(a, " *()'") {
			a = `"` + a + `"`
		}
		out[i] = a
	}
	return strings.Join(out, " ")
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "demo:", err)
		os.Exit(1)
	}
}
