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
		// Skip the github/ subdirectory — handled by GenerateGitHubCI.
		if d.IsDir() {
			if strings.Contains(path, "deploy/github") {
				return fs.SkipDir
			}
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

// GenerateGitHubCI writes community files to the generated MCP server project.
// Files under .github/ (workflows, ISSUE_TEMPLATE, PULL_REQUEST_TEMPLATE) are
// placed in .github/; root-level files (CONTRIBUTING, SECURITY, LICENSE, etc.)
// are placed at the project root.
// Placeholder __BINARY_NAME__ is replaced with the actual binary name.
func (g *Generator) GenerateGitHubCI() error {
	binName := filepath.Base(g.outputDir)

	return fs.WalkDir(templatesFS, "deploy/github", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := templatesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}

		content := strings.ReplaceAll(string(data), "__BINARY_NAME__", binName)

		rel := strings.TrimPrefix(path, "deploy/github/")

		// Route to .github/ or project root.
		// Files under workflows/, ISSUE_TEMPLATE/, and PULL_REQUEST_TEMPLATE.md
		// go under .github/. Everything else (CONTRIBUTING, SECURITY, LICENSE,
		// CODE_OF_CONDUCT) goes to the project root.
		var outPath string
		if strings.HasPrefix(rel, "workflows/") || strings.HasPrefix(rel, "ISSUE_TEMPLATE/") || rel == "PULL_REQUEST_TEMPLATE.md" {
			outPath = filepath.Join(g.outputDir, ".github", rel)
		} else {
			outPath = filepath.Join(g.outputDir, rel)
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", filepath.Dir(outPath), err)
		}
		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
		return nil
	})
}
