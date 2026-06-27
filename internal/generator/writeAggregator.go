package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// aggregatorSrcPrefix is the import path used in the embedded aggregator source files.
const aggregatorSrcPrefix = "github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator"

// GenerateAggregator copies the aggregated tool engine source files into the
// generated project as internal/mcpaggregator/, rewriting import paths.
func (g *Generator) GenerateAggregator() error {
	destDir := filepath.Join(g.outputDir, "internal", "mcpaggregator")
	moduleName := BuildModuleName(g.outputDir)
	destImportPrefix := moduleName + "/internal/mcpaggregator"

	return fs.WalkDir(templatesFS, "mcpaggregator", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		content, err := templatesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded %s: %w", path, err)
		}

		// Rewrite import paths
		src := strings.ReplaceAll(string(content), aggregatorSrcPrefix, destImportPrefix)

		// Compute destination path: strip "mcpaggregator/" prefix
		relPath := strings.TrimPrefix(path, "mcpaggregator/")
		destPath := filepath.Join(destDir, relPath)

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", filepath.Dir(destPath), err)
		}
		if err := os.WriteFile(destPath, []byte(src), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", destPath, err)
		}
		return nil
	})
}
