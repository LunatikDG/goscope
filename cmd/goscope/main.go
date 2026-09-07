// Command goscope is the project's CLI entry point.
//
// Usage:
//
//	goscope trace <path-to-trace-file>
//
// trace parses a runtime/trace file (see cmd/traced) into steps, serves the
// same WASM demo `make serve` runs — pointed at that trace instead of the
// built-in pattern gallery — and opens it in the browser.
package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "trace" {
		usage()
		os.Exit(2)
	}

	if err := runTrace(os.Args[2]); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: goscope trace <path-to-trace-file>")
}
