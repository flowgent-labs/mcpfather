package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to create a temporary go.mod file
func createTempGoMod(t *testing.T, dir string, content string) string {
	t.Helper()
	path := filepath.Join(dir, goModFile)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp go.mod: %v", err)
	}
	return path
}

func TestParseModuleName(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name          string
		content       string
		expectedName  string
		expectError   bool
		errorContains string
	}{
		{"simple", "module example.com/project\n\ngo 1.20", "example.com/project", false, ""},
		{"quoted", "module \"example.com/project/v2\"\n", "example.com/project/v2", false, ""},
		{"single quoted", "module 'example.com/project/v3'\n", "example.com/project/v3", false, ""},
		{"with comment", "module example.com/project // my module\n", "example.com/project", false, ""},
		{"quoted with comment", "module \"example.com/project\" // my module\n", "example.com/project", false, ""},
		{"extra spaces", "  module    example.com/another   \n", "example.com/another", false, ""},
		{"no module line", "go 1.20\nrequire other/thing v1.0.0", "", true, "module declaration not found"},
		{"empty file", "", "", true, "module declaration not found"},
		{"unclosed quote", "module \"example.com/project\n", "", true, "unclosed quote"},
		{"empty module name", "module \n", "", true, "empty module name after 'module' keyword"},
		{"module keyword only", "module\n", "", true, "empty module name after 'module' keyword"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			goModPath := createTempGoMod(t, tempDir, tc.content)
			name, err := parseModuleName(goModPath)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Error message %q does not contain %q", err.Error(), tc.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if name != tc.expectedName {
					t.Errorf("Expected module name %q, got %q", tc.expectedName, name)
				}
			}
		})
	}

	t.Run("file not found", func(t *testing.T) {
		_, err := parseModuleName(filepath.Join(tempDir, "nonexistent.mod"))
		if err == nil {
			t.Errorf("Expected error for non-existent file, got nil")
		} else if !strings.Contains(err.Error(), "failed to open") {
			t.Errorf("Error message %q does not contain 'failed to open'", err.Error())
		}
	})
}

func TestBuildModuleName(t *testing.T) {
	testCases := []struct {
		outputDir string
		expected  string
	}{
		{"mymcpserver", "mymcpserver"},
		{"myproject", "myproject"},
		{"/path/to/myapp", "myapp"},
		{"output", "output"},
	}

	for _, tc := range testCases {
		t.Run(tc.outputDir, func(t *testing.T) {
			result := BuildModuleName(tc.outputDir)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestBuildImportPath(t *testing.T) {
	testCases := []struct {
		name     string
		output   string
		expected string
	}{
		{"simple", "mymcpserver", "mymcpserver/internal/mcptools"},
		{"nested", "output/myserver", "myserver/internal/mcptools"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := BuildImportPath(tc.output)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestBuildServerImportPath(t *testing.T) {
	testCases := []struct {
		name     string
		output   string
		expected string
	}{
		{"simple", "mymcpserver", "mymcpserver/internal/mcpserver"},
		{"nested", "output/myserver", "myserver/internal/mcpserver"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := BuildServerImportPath(tc.output)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestGenerateGoMod(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "myserver")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	err := GenerateGoMod(outputDir)
	if err != nil {
		t.Fatalf("GenerateGoMod failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "go.mod"))
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	goModContent := string(content)
	if !strings.Contains(goModContent, "module myserver") {
		t.Errorf("Expected 'module myserver' in go.mod, got:\n%s", goModContent)
	}
	if !strings.Contains(goModContent, "github.com/mark3labs/mcp-go") {
		t.Errorf("Expected 'github.com/mark3labs/mcp-go' in go.mod, got:\n%s", goModContent)
	}
}
