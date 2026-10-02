package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateHelpers_NoErrorOnSuccess(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "generator_helpers_test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Clean up the temporary directory

	// Instantiate the Generator with the temporary output directory
	g := &Generator{
		PackageName: "mytools", // Use a specific package name for the test
		outputDir:   tmpDir,
	}

	// Call the function under test
	err = g.GenerateHelpers()

	// Assert that no error occurred
	if err != nil {
		t.Errorf("GenerateHelpers returned an unexpected error: %v", err)
	}

	// Optional: You could still check for the *existence* of the file
	// to ensure the writeFileContent call was at least attempted.
	expectedFilePath := filepath.Join(tmpDir, "pkg", "helpers", "client.go")
	if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
		t.Errorf("expected generated file %s to exist, but it does not", expectedFilePath)
	}

	content, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatalf("read generated client: %v", err)
	}
	generated := string(content)
	for _, want := range []string{
		"httpMiddlewareRegistry = mustBuiltinHTTPMiddlewareRegistry()",
		"http.DefaultTransport.(*http.Transport).Clone()",
		"transport.MaxIdleConnsPerHost = config.MaxIdleConnsPerHost",
		"transport.MaxConnsPerHost = config.MaxConnsPerHost",
		"doResilientUpstreamRequest(ctx, pool, retry, req)",
	} {
		if !strings.Contains(generated, want) {
			t.Errorf("generated client missing %q", want)
		}
	}
	if got := strings.Count(generated, "doPooledUpstreamRequest(req)"); got != 4 {
		t.Errorf("default pooled client call count = %d, want 4", got)
	}
	if got := strings.Count(generated, "&http.Client"); got != 1 {
		t.Errorf("http.Client construction count = %d, want only the shared client factory", got)
	}
}
