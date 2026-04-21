package generator

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxSearchDepth = 50
	goModFile      = "go.mod"
)

// BuildModuleName computes the module name from the output directory basename.
// e.g. outputDir "mymcpserver" → "mymcpserver.com"
func BuildModuleName(outputDir string) string {
	base := filepath.Base(filepath.Clean(outputDir))
	return base + ".com"
}

// BuildImportPath returns the import path for the mcptools package within the
// generated standalone project (e.g. "mymcpserver.com/internal/mcptools").
func BuildImportPath(outputDir string) (string, error) {
	moduleName := BuildModuleName(outputDir)
	return moduleName + "/internal/mcptools", nil
}

// BuildServerImportPath returns the import path for the mcpserver package.
func BuildServerImportPath(outputDir string) (string, error) {
	moduleName := BuildModuleName(outputDir)
	return moduleName + "/internal/mcpserver", nil
}

// GenerateGoMod creates a go.mod file in the output directory for the standalone project.
func GenerateGoMod(outputDir string) error {
	moduleName := BuildModuleName(outputDir)
	content := fmt.Sprintf("module %s\n\ngo 1.21\n\nrequire github.com/mark3labs/mcp-go v0.48.0\n", moduleName)

	goModPath := filepath.Join(outputDir, "go.mod")
	return os.WriteFile(goModPath, []byte(content), 0644)
}

// findModulePath searches upward from startDir to find the go.mod file
func findModulePath(startDir string) (string, string, error) {
	currentDir := filepath.Clean(startDir)
	for depth := 0; depth < maxSearchDepth; depth++ {
		goModPath := filepath.Join(currentDir, goModFile)

		// Check if go.mod exists
		if _, err := os.Stat(goModPath); err == nil {
			moduleName, err := parseModuleName(goModPath)
			if err != nil {
				return "", "", err
			}
			return moduleName, currentDir, nil
		} else if !os.IsNotExist(err) {
			return "", "", fmt.Errorf("error checking for go.mod: %w", err)
		}

		// Move up one directory
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			break // Reached filesystem root
		}
		currentDir = parent
	}

	return "", "", fmt.Errorf("no go.mod found in %d levels from %s",
		maxSearchDepth, startDir)
}

// parseModuleName extracts the module name from go.mod file
func parseModuleName(goModPath string) (string, error) {
	file, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to open %s: %w", goModPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module") {
			modulePart := strings.TrimSpace(strings.TrimPrefix(line, "module"))
			if i := strings.Index(modulePart, "//"); i != -1 {
				modulePart = strings.TrimSpace(modulePart[:i])
			}

			if len(modulePart) > 0 && (modulePart[0] == '"' || modulePart[0] == '\'') {
				quote := modulePart[0]
				end := strings.IndexByte(modulePart[1:], quote)
				if end == -1 {
					return "", fmt.Errorf("unclosed quote in module declaration in %s: %s", goModPath, line)
				}
				return modulePart[1 : end+1], nil
			}

			fields := strings.Fields(modulePart)
			if len(fields) == 0 {
				return "", fmt.Errorf("empty module name after 'module' keyword in %s: %s", goModPath, line)
			}
			return fields[0], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error scanning go.mod: %w", err)
	}

	return "", fmt.Errorf("module declaration not found in %s", goModPath)
}
