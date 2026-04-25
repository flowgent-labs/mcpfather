package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/lyeslabs/mcpgen/internal/generator"
)

var (
	inputFile  string
	outputDir  string
	validation bool
	includes   string
)

func init() {
	flag.StringVar(&inputFile, "i", "", "Path to the OpenAPI specification file (JSON or YAML)")
	flag.StringVar(&inputFile, "input", "", "Path to the OpenAPI specification file (JSON or YAML)")
	flag.StringVar(&outputDir, "o", "", "Path to the output MCP server directory")
	flag.StringVar(&outputDir, "output", "", "Path to the output MCP server directory")
	flag.BoolVar(&validation, "validation", false, "Enable OpenAPI validation")
	flag.StringVar(&includes, "includes", "", "Comma-separated list of includes for the generated code")
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: mcpgen [OPTIONS]

Options:
  -i, --input       Path to the OpenAPI specification file (JSON or YAML)
  -o, --output      Path to the output MCP server directory
  --includes        Comma-separated list of includes for the generated code
  --validation      Enable OpenAPI validation
`)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --input is required")
		usage()
		os.Exit(1)
	}

	if outputDir == "" {
		fmt.Fprintln(os.Stderr, "Error: --output is required")
		usage()
		os.Exit(1)
	}

	// Create the output directory if it doesn't exist
	if outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory '%s': %v\n", outputDir, err)
			os.Exit(1)
		}
	}

	gen, err := generator.NewGenerator(inputFile, validation, "", outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating generator: %v\n", err)
		os.Exit(1)
	}

	// Generate the HTTP client (optional)
	if includes != "" {
		if err := gen.GenerateHTTPClient(strings.Split(includes, ",")); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating HTTP client: %v\n", err)
			os.Exit(1)
		}
	}

	// Generate the MCP server
	if err := gen.GenerateMCP(); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating MCP: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated MCP server in: %s\n", outputDir)
}
