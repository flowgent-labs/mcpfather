// Binary gen-config-dsl-schema generates the JSON Schema for
// mcpfather virtual tool configuration and writes it to disk.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/flowgent-labs/mcpfather/pkg/dslschema"
)

func main() {
	output := flag.String("output", "", "Path to write the schema JSON (default: stdout)")
	flag.Parse()

	if *output != "" {
		if err := dslschema.Write(*output); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write schema: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Schema written to %s\n", *output)
		return
	}

	b, err := dslschema.Encode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode schema: %v\n", err)
		os.Exit(1)
	}
	os.Stdout.Write(b)
	fmt.Println()
}
