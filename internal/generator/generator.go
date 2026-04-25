package generator

import (
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/lyeslabs/mcpgen/internal/converter"
)

type Generator struct {
	specPath     string
	PackageName  string
	outputDir    string
	converter    converter.ConverterInterface
	spec         *openapi3.T
	includePaths []string
	excludePaths []string
	verbose      bool
}

func NewGenerator(specPath string, validation bool, packageName string, outputDir string, includePaths []string, excludePaths []string, verbose bool) (*Generator, error) {
	parser := converter.NewParser(validation)
	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] parsing OpenAPI spec: %s\n", specPath)
	}
	err := parser.ParseFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("error parsing OpenAPI specification: %w", err)
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] OpenAPI spec parsed successfully: %s\n", parser.GetDocument().Info.Title)
	}

	conv, err := converter.NewConverter(parser, includePaths, excludePaths, verbose)
	if err != nil {
		return nil, fmt.Errorf("error creating converter: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] converter initialized (includes=%d, excludes=%d)\n", len(includePaths), len(excludePaths))
	}

	return &Generator{
		specPath:     specPath,
		converter:    conv,
		spec:         parser.GetDocument(),
		outputDir:    outputDir,
		PackageName:  packageName,
		includePaths: includePaths,
		excludePaths: excludePaths,
		verbose:      verbose,
	}, nil
}
