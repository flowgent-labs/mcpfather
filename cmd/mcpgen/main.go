package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lyeslabs/mcpgen/internal/generator"
)

var (
	inputFile  string
	outputDir  string
	validation bool
	includes   string
	excludes   string
)

func init() {
	flag.StringVar(&inputFile, "i", "", "Path to the OpenAPI specification file (JSON or YAML)")
	flag.StringVar(&inputFile, "input", "", "Path to the OpenAPI specification file (JSON or YAML)")
	flag.StringVar(&outputDir, "o", "", "Path to the output MCP server directory")
	flag.StringVar(&outputDir, "output", "", "Path to the output MCP server directory")
	flag.BoolVar(&validation, "validation", false, "Enable OpenAPI validation")
	flag.StringVar(&includes, "I", "", "Comma-separated OpenAPI paths to include")
	flag.StringVar(&includes, "includes", "", "Comma-separated OpenAPI paths to include")
	flag.StringVar(&excludes, "e", "", "Comma-separated OpenAPI paths to exclude")
	flag.StringVar(&excludes, "excludes", "", "Comma-separated OpenAPI paths to exclude")
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: mcpgen [OPTIONS]

Options:
  -i, --input       Path to the OpenAPI specification file (JSON or YAML)
  -o, --output      Path to the output MCP server directory
  -I, --includes    Comma-separated OpenAPI paths to include
  -e, --excludes    Comma-separated OpenAPI paths to exclude
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

	// Parse include/exclude lists
	var includePaths, excludePaths []string
	if includes != "" {
		includePaths = strings.Split(includes, ",")
	}
	if excludes != "" {
		excludePaths = strings.Split(excludes, ",")
	}

	// Create the output directory if it doesn't exist
	if outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory '%s': %v\n", outputDir, err)
			os.Exit(1)
		}
	}

	gen, err := generator.NewGenerator(inputFile, validation, "", outputDir, includePaths, excludePaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating generator: %v\n", err)
		os.Exit(1)
	}

	// Generate the MCP server
	if err := gen.GenerateMCP(); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating MCP: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated MCP server in: %s\n\n", outputDir)
	fmt.Printf("To build and run:\n")
	fmt.Printf("  cd %s\n", outputDir)
	fmt.Printf("  go mod tidy\n")
	fmt.Printf("  go build -o %s .\n", filepath.Base(outputDir))
	fmt.Printf("  ./%s --transport http --port 8080\n", filepath.Base(outputDir))
}
