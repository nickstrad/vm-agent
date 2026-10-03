// Scaffold for a new prototype. Copy this folder, rename the module in go.mod,
// and replace the subcommands. See ../AGENTS.md.
//
// Convention: one binary, several subcommands, and a `demo` subcommand that
// walks the user through the approach by re-running the binary itself.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: proto demo|hello [name]")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "demo":
		runDemo()
	case "hello":
		name := "world"
		if len(os.Args) > 2 {
			name = os.Args[2]
		}
		fmt.Printf("hello, %s\n", name)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		os.Exit(1)
	}
}

func runDemo() {
	d := NewDemo("proto")
	defer d.Close()

	d.Step("Show the outcome the user should judge")
	d.Run("hello", "prototype")

	d.Step("Show a rejection or failure path as well as the happy path")
	d.Run("nope")

	d.Step("Leave something inspectable in the working directory")
	d.WriteFile(filepath.Join(d.Work, "note.txt"), "run with KEEP=1 to keep this directory\n", 0o600)
	d.Tool("ls", "-l", d.Work)
}
