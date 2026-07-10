package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// GenerateDeploy writes the deploy/ directory (helm chart + Dockerfile) to the
// generated MCP server project. Placeholder __BINARY_NAME__ is replaced with the
// actual binary name throughout all files.
func (g *Generator) GenerateDeploy() error {
	binName := filepath.Base(g.outputDir)

	return fs.WalkDir(templatesFS, "deploy", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dir := filepath.Join(g.outputDir, path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("create dir %s: %w", dir, err)
			}
			return nil
		}

		data, err := templatesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}

		content := strings.ReplaceAll(string(data), "__BINARY_NAME__", binName)

		outPath := filepath.Join(g.outputDir, path)
		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
		return nil
	})
}
