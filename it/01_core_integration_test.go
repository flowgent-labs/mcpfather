package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const specFixture = "testdata/minimal_spec.yaml"

// testProxyEnv returns the proxy URL and env vars for build subcommands.
// It checks MCPGEN_TEST_PROXY first, then HTTPS_PROXY.
// If neither is set, no proxy is configured — suitable for environments
// where Go can reach module proxies directly.
func testProxyEnv(t *testing.T) (proxyURL string, envVars []string) {
	t.Helper()
	proxyURL = os.Getenv("MCPGEN_TEST_PROXY")
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTPS_PROXY")
	}
	if proxyURL == "" {
		t.Logf("[proxy] MCPGEN_TEST_PROXY and HTTPS_PROXY not set — build commands will use direct network")
		return "", nil
	}
	t.Logf("[proxy] MCPGEN_TEST_PROXY=%q HTTPS_PROXY=%q → using %q for build commands",
		os.Getenv("MCPGEN_TEST_PROXY"), os.Getenv("HTTPS_PROXY"), proxyURL)
	return proxyURL, []string{"HTTPS_PROXY=" + proxyURL}
}

// mcpgenBin returns the path to the mcpgen binary, building it if needed.
func mcpgenBin(t *testing.T) string {
	t.Helper()
	root, err := findRepoRoot()
	if err != nil {
		t.Fatalf("cannot find repo root: %v", err)
	}
	bin := filepath.Join(root, "bin", "mcpgen")
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		_, proxyEnv := testProxyEnv(t)
		cmd := exec.Command("make", "-C", root, "build")
		cmd.Env = append(os.Environ(), proxyEnv...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("make build failed: %v\n%s", err, out)
		}
	}
	return bin
}

// findRepoRoot walks up from the test file to find the repo root.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for depth := 0; depth < 20; depth++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found")
}

// genProject runs mcpgen and returns the output directory path.
func genProject(t *testing.T, includes, excludes string) string {
	t.Helper()
	return genProjectWithSpec(t, specFixture, includes, excludes)
}

// genProjectWithSpec runs mcpgen with a custom spec file (relative to its/).
func genProjectWithSpec(t *testing.T, specFile, includes, excludes string) string {
	t.Helper()
	bin := mcpgenBin(t)
	dir := t.TempDir()
	args := []string{"-i", filepath.Join(repoRoot(t), "it", specFile), "-o", dir}
	if includes != "" {
		args = append(args, "--includes", includes)
	}
	if excludes != "" {
		args = append(args, "--excludes", excludes)
	}
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpgen failed: %v\n%s", err, out)
	}
	return dir
}

func repoRoot(t *testing.T) string {
	t.Helper()
	r, err := findRepoRoot()
	if err != nil {
		t.Fatalf("cannot find repo root: %v", err)
	}
	return r
}

// buildServer runs go mod tidy + go build in the generated project dir.
func buildServer(t *testing.T, projectDir string) string {
	t.Helper()
	_, proxyEnv := testProxyEnv(t)
	binName := filepath.Base(projectDir)
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), proxyEnv...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, out)
	}
	cmd = exec.Command("go", "build", "-o", filepath.Join("bin", binName), ".")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), proxyEnv...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	return filepath.Join(projectDir, "bin", binName)
}

// mockUpstream starts an httptest server that records requests.
type mockUpstream struct {
	mu       sync.Mutex
	server   *httptest.Server
	requests []recordedRequest
}

type recordedRequest struct {
	Method        string
	URL           string
	Authorization string
	Headers       http.Header
	Body          []byte
}

func startMockUpstream(handler http.HandlerFunc) *mockUpstream {
	m := &mockUpstream{}
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		m.mu.Lock()
		m.requests = append(m.requests, recordedRequest{
			Method:        r.Method,
			URL:           r.URL.String(),
			Authorization: r.Header.Get("Authorization"),
			Headers:       r.Header.Clone(),
			Body:          body,
		})
		m.mu.Unlock()
		handler(w, r)
	}))
	return m
}

func (m *mockUpstream) Close() {
	m.server.Close()
}

func (m *mockUpstream) requestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

// okHandler returns a handler that writes a simple JSON response (no echo).
// This prevents sensitive headers from appearing in the response body at high verbosity.
func okHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}
}

// runCLI runs the generated server in CLI mode and returns stdout+stderr.
func runCLI(t *testing.T, binPath string, env []string, args ...string) (string, string) {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Env = append(os.Environ(), env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	_ = cmd.Run()
	return stdout.String(), stderr.String()
}

// ---------------------------------------------------------------------------
// 1. Generator CLI validation
// ---------------------------------------------------------------------------

func TestGenerator_Includes_NonExistentOperationId_Errors(t *testing.T) {
	bin := mcpgenBin(t)
	dir := t.TempDir()
	spec := filepath.Join(repoRoot(t), "it", specFixture)

	cmd := exec.Command(bin, "-i", spec, "-o", dir, "--includes", "nonExistentOp")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for non-existent operationId, got success")
	}
	if !strings.Contains(string(out), "nonExistentOp") {
		t.Errorf("error message should mention the bad operationId, got: %s", out)
	}
	if !strings.Contains(string(out), "does not exist") {
		t.Errorf("error message should say 'does not exist', got: %s", out)
	}
}

func TestGenerator_Excludes_NonExistentOperationId_Errors(t *testing.T) {
	bin := mcpgenBin(t)
	dir := t.TempDir()
	spec := filepath.Join(repoRoot(t), "it", specFixture)

	cmd := exec.Command(bin, "-i", spec, "-o", dir, "--excludes", "alsoFake")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for non-existent operationId, got success")
	}
	if !strings.Contains(string(out), "alsoFake") {
		t.Errorf("error message should mention the bad operationId, got: %s", out)
	}
}

func TestGenerator_ValidOperationId_Succeeds(t *testing.T) {
	bin := mcpgenBin(t)
	dir := t.TempDir()
	spec := filepath.Join(repoRoot(t), "it", specFixture)

	cmd := exec.Command(bin, "-i", spec, "-o", dir, "--includes", "echoHeaders,sayHello")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected success, got: %v\n%s", err, out)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "pkg", "mcptools", "*.go"))
	names := make(map[string]bool)
	for _, f := range files {
		names[filepath.Base(f)] = true
	}
	if !names["EchoHeaders.go"] {
		t.Error("expected EchoHeaders.go to be generated")
	}
	if !names["SayHello.go"] {
		t.Error("expected SayHello.go to be generated")
	}
	if names["DownloadReport.go"] {
		t.Error("DownloadReport.go should NOT be generated (not included)")
	}
}

// TestGenerator_VeryLongOperationId_Succeeds tests the common enterprise
// scenario where operationIds are extremely long with dash/underscore
// separators (e.g. auto-generated from API gateways). The generator must:
//  1. Convert to PascalCase correctly (dashes/underscores → word boundaries)
//  2. Truncate to ≤125 chars with a hash suffix to keep the Go identifier unique
//  3. Produce a buildable server where the tool is registered under its truncated name
func TestGenerator_VeryLongOperationId_Succeeds(t *testing.T) {
	longOpID := "get-a-very-long-operation-id-with-dashes-and_underscores_that_exceeds_the_maximum_tool_name_limit_set_by_opencode_and_other_mcp_integrations_in_the_enterprise_environment"
	spec := filepath.Join(repoRoot(t), "it", "testdata", "oas3.1_spec.yaml")
	if _, err := os.Stat(spec); os.IsNotExist(err) {
		t.Skipf("Blogs OAS 3.1 spec not found at %s", spec)
	}

	// Generate with just this long operationId
	bin := mcpgenBin(t)
	dir := t.TempDir()
	cmd := exec.Command(bin, "-i", spec, "-o", dir, "--includes", longOpID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpgen failed for very-long operationId: %v\n%s", err, out)
	}

	// The tool file should exist (with a truncated, hash-suffixed name).
	// Exclude registry.go which is always generated alongside tools.
	var toolFiles []string
	files, _ := filepath.Glob(filepath.Join(dir, "pkg", "mcptools", "*.go"))
	for _, f := range files {
		if filepath.Base(f) != "registry.go" {
			toolFiles = append(toolFiles, f)
		}
	}
	if len(toolFiles) != 1 {
		t.Fatalf("expected exactly 1 tool file, got %d: %v", len(toolFiles), toolFiles)
	}
	toolFileName := filepath.Base(toolFiles[0])
	toolName := strings.TrimSuffix(toolFileName, ".go")
	t.Logf("generated tool file: %s (name length: %d)", toolFileName, len(toolName))

	// Tool name must be ≤125 chars (MCP limit)
	if len(toolName) > 125 {
		t.Errorf("tool name %q is %d chars, exceeds 125-char limit", toolName, len(toolName))
	}
	// Must retain a recognisable prefix from the original operationId
	if !strings.HasPrefix(strings.ToLower(toolName), "getaverylong") {
		t.Errorf("tool name %q doesn't start with expected PascalCase prefix of original operationId", toolName)
	}

	// Build and smoke-test against mock upstream
	binPath := buildServer(t, dir)
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	stdout, _ := runCLI(t, binPath,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_TOKEN=test-token",
		},
		"-t", "cli", toolName, "--id=12345",
	)

	if !strings.Contains(stdout, `"status":"ok"`) {
		t.Errorf("expected upstream response, got: %s", stdout)
	}
	if len(mock.requests) == 0 {
		t.Fatal("no request reached mock upstream")
	}
}

// ---------------------------------------------------------------------------
// 2. Auth / token behaviour
// ---------------------------------------------------------------------------

func TestAuth_BasicPrefixPreserved(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, _ = runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_TOKEN=Basic myCredential123",
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Authorization; got != "Basic myCredential123" {
		t.Errorf("Authorization = %q, want %q", got, "Basic myCredential123")
	}
}

func TestAuth_BearerPrefixPreserved(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, _ = runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_TOKEN=Bearer secretToken999",
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Authorization; got != "Bearer secretToken999" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer secretToken999")
	}
}

func TestAuth_NoPrefixDefaultsToBearer(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, _ = runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_TOKEN=plainToken",
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Authorization; got != "Bearer plainToken" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer plainToken")
	}
}

func TestAuth_TokenFileFallback(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	tokenFile := filepath.Join(t.TempDir(), "my-token.txt")
	if err := os.WriteFile(tokenFile, []byte("fileToken123"), 0600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	_, _ = runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_TOKEN=",
			"MCP_UPSTREAM_TOKEN_FILE=" + tokenFile,
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Authorization; got != "Bearer fileToken123" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer fileToken123")
	}
}

func TestAuth_TokenFileWithBasicPrefix(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	tokenFile := filepath.Join(t.TempDir(), "my-token.txt")
	if err := os.WriteFile(tokenFile, []byte("Basic fileBasic123"), 0600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	_, _ = runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_TOKEN=",
			"MCP_UPSTREAM_TOKEN_FILE=" + tokenFile,
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Authorization; got != "Basic fileBasic123" {
		t.Errorf("Authorization = %q, want %q", got, "Basic fileBasic123")
	}
}

func TestAuth_CookieFromEnv(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, _ = runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_COOKIE=JSESSIONID=abc123",
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Headers.Get("Cookie"); got != "JSESSIONID=abc123" {
		t.Errorf("Cookie = %q, want %q", got, "JSESSIONID=abc123")
	}
}

func TestAuth_CookieFileFallback(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	cookieFile := filepath.Join(t.TempDir(), "my-cookie.txt")
	if err := os.WriteFile(cookieFile, []byte("JSESSIONID=fileSession456"), 0600); err != nil {
		t.Fatalf("failed to write cookie file: %v", err)
	}

	_, _ = runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_COOKIE=",
			"MCP_UPSTREAM_COOKIE_FILE=" + cookieFile,
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Headers.Get("Cookie"); got != "JSESSIONID=fileSession456" {
		t.Errorf("Cookie = %q, want %q", got, "JSESSIONID=fileSession456")
	}
}

func TestAuth_CookieAndTokenBothSet(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, _ = runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_TOKEN=Bearer secretToken999",
			"MCP_UPSTREAM_COOKIE=JSESSIONID=abc123",
		},
		"-t", "cli", "EchoHeaders",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if got := mock.requests[0].Authorization; got != "Bearer secretToken999" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer secretToken999")
	}
	if got := mock.requests[0].Headers.Get("Cookie"); got != "JSESSIONID=abc123" {
		t.Errorf("Cookie = %q, want %q", got, "JSESSIONID=abc123")
	}
}

// ---------------------------------------------------------------------------
// 3. Logging behaviour
// ---------------------------------------------------------------------------

// TestLogging_AuthHeaderRedactedByDefault verifies that at -v 10 the
// Authorization header VALUE is shown as "***" in the upstream request log.
// Uses okHandler so the response body does NOT echo the token back.
func TestLogging_AuthHeaderRedactedByDefault(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, stderr := runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_TOKEN=secretSauce",
		},
		"-t", "cli", "-v", "10", "EchoHeaders",
	)

	// The header line should show "Authorization: ***"
	if !strings.Contains(stderr, "Authorization: ***") {
		t.Error("expected 'Authorization: ***' in upstream request log, but not found. stderr:\n" + stderr)
	}
	// The raw token value must NOT appear as a header value in the upstream request log
	// (it appears as "Authorization: ***" not "Authorization: Bearer secretSauce")
	if strings.Contains(stderr, "Bearer secretSauce") {
		t.Error("Authorization value should be redacted, but 'Bearer secretSauce' appears in log. stderr:\n" + stderr)
	}
}

// TestLogging_AuthHeaderPrintedWhenEnvSet verifies that setting
// MCP_LOG_PRINT_AUTHORIZATION=true makes the Authorization value visible.
func TestLogging_AuthHeaderPrintedWhenEnvSet(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, stderr := runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_TOKEN=visibleToken",
			"MCP_LOG_PRINT_AUTHORIZATION=true",
		},
		"-t", "cli", "-v", "10", "EchoHeaders",
	)

	// With MCP_LOG_PRINT_AUTHORIZATION=true, the token should appear
	if !strings.Contains(stderr, "visibleToken") {
		t.Error("expected Authorization value to be visible when MCP_LOG_PRINT_AUTHORIZATION=true. stderr:\n" + stderr)
	}
}

// TestLogging_CookieRedactedByDefault verifies that the Cookie header value is
// shown as "***" in upstream request logs at -v 10.
func TestLogging_CookieRedactedByDefault(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, stderr := runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_COOKIE=JSESSIONID=secretSession",
		},
		"-t", "cli", "-v", "10", "EchoHeaders",
	)

	// The header line should show "Cookie: ***"
	if !strings.Contains(stderr, "Cookie: ***") {
		t.Error("expected 'Cookie: ***' in upstream request log, but not found. stderr:\n" + stderr)
	}
	// The raw cookie value must NOT appear
	if strings.Contains(stderr, "secretSession") {
		t.Error("Cookie value should be redacted, but 'secretSession' appears in log. stderr:\n" + stderr)
	}
}

// TestLogging_CookiePrintedWhenEnvSet verifies that setting
// MCP_LOG_PRINT_AUTHORIZATION=true makes the Cookie value visible.
func TestLogging_CookiePrintedWhenEnvSet(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, stderr := runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_COOKIE=JSESSIONID=visibleSession",
			"MCP_LOG_PRINT_AUTHORIZATION=true",
		},
		"-t", "cli", "-v", "10", "EchoHeaders",
	)

	// With MCP_LOG_PRINT_AUTHORIZATION=true, the cookie value should appear
	if !strings.Contains(stderr, "visibleSession") {
		t.Error("expected Cookie value to be visible when MCP_LOG_PRINT_AUTHORIZATION=true. stderr:\n" + stderr)
	}
}

// TestLogging_NonAuthHeadersPrinted verifies that non-Authorization headers are
// printed at high verbosity. For a GET request without body, we check the method
// and URL are logged.
func TestLogging_NonAuthHeadersPrinted(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	_, stderr := runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_UPSTREAM_TOKEN=someToken",
		},
		"-t", "cli", "-v", "10", "EchoHeaders",
	)

	// At verbosity >= 2, method and URL are logged
	if !strings.Contains(stderr, "GET "+mock.server.URL) {
		t.Error("expected upstream method and URL in verbose logs. stderr:\n" + stderr)
	}
}

// ---------------------------------------------------------------------------
// 4. Transport mode consistency (CLI vs HTTP)
// ---------------------------------------------------------------------------

// mcpHTTPCall sends an MCP JSON-RPC request via HTTP Streamable transport.
// It first calls initialize to get a session ID, then uses that for subsequent calls.
func mcpHTTPCall(t *testing.T, baseURL string, method string, params map[string]interface{}) (*http.Response, string) {
	t.Helper()

	// Step 1: initialize
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}
	body, _ := json.Marshal(initReq)
	resp, err := http.Post(baseURL+"/mcp", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("initialize request failed: %v", err)
	}
	sessionID := resp.Header.Get("Mcp-Session-Id")
	resp.Body.Close()

	// Step 2: send initialized notification
	if sessionID != "" {
		notifReq := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "notifications/initialized",
		}
		body, _ = json.Marshal(notifReq)
		req, _ := http.NewRequest("POST", baseURL+"/mcp", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Mcp-Session-Id", sessionID)
		r, err := http.DefaultClient.Do(req)
		if err == nil {
			r.Body.Close()
		}
	}

	// Step 3: send the actual request
	mcpReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  method,
		"params":  params,
	}
	body, _ = json.Marshal(mcpReq)
	req, _ := http.NewRequest("POST", baseURL+"/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("MCP %s request failed: %v", method, err)
	}
	return resp, sessionID
}

// waitForServer polls the MCP endpoint until the server responds or times out.
func waitForServer(t *testing.T, baseURL string) {
	t.Helper()
	for i := 0; i < 100; i++ {
		resp, err := http.Post(baseURL+"/mcp", "application/json", strings.NewReader(`{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"probe","version":"1"}}}`))
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("HTTP server did not become ready after 5s")
}

// TestAuth_HTTPTransportMatchesCLI verifies that the HTTP transport sends the
// same Authorization header as CLI mode when using MCP_UPSTREAM_TOKEN.
func TestAuth_HTTPTransportMatchesCLI(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	port := "19876"

	cmd := exec.Command(bin, "--transport", "http", "--port", port, "-v", "1")
	cmd.Env = append(os.Environ(),
		"MCP_UPSTREAM_ENDPOINT="+mock.server.URL,
		"MCP_UPSTREAM_TOKEN=Basic httpToken456",
	)
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start HTTP server: %v", err)
	}
	defer func() {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}()

	baseURL := "http://localhost:" + port
	waitForServer(t, baseURL)

	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "EchoHeaders",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()

	time.Sleep(200 * time.Millisecond)

	if len(mock.requests) == 0 {
		t.Fatalf("no request reached the mock upstream. stderr:\n%s", stderrBuf.String())
	}
	if got := mock.requests[0].Authorization; got != "Basic httpToken456" {
		t.Errorf("HTTP transport: Authorization = %q, want %q", got, "Basic httpToken456")
	}
}

// TestLogging_HTTPTransportRedactsAuthByDefault verifies that the HTTP
// transport also redacts Authorization in upstream request logs by default.
func TestLogging_HTTPTransportRedactsAuthByDefault(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders", ""))
	port := "19878"

	cmd := exec.Command(bin, "--transport", "http", "--port", port, "-v", "10")
	cmd.Env = append(os.Environ(),
		"MCP_UPSTREAM_ENDPOINT="+mock.server.URL,
		"MCP_UPSTREAM_TOKEN=shouldBeHidden",
	)
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start HTTP server: %v", err)
	}
	defer func() {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}()

	baseURL := "http://localhost:" + port
	waitForServer(t, baseURL)

	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "EchoHeaders",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()

	time.Sleep(200 * time.Millisecond)

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "Authorization: ***") {
		t.Error("expected 'Authorization: ***' in HTTP transport upstream logs. stderr:\n" + stderr)
	}
	if strings.Contains(stderr, "shouldBeHidden") {
		t.Error("token value should NOT appear in logs. stderr:\n" + stderr)
	}
}

// ---------------------------------------------------------------------------
// 5. Binary download
// ---------------------------------------------------------------------------

func TestDownload_BinaryFileSavedLocally(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=report.pdf")
		w.Write([]byte("fake-binary-pdf-content"))
	})
	defer mock.Close()

	bin := buildServer(t, genProject(t, "downloadReport", ""))
	downloadDir := filepath.Join(t.TempDir(), "downloads")
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		t.Fatalf("failed to create download dir: %v", err)
	}

	stdout, _ := runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_SERVER_DOWNLOAD_DIR=" + downloadDir,
		},
		"-t", "cli", "DownloadReport",
	)

	if !strings.Contains(stdout, "Saved to:") {
		t.Fatalf("expected 'Saved to:' in stdout, got: %s", stdout)
	}

	entries, err := os.ReadDir(downloadDir)
	if err != nil {
		t.Fatalf("cannot read download dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no files in download directory")
	}
	found := false
	for _, e := range entries {
		if strings.Contains(e.Name(), "report") {
			found = true
			data, _ := os.ReadFile(filepath.Join(downloadDir, e.Name()))
			if string(data) != "fake-binary-pdf-content" {
				t.Errorf("downloaded content = %q, want %q", string(data), "fake-binary-pdf-content")
			}
		}
	}
	if !found {
		t.Error("downloaded report file not found")
	}
}

func TestDownload_NoContentDisposition_UsesDefaultName(t *testing.T) {
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write([]byte("fake-zip-content"))
	})
	defer mock.Close()

	bin := buildServer(t, genProject(t, "downloadReport", ""))
	downloadDir := filepath.Join(t.TempDir(), "downloads")
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		t.Fatalf("failed to create download dir: %v", err)
	}

	stdout, _ := runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_SERVER_DOWNLOAD_DIR=" + downloadDir,
		},
		"-t", "cli", "DownloadReport",
	)

	if !strings.Contains(stdout, "Saved to:") {
		t.Fatalf("expected 'Saved to:' in stdout, got: %s", stdout)
	}
	// When no Content-Disposition is set, DetermineFileName falls back to
	// the URL path last segment ("download" from /download endpoint).
	if !strings.Contains(stdout, "download") {
		t.Errorf("expected filename derived from URL path or content-type, got: %s", stdout)
	}

	entries, err := os.ReadDir(downloadDir)
	if err != nil {
		t.Fatalf("cannot read download dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no files in download directory")
	}
	found := false
	for _, e := range entries {
		if strings.Contains(e.Name(), "download") {
			found = true
			data, _ := os.ReadFile(filepath.Join(downloadDir, e.Name()))
			if string(data) != "fake-zip-content" {
				t.Errorf("downloaded content = %q, want %q", string(data), "fake-zip-content")
			}
		}
	}
	if !found {
		t.Error("downloaded file not found")
	}
}

// ---------------------------------------------------------------------------
// 5b. Real binary download (external endpoint)
// ---------------------------------------------------------------------------

// TestDownload_BinaryWithKnownSize tests binary download with a known
// content size. Uses a local mock so the test works without internet access.
func TestDownload_BinaryWithKnownSize(t *testing.T) {
	// 1024 bytes of deterministic pseudo-binary content
	payload := make([]byte, 1024)
	for i := range payload {
		payload[i] = byte(i % 256)
	}
	mock := startMockUpstream(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=random.bin")
		w.Write(payload)
	})
	defer mock.Close()

	bin := buildServer(t, genProjectWithSpec(t, "testdata/binary_spec.yaml", "downloadBytes", ""))
	downloadDir := filepath.Join(t.TempDir(), "downloads")
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		t.Fatalf("failed to create download dir: %v", err)
	}

	stdout, _ := runCLI(t, bin,
		[]string{
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
			"MCP_SERVER_DOWNLOAD_DIR=" + downloadDir,
		},
		"-t", "cli", "DownloadBytes",
	)

	if !strings.Contains(stdout, "Saved to:") {
		t.Fatalf("expected 'Saved to:' in stdout, got: %s", stdout)
	}

	entries, err := os.ReadDir(downloadDir)
	if err != nil {
		t.Fatalf("cannot read download dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no files in download directory")
	}
	for _, e := range entries {
		info, _ := e.Info()
		t.Logf("Downloaded file: %s (%d bytes)", e.Name(), info.Size())
		if info.Size() != 1024 {
			t.Errorf("expected 1024 bytes, got %d", info.Size())
		}
		data, _ := os.ReadFile(filepath.Join(downloadDir, e.Name()))
		if len(data) != 1024 {
			t.Errorf("expected 1024 bytes on disk, got %d", len(data))
		}
	}
}

// ---------------------------------------------------------------------------
// 6. CLI argument passing
// ---------------------------------------------------------------------------

func TestCLI_QueryParamsPassedToUpstream(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "sayHello", ""))
	_, _ = runCLI(t, bin,
		[]string{"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL},
		"-t", "cli", "SayHello", "--name=World",
	)

	if len(mock.requests) == 0 {
		t.Fatal("no request reached the mock upstream")
	}
	if !strings.Contains(mock.requests[0].URL, "name=World") {
		t.Errorf("query param 'name=World' not found in URL: %s", mock.requests[0].URL)
	}
}

func TestCLI_ListShowsTools(t *testing.T) {
	bin := buildServer(t, genProject(t, "echoHeaders,sayHello", ""))
	stdout, _ := runCLI(t, bin, nil, "-t", "cli", "list")

	if !strings.Contains(stdout, "EchoHeaders") {
		t.Error("expected EchoHeaders in tool list")
	}
	if !strings.Contains(stdout, "SayHello") {
		t.Error("expected SayHello in tool list")
	}
}

// ---------------------------------------------------------------------------
// 7. Cyclic $ref detection (regression: LinkGroup.groups → LinkGroup)
// ---------------------------------------------------------------------------

const cyclicSpecFixture = "testdata/cyclic_spec.yaml"

// TestCyclicRef_GenerationSucceeds verifies that mcpgen does NOT hang when the
// OpenAPI spec contains a self-referencing schema (LinkGroup.groups → LinkGroup).
// Before the cycle-detection fix, the recursive schema walkers would recurse
// infinitely and the process would OOM or hang.
func TestCyclicRef_GenerationSucceeds(t *testing.T) {
	bin := mcpgenBin(t)
	dir := t.TempDir()
	spec := filepath.Join(repoRoot(t), "it", cyclicSpecFixture)

	cmd := exec.Command(bin, "-i", spec, "-o", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpgen failed for cyclic spec: %v\n%s", err, out)
	}

	// Both tools should be generated
	expected := []string{"ListItems.go", "HealthCheck.go"}
	for _, name := range expected {
		fp := filepath.Join(dir, "pkg", "mcptools", name)
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			t.Errorf("expected tool file %s was not generated", name)
		}
	}
}

// TestCyclicRef_ResponseTemplateHasCyclicMarker verifies that the generated
// response template for a cyclic schema contains the [cyclic reference] marker.
func TestCyclicRef_ResponseTemplateHasCyclicMarker(t *testing.T) {
	bin := mcpgenBin(t)
	dir := t.TempDir()
	spec := filepath.Join(repoRoot(t), "it", cyclicSpecFixture)

	cmd := exec.Command(bin, "-i", spec, "-o", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpgen failed: %v\n%s", err, out)
	}

	// Read the ListItems tool file which has the cyclic LinkGroup schema
	toolFile := filepath.Join(dir, "pkg", "mcptools", "ListItems.go")
	data, err := os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("failed to read ListItems.go: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "[cyclic reference]") {
		t.Error("expected '[cyclic reference]' marker in ListItems.go response template, but not found")
	}
}

// TestCyclicRef_NonCyclicSchemaNoSpuriousMarker verifies that non-cyclic schemas
// do NOT get a false-positive [cyclic reference] marker. The HealthCheck tool uses
// a simple HealthStatus schema with no self-references.
func TestCyclicRef_NonCyclicSchemaNoSpuriousMarker(t *testing.T) {
	bin := mcpgenBin(t)
	dir := t.TempDir()
	spec := filepath.Join(repoRoot(t), "it", cyclicSpecFixture)

	cmd := exec.Command(bin, "-i", spec, "-o", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcpgen failed: %v\n%s", err, out)
	}

	// Read the HealthCheck tool file which uses a flat HealthStatus schema
	toolFile := filepath.Join(dir, "pkg", "mcptools", "HealthCheck.go")
	data, err := os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("failed to read HealthCheck.go: %v", err)
	}
	content := string(data)

	if strings.Contains(content, "[cyclic reference]") {
		t.Error("HealthCheck.go should NOT contain '[cyclic reference]' — false positive for acyclic schema")
	}

	// The response template should still describe the status and uptime fields
	if !strings.Contains(content, "status") {
		t.Error("expected 'status' field in HealthCheck response template")
	}
	if !strings.Contains(content, "uptime") {
		t.Error("expected 'uptime' field in HealthCheck response template")
	}
}

// TestCyclicRef_BuildsAndRuns verifies that a server generated from a cyclic spec
// builds successfully and can invoke a tool against a mock upstream at runtime.
func TestCyclicRef_BuildsAndRuns(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	binPath := buildServer(t, genProjectWithSpec(t, cyclicSpecFixture, "", ""))
	stdout, _ := runCLI(t, binPath,
		[]string{"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL},
		"-t", "cli", "HealthCheck",
	)

	if !strings.Contains(stdout, `"status":"ok"`) {
		t.Errorf("expected upstream response, got: %s", stdout)
	}
	if len(mock.requests) == 0 {
		t.Fatal("no request reached mock upstream")
	}
}

// TestRegression_MinimalSpecResponseTemplate is a regression test ensuring the
// cycle-detection changes did not alter the output for non-cyclic schemas. The
// response template for echoHeaders must contain its usual structure.
func TestRegression_MinimalSpecResponseTemplate(t *testing.T) {
	dir := genProject(t, "echoHeaders", "")

	toolFile := filepath.Join(dir, "pkg", "mcptools", "EchoHeaders.go")
	data, err := os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("failed to read EchoHeaders.go: %v", err)
	}
	content := string(data)

	// The response template should describe the response structure
	if !strings.Contains(content, "# API Response Information") {
		t.Error("expected '# API Response Information' header in response template")
	}
	if !strings.Contains(content, "**Status Code:** 200") {
		t.Error("expected '**Status Code:** 200' in response template")
	}
	if !strings.Contains(content, "**Content-Type:** application/json") {
		t.Error("expected '**Content-Type:** application/json' in response template")
	}
	// Must NOT have spurious cyclic markers
	if strings.Contains(content, "[cyclic reference]") {
		t.Error("EchoHeaders.go should NOT contain '[cyclic reference]' — regression in cycle detection")
	}
}

// TestRegression_SayHelloRequestSchema verifies the request arg schema for a
// tool with query parameters is still generated correctly after cycle-detection
// changes (visited map is threaded through requestArgsSchema path).
func TestRegression_SayHelloRequestSchema(t *testing.T) {
	dir := genProject(t, "sayHello", "")

	toolFile := filepath.Join(dir, "pkg", "mcptools", "SayHello.go")
	data, err := os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("failed to read SayHello.go: %v", err)
	}
	content := string(data)

	// The InputSchema must describe the "name" query parameter.
	// The JSON schema is embedded as an escaped Go string literal,
	// so we match the escaped form: \"name\" and \"type\": \"string\".
	if !strings.Contains(content, `\"name\"`) {
		t.Errorf("expected 'name' property in InputSchema, content:\n%s", content)
	}
	if !strings.Contains(content, `\"type\": \"string\"`) {
		t.Errorf("expected 'type: string' in InputSchema for name parameter, content:\n%s", content)
	}
}

// TestRegression_FullBuildAndCLI verifies the full end-to-end flow still works:
// generate → build → CLI invoke with the minimal spec. This is the broadest
// regression smoke test for the cycle-detection changes.
func TestRegression_FullBuildAndCLI(t *testing.T) {
	mock := startMockUpstream(okHandler())
	defer mock.Close()

	bin := buildServer(t, genProject(t, "echoHeaders,sayHello", ""))
	stdout, _ := runCLI(t, bin,
		[]string{"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL},
		"-t", "cli", "SayHello", "--name=RegressionTest",
	)

	if !strings.Contains(stdout, `"status":"ok"`) {
		t.Errorf("expected upstream response, got: %s", stdout)
	}
	if len(mock.requests) == 0 {
		t.Fatal("no request reached mock upstream")
	}
	if !strings.Contains(mock.requests[0].URL, "name=RegressionTest") {
		t.Errorf("query param 'name=RegressionTest' not found in URL: %s", mock.requests[0].URL)
	}
}

func TestE2E_Core_AuthForwarding(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "test-bearer-token-abc123", "")
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	data := mustJSON(t, result)

	auth, _ := data["authorization"].(string)
	if !strings.Contains(auth, "test-bearer-token-abc123") {
		t.Errorf("Authorization header should contain token, got: %q", auth)
	}
	if !strings.HasPrefix(auth, "Bearer ") {
		t.Errorf("Authorization should have Bearer prefix, got: %q", auth)
	}
	if data["status"] != "ok" {
		t.Errorf("status = %v, want ok", data["status"])
	}
}

// ---------------------------------------------------------------------------
// Core Test 2: Bearer prefix is not duplicated
// ---------------------------------------------------------------------------

func TestE2E_Core_AuthNoDoublePrefix(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "Bearer already-prefixed-token", "")
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	data := mustJSON(t, result)

	auth, _ := data["authorization"].(string)
	if strings.Count(auth, "Bearer") > 1 {
		t.Errorf("Bearer prefix should not be duplicated, got: %q", auth)
	}
	if !strings.Contains(auth, "already-prefixed-token") {
		t.Errorf("Token should be preserved, got: %q", auth)
	}
}

// ---------------------------------------------------------------------------
// Core Test 3: Cookie forwarding via MCP_UPSTREAM_COOKIE
// ---------------------------------------------------------------------------

func TestE2E_Core_CookieForwarding(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoHeadersScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "JSESSIONID=abc123test")
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	data := mustJSON(t, result)
	headers, ok := data["headers"].(map[string]interface{})
	if !ok {
		t.Fatal("headers not found in response")
	}
	cookie, _ := headers["Cookie"].(string)
	if !strings.Contains(cookie, "JSESSIONID=abc123test") {
		t.Errorf("Cookie header should contain JSESSIONID, got: %q", cookie)
	}
}

// ---------------------------------------------------------------------------
// Core Test 4: MCP session ID forwarding is always-on
// ---------------------------------------------------------------------------

func TestE2E_Core_SessionForwardingDisabledByDefault(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoHeadersScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	// enable_mcp_session_in_forwarding is deprecated and always-on;
	// X-MCP-Session-ID is always forwarded regardless of config.
	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	data := mustJSON(t, result)
	headers, _ := data["headers"].(map[string]interface{})

	if _, ok := headers["X-Mcp-Session-Id"]; !ok {
		t.Error("X-Mcp-Session-Id should always be forwarded (behavior is always-on)")
	}
}

// ---------------------------------------------------------------------------
// Core Test 5: MCP-Session-Id in client request should not leak to upstream
// ---------------------------------------------------------------------------

func TestE2E_Core_SessionNotLeaked(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoHeadersScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// The MCP server receives Mcp-Session-Id from the client, but should NOT
	// forward it to the upstream.
	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{})
	data := mustJSON(t, result)
	headers, _ := data["headers"].(map[string]interface{})

	if _, ok := headers["Mcp-Session-Id"]; ok {
		t.Error("Mcp-Session-Id should NOT be forwarded to upstream")
	}
}

// ---------------------------------------------------------------------------
// Core Test 6: Content-type handling — JSON response passes through
// ---------------------------------------------------------------------------

func TestE2E_Core_ContentTypeJSON(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterContentTypeScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{
		"format": "json",
	})
	data := mustJSON(t, result)

	if tv, _ := data["type"].(string); tv != "json" {
		t.Errorf("expected JSON response type 'json', got %q", tv)
	}
	// Verify the data field is present too
	if dv, _ := data["data"].(string); dv != "json-response" {
		t.Errorf("expected data field 'json-response', got %q", dv)
	}
}

// ---------------------------------------------------------------------------
// Core Test 7: Binary content-type detection
// ---------------------------------------------------------------------------

func TestE2E_Core_BinaryContentType(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterContentTypeScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "downloadReport", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	result := callNativeTool(t, baseURL, "DownloadReport", map[string]interface{}{})
	// Binary download should be saved to file, result contains "Saved to:"
	if !strings.Contains(result, "Saved to:") {
		t.Errorf("binary download should save to file, got: %s", trimMsg(result, 200))
	}
}

// ---------------------------------------------------------------------------
// Core Test 8: Path parameter substitution — no scientific notation
// ---------------------------------------------------------------------------

func TestE2E_Core_PathParamSubstitution(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterPathParamScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{
		"name": "test-user",
		"age":  float64(30),
	})
	data := mustJSON(t, result)

	qn, _ := data["queryName"].(string)
	if qn != "test-user" {
		t.Errorf("query param name = %q, want test-user", qn)
	}
	qa, _ := data["queryAge"].(string)
	if qa != "30" {
		t.Errorf("query param age = %q, want 30", qa)
	}
	// Float64 should NOT become scientific notation
	if strings.Contains(qa, "e") || strings.Contains(qa, "E") {
		t.Errorf("query param age should not be in scientific notation: %q", qa)
	}
}

// ---------------------------------------------------------------------------
// Core Test 9: Upstream error handling (4xx)
// ---------------------------------------------------------------------------

func TestE2E_Core_UpstreamErrorHandling(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterErrorScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name": "EchoHeaders",
		"arguments": map[string]interface{}{
			"status": "400",
		},
	})
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	// Upstream returns 400, MCP server should propagate the error
	if !strings.Contains(string(bodyBytes), "Bad Request") {
		t.Logf("error response body: %s", string(bodyBytes))
	}
}

// ---------------------------------------------------------------------------
// Core Test 10: XML content-type is treated as text (not binary)
// ---------------------------------------------------------------------------

func TestE2E_Core_ContentTypeXML(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterContentTypeScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders", "")
	homeDir := t.TempDir()

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	result := callNativeTool(t, baseURL, "EchoHeaders", map[string]interface{}{
		"format": "xml",
	})
	// XML response should be returned as text (not saved to file)
	if strings.Contains(result, "Saved to:") {
		t.Error("XML should be treated as text, not binary download")
	}
	// Should contain XML content
	if !strings.Contains(result, "root") && !strings.Contains(result, "status") {
		t.Errorf("XML response should contain content, got: %s", trimMsg(result, 200))
	}
}

// ---------------------------------------------------------------------------
// Core Test 11: Chained native tools via virtual config
// ---------------------------------------------------------------------------

func TestE2E_Core_ChainedNativeTools(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	virtConfig := `
virtualTools:
  - name: virt_chain
    description: Chain echo and greet
    inputSchema:
      type: object
      properties:
        name:
          type: string
      required:
        - name
    pipeline:
      - id: echo
        kind: call
        spec:
          tool: EchoHeaders
          args: {}
      - id: greet
        kind: call
        spec:
          tool: SayHello
          args:
            name: $input.name
      - id: done
        kind: return
        spec:
          from: $greet
`
	writeCoreVirtualConfig(t, homeDir, binaryName, virtConfig)

	cleanup, baseURL := startVirtualTestServer(t, dir, mock.server.URL, homeDir)
	defer cleanup()

	result := mcpCallVirtualTool(t, baseURL, "virt_chain", map[string]interface{}{
		"name": "Alice",
	})

	data := mustJSON(t, result)
	greeting, _ := data["greeting"].(string)
	if !strings.Contains(greeting, "Alice") {
		t.Errorf("greeting should mention Alice, got %q", greeting)
	}
}

// ---------------------------------------------------------------------------
// Core Test 12: Config-based tool activation — backward compatibility
// ---------------------------------------------------------------------------

// TestConfig_NoConfigFile_AllToolsEnabled verifies that when no config.yaml
// exists, all native tools are registered and callable (backward compatible).
func TestConfig_NoConfigFile_AllToolsEnabled(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	// Intentionally do NOT write any config file

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// All 3 tools should be callable
	for _, toolName := range []string{"EchoHeaders", "SayHello", "DownloadReport"} {
		resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      toolName,
			"arguments": map[string]interface{}{},
		})
		resp.Body.Close()
	}

	// Also verify via CLI list
	binPath := buildServer(t, dir)
	stdout, _ := runCLI(t, binPath,
		[]string{
			"HOME=" + homeDir,
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
		},
		"-t", "cli", "list",
	)
	for _, name := range []string{"EchoHeaders", "SayHello", "DownloadReport"} {
		if !strings.Contains(stdout, name) {
			t.Errorf("CLI list should contain %q when no config file exists", name)
		}
	}
}

// TestConfig_AllNativeToolsByDefault_True_WithExcludes verifies that when
// all-native-tools-by-default is true, all tools are available except those
// listed in excludes.
func TestConfig_AllNativeToolsByDefault_True_WithExcludes(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	cfg := `
tools:
  expose:
    all-native-tools-by-default: true
    excludes:
      - DownloadReport
`
	writeCoreVirtualConfig(t, homeDir, binaryName, cfg)

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// EchoHeaders and SayHello should be callable
	for _, toolName := range []string{"EchoHeaders", "SayHello"} {
		resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      toolName,
			"arguments": map[string]interface{}{},
		})
		resp.Body.Close()
	}

	// DownloadReport should NOT be callable (excluded)
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "DownloadReport",
		"arguments": map[string]interface{}{},
	})
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(bodyBytes), "not found") && !strings.Contains(string(bodyBytes), "unknown tool") {
		t.Errorf("DownloadReport should be excluded, but got response: %s", trimMsg(string(bodyBytes), 300))
	}

	// CLI list should not show DownloadReport
	binPath := buildServer(t, dir)
	stdout, _ := runCLI(t, binPath,
		[]string{
			"HOME=" + homeDir,
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
		},
		"-t", "cli", "list",
	)
	if !strings.Contains(stdout, "EchoHeaders") {
		t.Error("CLI list should contain EchoHeaders")
	}
	if !strings.Contains(stdout, "SayHello") {
		t.Error("CLI list should contain SayHello")
	}
	if strings.Contains(stdout, "DownloadReport") {
		t.Error("CLI list should NOT contain excluded DownloadReport")
	}
}

// TestConfig_AllNativeToolsByDefault_False_WithIncludes verifies that when
// all-native-tools-by-default is false (the default), only tools listed in
// includes are available.
func TestConfig_AllNativeToolsByDefault_False_WithIncludes(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	cfg := `
tools:
  expose:
    all-native-tools-by-default: false
    includes:
      - EchoHeaders
      - SayHello
`
	writeCoreVirtualConfig(t, homeDir, binaryName, cfg)

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// EchoHeaders and SayHello should be callable
	for _, toolName := range []string{"EchoHeaders", "SayHello"} {
		resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      toolName,
			"arguments": map[string]interface{}{},
		})
		resp.Body.Close()
	}

	// DownloadReport should NOT be callable (not in includes)
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "DownloadReport",
		"arguments": map[string]interface{}{},
	})
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(bodyBytes), "not found") && !strings.Contains(string(bodyBytes), "unknown tool") {
		t.Errorf("DownloadReport should not be available, but got response: %s", trimMsg(string(bodyBytes), 300))
	}
}

// TestConfig_AllNativeToolsByDefault_False_Default_EmptyIncludes_NoTools
// verifies that when all-native-tools-by-default is false (or omitted) and
// includes is empty, no native tools are registered.
func TestConfig_AllNativeToolsByDefault_False_Default_EmptyIncludes_NoTools(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	cfg := `
tools:
  expose:
    all-native-tools-by-default: false
    includes: []
`
	writeCoreVirtualConfig(t, homeDir, binaryName, cfg)

	binPath := buildServer(t, dir)
	stdout, _ := runCLI(t, binPath,
		[]string{
			"HOME=" + homeDir,
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
		},
		"-t", "cli", "list",
	)

	// No native tools should be listed
	if strings.Contains(stdout, "EchoHeaders") {
		t.Error("CLI list should NOT contain EchoHeaders when all-native-tools-by-default is false and includes is empty")
	}
	if strings.Contains(stdout, "SayHello") {
		t.Error("CLI list should NOT contain SayHello when all-native-tools-by-default is false and includes is empty")
	}
}

// TestConfig_IncludesAndExcludes_Conflict_ServerFailsToStart verifies that
// when a tool appears in both includes and excludes, the server exits with
// an error at startup.
func TestConfig_IncludesAndExcludes_Conflict_ServerFailsToStart(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	cfg := `
tools:
  expose:
    all-native-tools-by-default: true
    includes:
      - EchoHeaders
    excludes:
      - EchoHeaders
`
	writeCoreVirtualConfig(t, homeDir, binaryName, cfg)

	binPath := buildServer(t, dir)

	// Start the server — it should fail due to conflict
	port := fmt.Sprintf("%d", 19000+(time.Now().UnixNano()%1000))
	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "1")
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"MCP_UPSTREAM_ENDPOINT="+mock.server.URL,
	)
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	err := cmd.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	// Wait for the process to exit (it should exit quickly with an error)
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case waitErr := <-done:
		if waitErr == nil {
			t.Error("server should have exited with an error due to config conflict, but exited successfully")
		}
		stderr := stderrBuf.String()
		if !strings.Contains(stderr, "EchoHeaders") {
			t.Errorf("stderr should mention the conflicting tool name. stderr:\n%s", stderr)
		}
		if !strings.Contains(stderr, "both includes and excludes") || !strings.Contains(stderr, "includes and excludes") {
			t.Errorf("stderr should mention includes/excludes conflict. stderr:\n%s", stderr)
		}
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		cmd.Wait()
		t.Fatal("server did not exit within 5 seconds — expected immediate failure on config conflict")
	}
}

// TestConfig_IncludesAndExcludes_NoConflict_Success verifies that includes
// and excludes can be used together without conflict, and the resulting
// tool set is correct.
func TestConfig_IncludesAndExcludes_NoConflict_Success(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	cfg := `
tools:
  expose:
    all-native-tools-by-default: true
    includes:
      - EchoHeaders
    excludes:
      - DownloadReport
`
	writeCoreVirtualConfig(t, homeDir, binaryName, cfg)

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// EchoHeaders: in includes (explicitly added) + all-by-default
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "EchoHeaders",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()

	// SayHello: enabled by all-native-tools-by-default, not excluded
	resp, _ = mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "SayHello",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()

	// DownloadReport: excluded
	resp, _ = mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "DownloadReport",
		"arguments": map[string]interface{}{},
	})
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(bodyBytes), "not found") && !strings.Contains(string(bodyBytes), "unknown tool") {
		t.Errorf("DownloadReport should be excluded, but got: %s", trimMsg(string(bodyBytes), 300))
	}
}

// TestConfig_ExposeIncludes_WithAllDefaultFalse_AddsTools verifies that
// includes adds tools even when all-native-tools-by-default is false.
func TestConfig_ExposeIncludes_WithAllDefaultFalse_AddsTools(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	cfg := `
tools:
  expose:
    all-native-tools-by-default: false
    includes:
      - EchoHeaders
    excludes:
      - SayHello
`
	writeCoreVirtualConfig(t, homeDir, binaryName, cfg)

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// EchoHeaders: explicitly included, should work
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "EchoHeaders",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()

	// SayHello: excluded (no-op since it wasn't activated by default anyway)
	resp, _ = mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "SayHello",
		"arguments": map[string]interface{}{},
	})
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(bodyBytes), "not found") && !strings.Contains(string(bodyBytes), "unknown tool") {
		t.Errorf("SayHello should not be available, but got: %s", trimMsg(string(bodyBytes), 300))
	}

	// DownloadReport: not in includes, not in excludes, not activated by default
	resp2, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "DownloadReport",
		"arguments": map[string]interface{}{},
	})
	defer resp2.Body.Close()
	bodyBytes, _ = io.ReadAll(resp2.Body)
	if !strings.Contains(string(bodyBytes), "not found") {
		t.Errorf("DownloadReport should not be available, but got: %s", trimMsg(string(bodyBytes), 300))
	}
}

// TestConfig_CLI_ListRespectsConfig verifies that CLI list honors the
// expose config with includes/excludes.
func TestConfig_CLI_ListRespectsConfig(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	cfg := `
tools:
  expose:
    all-native-tools-by-default: false
    includes:
      - EchoHeaders
`
	writeCoreVirtualConfig(t, homeDir, binaryName, cfg)

	binPath := buildServer(t, dir)
	stdout, _ := runCLI(t, binPath,
		[]string{
			"HOME=" + homeDir,
			"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
		},
		"-t", "cli", "list",
	)

	if !strings.Contains(stdout, "EchoHeaders") {
		t.Error("CLI list should contain EchoHeaders (included)")
	}
	if strings.Contains(stdout, "SayHello") {
		t.Error("CLI list should NOT contain SayHello (not included)")
	}
	if strings.Contains(stdout, "DownloadReport") {
		t.Error("CLI list should NOT contain DownloadReport (not included)")
	}
}

// TestConfig_CLI_CallRespectsConfig verifies that CLI call rejects tools
// that are not enabled by the expose config.
func TestConfig_CLI_CallRespectsConfig(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	cfg := `
tools:
  expose:
    all-native-tools-by-default: false
    includes:
      - EchoHeaders
`
	writeCoreVirtualConfig(t, homeDir, binaryName, cfg)

	binPath := buildServer(t, dir)
	env := []string{
		"HOME=" + homeDir,
		"MCP_UPSTREAM_ENDPOINT=" + mock.server.URL,
	}

	// EchoHeaders is enabled, should succeed
	_, stderr := runCLI(t, binPath, env, "-t", "cli", "EchoHeaders")
	if strings.Contains(stderr, "not enabled") {
		t.Error("EchoHeaders should be callable when included")
	}

	// SayHello is not enabled, should be rejected
	_, stderr = runCLI(t, binPath, env, "-t", "cli", "SayHello", "--name=Test")
	if !strings.Contains(stderr, "not enabled") {
		t.Errorf("SayHello should be rejected as not enabled, but got stderr: %s", stderr)
	}
}

// TestConfig_EmptyExpose_AllToolsAvailable verifies backward compatibility:
// when the config file exists but has no tools.expose section, all tools
// are still available.
func TestConfig_EmptyExpose_AllToolsAvailable(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	mock.RegisterGreetingScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	// Config file with upstream section only, no tools.expose
	cfg := `
upstream:
  enable_mcp_session_in_forwarding: false
`
	writeCoreVirtualConfig(t, homeDir, binaryName, cfg)

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// All 3 tools should still be callable
	for _, toolName := range []string{"EchoHeaders", "SayHello", "DownloadReport"} {
		resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      toolName,
			"arguments": map[string]interface{}{},
		})
		resp.Body.Close()
	}
}

// TestConfig_IncludesNotFoundInRegistry_NoError verifies that if an
// includes entry names a tool not in the registry, it is silently ignored
// (only tools that actually exist are activated). A warning should not
// prevent the server from starting.
func TestConfig_IncludesNotFoundInRegistry_NoError(t *testing.T) {
	mock := NewCoreMockService()
	mock.RegisterEchoAuthScenario()
	_ = mock.Start()
	defer mock.Close()

	dir := genProject(t, "echoHeaders,sayHello,downloadReport", "")
	homeDir := t.TempDir()
	binaryName := filepath.Base(dir)

	cfg := `
tools:
  expose:
    all-native-tools-by-default: false
    includes:
      - EchoHeaders
      - NonExistentTool
`
	writeCoreVirtualConfig(t, homeDir, binaryName, cfg)

	cleanup, baseURL := startCoreForwardTestServer(t, dir, mock.server.URL, homeDir, "", "")
	defer cleanup()

	// EchoHeaders should still be callable (NonExistentTool is silently ignored)
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      "EchoHeaders",
		"arguments": map[string]interface{}{},
	})
	resp.Body.Close()
}

func callNativeTool(t *testing.T, baseURL string, toolName string, args map[string]interface{}) string {
	t.Helper()
	resp, _ := mcpHTTPCall(t, baseURL, "tools/call", map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	})
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var rpcResp struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	json.Unmarshal(body, &rpcResp)

	if rpcResp.Error != nil {
		return fmt.Sprintf("MCP error: %s", rpcResp.Error.Message)
	}
	if len(rpcResp.Result.Content) > 0 {
		return rpcResp.Result.Content[0].Text
	}
	return ""
}

// startCoreForwardTestServer builds and starts a native MCP server with
// custom environment for authentication/header forwarding tests.
func startCoreForwardTestServer(t *testing.T, projectDir, mockURL, homeDir, token, cookie string) (cleanup func(), baseURL string) {
	t.Helper()
	binPath := buildServer(t, projectDir)
	port := fmt.Sprintf("%d", 19000+(time.Now().UnixNano()%1000))

	cmd := exec.Command(binPath, "--transport", "http", "--port", port, "-v", "1")
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"MCP_UPSTREAM_ENDPOINT="+mockURL,
	)
	if token != "" {
		cmd.Env = append(cmd.Env, "MCP_UPSTREAM_TOKEN="+token)
	}
	if cookie != "" {
		cmd.Env = append(cmd.Env, "MCP_UPSTREAM_COOKIE="+cookie)
	}

	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start HTTP server: %v", err)
	}

	cleanup = func() {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}
	baseURL = "http://localhost:" + port
	waitForServer(t, baseURL)
	return
}

// writeCoreVirtualConfig writes an virtual tools config for core tests.
func writeCoreVirtualConfig(t *testing.T, homeDir, binaryName, yamlContent string) {
	t.Helper()
	configDir := filepath.Join(homeDir, "."+binaryName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
}

func trimMsg(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
